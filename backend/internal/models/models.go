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
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name         string             `bson:"name" json:"name"`
	Set          string             `bson:"set" json:"set"`
	Game         string             `bson:"game" json:"game"` // Pokemon, Yu-Gi-Oh, Magic, etc.
	Category     string             `bson:"category" json:"category"` // "card" or "sealed"
	Rarity       string             `bson:"rarity,omitempty" json:"rarity,omitempty"`
	Number       string             `bson:"number,omitempty" json:"number,omitempty"`
	ImageURL     string             `bson:"image_url" json:"image_url"`
	Description  string             `bson:"description,omitempty" json:"description,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`

	// Current market data
	CurrentPrice float64   `bson:"current_price" json:"current_price"`
	AllTimeHigh  float64   `bson:"all_time_high" json:"all_time_high"`
	AllTimeLow   float64   `bson:"all_time_low" json:"all_time_low"`
	ATHDate      time.Time `bson:"ath_date" json:"ath_date"`
	ATLDate      time.Time `bson:"atl_date" json:"atl_date"`

	// Search and categorization
	SearchTerms  []string `bson:"search_terms" json:"search_terms"`
	Tags         []string `bson:"tags,omitempty" json:"tags,omitempty"`
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
	CardID          primitive.ObjectID `bson:"card_id" json:"card_id"`
	Date            time.Time          `bson:"date" json:"date"`
	OpenPrice       float64            `bson:"open_price" json:"open_price"`
	ClosePrice      float64            `bson:"close_price" json:"close_price"`
	HighPrice       float64            `bson:"high_price" json:"high_price"`
	LowPrice        float64            `bson:"low_price" json:"low_price"`
	Volume          int                `bson:"volume" json:"volume"`
	WeightedAvgPrice float64           `bson:"weighted_avg_price" json:"weighted_avg_price"`
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
	Prices     []PricePoint              `json:"prices"`
	MarketData []MarketData              `json:"market_data"`
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
	ChartsCreated    int `json:"charts_created"`
	IndicatorsUsed   int `json:"indicators_used"`
	MaxIndicators    int `json:"max_indicators"`
	DaysActive       int `json:"days_active"`
	LastActiveDate   time.Time `json:"last_active_date"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"firstName" validate:"required,min=2"`  // ← Changed to camelCase
	LastName  string `json:"lastName" validate:"required,min=2"`   // ← Changed to camelCase
}

// AuthResponse represents authentication response
type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Details map[string]interface{} `json:"details,omitempty"`
}