package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/alerts"
	"github.com/jamesc159/monmetrics/internal/models"
)

// getMaxAlerts returns the active alert limit for a user type
func getMaxAlerts(userType string) int {
	if userType == "paid" {
		return 25
	}
	return 3
}

// GetAlerts lists the user's alerts, optionally filtered by ?card_id=
func (h *Handlers) GetAlerts(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	filter := bson.M{"user_id": userID}
	if raw := r.URL.Query().Get("card_id"); raw != "" {
		cardID, err := primitive.ObjectIDFromHex(raw)
		if err != nil {
			h.sendError(w, "Invalid card ID", http.StatusBadRequest, nil)
			return
		}
		filter["card_id"] = cardID
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := h.db.Collection("price_alerts").Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		h.sendError(w, "Error retrieving alerts", http.StatusInternalServerError, nil)
		return
	}
	list := make([]models.PriceAlert, 0)
	if err := cur.All(ctx, &list); err != nil {
		h.sendError(w, "Error decoding alerts", http.StatusInternalServerError, nil)
		return
	}
	if err := h.attachAlertCards(ctx, list); err != nil {
		h.sendError(w, "Error retrieving cards", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, list)
}

// CreateAlert creates a new price alert
func (h *Handlers) CreateAlert(w http.ResponseWriter, r *http.Request) {
	userID, claims, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	var req models.PriceAlertRequest
	if err := decodeJSON(w, r, &req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest, nil)
		return
	}
	if err := validateAlert(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}
	cardID, err := primitive.ObjectIDFromHex(req.CardID)
	if err != nil {
		h.sendError(w, "Invalid card ID", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := h.cardExists(ctx, cardID)
	if err != nil {
		h.sendError(w, "Error validating card", http.StatusInternalServerError, nil)
		return
	}
	if !exists {
		h.sendError(w, "Card not found", http.StatusBadRequest, nil)
		return
	}
	if ok := h.checkAlertLimit(w, ctx, userID, claims.UserType); !ok {
		return
	}

	now := time.Now().UTC()
	alert := models.PriceAlert{UserID: userID, CardID: cardID, Active: true, CreatedAt: now, UpdatedAt: now}
	applyAlertRequest(&alert, &req)
	h.setAlertBaseline(ctx, &alert)

	res, err := h.db.Collection("price_alerts").InsertOne(ctx, alert)
	if err != nil {
		h.sendError(w, "Error saving alert", http.StatusInternalServerError, nil)
		return
	}
	alert.ID = res.InsertedID.(primitive.ObjectID)
	list := []models.PriceAlert{alert}
	if err := h.attachAlertCards(ctx, list); err != nil {
		h.sendError(w, "Error retrieving card", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusCreated, list[0])
}

// UpdateAlert edits an alert; setting active=true on an inactive alert re-arms it
func (h *Handlers) UpdateAlert(w http.ResponseWriter, r *http.Request) {
	userID, claims, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	alertID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid alert ID", http.StatusBadRequest, nil)
		return
	}
	var req models.PriceAlertRequest
	if err := decodeJSON(w, r, &req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest, nil)
		return
	}
	if err := validateAlert(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	coll := h.db.Collection("price_alerts")
	var alert models.PriceAlert
	err = coll.FindOne(ctx, bson.M{"_id": alertID, "user_id": userID}).Decode(&alert)
	if err == mongo.ErrNoDocuments {
		h.sendError(w, "Alert not found", http.StatusNotFound, nil)
		return
	}
	if err != nil {
		h.sendError(w, "Error retrieving alert", http.StatusInternalServerError, nil)
		return
	}

	active := alert.Active
	if req.Active != nil {
		active = *req.Active
	}
	if active && !alert.Active {
		if ok := h.checkAlertLimit(w, ctx, userID, claims.UserType); !ok {
			return
		}
	}

	applyAlertRequest(&alert, &req)
	alert.Active = active
	alert.UpdatedAt = time.Now().UTC()
	h.setAlertBaseline(ctx, &alert)

	set := bson.M{
		"source": alert.Source, "condition": alert.Condition, "target_price": alert.TargetPrice,
		"pct": alert.Pct, "days": alert.Days, "ema_period": alert.EMAPeriod, "direction": alert.Direction,
		"mode": alert.Mode, "cooldown_hours": alert.CooldownHours, "notify_email": alert.NotifyEmail,
		"note": alert.Note, "active": alert.Active, "updated_at": alert.UpdatedAt,
	}
	unset := bson.M{}
	if alert.LastPrice != nil {
		set["last_price"] = *alert.LastPrice
	} else {
		unset["last_price"] = ""
	}
	if alert.LastEMA != nil {
		set["last_ema"] = *alert.LastEMA
	} else {
		unset["last_ema"] = ""
	}
	update := bson.M{"$set": set}
	if len(unset) > 0 {
		update["$unset"] = unset
	}
	if _, err := coll.UpdateOne(ctx, bson.M{"_id": alertID, "user_id": userID}, update); err != nil {
		h.sendError(w, "Error updating alert", http.StatusInternalServerError, nil)
		return
	}
	list := []models.PriceAlert{alert}
	if err := h.attachAlertCards(ctx, list); err != nil {
		h.sendError(w, "Error retrieving card", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, list[0])
}

// DeleteAlert removes a user's alert
func (h *Handlers) DeleteAlert(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	alertID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid alert ID", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := h.db.Collection("price_alerts").DeleteOne(ctx, bson.M{"_id": alertID, "user_id": userID})
	if err != nil {
		h.sendError(w, "Error deleting alert", http.StatusInternalServerError, nil)
		return
	}
	if res.DeletedCount == 0 {
		h.sendError(w, "Alert not found", http.StatusNotFound, nil)
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"message": "Alert deleted"})
}

func (h *Handlers) checkAlertLimit(w http.ResponseWriter, ctx context.Context, userID primitive.ObjectID, userType string) bool {
	limit := getMaxAlerts(userType)
	count, err := h.db.Collection("price_alerts").CountDocuments(ctx, bson.M{"user_id": userID, "active": true})
	if err != nil {
		h.sendError(w, "Error checking alerts", http.StatusInternalServerError, nil)
		return false
	}
	if count >= int64(limit) {
		h.sendError(w, "Active alert limit reached ("+strconv.Itoa(limit)+")", http.StatusBadRequest,
			map[string]interface{}{"max_alerts": limit})
		return false
	}
	return true
}

func applyAlertRequest(a *models.PriceAlert, req *models.PriceAlertRequest) {
	a.Source = req.Source
	a.Condition = req.Condition
	a.TargetPrice = req.TargetPrice
	a.Pct = req.Pct
	a.Days = req.Days
	a.EMAPeriod = req.EMAPeriod
	a.Direction = req.Direction
	a.Mode = req.Mode
	a.CooldownHours = req.CooldownHours
	a.NotifyEmail = req.NotifyEmail
	a.Note = req.Note
}

// setAlertBaseline records the current price so the alert only fires on future moves.
func (h *Handlers) setAlertBaseline(ctx context.Context, a *models.PriceAlert) {
	a.LastPrice, a.LastEMA = nil, nil
	series, err := alerts.LoadSeries(ctx, h.db, a.CardID, a.Source, alerts.WindowDays(a))
	if err != nil {
		return
	}
	snap, ok := alerts.BuildSnapshot(a, series, time.Now().UTC())
	if !ok {
		return
	}
	a.LastPrice = &snap.Price
	a.LastEMA = snap.EMA
}

func (h *Handlers) attachAlertCards(ctx context.Context, list []models.PriceAlert) error {
	ids := make([]primitive.ObjectID, 0, len(list))
	for _, a := range list {
		ids = append(ids, a.CardID)
	}
	cards, err := h.cardsByID(ctx, ids)
	if err != nil {
		return err
	}
	for i := range list {
		if c, ok := cards[list[i].CardID]; ok {
			list[i].CardName = c.Name
			list[i].CardImageURL = c.ImageURL
		}
	}
	return nil
}

// GetNotifications lists recent notifications; ?unread=1 limits to unread
func (h *Handlers) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	limit := int64(20)
	if raw := r.URL.Query().Get("limit"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 100 {
			h.sendError(w, "limit must be 1-100", http.StatusBadRequest, nil)
			return
		}
		limit = int64(n)
	}
	filter := bson.M{"user_id": userID}
	if r.URL.Query().Get("unread") == "1" {
		filter["read"] = false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	coll := h.db.Collection("notifications")
	cur, err := coll.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(limit))
	if err != nil {
		h.sendError(w, "Error retrieving notifications", http.StatusInternalServerError, nil)
		return
	}
	list := make([]models.Notification, 0)
	if err := cur.All(ctx, &list); err != nil {
		h.sendError(w, "Error decoding notifications", http.StatusInternalServerError, nil)
		return
	}
	unread, err := coll.CountDocuments(ctx, bson.M{"user_id": userID, "read": false})
	if err != nil {
		h.sendError(w, "Error counting notifications", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, map[string]interface{}{"notifications": list, "unread_count": unread})
}

// MarkNotificationRead marks one notification read
func (h *Handlers) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	id, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid notification ID", http.StatusBadRequest, nil)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := h.db.Collection("notifications").UpdateOne(ctx,
		bson.M{"_id": id, "user_id": userID}, bson.M{"$set": bson.M{"read": true}})
	if err != nil {
		h.sendError(w, "Error updating notification", http.StatusInternalServerError, nil)
		return
	}
	if res.MatchedCount == 0 {
		h.sendError(w, "Notification not found", http.StatusNotFound, nil)
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// MarkAllNotificationsRead marks every unread notification read
func (h *Handlers) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.db.Collection("notifications").UpdateMany(ctx,
		bson.M{"user_id": userID, "read": false}, bson.M{"$set": bson.M{"read": true}}); err != nil {
		h.sendError(w, "Error updating notifications", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}
