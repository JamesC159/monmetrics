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
	SearchTerms  []string   `bson:"search_terms" json:"search_terms"`
	Tags         []string   `bson:"tags,omitempty" json:"tags,omitempty"`
	EbayQuery    string     `bson:"ebay_query,omitempty" json:"ebay_query,omitempty"`
	EbayExclude  []string   `bson:"ebay_exclude,omitempty" json:"ebay_exclude,omitempty"`
	EbaySyncedAt *time.Time `bson:"ebay_synced_at,omitempty" json:"ebay_synced_at,omitempty"`

	// Catalog provenance (populated by the automated catalog ingester)
	CatalogSource     string     `bson:"catalog_source,omitempty" json:"catalog_source,omitempty"`
	ExternalID        string     `bson:"external_id,omitempty" json:"external_id,omitempty"`
	ExternalGroupID   string     `bson:"external_group_id,omitempty" json:"external_group_id,omitempty"`
	ExternalCategory  string     `bson:"external_category_id,omitempty" json:"external_category_id,omitempty"`
	ProductType       string     `bson:"product_type,omitempty" json:"product_type,omitempty"` // "single", "booster_box", "etb", ...
	Language          string     `bson:"language,omitempty" json:"language,omitempty"`
	ProductURL        string     `bson:"product_url,omitempty" json:"product_url,omitempty"`
	MarketPrice       float64    `bson:"market_price,omitempty" json:"market_price,omitempty"`
	MarketPriceSynced *time.Time `bson:"market_price_synced_at,omitempty" json:"market_price_synced_at,omitempty"`
	Printings         []Printing `bson:"printings,omitempty" json:"printings,omitempty"` // best first; the first priced one is the headline
	CatalogSyncedAt   *time.Time `bson:"catalog_synced_at,omitempty" json:"catalog_synced_at,omitempty"`
	Active            bool       `bson:"active" json:"active"`

	// Scrape scheduling (tiered priority queue)
	ScrapeTier     int        `bson:"scrape_tier,omitempty" json:"scrape_tier,omitempty"`
	NextScrapeAt   *time.Time `bson:"next_scrape_at,omitempty" json:"next_scrape_at,omitempty"`
	LastScrapeAt   *time.Time `bson:"last_scrape_at,omitempty" json:"last_scrape_at,omitempty"`
	ScrapeFailures int        `bson:"scrape_failures,omitempty" json:"scrape_failures,omitempty"`
	LastSaleCount  int        `bson:"last_sale_count,omitempty" json:"last_sale_count,omitempty"`

	// Popularity and ranking (based on 6-month metrics)
	PopularityRank int `bson:"popularity_rank,omitempty" json:"popularity_rank,omitempty"`
}

// Printing is one TCGplayer printing of a product (Normal, Foil, Reverse Holofoil, ...).
// Series is its price-series key, the grade its price points are stored under.
// MarketPrice is absent when TCGplayer has no figure for that printing.
type Printing struct {
	Series      string  `bson:"series" json:"series"`
	MarketPrice float64 `bson:"market_price,omitempty" json:"market_price,omitempty"`
}

// Scrape tiers control how often the sold-listing scraper revisits a card.
const (
	ScrapeTierNone     = 0 // never scheduled (inactive / repeatedly empty)
	ScrapeTierHot      = 1
	ScrapeTierWarm     = 2
	ScrapeTierColdTail = 3
)

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

// GameSummary is a catalog browse entry for one game.
type GameSummary struct {
	Game        string `bson:"game" json:"game"`
	CardCount   int    `bson:"card_count" json:"card_count"`
	SealedCount int    `bson:"sealed_count" json:"sealed_count"`
	PricedCount int    `bson:"priced_count" json:"priced_count"`
	ImageURL    string `bson:"image_url" json:"image_url"`
}

// SetSummary is a catalog browse entry for one set within a game.
// PricedCount is how many products already have scraped sales data.
type SetSummary struct {
	Set             string `bson:"set" json:"set"`
	Game            string `bson:"game" json:"game"`
	ExternalGroupID string `bson:"external_group_id,omitempty" json:"external_group_id,omitempty"`
	ProductCount    int    `bson:"product_count" json:"product_count"`
	PricedCount     int    `bson:"priced_count" json:"priced_count"`
	ImageURL        string `bson:"image_url" json:"image_url"`
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
