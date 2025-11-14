package middleware

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
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

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		clients:  make(map[string]*clientLimiter),
		requests: requests,
		window:   window,
	}
}

// Allow checks if a client is allowed to make a request
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

// RateLimit creates a rate limiting middleware
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
