package models

import "time"

// ═══════════════════════════════════════════════════════════════════════════════
// REQUEST MODELS - Incoming API Requests
// ═══════════════════════════════════════════════════════════════════════════════

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// SearchParams represents search query parameters
type SearchParams struct {
	Query    string `json:"q,omitempty"`
	Game     string `json:"game,omitempty"`
	Category string `json:"category,omitempty"`
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// ═══════════════════════════════════════════════════════════════════════════════
// RESPONSE MODELS - Outgoing API Responses
// ═══════════════════════════════════════════════════════════════════════════════

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Success bool                   `json:"success"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// HealthResponse represents a health check response
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}
