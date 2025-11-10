package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"github.com/jamesc159/monmetrics/configs"
	"github.com/jamesc159/monmetrics/internal/models"
)

type Handlers struct {
	db     *mongo.Database
	config *configs.Config
}

func New(db *mongo.Database, config *configs.Config) *Handlers {
	return &Handlers{
		db:     db,
		config: config,
	}
}

// Health check endpoint
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("🏥 Health check request from %s\n", r.RemoteAddr)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test database connection
	err := h.db.Client().Ping(ctx, nil)
	status := "healthy"
	if err != nil {
		status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Printf("❌ Database ping failed: %v\n", err)
	} else {
		fmt.Printf("✅ Database ping successful\n")
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("❌ Error encoding health response: %v\n", err)
	} else {
		fmt.Printf("✅ Health check response sent\n")
	}
}

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

// Authentication handlers

// Register handles user registration
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := h.validateRegisterRequest(&req); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	collection := h.db.Collection("users")
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		h.sendError(w, "User already exists", http.StatusConflict, nil)
		return
	}

	// Hash password
	passwordHash, err := h.hashPassword(req.Password)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		http.Error(w, "Error processing registration", http.StatusInternalServerError)
		return
	}

	// Create new user
	user := models.User{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		UserType:     "free",
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
		IsActive:     true,
	}

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		fmt.Printf("Error creating user: %v\n", err)
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	// Generate JWT token
	token, err := h.generateJWT(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login handles user authentication
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user by email
	collection := h.db.Collection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		h.sendError(w, "Invalid credentials", http.StatusUnauthorized, nil)
		return
	}

	// Verify password
	if !h.verifyPassword(req.Password, user.PasswordHash) {
		h.sendError(w, "Invalid credentials", http.StatusUnauthorized, nil)
		return
	}

	// Update last login
	collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{
		"$set": bson.M{
			"last_login_at": time.Now().UTC(),
		},
	})

	// Generate JWT token
	token, err := h.generateJWT(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// For JWT-based authentication, logout is typically handled client-side
	// The token is removed from client storage
	// In a production environment, you might want to implement a token blacklist
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
		"success": "true",
	})
}

// Dashboard handlers

