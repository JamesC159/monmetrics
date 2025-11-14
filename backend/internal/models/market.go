package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
