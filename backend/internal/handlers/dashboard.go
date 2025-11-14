package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jamesc159/monmetrics/internal/middleware"
	"github.com/jamesc159/monmetrics/internal/models"
)

// GetDashboard retrieves user dashboard data
func (h *Handlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get saved charts
	chartsCollection := h.db.Collection("saved_charts")
	cursor, err := chartsCollection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		http.Error(w, "Error retrieving charts", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var savedCharts []models.SavedChart
	if err = cursor.All(ctx, &savedCharts); err != nil {
		http.Error(w, "Error decoding charts", http.StatusInternalServerError)
		return
	}

	// TODO: Implement recently viewed cards (would need to track user views)
	recentlyViewed := []models.Card{}

	// Calculate user stats
	userStats := models.UserStats{
		ChartsCreated:  len(savedCharts),
		IndicatorsUsed: calculateIndicatorsUsed(savedCharts),
		MaxIndicators:  getMaxIndicators(claims.UserType),
	}

	dashboard := models.Dashboard{
		SavedCharts:    savedCharts,
		RecentlyViewed: recentlyViewed,
		UserStats:      userStats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// SaveChart saves a new chart for the user
func (h *Handlers) SaveChart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var req models.SavedChart
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate indicators count based on user type
	maxIndicators := getMaxIndicators(claims.UserType)
	if len(req.Indicators) > maxIndicators {
		h.sendError(w, fmt.Sprintf("Exceeded maximum indicators limit (%d)", maxIndicators), http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set user ID and timestamps
	req.UserID = userID
	req.CreatedAt = time.Now().UTC()
	req.UpdatedAt = time.Now().UTC()

	collection := h.db.Collection("saved_charts")
	result, err := collection.InsertOne(ctx, req)
	if err != nil {
		http.Error(w, "Error saving chart", http.StatusInternalServerError)
		return
	}

	req.ID = result.InsertedID.(primitive.ObjectID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(req)
}

// GetSavedCharts retrieves user's saved charts
func (h *Handlers) GetSavedCharts(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.db.Collection("saved_charts")
	cursor, err := collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		http.Error(w, "Error retrieving charts", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var charts []models.SavedChart
	if err = cursor.All(ctx, &charts); err != nil {
		http.Error(w, "Error decoding charts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(charts)
}

// DeleteChart deletes a user's saved chart
func (h *Handlers) DeleteChart(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	chartIDStr := r.PathValue("id")
	chartID, err := primitive.ObjectIDFromHex(chartIDStr)
	if err != nil {
		http.Error(w, "Invalid chart ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := h.db.Collection("saved_charts")
	result, err := collection.DeleteOne(ctx, bson.M{
		"_id":     chartID,
		"user_id": userID, // Ensure user can only delete their own charts
	})
	if err != nil {
		http.Error(w, "Error deleting chart", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Chart not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Chart deleted successfully"})
}

// Helper functions for dashboard

// calculateIndicatorsUsed counts unique indicators across all saved charts
func calculateIndicatorsUsed(charts []models.SavedChart) int {
	indicatorMap := make(map[string]bool)
	for _, chart := range charts {
		for _, indicator := range chart.Indicators {
			indicatorMap[indicator.Type] = true
		}
	}
	return len(indicatorMap)
}

// getMaxIndicators returns the maximum number of indicators based on user type
func getMaxIndicators(userType string) int {
	if userType == "paid" {
		return 10
	}
	return 3 // free users
}
