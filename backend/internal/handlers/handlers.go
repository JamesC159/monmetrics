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
	"strings"
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
	token, err := h.generateJWT(&user)
	if err != nil {
		h.sendError(w, "Failed to generate token", http.StatusInternalServerError, nil)
		return
	}

	// Return response
	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
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

	// Update last login
	now := time.Now().UTC()
	collection.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{
		"$set": bson.M{"last_login_at": now},
	})
	user.LastLoginAt = &now

	// Generate JWT token
	token, err := h.generateJWT(&user)
	if err != nil {
		h.sendError(w, "Failed to generate token", http.StatusInternalServerError, nil)
		return
	}

	// Return response
	response := models.AuthResponse{
		Token: token,
		User:  user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout user (simple implementation)
func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	// In a more complex implementation, you might invalidate the token
	// For now, we'll just return success as the client will remove the token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// Search cards
func (h *Handlers) SearchCards(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	game := r.URL.Query().Get("game")
	category := r.URL.Query().Get("category")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	// Parse pagination
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

	// Build search filter
	filter := bson.M{}

	if query != "" {
		filter["$text"] = bson.M{"$search": query}
	}

	if game != "" {
		filter["game"] = game
	}

	if category != "" {
		filter["category"] = category
	}

	collection := h.db.Collection("cards")

	// Get total count
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		h.sendError(w, "Search failed", http.StatusInternalServerError, nil)
		return
	}

	// Find cards with pagination
	opts := options.Find()
	opts.SetSkip(int64((page - 1) * limit))
	opts.SetLimit(int64(limit))
	opts.SetSort(bson.D{{"current_price", -1}}) // Sort by price descending

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		h.sendError(w, "Search failed", http.StatusInternalServerError, nil)
		return
	}
	defer cursor.Close(ctx)

	var cards []models.Card
	if err := cursor.All(ctx, &cards); err != nil {
		h.sendError(w, "Failed to decode results", http.StatusInternalServerError, nil)
		return
	}

	// Calculate total pages
	totalPages := int((total + int64(limit) - 1) / int64(limit))

	result := models.SearchResult{
		Cards:      cards,
		Total:      int(total),
		Page:       page,
		PerPage:    limit,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Get single card
func (h *Handlers) GetCard(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Card ID required", http.StatusBadRequest)
		return
	}

	cardID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var card models.Card
	collection := h.db.Collection("cards")
	err = collection.FindOne(ctx, bson.M{"_id": cardID}).Decode(&card)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Card not found", http.StatusNotFound)
			return
		}
		h.sendError(w, "Failed to fetch card", http.StatusInternalServerError, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}

