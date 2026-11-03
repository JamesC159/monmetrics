package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SoldSale struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ItemID        string             `bson:"item_id" json:"item_id"`
	CardID        primitive.ObjectID `bson:"card_id" json:"card_id"`
	Source        string             `bson:"source" json:"source"`
	Title         string             `bson:"title" json:"title"`
	Price         float64            `bson:"price" json:"price"`
	Shipping      float64            `bson:"shipping" json:"shipping"`
	Currency      string             `bson:"currency" json:"currency"`
	SoldAt        time.Time          `bson:"sold_at" json:"sold_at"`
	Grade         string             `bson:"grade" json:"grade"`
	BestOffer     bool               `bson:"best_offer" json:"best_offer"`
	URL           string             `bson:"url" json:"url"`
	Excluded      bool               `bson:"excluded" json:"excluded"`
	ExcludeReason string             `bson:"exclude_reason,omitempty" json:"exclude_reason,omitempty"`
	ScrapedAt     time.Time          `bson:"scraped_at" json:"scraped_at"`
}

type PriceSummary struct {
	CardID    primitive.ObjectID          `bson:"card_id" json:"card_id"`
	Source    string                      `bson:"source" json:"source"`
	Grade     string                      `bson:"grade" json:"grade"`
	UpdatedAt time.Time                   `bson:"updated_at" json:"updated_at"`
	Windows   map[string]PriceWindowStats `bson:"windows" json:"windows"`
}

type PriceWindowStats struct {
	Count       int       `bson:"count" json:"count"`
	Average     float64   `bson:"average" json:"average"`
	TrimmedMean float64   `bson:"trimmed_mean" json:"trimmed_mean"`
	Median      float64   `bson:"median" json:"median"`
	Min         float64   `bson:"min" json:"min"`
	Max         float64   `bson:"max" json:"max"`
	ChangePct   float64   `bson:"change_pct" json:"change_pct"`
	From        time.Time `bson:"from" json:"from"`
	To          time.Time `bson:"to" json:"to"`
}
