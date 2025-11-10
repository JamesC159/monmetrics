package main

import (
	"context"
	"fmt"
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

	// Health check endpoint (no middleware needed)
	mux.HandleFunc("GET /health", h.Health)

	// Public API routes
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("GET /cards/search", h.SearchCards)
	apiMux.HandleFunc("GET /cards/{id}", h.GetCard)
	apiMux.HandleFunc("GET /cards/{id}/prices", h.GetCardPrices)

	// Featured content and organized search
	apiMux.HandleFunc("GET /featured-content", h.GetFeaturedContent)
	apiMux.HandleFunc("GET /cards/by-game", h.GetCardsByGame)
	apiMux.HandleFunc("GET /sealed/by-game", h.GetSealedByGame)

	// Auth routes (public)
	apiMux.HandleFunc("POST /auth/register", h.Register)
	apiMux.HandleFunc("POST /auth/login", h.Login)
	apiMux.HandleFunc("POST /auth/logout", h.Logout)

	// Protected routes (require authentication)
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("GET /user/dashboard", h.GetDashboard)
	protectedMux.HandleFunc("POST /user/charts", h.SaveChart)
	protectedMux.HandleFunc("GET /user/charts", h.GetSavedCharts)
	protectedMux.HandleFunc("DELETE /user/charts/{id}", h.DeleteChart)

	// Apply middleware stack to public API routes
	api := middleware.Chain(
		middleware.CORS(config.CORSOrigins),
		middleware.SecurityHeaders(),
		middleware.RateLimit(config.RateLimitRequests, config.RateLimitWindow),
		middleware.RequestLogger(),
	)(apiMux)

	// Apply middleware stack to protected routes (includes auth)
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

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 MonMetrics server starting on port %s", config.Port)
		log.Printf("📊 Environment: %s", config.Environment)
		log.Printf("🗄️  Database: %s", config.DBName)
		log.Printf("🌐 CORS Origins: %v", config.CORSOrigins)
		log.Printf("⚡ Rate Limit: %d requests per %v", config.RateLimitRequests, config.RateLimitWindow)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Print available endpoints
	fmt.Println("\n📡 Available Endpoints:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("🏥 Health Check:     GET  http://localhost:%s/health\n", config.Port)
	fmt.Println("\n🔓 Public API:")
	fmt.Printf("🔍 Search Cards:     GET  http://localhost:%s/api/cards/search\n", config.Port)
	fmt.Printf("📋 Get Card:         GET  http://localhost:%s/api/cards/{id}\n", config.Port)
	fmt.Printf("📈 Card Prices:      GET  http://localhost:%s/api/cards/{id}/prices\n", config.Port)
	fmt.Printf("🎪 Featured Content: GET  http://localhost:%s/api/featured-content\n", config.Port)
	fmt.Printf("🎮 Cards by Game:    GET  http://localhost:%s/api/cards/by-game\n", config.Port)
	fmt.Printf("📦 Sealed by Game:   GET  http://localhost:%s/api/sealed/by-game\n", config.Port)
	fmt.Printf("👤 Register:         POST http://localhost:%s/api/auth/register\n", config.Port)
	fmt.Printf("🔑 Login:            POST http://localhost:%s/api/auth/login\n", config.Port)
	fmt.Printf("🚪 Logout:           POST http://localhost:%s/api/auth/logout\n", config.Port)
	fmt.Println("\n🔒 Protected API (requires authentication):")
	fmt.Printf("📊 Dashboard:        GET  http://localhost:%s/api/protected/user/dashboard\n", config.Port)
	fmt.Printf("💾 Save Chart:       POST http://localhost:%s/api/protected/user/charts\n", config.Port)
	fmt.Printf("📋 Get Charts:       GET  http://localhost:%s/api/protected/user/charts\n", config.Port)
	fmt.Printf("🗑️  Delete Chart:     DEL  http://localhost:%s/api/protected/user/charts/{id}\n", config.Port)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("🎯 Frontend URL:     http://localhost:3000\n")
	fmt.Println("\n✅ Server is ready to accept connections!")

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	} else {
		log.Println("✅ Server gracefully stopped")
	}
}
