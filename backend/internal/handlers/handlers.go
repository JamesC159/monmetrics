package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jamesc159/monmetrics/configs"
	"github.com/jamesc159/monmetrics/internal/marketplace"
	"github.com/jamesc159/monmetrics/internal/middleware"
	"github.com/jamesc159/monmetrics/internal/secure"
)

// Handlers holds the database and configuration for all handler methods
type Handlers struct {
	db     *mongo.Database
	config *configs.Config
	market *marketplace.Registry
	cipher *secure.Cipher
}

// New creates a new Handlers instance
func New(db *mongo.Database, config *configs.Config) (*Handlers, error) {
	cipher, err := secure.NewCipher(config.TokenEncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("token encryption: %w", err)
	}
	market, err := marketplace.NewRegistry(config.MarketplaceMode, db, config.APIBaseURL)
	if err != nil {
		return nil, err
	}
	return &Handlers{
		db:     db,
		config: config,
		market: market,
		cipher: cipher,
	}, nil
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

// userIDFromRequest extracts the authenticated user's ID and claims set by AuthRequired
func userIDFromRequest(r *http.Request) (primitive.ObjectID, *middleware.Claims, error) {
	claims, ok := r.Context().Value(middleware.ClaimsKey).(*middleware.Claims)
	if !ok || claims == nil {
		return primitive.NilObjectID, nil, errors.New("unauthorized")
	}
	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		return primitive.NilObjectID, nil, errors.New("invalid user id")
	}
	return userID, claims, nil
}

// sendJSON writes a JSON response with the given status code
func sendJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		fmt.Printf("❌ Error encoding response: %v\n", err)
	}
}

// decodeJSON decodes a size-limited JSON request body, rejecting unknown fields
func decodeJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

// sendError sends a standardized error response
func (h *Handlers) sendError(w http.ResponseWriter, message string, statusCode int, data map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   message,
		"success": false,
	}

	for k, v := range data {
		response[k] = v
	}

	json.NewEncoder(w).Encode(response)
}
