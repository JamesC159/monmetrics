package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/models"
)

// GetDashboard retrieves user dashboard data
func (h *Handlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID, claims, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	if err := h.db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			h.sendError(w, "User not found", http.StatusUnauthorized, nil)
			return
		}
		h.sendError(w, "Error retrieving user", http.StatusInternalServerError, nil)
		return
	}

	savedCharts, err := h.loadCharts(ctx, bson.M{"user_id": userID})
	if err != nil {
		h.sendError(w, "Error retrieving charts", http.StatusInternalServerError, nil)
		return
	}

	favCount, err := h.db.Collection("favorites").CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		fmt.Printf("Warning: could not count favorites: %v\n", err)
	}

	portfolio, err := h.buildPortfolio(ctx, userID, portfolioFilter{})
	if err != nil {
		fmt.Printf("Warning: could not build portfolio summary: %v\n", err)
	}

	dashboard := models.Dashboard{
		User:           &user,
		SavedCharts:    savedCharts,
		RecentlyViewed: []models.Card{},
		UserStats: models.UserStats{
			ChartsCreated:  len(savedCharts),
			IndicatorsUsed: calculateIndicatorsUsed(savedCharts),
			MaxIndicators:  getMaxIndicators(claims.UserType),
		},
		FavoritesCount: int(favCount),
	}
	if portfolio != nil {
		dashboard.PortfolioSummary = &portfolio.Summary
	}

	sendJSON(w, http.StatusOK, dashboard)
}

// loadCharts returns matching charts, newest first, with a card snapshot attached.
func (h *Handlers) loadCharts(ctx context.Context, filter bson.M) ([]models.SavedChart, error) {
	cursor, err := h.db.Collection("saved_charts").Find(ctx, filter,
		options.Find().SetSort(bson.D{{Key: "updated_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	charts := make([]models.SavedChart, 0)
	if err := cursor.All(ctx, &charts); err != nil {
		return nil, err
	}

	ids := make([]primitive.ObjectID, 0, len(charts))
	for _, c := range charts {
		ids = append(ids, c.CardID)
	}
	cards, err := h.cardsByID(ctx, ids)
	if err != nil {
		return nil, err
	}
	for i := range charts {
		if charts[i].Indicators == nil {
			charts[i].Indicators = []models.ChartIndicator{}
		}
		if charts[i].Source == "" {
			charts[i].Source = "all"
		}
		if charts[i].ChartType == "" {
			charts[i].ChartType = "line"
		}
		if charts[i].Drawings == nil {
			charts[i].Drawings = []models.ChartDrawing{}
		}
		if card, ok := cards[charts[i].CardID]; ok {
			charts[i].CardName = card.Name
			charts[i].CardImageURL = card.ImageURL
			charts[i].CardGame = card.Game
		}
	}
	return charts, nil
}

// parseChartRequest decodes and validates a chart payload, confirming the card exists.
func (h *Handlers) parseChartRequest(w http.ResponseWriter, r *http.Request, ctx context.Context, userType string) (*models.SavedChartRequest, primitive.ObjectID, bool) {
	var req models.SavedChartRequest
	if err := decodeJSON(w, r, &req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest, nil)
		return nil, primitive.NilObjectID, false
	}
	if err := validateSavedChart(&req, getMaxIndicators(userType)); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest, nil)
		return nil, primitive.NilObjectID, false
	}
	cardID, err := primitive.ObjectIDFromHex(req.CardID)
	if err != nil {
		h.sendError(w, "Invalid card ID", http.StatusBadRequest, nil)
		return nil, primitive.NilObjectID, false
	}
	exists, err := h.cardExists(ctx, cardID)
	if err != nil {
		h.sendError(w, "Error validating card", http.StatusInternalServerError, nil)
		return nil, primitive.NilObjectID, false
	}
	if !exists {
		h.sendError(w, "Card not found", http.StatusBadRequest, nil)
		return nil, primitive.NilObjectID, false
	}
	return &req, cardID, true
}

// SaveChart saves a new chart for the user
func (h *Handlers) SaveChart(w http.ResponseWriter, r *http.Request) {
	userID, claims, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, cardID, ok := h.parseChartRequest(w, r, ctx, claims.UserType)
	if !ok {
		return
	}

	now := time.Now().UTC()
	chart := models.SavedChart{
		UserID:      userID,
		CardID:      cardID,
		Name:        req.Name,
		Description: req.Description,
		Indicators:  req.Indicators,
		TimeRange:   req.TimeRange,
		Source:      req.Source,
		ChartType:   req.ChartType,
		ShowVolume:  req.ShowVolume,
		Drawings:    req.Drawings,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := h.db.Collection("saved_charts").InsertOne(ctx, chart)
	if err != nil {
		h.sendError(w, "Error saving chart", http.StatusInternalServerError, nil)
		return
	}
	chart.ID = result.InsertedID.(primitive.ObjectID)

	sendJSON(w, http.StatusCreated, chart)
}

// GetChart retrieves one of the user's saved charts
func (h *Handlers) GetChart(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	chartID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid chart ID", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	charts, err := h.loadCharts(ctx, bson.M{"_id": chartID, "user_id": userID})
	if err != nil {
		h.sendError(w, "Error retrieving chart", http.StatusInternalServerError, nil)
		return
	}
	if len(charts) == 0 {
		h.sendError(w, "Chart not found", http.StatusNotFound, nil)
		return
	}
	sendJSON(w, http.StatusOK, charts[0])
}

// UpdateChart replaces the editable fields of a saved chart
func (h *Handlers) UpdateChart(w http.ResponseWriter, r *http.Request) {
	userID, claims, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	chartID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid chart ID", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, cardID, ok := h.parseChartRequest(w, r, ctx, claims.UserType)
	if !ok {
		return
	}

	var updated models.SavedChart
	err = h.db.Collection("saved_charts").FindOneAndUpdate(ctx,
		bson.M{"_id": chartID, "user_id": userID},
		bson.M{"$set": bson.M{
			"card_id":     cardID,
			"name":        req.Name,
			"description": req.Description,
			"indicators":  req.Indicators,
			"time_range":  req.TimeRange,
			"source":      req.Source,
			"chart_type":  req.ChartType,
			"show_volume": req.ShowVolume,
			"drawings":    req.Drawings,
			"updated_at":  time.Now().UTC(),
		}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)
	if err == mongo.ErrNoDocuments {
		h.sendError(w, "Chart not found", http.StatusNotFound, nil)
		return
	}
	if err != nil {
		h.sendError(w, "Error updating chart", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, updated)
}

// GetSavedCharts retrieves user's saved charts
func (h *Handlers) GetSavedCharts(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	charts, err := h.loadCharts(ctx, bson.M{"user_id": userID})
	if err != nil {
		h.sendError(w, "Error retrieving charts", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, charts)
}

// DeleteChart deletes a user's saved chart
func (h *Handlers) DeleteChart(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	chartID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid chart ID", http.StatusBadRequest, nil)
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
		h.sendError(w, "Error deleting chart", http.StatusInternalServerError, nil)
		return
	}

	if result.DeletedCount == 0 {
		h.sendError(w, "Chart not found", http.StatusNotFound, nil)
		return
	}

	sendJSON(w, http.StatusOK, map[string]string{"message": "Chart deleted successfully"})
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