// GetDashboard retrieves user dashboard data
func (h *Handlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(*Claims)
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
	claims, ok := r.Context().Value("claims").(*Claims)
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
	claims, ok := r.Context().Value("claims").(*Claims)
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
	claims, ok := r.Context().Value("claims").(*Claims)
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

// Featured content and organized search handlers

// GetFeaturedContent retrieves active featured content for the carousel
func (h *Handlers) GetFeaturedContent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.db.Collection("featured_content")

	// Get active featured content, sorted by priority
	filter := bson.M{
		"active": true,
		"$or": []bson.M{
			{"expires_at": bson.M{"$exists": false}},
			{"expires_at": bson.M{"$gt": time.Now()}},
		},
	}

	// Use bson.D for ordered sort (priority first, then created_at)
	findOptions := options.Find().SetSort(bson.D{{"priority", -1}, {"created_at", -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		fmt.Printf("Error retrieving featured content: %v\n", err)
		// Return empty array instead of error if collection doesn't exist
		featuredContent := make([]models.FeaturedContent, 0)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(featuredContent)
		return
	}
	defer cursor.Close(ctx)

	featuredContent := make([]models.FeaturedContent, 0)
	if err = cursor.All(ctx, &featuredContent); err != nil {
		fmt.Printf("Error decoding featured content: %v\n", err)
		// Return empty array instead of error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(featuredContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(featuredContent)
}

// GetCardsByGame retrieves cards organized by game with popularity sorting
func (h *Handlers) GetCardsByGame(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := h.db.Collection("cards")

	// Get all unique games
	games, err := collection.Distinct(ctx, "game", bson.M{"category": "card"})
	if err != nil {
		fmt.Printf("Error retrieving games: %v\n", err)
		// Return empty array instead of error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.GameCardGroup{})
		return
	}

	fmt.Printf("Found %d unique games for cards\n", len(games))

	// For each game, get top cards by popularity (limit to 12 for 2 rows of 6)
	gameGroups := make([]models.GameCardGroup, 0)

	for _, game := range games {
		gameStr, ok := game.(string)
		if !ok {
			continue
		}

		// Get total count
		totalCount, err := collection.CountDocuments(ctx, bson.M{
			"game":     gameStr,
			"category": "card",
		})
		if err != nil {
			fmt.Printf("Error counting cards for game %s: %v\n", gameStr, err)
			continue
		}

		// Get top 12 cards by popularity
		findOptions := options.Find().
			SetSort(bson.D{{"popularity_rank", 1}, {"current_price", -1}}).
			SetLimit(12)

		cursor, err := collection.Find(ctx, bson.M{
			"game":     gameStr,
			"category": "card",
		}, findOptions)

		if err != nil {
			fmt.Printf("Error retrieving cards for game %s: %v\n", gameStr, err)
			continue
		}

		var cards []models.Card
		if err = cursor.All(ctx, &cards); err != nil {
			cursor.Close(ctx)
			fmt.Printf("Error decoding cards for game %s: %v\n", gameStr, err)
			continue
		}
		cursor.Close(ctx)

		gameGroups = append(gameGroups, models.GameCardGroup{
			Game:       gameStr,
			Category:   "card",
			Cards:      cards,
			TotalCount: int(totalCount),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameGroups)
}

// GetSealedByGame retrieves sealed products organized by game with popularity sorting
func (h *Handlers) GetSealedByGame(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := h.db.Collection("cards")

	// Get all unique games that have sealed products
	games, err := collection.Distinct(ctx, "game", bson.M{"category": "sealed"})
	if err != nil {
		fmt.Printf("Error retrieving games: %v\n", err)
		// Return empty array instead of error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.GameCardGroup{})
		return
	}

	fmt.Printf("Found %d unique games for sealed\n", len(games))

	// For each game, get top sealed products by popularity (limit to 12 for 2 rows of 6)
	gameGroups := make([]models.GameCardGroup, 0)

	for _, game := range games {
		gameStr, ok := game.(string)
		if !ok {
			continue
		}

		// Get total count
		totalCount, err := collection.CountDocuments(ctx, bson.M{
			"game":     gameStr,
			"category": "sealed",
		})
		if err != nil {
			fmt.Printf("Error counting sealed for game %s: %v\n", gameStr, err)
			continue
		}

		// Get top 12 sealed products by popularity
		findOptions := options.Find().
			SetSort(bson.D{{"popularity_rank", 1}, {"current_price", -1}}).
			SetLimit(12)

		cursor, err := collection.Find(ctx, bson.M{
			"game":     gameStr,
			"category": "sealed",
		}, findOptions)

		if err != nil {
			fmt.Printf("Error retrieving sealed for game %s: %v\n", gameStr, err)
			continue
		}

		var cards []models.Card
		if err = cursor.All(ctx, &cards); err != nil {
			cursor.Close(ctx)
			fmt.Printf("Error decoding sealed for game %s: %v\n", gameStr, err)
			continue
		}
		cursor.Close(ctx)

		gameGroups = append(gameGroups, models.GameCardGroup{
			Game:       gameStr,
			Category:   "sealed",
			Cards:      cards,
			TotalCount: int(totalCount),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameGroups)
}

// Helper functions

// normalizeEmail converts email to lowercase and trims whitespace
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// sanitizeName removes dangerous characters and normalizes whitespace
func sanitizeName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)

	// Remove control characters and normalize spaces
	var result strings.Builder
	prevSpace := false
	for _, r := range name {
		if unicode.IsControl(r) {
			continue
		}
		if unicode.IsSpace(r) {
			if !prevSpace {
				result.WriteRune(' ')
				prevSpace = true
			}
			continue
		}
		result.WriteRune(r)
		prevSpace = false
	}

	return result.String()
}

// validateEmail validates email format using regex
func validateEmail(email string) bool {
	// RFC 5322 simplified email regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_` + "`" + `{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	return emailRegex.MatchString(email) && len(email) <= 254
}

// validatePassword checks password strength according to OWASP guidelines
func validatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	if len(password) > 128 {
		return fmt.Errorf("password must not exceed 128 characters")
	}

	// Check for at least one uppercase, lowercase, digit, and special character
	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, one number, and one special character")
	}

	return nil
}

// validateName checks if name is valid
func validateName(name string) error {
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}

	if len(name) > 50 {
		return fmt.Errorf("name must not exceed 50 characters")
	}

	// Check for at least one letter
	hasLetter := false
	for _, r := range name {
		if unicode.IsLetter(r) {
			hasLetter = true
			break
		}
	}

	if !hasLetter {
		return fmt.Errorf("name must contain at least one letter")
	}

	return nil
}

func (h *Handlers) validateRegisterRequest(req *models.RegisterRequest) error {
	// Normalize inputs
	req.Email = normalizeEmail(req.Email)
	req.FirstName = sanitizeName(req.FirstName)
	req.LastName = sanitizeName(req.LastName)

	// Validate all fields are present
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return fmt.Errorf("all fields are required")
	}

	// Validate email format
	if !validateEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	// Validate password strength
	if err := validatePassword(req.Password); err != nil {
		return err
	}

	// Validate names
	if err := validateName(req.FirstName); err != nil {
		return fmt.Errorf("invalid first name: %v", err)
	}

	if err := validateName(req.LastName); err != nil {
		return fmt.Errorf("invalid last name: %v", err)
	}

	return nil
}

func (h *Handlers) hashPassword(password string) (string, error) {
	// Use bcrypt with cost factor 12 (OWASP recommended minimum is 10)
	// Cost 12 provides good balance between security and performance
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func (h *Handlers) verifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (h *Handlers) generateJWT(user models.User) (string, error) {
	// JWT implementation using HMAC-SHA256 (consistent with middleware)
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}

	payload := map[string]interface{}{
		"user_id":   user.ID.Hex(),
		"email":     user.Email,
		"user_type": user.UserType,
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
	}

	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	message := headerB64 + "." + payloadB64

	// Sign with HMAC-SHA256
	hasher := hmac.New(sha256.New, h.config.JWTSecret)
	hasher.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))

	return message + "." + signature, nil
}

func (h *Handlers) sendError(w http.ResponseWriter, message string, statusCode int, data map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   message,
		"success": false,
	}

	if data != nil {
		for k, v := range data {
			response[k] = v
		}
	}

	json.NewEncoder(w).Encode(response)
}

func calculateIndicatorsUsed(charts []models.SavedChart) int {
	indicatorMap := make(map[string]bool)
	for _, chart := range charts {
		for _, indicator := range chart.Indicators {
			indicatorMap[indicator.Type] = true
		}
	}
	return len(indicatorMap)
}

func getMaxIndicators(userType string) int {
	if userType == "paid" {
		return 10
	}
	return 3 // free users
}

// JWT Claims structure for middleware
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	Exp      int64  `json:"exp"`
}
