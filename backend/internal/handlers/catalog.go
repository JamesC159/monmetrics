package handlers

import (
	"context"
	"net/http"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jamesc159/monmetrics/internal/features"
	"github.com/jamesc159/monmetrics/internal/models"
)

// maxSetsPerResponse caps the set list so a game with a huge catalog cannot blow up a response.
const maxSetsPerResponse = 500

var validCatalogCategories = map[string]bool{"card": true, "sealed": true}

// countIf builds a $sum accumulator that counts documents matching cond.
func countIf(cond bson.M) bson.M {
	return bson.M{"$sum": bson.M{"$cond": bson.A{cond, 1, 0}}}
}

// activeCatalogMatch is the shared base filter: ignore soft-deleted catalog rows.
func activeCatalogMatch() bson.M {
	return bson.M{"active": bson.M{"$ne": false}}
}

// applyEnabledGames constrains a query to the games this deployment exposes.
// Wrapped in $and rather than assigning filter["game"], which a caller-supplied
// ?game= filter would otherwise overwrite and escape the flag.
func (h *Handlers) applyEnabledGames(filter bson.M) bson.M {
	if h.features != nil && h.features.Enabled(features.KeyAllGames) {
		return filter
	}
	if h.config == nil || len(h.config.EnabledGames) == 0 {
		return filter
	}
	patterns := make([]primitive.Regex, 0, len(h.config.EnabledGames))
	for _, game := range h.config.EnabledGames {
		patterns = append(patterns, primitive.Regex{Pattern: regexp.QuoteMeta(game), Options: "i"})
	}
	games := bson.M{"game": bson.M{"$in": patterns}}
	if len(filter) == 0 {
		return games
	}
	return bson.M{"$and": []bson.M{filter, games}}
}

// GetGames lists every game present in the catalog with per-category counts.
func (h *Handlers) GetGames(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	match := activeCatalogMatch()
	match["game"] = bson.M{"$nin": bson.A{nil, ""}}

	pipeline := []bson.M{
		{"$match": h.applyEnabledGames(match)},
		{"$group": bson.M{
			"_id":          "$game",
			"card_count":   countIf(bson.M{"$eq": bson.A{"$category", "card"}}),
			"sealed_count": countIf(bson.M{"$eq": bson.A{"$category", "sealed"}}),
			"priced_count": countIf(bson.M{"$gt": bson.A{"$current_price", 0}}),
			"image_url":    bson.M{"$first": "$image_url"},
		}},
		{"$project": bson.M{
			"_id":          0,
			"game":         "$_id",
			"card_count":   1,
			"sealed_count": 1,
			"priced_count": 1,
			"image_url":    1,
		}},
		{"$sort": bson.M{"game": 1}},
	}

	cursor, err := h.db.Collection("cards").Aggregate(ctx, pipeline)
	if err != nil {
		h.sendError(w, "Error retrieving games", http.StatusInternalServerError, nil)
		return
	}
	defer cursor.Close(ctx)

	games := make([]models.GameSummary, 0)
	if err := cursor.All(ctx, &games); err != nil {
		h.sendError(w, "Error decoding games", http.StatusInternalServerError, nil)
		return
	}

	sendJSON(w, http.StatusOK, games)
}

// GetSets lists the sets/expansions in the catalog, optionally scoped to a game or category.
func (h *Handlers) GetSets(w http.ResponseWriter, r *http.Request) {
	game := r.URL.Query().Get("game")
	category := r.URL.Query().Get("category")
	if category != "" && !validCatalogCategories[category] {
		h.sendError(w, "Invalid category", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	match := activeCatalogMatch()
	match["set"] = bson.M{"$nin": bson.A{nil, ""}}
	if game != "" {
		match["game"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(game), Options: "i"}}
	}
	if category != "" {
		match["category"] = category
	}

	pipeline := []bson.M{
		{"$match": h.applyEnabledGames(match)},
		{"$group": bson.M{
			"_id":               bson.M{"set": "$set", "game": "$game"},
			"external_group_id": bson.M{"$first": "$external_group_id"},
			"product_count":     bson.M{"$sum": 1},
			"priced_count":      countIf(bson.M{"$gt": bson.A{"$current_price", 0}}),
			"image_url":         bson.M{"$first": "$image_url"},
		}},
		{"$project": bson.M{
			"_id":               0,
			"set":               "$_id.set",
			"game":              "$_id.game",
			"external_group_id": 1,
			"product_count":     1,
			"priced_count":      1,
			"image_url":         1,
		}},
		{"$sort": bson.M{"priced_count": -1, "product_count": -1, "set": 1}},
		{"$limit": maxSetsPerResponse},
	}

	cursor, err := h.db.Collection("cards").Aggregate(ctx, pipeline)
	if err != nil {
		h.sendError(w, "Error retrieving sets", http.StatusInternalServerError, nil)
		return
	}
	defer cursor.Close(ctx)

	sets := make([]models.SetSummary, 0)
	if err := cursor.All(ctx, &sets); err != nil {
		h.sendError(w, "Error decoding sets", http.StatusInternalServerError, nil)
		return
	}

	sendJSON(w, http.StatusOK, sets)
}
