package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/models"
)

const maxWarmupDays = 400

// validGradePattern accepts both eBay grades (raw, psa-10, bgs-9.5) and TCGplayer
// printings (normal, foil, 1st-edition, reverse-holofoil). Values are used only as exact
// match terms, so the charset bound is about rejecting junk, not escaping.
var validGradePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{0,31}$`)

// searchSorts whitelists the sort orders a client may request.
var searchSorts = map[string]bson.D{
	"relevance":  {{Key: "current_price", Value: -1}, {Key: "updated_at", Value: -1}},
	"price_desc": {{Key: "current_price", Value: -1}},
	"price_asc":  {{Key: "current_price", Value: 1}},
	"name":       {{Key: "name", Value: 1}},
	"recent":     {{Key: "updated_at", Value: -1}},
}

// searchFilters carries the validated query parameters for a catalog search.
type searchFilters struct {
	Query       string
	Game        string
	Category    string
	Set         string
	ProductType string
	Rarity      string
	PricedOnly  bool
}

// buildSearchFilter creates a robust search filter without relying on text indexes
func (h *Handlers) buildSearchFilter(f searchFilters) bson.M {
	filter := bson.M{"active": bson.M{"$ne": false}}

	// Handle text search using regex patterns (more reliable)
	if query := strings.TrimSpace(f.Query); query != "" {
		// User input is escaped: it is pattern data, never a pattern.
		escapedQuery := regexp.QuoteMeta(query)
		orConditions := []bson.M{
			{"name": bson.M{"$regex": primitive.Regex{Pattern: escapedQuery, Options: "i"}}},
			{"set": bson.M{"$regex": primitive.Regex{Pattern: escapedQuery, Options: "i"}}},
			{"game": bson.M{"$regex": primitive.Regex{Pattern: escapedQuery, Options: "i"}}},
		}

		// Indexed token match: the only branch that scales past a small catalog.
		lowerQuery := strings.ToLower(query)
		tokens := []string{lowerQuery}
		words := strings.Fields(lowerQuery)
		if len(words) > 1 {
			tokens = append(tokens, words...)
		}
		orConditions = append(orConditions, bson.M{"search_terms": bson.M{"$in": tokens}})

		if len(words) > 1 {
			for _, word := range words {
				if len(word) > 2 { // Only search for words longer than 2 characters
					orConditions = append(orConditions,
						bson.M{"name": bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(word), Options: "i"}}},
					)
				}
			}
		}

		filter["$or"] = orConditions
	}

	if f.Game != "" {
		filter["game"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(f.Game), Options: "i"}}
	}
	if f.Category != "" {
		filter["category"] = f.Category
	}
	if f.Set != "" {
		filter["set"] = bson.M{"$regex": primitive.Regex{Pattern: "^" + regexp.QuoteMeta(f.Set) + "$", Options: "i"}}
	}
	if f.ProductType != "" {
		filter["product_type"] = f.ProductType
	}
	if f.Rarity != "" {
		filter["rarity"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(f.Rarity), Options: "i"}}
	}
	if f.PricedOnly {
		filter["current_price"] = bson.M{"$gt": 0}
	}

	return h.applyEnabledGames(filter)
}

// SearchCards searches for cards based on query parameters
func (h *Handlers) SearchCards(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filters := searchFilters{
		Query:       q.Get("q"),
		Game:        q.Get("game"),
		Category:    q.Get("category"),
		Set:         q.Get("set"),
		ProductType: q.Get("product_type"),
		Rarity:      q.Get("rarity"),
		PricedOnly:  q.Get("priced") == "true",
	}
	if filters.Category != "" && !validCatalogCategories[filters.Category] {
		h.sendError(w, "Invalid category", http.StatusBadRequest, nil)
		return
	}

	sortKey := q.Get("sort")
	if sortKey == "" {
		sortKey = "relevance"
	}
	sort, ok := searchSorts[sortKey]
	if !ok {
		h.sendError(w, "Invalid sort", http.StatusBadRequest, nil)
		return
	}

	// Parse pagination parameters
	page := 1
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		page = p
	}

	limit := 20
	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 && l <= 100 {
		limit = l
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.db.Collection("cards")

	// Build search filter using the robust method
	filter := h.buildSearchFilter(filters)

	// Count total results with improved error handling
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		fmt.Printf("Error counting documents: %v\n", err)
		// Instead of returning an error immediately, try a simpler filter

		// If there's a complex query causing issues, try with just basic filters
		fallbackFilter := bson.M{}
		if filters.Game != "" {
			fallbackFilter["game"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(filters.Game), Options: "i"}}
		}
		if filters.Category != "" {
			fallbackFilter["category"] = filters.Category
		}

		// If we still have a query but the complex search failed, try simple name search
		if cleanQuery := strings.TrimSpace(filters.Query); cleanQuery != "" {
			fallbackFilter["name"] = bson.M{"$regex": primitive.Regex{Pattern: regexp.QuoteMeta(cleanQuery), Options: "i"}}
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
	findOptions.SetSort(sort)

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
	h.stripHiddenImages(cards)

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
	if h.imageHidden(card.Game) {
		card.ImageURL = ""
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
	grade := r.URL.Query().Get("grade")
	if grade == "" {
		grade = "raw"
	}
	if !validGradePattern.MatchString(grade) {
		h.sendError(w, "Invalid grade", http.StatusBadRequest, nil)
		return
	}
	source := r.URL.Query().Get("source")
	if source != "" && !validChartSources[source] {
		h.sendError(w, "Invalid source", http.StatusBadRequest, nil)
		return
	}

	// Parse time range
	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "30d"
	}

	// Calculate start date based on range
	var startDate time.Time
	now := time.Now().UTC()

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

	// Extra history before the visible range lets clients seed indicators such as EMA
	queryStart := startDate
	if wd := r.URL.Query().Get("warmup_days"); wd != "" {
		warmup, err := strconv.Atoi(wd)
		if err != nil || warmup < 0 {
			h.sendError(w, "warmup_days must be a non-negative integer", http.StatusBadRequest, nil)
			return
		}
		if warmup > maxWarmupDays {
			warmup = maxWarmupDays
		}
		queryStart = startDate.AddDate(0, 0, -warmup)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get price history
	pricesCollection := h.db.Collection("prices")

	filter := bson.M{
		"card_id": objectID,
		"timestamp": bson.M{
			"$gte": queryStart.UTC().Truncate(24 * time.Hour),
			"$lte": now,
		},
	}
	if grade == "raw" {
		filter["$or"] = bson.A{
			bson.M{"grade": "raw"}, bson.M{"grade": ""}, bson.M{"grade": bson.M{"$exists": false}},
		}
	} else {
		filter["grade"] = grade
	}
	if source != "" && source != "all" {
		filter["source"] = source
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
	if prices == nil {
		prices = []models.PricePoint{}
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
	marketFilter := bson.M{
		"card_id": objectID,
		"date": bson.M{
			"$gte": startDate.UTC().Truncate(24 * time.Hour),
			"$lte": now,
		},
	}
	if grade == "raw" {
		marketFilter["$or"] = bson.A{
			bson.M{"grade": "raw"}, bson.M{"grade": ""}, bson.M{"grade": bson.M{"$exists": false}},
		}
	} else {
		marketFilter["grade"] = grade
	}
	if source != "" && source != "all" {
		marketFilter["source"] = source
	}
	marketCursor, err := marketDataCollection.Find(ctx, marketFilter)

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
		"prices":       prices,
		"listings":     listings,
		"market_data":  marketData,
		"card_id":      objectID.Hex(),
		"time_range":   timeRange,
		"visible_from": startDate.UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
