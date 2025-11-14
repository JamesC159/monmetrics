package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
