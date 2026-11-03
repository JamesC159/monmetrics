package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jamesc159/monmetrics/internal/features"
	"github.com/jamesc159/monmetrics/internal/models"
)

func (h *Handlers) GetPriceSummary(w http.ResponseWriter, r *http.Request) {
	if !h.features.Enabled(features.KeyPriceSummary) {
		sendJSON(w, http.StatusOK, []models.PriceSummary{})
		return
	}

	cardID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}
	grade := r.URL.Query().Get("grade")
	if grade == "" {
		grade = "raw"
	}
	if !validGradePattern.MatchString(grade) {
		h.sendError(w, "Invalid grade", http.StatusBadRequest, nil)
		return
	}
	source := r.URL.Query().Get("source")
	if source == "" {
		source = "ebay"
	}
	if !validChartSources[source] {
		h.sendError(w, "Invalid source", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	filter := bson.M{"card_id": cardID, "grade": grade}
	if source != "all" {
		filter["source"] = source
	}
	cursor, err := h.db.Collection("price_summaries").Find(ctx, filter)
	if err != nil {
		http.Error(w, "Error retrieving price summaries", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	var summaries []models.PriceSummary
	if err := cursor.All(ctx, &summaries); err != nil {
		http.Error(w, "Error decoding price summaries", http.StatusInternalServerError)
		return
	}
	if summaries == nil {
		summaries = []models.PriceSummary{}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summaries); err != nil {
		h.sendError(w, "Could not encode price summaries", http.StatusInternalServerError, nil)
	}
}
