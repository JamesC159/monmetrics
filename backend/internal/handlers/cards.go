package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/models"
)

// buildSearchFilter creates a robust search filter without relying on text indexes
func (h *Handlers) buildSearchFilter(query, game, category string) bson.M {
	filter := bson.M{}

	// Handle text search using regex patterns (more reliable)
	if query != "" {
		cleanQuery := strings.TrimSpace(query)
		if cleanQuery != "" {
			// Create an array of regex patterns for different search strategies
			orConditions := []bson.M{
				{"name": bson.M{"$regex": primitive.Regex{Pattern: cleanQuery, Options: "i"}}},
				{"set": bson.M{"$regex": primitive.Regex{Pattern: cleanQuery, Options: "i"}}},
				{"game": bson.M{"$regex": primitive.Regex{Pattern: cleanQuery, Options: "i"}}},
			}

			// Add search terms array match
			lowerQuery := strings.ToLower(cleanQuery)
			orConditions = append(orConditions, bson.M{"search_terms": bson.M{"$in": []string{lowerQuery}}})

			// Split query into words for partial matching
			words := strings.Fields(cleanQuery)
			if len(words) > 1 {
				for _, word := range words {
					if len(word) > 2 { // Only search for words longer than 2 characters
						orConditions = append(orConditions,
							bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: word, Options: "i"}}},
						)
					}
				}
			}

			filter["$or"] = orConditions
		}
	}

	// Game filter
	if game != "" {
		filter["game"] = bson.M{"$regex": primitive.Regex{Pattern: game, Options: "i"}}
	}

	// Category filter
	if category != "" {
		filter["category"] = category
	}

	return filter
}

// SearchCards searches for cards based on query parameters
func (h *Handlers) SearchCards(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	game := r.URL.Query().Get("game")
	category := r.URL.Query().Get("category")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Parse pagination parameters
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.db.Collection("cards")

	// Build search filter using the robust method
	filter := h.buildSearchFilter(query, game, category)

	// Count total results with improved error handling
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		fmt.Printf("Error counting documents: %v\n", err)
		// Instead of returning an error immediately, try a simpler filter

		// If there's a complex query causing issues, try with just basic filters
		fallbackFilter := bson.M{}
		if game != "" {
			fallbackFilter["game"] = bson.M{"$regex": primitive.Regex{Pattern: game, Options: "i"}}
		}
		if category != "" {
			fallbackFilter["category"] = category
		}

		// If we still have a query but the complex search failed, try simple name search
		if query != "" {
			cleanQuery := strings.TrimSpace(query)
			if cleanQuery != "" {
				fallbackFilter["name"] = bson.M{"$regex": primitive.Regex{Pattern: cleanQuery, Options: "i"}}
			}
		}

		// Try the fallback count
		total, err = collection.CountDocuments(ctx, fallbackFilter)
		if err != nil {
			fmt.Printf("Error counting documents (fallback): %v\n", err)
			http.Error(w, "Error retrieving search results", http.StatusInternalServerError)
			return
		}

		// Use the fallback filter for the actual search too
		filter = fallbackFilter
	}

	// Calculate pagination
	skip := (page - 1) * limit
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Build find options
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))
	findOptions.SetSort(bson.M{"updated_at": -1})

	// Execute search
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		fmt.Printf("Error executing search: %v\n", err)
		http.Error(w, "Error executing search", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	// Initialize as empty slice instead of nil to ensure JSON serializes as [] not null
	cards := make([]models.Card, 0)
	if err = cursor.All(ctx, &cards); err != nil {
		fmt.Printf("Error decoding results: %v\n", err)
		http.Error(w, "Error decoding results", http.StatusInternalServerError)
		return
	}

	// Build response
	response := models.SearchResult{
		Cards:      cards,
		Total:      int(total),
		Page:       page,
		PerPage:    limit,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCard retrieves a specific card by ID
func (h *Handlers) GetCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Card ID required", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := h.db.Collection("cards")

	var card models.Card
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Card not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Error retrieving card", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}

// GetCardPrices retrieves price history for a specific card
func (h *Handlers) GetCardPrices(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Card ID required", http.StatusBadRequest)
		return
	}

	objectID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	// Parse time range
	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "30d"
	}

	// Calculate start date based on range
	var startDate time.Time
	now := time.Now()

	switch timeRange {
	case "1d":
		startDate = now.AddDate(0, 0, -1)
	case "7d":
		startDate = now.AddDate(0, 0, -7)
	case "30d":
		startDate = now.AddDate(0, 0, -30)
	case "90d":
		startDate = now.AddDate(0, 0, -90)
	case "1y":
		startDate = now.AddDate(-1, 0, 0)
	case "5y":
		startDate = now.AddDate(-5, 0, 0)
	default:
		startDate = now.AddDate(0, 0, -30) // Default to 30 days
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get price history
	pricesCollection := h.db.Collection("prices")

	filter := bson.M{
		"card_id": objectID,
		"timestamp": bson.M{
			"$gte": startDate,
			"$lte": now,
		},
	}

	// Sort by timestamp
	findOptions := options.Find().SetSort(bson.M{"timestamp": 1})

	cursor, err := pricesCollection.Find(ctx, filter, findOptions)
	if err != nil {
		http.Error(w, "Error retrieving price history", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var prices []models.PricePoint
	if err = cursor.All(ctx, &prices); err != nil {
		http.Error(w, "Error decoding price history", http.StatusInternalServerError)
		return
	}

	// Get current listings
	listingsCollection := h.db.Collection("listings")
	listingsCursor, err := listingsCollection.Find(ctx, bson.M{"card_id": objectID})
	if err != nil {
		// Log error but continue - listings are optional
		fmt.Printf("Warning: Could not retrieve listings: %v\n", err)
	}

	var listings []models.Listing
	if listingsCursor != nil {
		defer listingsCursor.Close(ctx)
		if err = listingsCursor.All(ctx, &listings); err != nil {
			fmt.Printf("Warning: Could not decode listings: %v\n", err)
			listings = []models.Listing{} // Ensure we have an empty slice
		}
	}

	// Get market data
	marketDataCollection := h.db.Collection("market_data")
	marketCursor, err := marketDataCollection.Find(ctx, bson.M{
		"card_id": objectID,
		"date": bson.M{
			"$gte": startDate,
			"$lte": now,
		},
	})

	var marketData []models.MarketData
	if err != nil {
		fmt.Printf("Warning: Could not retrieve market data: %v\n", err)
		marketData = []models.MarketData{} // Ensure we have an empty slice
	} else {
		defer marketCursor.Close(ctx)
		if err = marketCursor.All(ctx, &marketData); err != nil {
			fmt.Printf("Warning: Could not decode market data: %v\n", err)
			marketData = []models.MarketData{} // Ensure we have an empty slice
		}
	}

	// Build response
	response := map[string]interface{}{
		"prices":      prices,
		"listings":    listings,
		"market_data": marketData,
		"card_id":     objectID.Hex(),
		"time_range":  timeRange,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
