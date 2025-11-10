package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	FirstName    string             `bson:"first_name" json:"first_name"`
	LastName     string             `bson:"last_name" json:"last_name"`
	UserType     string             `bson:"user_type" json:"user_type"` // "free" or "paid"
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	IsActive     bool               `bson:"is_active" json:"is_active"`
	LastLoginAt  *time.Time         `bson:"last_login_at,omitempty" json:"last_login_at,omitempty"`
}

// Card represents a trading card or sealed product
type Card struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Set         string             `bson:"set" json:"set"`
	Game        string             `bson:"game" json:"game"`         // Pokemon, Yu-Gi-Oh, Magic, etc.
	Category    string             `bson:"category" json:"category"` // "card" or "sealed"
	Rarity      string             `bson:"rarity,omitempty" json:"rarity,omitempty"`
	Number      string             `bson:"number,omitempty" json:"number,omitempty"`
	ImageURL    string             `bson:"image_url" json:"image_url"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`

	// Current market data
	CurrentPrice float64   `bson:"current_price" json:"current_price"`
	AllTimeHigh  float64   `bson:"all_time_high" json:"all_time_high"`
	AllTimeLow   float64   `bson:"all_time_low" json:"all_time_low"`
	ATHDate      time.Time `bson:"ath_date" json:"ath_date"`
	ATLDate      time.Time `bson:"atl_date" json:"atl_date"`

	// Search and categorization
	SearchTerms []string `bson:"search_terms" json:"search_terms"`
	Tags        []string `bson:"tags,omitempty" json:"tags,omitempty"`

	// Popularity and ranking (based on 6-month metrics)
	PopularityRank int `bson:"popularity_rank,omitempty" json:"popularity_rank,omitempty"`
}

// PricePoint represents a single price data point
type PricePoint struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CardID    primitive.ObjectID `bson:"card_id" json:"card_id"`
	Price     float64            `bson:"price" json:"price"`
	Volume    int                `bson:"volume,omitempty" json:"volume,omitempty"`
	Source    string             `bson:"source" json:"source"` // "ebay", "tcgplayer"
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// SavedChart represents a user's saved chart configuration
type SavedChart struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	CardID      primitive.ObjectID `bson:"card_id" json:"card_id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Indicators  []ChartIndicator   `bson:"indicators" json:"indicators"`
	TimeRange   string             `bson:"time_range" json:"time_range"` // "1d", "7d", "30d", "90d", "1y", "5y"
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// ChartIndicator represents a technical indicator configuration
type ChartIndicator struct {
	Type       string                 `bson:"type" json:"type"` // "bollinger", "rsi", "sma", "ema", etc.
	Parameters map[string]interface{} `bson:"parameters" json:"parameters"`
	Color      string                 `bson:"color,omitempty" json:"color,omitempty"`
	Visible    bool                   `bson:"visible" json:"visible"`
}

// MarketData represents aggregated market data for a card
type MarketData struct {
	CardID           primitive.ObjectID `bson:"card_id" json:"card_id"`
	Date             time.Time          `bson:"date" json:"date"`
	OpenPrice        float64            `bson:"open_price" json:"open_price"`
	ClosePrice       float64            `bson:"close_price" json:"close_price"`
	HighPrice        float64            `bson:"high_price" json:"high_price"`
	LowPrice         float64            `bson:"low_price" json:"low_price"`
	Volume           int                `bson:"volume" json:"volume"`
	WeightedAvgPrice float64            `bson:"weighted_avg_price" json:"weighted_avg_price"`
}

// Listing represents a current marketplace listing
type Listing struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CardID    primitive.ObjectID `bson:"card_id" json:"card_id"`
	Title     string             `bson:"title" json:"title"`
	Price     float64            `bson:"price" json:"price"`
	Quantity  int                `bson:"quantity" json:"quantity"`
	Condition string             `bson:"condition" json:"condition"`
	Seller    string             `bson:"seller" json:"seller"`
	Source    string             `bson:"source" json:"source"` // "ebay", "tcgplayer"
	ImageURL  string             `bson:"image_url,omitempty" json:"image_url,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// SearchResult represents search results for cards
type SearchResult struct {
	Cards      []Card `json:"cards"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
}

// PriceHistory represents historical price data with indicators
type PriceHistory struct {
	Prices     []PricePoint                `json:"prices"`
	MarketData []MarketData                `json:"market_data"`
	Indicators map[string][]IndicatorPoint `json:"indicators,omitempty"`
}

// IndicatorPoint represents a calculated indicator value
type IndicatorPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Label     string    `json:"label,omitempty"`
}

// Dashboard represents user dashboard data
type Dashboard struct {
	SavedCharts    []SavedChart `json:"saved_charts"`
	RecentlyViewed []Card       `json:"recently_viewed"`
	UserStats      UserStats    `json:"user_stats"`
}

// UserStats represents user statistics
type UserStats struct {
	ChartsCreated  int `json:"charts_created"`
	IndicatorsUsed int `json:"indicators_used"`
	MaxIndicators  int `json:"max_indicators"`
}

// Request/Response Types

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

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// SearchParams represents search query parameters
type SearchParams struct {
	Query    string `json:"q,omitempty"`
	Game     string `json:"game,omitempty"`
	Category string `json:"category,omitempty"`
	Page     int    `json:"page,omitempty"`
	Limit    int    `json:"limit,omitempty"`
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

// FeaturedContent represents carousel content types
type FeaturedContent struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Type        string              `bson:"type" json:"type"` // "product", "market_mover", "news", "pickup", "sponsored"
	Title       string              `bson:"title" json:"title"`
	Description string              `bson:"description,omitempty" json:"description,omitempty"`
	ImageURL    string              `bson:"image_url" json:"image_url"`
	CardID      *primitive.ObjectID `bson:"card_id,omitempty" json:"card_id,omitempty"`
	Link        string              `bson:"link,omitempty" json:"link,omitempty"`
	Priority    int                 `bson:"priority" json:"priority"` // Higher = shown first
	Active      bool                `bson:"active" json:"active"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
	ExpiresAt   *time.Time          `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	// Market mover specific fields
	PriceChange      float64 `bson:"price_change,omitempty" json:"price_change,omitempty"`             // Percentage
	PriceChangeValue float64 `bson:"price_change_value,omitempty" json:"price_change_value,omitempty"` // Dollar amount
}

// GameCardGroup represents cards/sealed grouped by game
type GameCardGroup struct {
	Game       string `json:"game"`
	Category   string `json:"category"` // "card" or "sealed"
	Cards      []Card `json:"cards"`
	TotalCount int    `json:"total_count"`
}
