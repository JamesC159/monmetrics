package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/models"
)

// GetFeaturedContent retrieves active featured content for the carousel
func (h *Handlers) GetFeaturedContent(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := h.db.Collection("featured_content")

	// Get active featured content, sorted by priority
	filter := bson.M{
		"active": true,
		"$or": []bson.M{
			{"expires_at": bson.M{"$exists": false}},
			{"expires_at": bson.M{"$gt": time.Now()}},
		},
	}

	// Use bson.D for ordered sort (priority first, then created_at)
	findOptions := options.Find().SetSort(bson.D{{Key: "priority", Value: -1}, {Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		fmt.Printf("Error retrieving featured content: %v\n", err)
		// Return empty array instead of error if collection doesn't exist
		featuredContent := make([]models.FeaturedContent, 0)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(featuredContent)
		return
	}
	defer cursor.Close(ctx)

	featuredContent := make([]models.FeaturedContent, 0)
	if err = cursor.All(ctx, &featuredContent); err != nil {
		fmt.Printf("Error decoding featured content: %v\n", err)
		// Return empty array instead of error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(featuredContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(featuredContent)
}

// GetCardsByGame retrieves cards organized by game with popularity sorting
func (h *Handlers) GetCardsByGame(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := h.db.Collection("cards")

	// Get all unique games
	games, err := collection.Distinct(ctx, "game", bson.M{"category": "card"})
	if err != nil {
		fmt.Printf("Error retrieving games: %v\n", err)
		// Return empty array instead of error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.GameCardGroup{})
		return
	}

	fmt.Printf("Found %d unique games for cards\n", len(games))

	// For each game, get top cards by popularity (limit to 12 for 2 rows of 6)
	gameGroups := make([]models.GameCardGroup, 0)

	for _, game := range games {
		gameStr, ok := game.(string)
		if !ok {
			continue
		}

		// Get total count
		totalCount, err := collection.CountDocuments(ctx, bson.M{
			"game":     gameStr,
			"category": "card",
		})
		if err != nil {
			fmt.Printf("Error counting cards for game %s: %v\n", gameStr, err)
			continue
		}

		// Get top 12 cards by popularity
		findOptions := options.Find().
			SetSort(bson.D{{Key: "popularity_rank", Value: 1}, {Key: "current_price", Value: -1}}).
			SetLimit(12)

		cursor, err := collection.Find(ctx, bson.M{
			"game":     gameStr,
			"category": "card",
		}, findOptions)

		if err != nil {
			fmt.Printf("Error retrieving cards for game %s: %v\n", gameStr, err)
			continue
		}

		var cards []models.Card
		if err = cursor.All(ctx, &cards); err != nil {
			cursor.Close(ctx)
			fmt.Printf("Error decoding cards for game %s: %v\n", gameStr, err)
			continue
		}
		cursor.Close(ctx)

		gameGroups = append(gameGroups, models.GameCardGroup{
			Game:       gameStr,
			Category:   "card",
			Cards:      cards,
			TotalCount: int(totalCount),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameGroups)
}

// GetSealedByGame retrieves sealed products organized by game with popularity sorting
func (h *Handlers) GetSealedByGame(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := h.db.Collection("cards")

	// Get all unique games that have sealed products
	games, err := collection.Distinct(ctx, "game", bson.M{"category": "sealed"})
	if err != nil {
		fmt.Printf("Error retrieving games: %v\n", err)
		// Return empty array instead of error
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.GameCardGroup{})
		return
	}

	fmt.Printf("Found %d unique games for sealed\n", len(games))

	// For each game, get top sealed products by popularity (limit to 12 for 2 rows of 6)
	gameGroups := make([]models.GameCardGroup, 0)

	for _, game := range games {
		gameStr, ok := game.(string)
		if !ok {
			continue
		}

		// Get total count
		totalCount, err := collection.CountDocuments(ctx, bson.M{
			"game":     gameStr,
			"category": "sealed",
		})
		if err != nil {
			fmt.Printf("Error counting sealed for game %s: %v\n", gameStr, err)
			continue
		}

		// Get top 12 sealed products by popularity
		findOptions := options.Find().
			SetSort(bson.D{{Key: "popularity_rank", Value: 1}, {Key: "current_price", Value: -1}}).
			SetLimit(12)

		cursor, err := collection.Find(ctx, bson.M{
			"game":     gameStr,
			"category": "sealed",
		}, findOptions)

		if err != nil {
			fmt.Printf("Error retrieving sealed for game %s: %v\n", gameStr, err)
			continue
		}

		var cards []models.Card
		if err = cursor.All(ctx, &cards); err != nil {
			cursor.Close(ctx)
			fmt.Printf("Error decoding sealed for game %s: %v\n", gameStr, err)
			continue
		}
		cursor.Close(ctx)

		gameGroups = append(gameGroups, models.GameCardGroup{
			Game:       gameStr,
			Category:   "sealed",
			Cards:      cards,
			TotalCount: int(totalCount),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gameGroups)
}
