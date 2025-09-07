package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
)

// Middleware type for chaining
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in order
func Chain(middlewares ...Middleware) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			h = middlewares[i](h)
		}
		return h
	}
}

// CORS middleware with secure configuration
func CORS(allowedOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if origin == allowedOrigin {
					allowed = true
					break
				}
			}

			// For development, also allow localhost variations
			if !allowed && (origin == "http://localhost:3000" ||
				origin == "http://127.0.0.1:3000" ||
				origin == "http://localhost:5173" ||
				origin == "") {
				allowed = true
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else {
				// For development, be more permissive
				w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeaders implements OWASP security headers
func SecurityHeaders() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Prevent XSS attacks
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// HSTS for HTTPS
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

			// Content Security Policy
			csp := "default-src 'self'; " +
				"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
				"style-src 'self' 'unsafe-inline'; " +
				"img-src 'self' data: https:; " +
				"font-src 'self'; " +
				"connect-src 'self'; " +
				"media-src 'self'; " +
				"object-src 'none'; " +
				"child-src 'self'; " +
				"frame-ancestors 'none'; " +
				"form-action 'self'; " +
				"base-uri 'self';"
			w.Header().Set("Content-Security-Policy", csp)

			// Referrer Policy
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Feature Policy
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			next.ServeHTTP(w, r)
		})
	}
}

// Rate limiting implementation
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*clientLimiter
	requests int
	window   time.Duration
}

type clientLimiter struct {
	tokens     int
	lastRefill time.Time
}

func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients:  make(map[string]*clientLimiter),
		requests: requests,
		window:   window,
	}
}

func (rl *RateLimiter) Allow(clientIP string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientIP]

	if !exists {
		rl.clients[clientIP] = &clientLimiter{
			tokens:     rl.requests - 1,
			lastRefill: now,
		}
		return true
	}

	// Refill tokens based on time passed
	timePassed := now.Sub(client.lastRefill)
	tokensToAdd := int(timePassed / rl.window * time.Duration(rl.requests))

	if tokensToAdd > 0 {
		client.tokens += tokensToAdd
		if client.tokens > rl.requests {
			client.tokens = rl.requests
		}
		client.lastRefill = now
	}

	if client.tokens > 0 {
		client.tokens--
		return true
	}

	return false
}

func RateLimit(requests int, window time.Duration) Middleware {
	limiter := NewRateLimiter(requests, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !limiter.Allow(clientIP) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "Rate limit exceeded",
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// JWT Claims structure
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`
	Exp      int64  `json:"exp"`
}

// AuthRequired middleware for protected routes
func AuthRequired(jwtSecret []byte) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				http.Error(w, "Bearer token required", http.StatusUnauthorized)
				return
			}

			claims, err := validateJWT(tokenString, jwtSecret)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), "claims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Simple JWT validation (using HMAC-SHA256)
func validateJWT(tokenString string, secret []byte) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode and validate header (but don't store it since we don't use it)
	_, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid header encoding: %v", err)
	}

	// Decode payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding: %v", err)
	}

	// Decode signature
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid signature encoding: %v", err)
	}

	// Verify signature
	hasher := hmac.New(sha256.New, secret)
	hasher.Write([]byte(parts[0] + "." + parts[1]))
	expectedSignature := hasher.Sum(nil)

	if !hmac.Equal(signature, expectedSignature) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Parse claims
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("invalid claims: %v", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.Exp {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// Request logging middleware
func RequestLogger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap ResponseWriter to capture status code
			wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapper, r)

			log.Printf(
				"%s %s %d %v %s",
				r.Method,
				r.URL.Path,
				wrapper.statusCode,
				time.Since(start),
				getClientIP(r),
			)
		})
	}
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Get client IP address safely
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}