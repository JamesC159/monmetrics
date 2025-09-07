package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jamesc159/monmetrics/configs"
	"github.com/jamesc159/monmetrics/internal/database"
	"github.com/jamesc159/monmetrics/internal/handlers"
	"github.com/jamesc159/monmetrics/internal/middleware"
)

func main() {
	// Load configuration
	config := configs.Load()

	// Initialize database
	db, err := database.Connect(config.MongoURI, config.DBName)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Disconnect()

	// Initialize handlers
	h := handlers.New(db, config)

	// Setup router with middleware
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", h.Health)

	// API routes
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /cards/search", h.SearchCards)
	apiMux.HandleFunc("GET /cards/{id}", h.GetCard)
	apiMux.HandleFunc("GET /cards/{id}/prices", h.GetCardPrices)

	// Auth routes
	apiMux.HandleFunc("POST /auth/register", h.Register)
	apiMux.HandleFunc("POST /auth/login", h.Login)
	apiMux.HandleFunc("POST /auth/logout", h.Logout)

	// Protected routes
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /user/dashboard", h.GetDashboard)
	protectedMux.HandleFunc("POST /user/charts", h.SaveChart)
	protectedMux.HandleFunc("GET /user/charts", h.GetSavedCharts)
	protectedMux.HandleFunc("DELETE /user/charts/{id}", h.DeleteChart)

	// Apply middleware stack
	api := middleware.Chain(
		middleware.CORS(config.CORSOrigins),
		middleware.SecurityHeaders(),
		middleware.RateLimit(config.RateLimitRequests, config.RateLimitWindow),
		middleware.RequestLogger(),
	)(apiMux)

	protectedAPI := middleware.Chain(
		middleware.CORS(config.CORSOrigins),
		middleware.SecurityHeaders(),
		middleware.RateLimit(config.RateLimitRequests, config.RateLimitWindow),
		middleware.RequestLogger(),
		middleware.AuthRequired(config.JWTSecret),
	)(protectedMux)

	// Mount routes
	mux.Handle("/api/", http.StripPrefix("/api", api))
	mux.Handle("/api/protected/", http.StripPrefix("/api/protected", protectedAPI))

	// Static file serving for production
	if config.Environment == "production" {
		mux.Handle("/", http.FileServer(http.Dir("./static/")))
	}

	// Create server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
