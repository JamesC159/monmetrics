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

// UserStats represents user statistics
type UserStats struct {
	ChartsCreated  int `json:"charts_created"`
	IndicatorsUsed int `json:"indicators_used"`
	MaxIndicators  int `json:"max_indicators"`
}

// Dashboard represents user dashboard data
type Dashboard struct {
	SavedCharts    []SavedChart `json:"saved_charts"`
	RecentlyViewed []Card       `json:"recently_viewed"`
	UserStats      UserStats    `json:"user_stats"`
}
