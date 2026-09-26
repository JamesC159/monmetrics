package alerts

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jamesc159/monmetrics/internal/models"
	"github.com/jamesc159/monmetrics/internal/notify"
)

// LoadSeries returns chronological daily average prices for a card over the last days.
func LoadSeries(ctx context.Context, db *mongo.Database, cardID primitive.ObjectID, source string, days int) ([]DayPrice, error) {
	match := bson.M{"card_id": cardID, "timestamp": bson.M{"$gte": time.Now().UTC().AddDate(0, 0, -days)}}
	if source != "" && source != "all" {
		match["source"] = source
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: match}},
		{{Key: "$group", Value: bson.M{
			"_id": bson.M{"$dateTrunc": bson.M{"date": "$timestamp", "unit": "day"}},
			"avg": bson.M{"$avg": "$price"},
		}}},
		{{Key: "$sort", Value: bson.M{"_id": 1}}},
	}
	cur, err := db.Collection("prices").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var rows []struct {
		Day time.Time `bson:"_id"`
		Avg float64   `bson:"avg"`
	}
	if err := cur.All(ctx, &rows); err != nil {
		return nil, err
	}
	out := make([]DayPrice, len(rows))
	for i, r := range rows {
		out[i] = DayPrice{Day: r.Day, Price: r.Avg}
	}
	return out, nil
}

// Evaluator periodically checks active alerts and notifies users when they fire
type Evaluator struct {
	DB          *mongo.Database
	Sender      notify.Sender
	Interval    time.Duration
	FrontendURL string
}

// Run evaluates immediately and then on every tick until ctx is cancelled.
func (e *Evaluator) Run(ctx context.Context) {
	log.Printf("🔔 Price alert evaluator running every %v", e.Interval)
	e.tick(ctx)
	t := time.NewTicker(e.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			e.tick(ctx)
		}
	}
}

type seriesKey struct {
	card   primitive.ObjectID
	source string
}

func (e *Evaluator) tick(parent context.Context) {
	ctx, cancel := context.WithTimeout(parent, 2*time.Minute)
	defer cancel()

	cur, err := e.DB.Collection("price_alerts").Find(ctx, bson.M{"active": true})
	if err != nil {
		log.Printf("❌ alerts: load failed: %v", err)
		return
	}
	var active []models.PriceAlert
	if err := cur.All(ctx, &active); err != nil {
		log.Printf("❌ alerts: decode failed: %v", err)
		return
	}
	if len(active) == 0 {
		return
	}

	groups := map[seriesKey][]*models.PriceAlert{}
	for i := range active {
		a := &active[i]
		k := seriesKey{a.CardID, a.Source}
		groups[k] = append(groups[k], a)
	}

	now := time.Now().UTC()
	var baselines []mongo.WriteModel
	fired := 0
	for k, list := range groups {
		window := 3
		for _, a := range list {
			window = max(window, WindowDays(a))
		}
		series, err := LoadSeries(ctx, e.DB, k.card, k.source, window)
		if err != nil {
			log.Printf("❌ alerts: series for %s failed: %v", k.card.Hex(), err)
			continue
		}
		for _, a := range list {
			snap, ok := BuildSnapshot(a, series, now)
			if !ok {
				continue
			}
			if ShouldFire(a, snap, now) {
				if e.fire(ctx, a, snap, now) {
					fired++
				}
				continue
			}
			set := bson.M{"last_price": snap.Price}
			if snap.EMA != nil {
				set["last_ema"] = *snap.EMA
			}
			baselines = append(baselines, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": a.ID, "active": true}).
				SetUpdate(bson.M{"$set": set}))
		}
	}
	if len(baselines) > 0 {
		if _, err := e.DB.Collection("price_alerts").BulkWrite(ctx, baselines); err != nil {
			log.Printf("❌ alerts: baseline update failed: %v", err)
		}
	}
	if fired > 0 {
		log.Printf("🔔 alerts: %d fired", fired)
	}
}

// fire records the trigger and notifies; it returns false if another evaluator already claimed it.
func (e *Evaluator) fire(ctx context.Context, a *models.PriceAlert, snap Snapshot, now time.Time) bool {
	filter := bson.M{"_id": a.ID, "active": true}
	if a.LastTriggeredAt != nil {
		filter["last_triggered_at"] = *a.LastTriggeredAt
	} else {
		filter["last_triggered_at"] = bson.M{"$exists": false}
	}
	set := bson.M{
		"last_triggered_at": now,
		"last_price":        snap.Price,
		"active":            a.Mode == models.AlertModeRecurring,
		"updated_at":        now,
	}
	if snap.EMA != nil {
		set["last_ema"] = *snap.EMA
	}
	res, err := e.DB.Collection("price_alerts").UpdateOne(ctx, filter, bson.M{"$set": set, "$inc": bson.M{"trigger_count": 1}})
	if err != nil {
		log.Printf("❌ alerts: claim %s failed: %v", a.ID.Hex(), err)
		return false
	}
	if res.ModifiedCount == 0 {
		return false
	}

	var card models.Card
	if err := e.DB.Collection("cards").FindOne(ctx, bson.M{"_id": a.CardID}).Decode(&card); err != nil {
		card.Name = "A card you follow"
	}
	title := fmt.Sprintf("Price alert: %s", card.Name)
	message := Describe(a, snap)

	alertID, cardID := a.ID, a.CardID
	n := models.Notification{UserID: a.UserID, AlertID: &alertID, CardID: &cardID, Title: title, Message: message, CreatedAt: now}
	if _, err := e.DB.Collection("notifications").InsertOne(ctx, n); err != nil {
		log.Printf("❌ alerts: notification insert failed: %v", err)
	}

	if a.NotifyEmail && e.Sender != nil {
		var user models.User
		if err := e.DB.Collection("users").FindOne(ctx, bson.M{"_id": a.UserID}).Decode(&user); err != nil {
			log.Printf("❌ alerts: user lookup failed: %v", err)
			return true
		}
		body := fmt.Sprintf("%s\n\n%s\n\nView the chart: %s/card/%s\n\nManage alerts: %s/dashboard\n",
			title, message, e.FrontendURL, a.CardID.Hex(), e.FrontendURL)
		if err := e.Sender.Send(ctx, notify.Message{To: user.Email, Subject: title, Body: body}); err != nil {
			log.Printf("❌ alerts: email to user %s failed: %v", a.UserID.Hex(), err)
		}
	}
	return true
}

// Describe renders a human-readable trigger message.
func Describe(a *models.PriceAlert, s Snapshot) string {
	price := fmt.Sprintf("$%.2f", s.Price)
	var msg string
	switch a.Condition {
	case models.AlertAbove:
		msg = fmt.Sprintf("Price rose to %s, crossing above $%.2f.", price, a.TargetPrice)
	case models.AlertBelow:
		msg = fmt.Sprintf("Price fell to %s, crossing below $%.2f.", price, a.TargetPrice)
	case models.AlertPctChange:
		change := 0.0
		if s.PastPrice != nil && *s.PastPrice > 0 {
			change = (s.Price / *s.PastPrice - 1) * 100
		}
		msg = fmt.Sprintf("Price moved %+.1f%% over %d days (now %s).", change, a.Days, price)
	case models.AlertEMACross:
		dir := "above"
		if s.EMA != nil && s.Price < *s.EMA {
			dir = "below"
		}
		msg = fmt.Sprintf("Price (%s) crossed %s its EMA %d.", price, dir, a.EMAPeriod)
	}
	if a.Note != "" {
		msg += " Note: " + a.Note
	}
	return msg
}
