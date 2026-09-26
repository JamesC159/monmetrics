package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	AlertAbove     = "above"
	AlertBelow     = "below"
	AlertPctChange = "pct_change"
	AlertEMACross  = "ema_cross"

	AlertModeOnce      = "once"
	AlertModeRecurring = "recurring"
)

// PriceAlert is a user-defined condition evaluated against a card's daily average price
type PriceAlert struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID          primitive.ObjectID `bson:"user_id" json:"user_id"`
	CardID          primitive.ObjectID `bson:"card_id" json:"card_id"`
	Source          string             `bson:"source" json:"source"`
	Condition       string             `bson:"condition" json:"condition"`
	TargetPrice     float64            `bson:"target_price,omitempty" json:"target_price,omitempty"`
	Pct             float64            `bson:"pct,omitempty" json:"pct,omitempty"`
	Days            int                `bson:"days,omitempty" json:"days,omitempty"`
	EMAPeriod       int                `bson:"ema_period,omitempty" json:"ema_period,omitempty"`
	Direction       string             `bson:"direction,omitempty" json:"direction,omitempty"` // "up", "down", "either"
	Mode            string             `bson:"mode" json:"mode"`
	CooldownHours   int                `bson:"cooldown_hours,omitempty" json:"cooldown_hours,omitempty"`
	NotifyEmail     bool               `bson:"notify_email" json:"notify_email"`
	Note            string             `bson:"note,omitempty" json:"note,omitempty"`
	Active          bool               `bson:"active" json:"active"`
	LastPrice       *float64           `bson:"last_price,omitempty" json:"last_price,omitempty"`
	LastEMA         *float64           `bson:"last_ema,omitempty" json:"last_ema,omitempty"`
	LastTriggeredAt *time.Time         `bson:"last_triggered_at,omitempty" json:"last_triggered_at,omitempty"`
	TriggerCount    int                `bson:"trigger_count" json:"trigger_count"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`

	// Response-only card snapshot
	CardName     string `bson:"-" json:"card_name,omitempty"`
	CardImageURL string `bson:"-" json:"card_image_url,omitempty"`
}

// PriceAlertRequest is the client payload for creating or updating an alert
type PriceAlertRequest struct {
	CardID        string  `json:"card_id"`
	Source        string  `json:"source"`
	Condition     string  `json:"condition"`
	TargetPrice   float64 `json:"target_price"`
	Pct           float64 `json:"pct"`
	Days          int     `json:"days"`
	EMAPeriod     int     `json:"ema_period"`
	Direction     string  `json:"direction"`
	Mode          string  `json:"mode"`
	CooldownHours int     `json:"cooldown_hours"`
	NotifyEmail   bool    `json:"notify_email"`
	Note          string  `json:"note"`
	Active        *bool   `json:"active,omitempty"`
}

// Notification is an in-app message, currently produced by triggered alerts
type Notification struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID  `bson:"user_id" json:"user_id"`
	AlertID   *primitive.ObjectID `bson:"alert_id,omitempty" json:"alert_id,omitempty"`
	CardID    *primitive.ObjectID `bson:"card_id,omitempty" json:"card_id,omitempty"`
	Title     string              `bson:"title" json:"title"`
	Message   string              `bson:"message" json:"message"`
	Read      bool                `bson:"read" json:"read"`
	CreatedAt time.Time           `bson:"created_at" json:"created_at"`
}
