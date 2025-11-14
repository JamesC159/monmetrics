package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

// SearchResult represents search results for cards
type SearchResult struct {
	Cards      []Card `json:"cards"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PerPage    int    `json:"per_page"`
	TotalPages int    `json:"total_pages"`
}

// GameCardGroup represents cards/sealed grouped by game
type GameCardGroup struct {
	Game       string `json:"game"`
	Category   string `json:"category"` // "card" or "sealed"
	Cards      []Card `json:"cards"`
	TotalCount int    `json:"total_count"`
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
