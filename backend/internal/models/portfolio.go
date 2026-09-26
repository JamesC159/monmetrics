package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Portfolio item types
const (
	ItemTypeRawCard    = "raw_card"
	ItemTypeGradedCard = "graded_card"
	ItemTypeSealed     = "sealed"
)

// Valuation sources
const (
	ValueSourceMarket         = "market"
	ValueSourceGradedEstimate = "graded_estimate"
	ValueSourceManual         = "manual"
	ValueSourceNone           = "none"
)

// Grading holds professional grading details for a graded card
type Grading struct {
	Company    string  `bson:"company" json:"company"` // PSA, BGS, CGC, SGC
	Grade      float64 `bson:"grade" json:"grade"`
	CertNumber string  `bson:"cert_number,omitempty" json:"cert_number,omitempty"`
}

// PortfolioItem is something a user owns: a raw card, graded card, or sealed product.
// CardID is nil for custom items that are not in the catalog.
type PortfolioItem struct {
	ID             primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID  `bson:"user_id" json:"user_id"`
	CardID         *primitive.ObjectID `bson:"card_id,omitempty" json:"card_id,omitempty"`
	ItemType       string              `bson:"item_type" json:"item_type"`
	CustomName     string              `bson:"custom_name,omitempty" json:"custom_name,omitempty"`
	CustomGame     string              `bson:"custom_game,omitempty" json:"custom_game,omitempty"`
	CustomSet      string              `bson:"custom_set,omitempty" json:"custom_set,omitempty"`
	CustomImageURL string              `bson:"custom_image_url,omitempty" json:"custom_image_url,omitempty"`
	Quantity       int                 `bson:"quantity" json:"quantity"`
	Condition      string              `bson:"condition,omitempty" json:"condition,omitempty"` // NM, LP, MP, HP, DMG
	Finish         string              `bson:"finish,omitempty" json:"finish,omitempty"`
	Language       string              `bson:"language,omitempty" json:"language,omitempty"`
	Grading        *Grading            `bson:"grading,omitempty" json:"grading,omitempty"`
	PurchasePrice  float64             `bson:"purchase_price" json:"purchase_price"` // per unit
	PurchaseDate   *time.Time          `bson:"purchase_date,omitempty" json:"purchase_date,omitempty"`
	PurchaseSource string              `bson:"purchase_source,omitempty" json:"purchase_source,omitempty"`
	ManualValue    *float64            `bson:"manual_value,omitempty" json:"manual_value,omitempty"` // per unit
	Notes          string              `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt      time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time           `bson:"updated_at" json:"updated_at"`
}

// PortfolioItemRequest is the client payload for creating/updating an item
type PortfolioItemRequest struct {
	CardID         string   `json:"card_id"`
	ItemType       string   `json:"item_type"`
	CustomName     string   `json:"custom_name"`
	CustomGame     string   `json:"custom_game"`
	CustomSet      string   `json:"custom_set"`
	CustomImageURL string   `json:"custom_image_url"`
	Quantity       int      `json:"quantity"`
	Condition      string   `json:"condition"`
	Finish         string   `json:"finish"`
	Language       string   `json:"language"`
	Grading        *Grading `json:"grading"`
	PurchasePrice  float64  `json:"purchase_price"`
	PurchaseDate   string   `json:"purchase_date"` // YYYY-MM-DD
	PurchaseSource string   `json:"purchase_source"`
	ManualValue    *float64 `json:"manual_value"`
	Notes          string   `json:"notes"`
}

// PortfolioItemView is an item with its catalog snapshot and computed valuation
type PortfolioItemView struct {
	PortfolioItem
	Card            *Card   `json:"card,omitempty"`
	CardUnavailable bool    `json:"card_unavailable,omitempty"`
	DisplayName     string  `json:"display_name"`
	DisplayGame     string  `json:"display_game"`
	DisplaySet      string  `json:"display_set"`
	DisplayImageURL string  `json:"display_image_url"`
	UnitValue       float64 `json:"unit_value"`
	TotalValue      float64 `json:"total_value"`
	CostBasis       float64 `json:"cost_basis"`
	GainLoss        float64 `json:"gain_loss"`
	GainLossPct     float64 `json:"gain_loss_pct"`
	ValueSource     string  `json:"value_source"`
	Multiplier      float64 `json:"multiplier"`
}

// AllocationSlice is one segment of the portfolio allocation breakdown
type AllocationSlice struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
	Count int     `json:"count"`
}

// PortfolioMover is a compact gainer/loser entry
type PortfolioMover struct {
	ID          primitive.ObjectID `json:"id"`
	DisplayName string             `json:"display_name"`
	GainLoss    float64            `json:"gain_loss"`
	GainLossPct float64            `json:"gain_loss_pct"`
}

// PortfolioSummary aggregates portfolio-wide totals
type PortfolioSummary struct {
	TotalValue    float64           `json:"total_value"`
	CostBasis     float64           `json:"cost_basis"`
	GainLoss      float64           `json:"gain_loss"`
	GainLossPct   float64           `json:"gain_loss_pct"`
	ItemCount     int               `json:"item_count"`
	UnitCount     int               `json:"unit_count"`
	UnvaluedCount int               `json:"unvalued_count"`
	ByGame        []AllocationSlice `json:"by_game"`
	ByItemType    []AllocationSlice `json:"by_item_type"`
	TopGainers    []PortfolioMover  `json:"top_gainers"`
	TopLosers     []PortfolioMover  `json:"top_losers"`
}

// PortfolioResponse is returned by the portfolio list endpoint
type PortfolioResponse struct {
	Items   []PortfolioItemView `json:"items"`
	Summary PortfolioSummary    `json:"summary"`
}
