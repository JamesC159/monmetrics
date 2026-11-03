package scraper

import (
	"context"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Price floors that place an unwatched item into a faster refresh tier.
const (
	HotPriceFloor  = 200.0
	WarmPriceFloor = 20.0
	maxFailureStep = 6
)

// Scheduler decides which catalog items the sold-listing scraper visits next.
// The catalog is far larger than one day of polite scraping, so items rotate
// through tiers instead of being scanned end to end.
type Scheduler struct {
	DB        *mongo.Database
	Intervals map[int]time.Duration
}

func (s *Scheduler) interval(tier int) time.Duration {
	if value, ok := s.Intervals[tier]; ok && value > 0 {
		return value
	}
	switch tier {
	case models.ScrapeTierHot:
		return 24 * time.Hour
	case models.ScrapeTierWarm:
		return 7 * 24 * time.Hour
	default:
		return 30 * 24 * time.Hour
	}
}

// DueCards returns up to limit items whose next scrape time has passed.
func (s *Scheduler) DueCards(ctx context.Context, now time.Time, limit int) ([]models.Card, error) {
	filter := bson.M{
		"active":      bson.M{"$ne": false},
		"scrape_tier": bson.M{"$ne": models.ScrapeTierNone},
		"category":    bson.M{"$in": []string{"card", "sealed"}},
		"$or": []bson.M{
			{"next_scrape_at": bson.M{"$lte": now.UTC()}},
			{"next_scrape_at": bson.M{"$exists": false}},
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "next_scrape_at", Value: 1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	cursor, err := s.DB.Collection("cards").Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var cards []models.Card
	if err := cursor.All(ctx, &cards); err != nil {
		return nil, err
	}
	return cards, nil
}

// MarkScraped records a successful pass and schedules the next one. Items that
// keep returning nothing back off so dead stock does not consume the budget.
func (s *Scheduler) MarkScraped(ctx context.Context, card models.Card, saleCount int, now time.Time) error {
	failures := 0
	if saleCount == 0 {
		failures = card.ScrapeFailures + 1
	}
	next := now.UTC().Add(backoff(s.interval(card.ScrapeTier), failures))
	_, err := s.DB.Collection("cards").UpdateOne(ctx, bson.M{"_id": card.ID}, bson.M{"$set": bson.M{
		"last_scrape_at":  now.UTC(),
		"next_scrape_at":  next,
		"scrape_failures": failures,
		"last_sale_count": saleCount,
	}})
	return err
}

// MarkFailed reschedules an item after a transport or parse error.
func (s *Scheduler) MarkFailed(ctx context.Context, card models.Card, now time.Time) error {
	failures := card.ScrapeFailures + 1
	_, err := s.DB.Collection("cards").UpdateOne(ctx, bson.M{"_id": card.ID}, bson.M{"$set": bson.M{
		"last_scrape_at":  now.UTC(),
		"next_scrape_at":  now.UTC().Add(backoff(s.interval(card.ScrapeTier), failures)),
		"scrape_failures": failures,
	}})
	return err
}

func backoff(base time.Duration, failures int) time.Duration {
	if failures <= 0 {
		return base
	}
	if failures > maxFailureStep {
		failures = maxFailureStep
	}
	return base * time.Duration(1<<uint(failures-1))
}

// AssignTiers recomputes refresh priority. Anything a user watches is hot; the
// rest is ranked by last known price.
func (s *Scheduler) AssignTiers(ctx context.Context, now time.Time) error {
	cards := s.DB.Collection("cards")
	base := bson.M{"active": bson.M{"$ne": false}, "category": bson.M{"$in": []string{"card", "sealed"}}}

	if _, err := cards.UpdateMany(ctx, base, bson.M{"$set": bson.M{"scrape_tier": models.ScrapeTierColdTail}}); err != nil {
		return err
	}
	warm := bson.M{"active": base["active"], "category": base["category"], "current_price": bson.M{"$gte": WarmPriceFloor}}
	if _, err := cards.UpdateMany(ctx, warm, bson.M{"$set": bson.M{"scrape_tier": models.ScrapeTierWarm}}); err != nil {
		return err
	}

	watched, err := s.watchedCardIDs(ctx)
	if err != nil {
		return err
	}
	hot := bson.M{"active": base["active"], "category": base["category"], "$or": []bson.M{
		{"current_price": bson.M{"$gte": HotPriceFloor}},
		{"_id": bson.M{"$in": watched}},
	}}
	if _, err := cards.UpdateMany(ctx, hot, bson.M{"$set": bson.M{"scrape_tier": models.ScrapeTierHot}}); err != nil {
		return err
	}

	// Pull promoted items forward so a tier change takes effect this cycle.
	cutoff := now.UTC().Add(s.interval(models.ScrapeTierHot))
	_, err = cards.UpdateMany(ctx,
		bson.M{"scrape_tier": models.ScrapeTierHot, "next_scrape_at": bson.M{"$gt": cutoff}},
		bson.M{"$set": bson.M{"next_scrape_at": cutoff}})
	return err
}

func (s *Scheduler) watchedCardIDs(ctx context.Context) ([]primitive.ObjectID, error) {
	unique := make(map[primitive.ObjectID]struct{})
	for _, collection := range []string{"favorites", "portfolio_items", "price_alerts"} {
		values, err := s.DB.Collection(collection).Distinct(ctx, "card_id", bson.M{})
		if err != nil {
			return nil, err
		}
		for _, value := range values {
			if id, ok := value.(primitive.ObjectID); ok {
				unique[id] = struct{}{}
			}
		}
	}
	ids := make([]primitive.ObjectID, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}
	return ids, nil
}
