package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Favorite marks a catalog card or sealed product on a user's watchlist
type Favorite struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	CardID    primitive.ObjectID `bson:"card_id" json:"card_id"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

// FavoriteView is a favorite enriched with card data and recent price change
type FavoriteView struct {
	ID             primitive.ObjectID `json:"id"`
	CardID         primitive.ObjectID `json:"card_id"`
	CreatedAt      time.Time          `json:"created_at"`
	Card           *Card              `json:"card,omitempty"`
	Change7d       float64            `json:"change_7d"`
	Change7dValue  float64            `json:"change_7d_value"`
	Change30d      float64            `json:"change_30d"`
	Change30dValue float64            `json:"change_30d_value"`
}

// FavoriteRequest is the payload for adding a favorite
type FavoriteRequest struct {
	CardID string `json:"card_id"`
}
