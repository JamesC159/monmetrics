package handlers

import (
	"context"
	"math"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/models"
)

const maxFavoritesPerUser = 200

// GetFavorites lists the user's favorites with card data and recent price change
func (h *Handlers) GetFavorites(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cur, err := h.db.Collection("favorites").Find(ctx, bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		h.sendError(w, "Error retrieving favorites", http.StatusInternalServerError, nil)
		return
	}
	defer cur.Close(ctx)

	favs := make([]models.Favorite, 0)
	if err := cur.All(ctx, &favs); err != nil {
		h.sendError(w, "Error decoding favorites", http.StatusInternalServerError, nil)
		return
	}

	ids := make([]primitive.ObjectID, 0, len(favs))
	for _, f := range favs {
		ids = append(ids, f.CardID)
	}
	cards, err := h.cardsByID(ctx, ids)
	if err != nil {
		h.sendError(w, "Error retrieving cards", http.StatusInternalServerError, nil)
		return
	}
	past7, err := h.averagePriceAround(ctx, ids, 7)
	if err != nil {
		h.sendError(w, "Error computing price change", http.StatusInternalServerError, nil)
		return
	}
	past30, err := h.averagePriceAround(ctx, ids, 30)
	if err != nil {
		h.sendError(w, "Error computing price change", http.StatusInternalServerError, nil)
		return
	}

	views := make([]models.FavoriteView, 0, len(favs))
	for _, f := range favs {
		v := models.FavoriteView{ID: f.ID, CardID: f.CardID, CreatedAt: f.CreatedAt}
		if card, ok := cards[f.CardID]; ok {
			v.Card = card
			v.Change7d, v.Change7dValue = priceChange(card.CurrentPrice, past7[f.CardID])
			v.Change30d, v.Change30dValue = priceChange(card.CurrentPrice, past30[f.CardID])
		}
		views = append(views, v)
	}
	sendJSON(w, http.StatusOK, views)
}

// AddFavorite adds a card to the user's favorites (idempotent)
func (h *Handlers) AddFavorite(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	var req models.FavoriteRequest
	if err := decodeJSON(w, r, &req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest, nil)
		return
	}
	cardID, err := primitive.ObjectIDFromHex(req.CardID)
	if err != nil {
		h.sendError(w, "Invalid card ID", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := h.cardExists(ctx, cardID)
	if err != nil {
		h.sendError(w, "Error validating card", http.StatusInternalServerError, nil)
		return
	}
	if !exists {
		h.sendError(w, "Card not found", http.StatusNotFound, nil)
		return
	}

	coll := h.db.Collection("favorites")
	count, err := coll.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		h.sendError(w, "Error checking favorites", http.StatusInternalServerError, nil)
		return
	}
	if count >= maxFavoritesPerUser {
		already, err := coll.CountDocuments(ctx, bson.M{"user_id": userID, "card_id": cardID})
		if err != nil {
			h.sendError(w, "Error checking favorites", http.StatusInternalServerError, nil)
			return
		}
		if already == 0 {
			h.sendError(w, "Favorites limit reached (200)", http.StatusBadRequest, nil)
			return
		}
	}

	res, err := coll.UpdateOne(ctx,
		bson.M{"user_id": userID, "card_id": cardID},
		bson.M{"$setOnInsert": bson.M{"user_id": userID, "card_id": cardID, "created_at": time.Now().UTC()}},
		options.Update().SetUpsert(true))
	if err != nil {
		h.sendError(w, "Error saving favorite", http.StatusInternalServerError, nil)
		return
	}
	status := http.StatusOK
	if res.UpsertedCount > 0 {
		status = http.StatusCreated
	}
	sendJSON(w, status, map[string]string{"card_id": cardID.Hex()})
}

// RemoveFavorite removes a card from the user's favorites
func (h *Handlers) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	cardID, err := primitive.ObjectIDFromHex(r.PathValue("cardId"))
	if err != nil {
		h.sendError(w, "Invalid card ID", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := h.db.Collection("favorites").DeleteOne(ctx, bson.M{"user_id": userID, "card_id": cardID})
	if err != nil {
		h.sendError(w, "Error removing favorite", http.StatusInternalServerError, nil)
		return
	}
	if res.DeletedCount == 0 {
		h.sendError(w, "Favorite not found", http.StatusNotFound, nil)
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"message": "Favorite removed"})
}

// averagePriceAround returns each card's average price in a 2-day window centered daysAgo days back.
func (h *Handlers) averagePriceAround(ctx context.Context, ids []primitive.ObjectID, daysAgo int) (map[primitive.ObjectID]float64, error) {
	out := map[primitive.ObjectID]float64{}
	if len(ids) == 0 {
		return out, nil
	}
	center := time.Now().AddDate(0, 0, -daysAgo)
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"card_id":   bson.M{"$in": ids},
			"timestamp": bson.M{"$gte": center.Add(-24 * time.Hour), "$lte": center.Add(24 * time.Hour)},
		}}},
		{{Key: "$group", Value: bson.M{"_id": "$card_id", "avg": bson.M{"$avg": "$price"}}}},
	}
	cur, err := h.db.Collection("prices").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var rows []struct {
		ID  primitive.ObjectID `bson:"_id"`
		Avg float64            `bson:"avg"`
	}
	if err := cur.All(ctx, &rows); err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ID] = row.Avg
	}
	return out, nil
}

func priceChange(current, past float64) (pct, value float64) {
	if past <= 0 {
		return 0, 0
	}
	value = current - past
	pct = value / past * 100
	return math.Round(pct*100) / 100, math.Round(value*100) / 100
}