// Get card prices
func (h *Handlers) GetCardPrices(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Card ID required", http.StatusBadRequest)
		return
	}

	cardID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
		return
	}

	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "30d"
	}

	// Calculate time filter
	var startTime time.Time
	switch timeRange {
	case "1d":
		startTime = time.Now().AddDate(0, 0, -1)
	case "7d":
		startTime = time.Now().AddDate(0, 0, -7)
	case "30d":
		startTime = time.Now().AddDate(0, 0, -30)
	case "90d":
		startTime = time.Now().AddDate(0, 0, -90)
	case "1y":
		startTime = time.Now().AddDate(-1, 0, 0)
	case "5y":
		startTime = time.Now().AddDate(-5, 0, 0)
	default:
		startTime = time.Now().AddDate(0, 0, -30)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get price points
	pricesCollection := h.db.Collection("price_points")
	filter := bson.M{
		"card_id":   cardID,
		"timestamp": bson.M{"$gte": startTime},
	}

	cursor, err := pricesCollection.Find(ctx, filter, options.Find().SetSort(bson.D{{"timestamp", 1}}))
	if err != nil {
		h.sendError(w, "Failed to fetch prices", http.StatusInternalServerError, nil)
		return
	}
	defer cursor.Close(ctx)

	var prices []models.PricePoint
	if err := cursor.All(ctx, &prices); err != nil {
		h.sendError(w, "Failed to decode prices", http.StatusInternalServerError, nil)
		return
	}

	// Get market data
	marketDataCollection := h.db.Collection("market_data")
	marketCursor, err := marketDataCollection.Find(ctx, bson.M{
		"card_id": cardID,
		"date":    bson.M{"$gte": startTime},
	}, options.Find().SetSort(bson.D{{"date", 1}}))
	if err != nil {
		h.sendError(w, "Failed to fetch market data", http.StatusInternalServerError, nil)
		return
	}
	defer marketCursor.Close(ctx)

	var marketData []models.MarketData
	if err := marketCursor.All(ctx, &marketData); err != nil {
		h.sendError(w, "Failed to decode market data", http.StatusInternalServerError, nil)
		return
	}

	result := models.PriceHistory{
		Prices:     prices,
		MarketData: marketData,
		Indicators: make(map[string][]models.IndicatorPoint),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Get user dashboard
func (h *Handlers) GetDashboard(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	userID, _ := primitive.ObjectIDFromHex(claims.UserID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get saved charts
	chartsCollection := h.db.Collection("saved_charts")
	cursor, err := chartsCollection.Find(ctx, bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{"created_at", -1}}).SetLimit(10))
	if err != nil {
		h.sendError(w, "Failed to fetch dashboard", http.StatusInternalServerError, nil)
		return
	}
	defer cursor.Close(ctx)

	var savedCharts []models.SavedChart
	if err := cursor.All(ctx, &savedCharts); err != nil {
		h.sendError(w, "Failed to decode charts", http.StatusInternalServerError, nil)
		return
	}

	dashboard := models.Dashboard{
		SavedCharts:    savedCharts,
		RecentlyViewed: []models.Card{}, // TODO: Implement recently viewed
		UserStats: models.UserStats{
			ChartsCreated: len(savedCharts),
			MaxIndicators: getMaxIndicators(claims.UserType),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dashboard)
}

// Save chart
func (h *Handlers) SaveChart(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	userID, _ := primitive.ObjectIDFromHex(claims.UserID)

	var chart models.SavedChart
	if err := json.NewDecoder(r.Body).Decode(&chart); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate indicator limits
	maxIndicators := getMaxIndicators(claims.UserType)
	if len(chart.Indicators) > maxIndicators {
		h.sendError(w, fmt.Sprintf("Maximum %d indicators allowed", maxIndicators), http.StatusBadRequest, nil)
		return
	}

	chart.UserID = userID
	chart.CreatedAt = time.Now().UTC()
	chart.UpdatedAt = time.Now().UTC()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := h.db.Collection("saved_charts")
	result, err := collection.InsertOne(ctx, chart)
	if err != nil {
		h.sendError(w, "Failed to save chart", http.StatusInternalServerError, nil)
		return
	}

	chart.ID = result.InsertedID.(primitive.ObjectID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(chart)
}

// Get saved charts
func (h *Handlers) GetSavedCharts(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	userID, _ := primitive.ObjectIDFromHex(claims.UserID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.db.Collection("saved_charts")
	cursor, err := collection.Find(ctx, bson.M{"user_id": userID},
		options.Find().SetSort(bson.D{{"created_at", -1}}))
	if err != nil {
		h.sendError(w, "Failed to fetch charts", http.StatusInternalServerError, nil)
		return
	}
	defer cursor.Close(ctx)

	var charts []models.SavedChart
	if err := cursor.All(ctx, &charts); err != nil {
		h.sendError(w, "Failed to decode charts", http.StatusInternalServerError, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(charts)
}

// Delete saved chart
func (h *Handlers) DeleteChart(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*Claims)
	userID, _ := primitive.ObjectIDFromHex(claims.UserID)

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
		"user_id": userID,
	})
	if err != nil {
		h.sendError(w, "Failed to delete chart", http.StatusInternalServerError, nil)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "Chart not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	Exp      int64  `json:"exp"`
}

func (h *Handlers) validateRegisterRequest(req *models.RegisterRequest) error {
	// Basic validation
	if len(req.Email) == 0 || !strings.Contains(req.Email, "@") {
		return fmt.Errorf("valid email required")
	}
	if len(req.Password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(req.FirstName) < 2 {
		return fmt.Errorf("first name must be at least 2 characters")
	}
	if len(req.LastName) < 2 {
		return fmt.Errorf("last name must be at least 2 characters")
	}
	return nil
}

func (h *Handlers) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password + "salt")) // In production, use proper salt
	return base64.StdEncoding.EncodeToString(hash[:])
}

func (h *Handlers) verifyPassword(password, hash string) bool {
	return h.hashPassword(password) == hash
}

func (h *Handlers) generateJWT(user *models.User) (string, error) {
	// Create claims
	claims := Claims{
		UserID:   user.ID.Hex(),
		Email:    user.Email,
		UserType: user.UserType,
		Exp:      time.Now().Add(24 * time.Hour).Unix(),
	}

	// Create header and payload
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerBytes, _ := json.Marshal(header)
	claimsBytes, _ := json.Marshal(claims)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	claimsB64 := base64.RawURLEncoding.EncodeToString(claimsBytes)

	// Create signature
	message := headerB64 + "." + claimsB64
	hasher := hmac.New(sha256.New, h.config.JWTSecret)
	hasher.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(hasher.Sum(nil))

	return message + "." + signature, nil
}

func (h *Handlers) sendError(w http.ResponseWriter, message string, status int, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := models.ErrorResponse{
		Error:   message,
		Details: details,
	}

	json.NewEncoder(w).Encode(response)
}

func getMaxIndicators(userType string) int {
	if userType == "paid" {
		return 10
	}
	return 3
}
