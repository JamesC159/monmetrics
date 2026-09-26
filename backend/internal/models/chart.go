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
	TimeRange   string             `bson:"time_range" json:"time_range"`           // "1d", "7d", "30d", "90d", "1y", "5y"
	Source      string             `bson:"source" json:"source"`                   // "all", "ebay", "tcgplayer"
	ChartType   string             `bson:"chart_type,omitempty" json:"chart_type"` // "line", "candle"
	ShowVolume  bool               `bson:"show_volume" json:"show_volume"`
	Drawings    []ChartDrawing     `bson:"drawings" json:"drawings"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`

	// Response-only card snapshot
	CardName     string `bson:"-" json:"card_name,omitempty"`
	CardImageURL string `bson:"-" json:"card_image_url,omitempty"`
	CardGame     string `bson:"-" json:"card_game,omitempty"`
}

// SavedChartRequest is the client payload for creating or updating a saved chart
type SavedChartRequest struct {
	CardID      string           `json:"card_id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Indicators  []ChartIndicator `json:"indicators"`
	TimeRange   string           `json:"time_range"`
	Source      string           `json:"source"`
	ChartType   string           `json:"chart_type"`
	ShowVolume  bool             `json:"show_volume"`
	Drawings    []ChartDrawing   `json:"drawings"`
}

// DrawingPoint anchors a drawing in data space; Time is unix seconds (UTC day start).
type DrawingPoint struct {
	Time  int64   `bson:"time" json:"time"`
	Price float64 `bson:"price" json:"price"`
}

// ChartDrawing is a user annotation on a chart
type ChartDrawing struct {
	ID        string         `bson:"id" json:"id"`
	Type      string         `bson:"type" json:"type"` // "trendline", "hline", "rect", "text", "freehand", "fib"
	Points    []DrawingPoint `bson:"points" json:"points"`
	Color     string         `bson:"color" json:"color"`
	Text      string         `bson:"text,omitempty" json:"text,omitempty"`
	LineWidth int            `bson:"line_width" json:"line_width"`
}

// ChartIndicator represents a technical indicator configuration
type ChartIndicator struct {
	Type       string                 `bson:"type" json:"type"` // "bollinger", "rsi", "sma", "ema", etc.
	Parameters map[string]interface{} `bson:"parameters" json:"parameters"`
	Color      string                 `bson:"color,omitempty" json:"color,omitempty"`
	Visible    bool                   `bson:"visible" json:"visible"`
}
