package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test database connection
	err := h.db.Client().Ping(ctx, nil)
	status := "healthy"
	if err != nil {
		status = "unhealthy"
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	// Build search filter
	filter := bson.M{}

	// Text search if query provided
	if query != "" {
		// Use MongoDB text search index
		filter["$text"] = bson.M{"$search": query}
	}

	// Game filter
	if game != "" {
		filter["game"] = bson.M{"$regex": primitive.Regex{Pattern: game, Options: "i"}}
	}

	// Category filter
	if category != "" {
		filter["category"] = category
	}

	// Count total results
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		http.Error(w, "Error counting documents", http.StatusInternalServerError)
		return
	}

	// Calculate pagination
	skip := (page - 1) * limit
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	// Build find options
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(limit))

	// Sort by text score if searching, otherwise by updated_at
	if query != "" {
		findOptions.SetSort(bson.M{"score": bson.M{"$meta": "textScore"}})
	} else {
		findOptions.SetSort(bson.M{"updated_at": -1})
	}

	// Execute search
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		http.Error(w, "Error searching cards", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var cards []models.Card
	if err = cursor.All(ctx, &cards); err != nil {
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
		// Don't fail if listings aren't found, just return empty
		fmt.Printf("Warning: Could not retrieve listings: %v\n", err)
	}

	var listings []interface{}
	if listingsCursor != nil {
		defer listingsCursor.Close(ctx)
		if err = listingsCursor.All(ctx, &listings); err != nil {
			fmt.Printf("Warning: Could not decode listings: %v\n", err)
			listings = []interface{}{} // Empty listings
		}
	}

	// Build response
	response := map[string]interface{}{
		"prices":   prices,
		"listings": listings,
		"range":    timeRange,
		"total":    len(prices),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Register new user
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := h.validateRegisterRequest(&req); err != nil {
		h.sendError(w, "Validation failed", http.StatusBadRequest, map[string]interface{}{
			"validation_errors": err.Error(),
		})
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
	passwordHash := h.hashPassword(req.Password)

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
		h.sendError(w, "Failed to create user", http.StatusInternalServerError, nil)
		return
	}

	user.ID = result.InsertedID.(primitive.ObjectID)

	// Generate JWT token
	token, err := h.generateJWT(user)
	if err != nil {
		h.sendError(w, "Failed to generate token", http.StatusInternalServerError, nil)
		return
	}

	// Return success response
	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Login user
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

	// Check if user is active
	if !user.IsActive {
		h.sendError(w, "Account is disabled", http.StatusUnauthorized, nil)
		return
	}

	// Update last login
	now := time.Now().UTC()
	user.LastLoginAt = &now
	collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{
		"$set": bson.M{
			"last_login_at": now,
			"updated_at":    now,
		},
	})

	// Generate JWT token
	token, err := h.generateJWT(user)
	if err != nil {
		h.sendError(w, "Failed to generate token", http.StatusInternalServerError, nil)
		return
	}

	// Return success response
	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout user
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// For JWT-based auth, logout is typically handled client-side
	// We could implement a token blacklist here if needed
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
}

// GetDashboard returns user dashboard data
func (h *Handlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
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

// Helper functions

func (h *Handlers) validateRegisterRequest(req *models.RegisterRequest) error {
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return fmt.Errorf("all fields are required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	return nil
}

func (h *Handlers) hashPassword(password string) string {
	// Simple hash for demo - use bcrypt in production
	hasher := hmac.New(sha256.New, h.config.JWTSecret)
	hasher.Write([]byte(password))
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

func (h *Handlers) verifyPassword(password, hash string) bool {
	return h.hashPassword(password) == hash
}

func (h *Handlers) generateJWT(user models.User) (string, error) {
	// Simple JWT implementation for demo
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