package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jamesc159/monmetrics/configs"
)

// Handlers holds the database and configuration for all handler methods
type Handlers struct {
	db     *mongo.Database
	config *configs.Config
}

// New creates a new Handlers instance
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
