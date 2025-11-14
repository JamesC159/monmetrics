package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
