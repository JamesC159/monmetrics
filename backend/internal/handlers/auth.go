package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jamesc159/monmetrics/internal/models"
)

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

// generateJWT creates a JWT token for the user
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
