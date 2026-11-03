package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	h.stripHiddenFeaturedImages(ctx, featuredContent)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(featuredContent)
}

// stripHiddenFeaturedImages blanks carousel art copied from a withheld game's card.
// Items without a card (news, sponsored) carry their own art and are left alone.
func (h *Handlers) stripHiddenFeaturedImages(ctx context.Context, items []models.FeaturedContent) {
	if h.config == nil || len(h.config.HiddenImageGames) == 0 {
		return
	}
	ids := make([]primitive.ObjectID, 0, len(items))
	for _, item := range items {
		if item.CardID != nil {
			ids = append(ids, *item.CardID)
		}
	}
	cards, err := h.cardsByID(ctx, ids)
	if err != nil {
		fmt.Printf("Error checking featured card games: %v\n", err)
	}
	for i := range items {
		if items[i].CardID == nil {
			continue
		}
		// Fail closed: if the lookup failed, withhold every card-linked image.
		card, ok := cards[*items[i].CardID]
		if err != nil || (ok && h.imageHidden(card.Game)) {
			items[i].ImageURL = ""
		}
	}
}

// GetCardsByGame retrieves cards organized by game with popularity sorting
func (h *Handlers) GetCardsByGame(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := h.db.Collection("cards")

	// Get all unique games
	games, err := collection.Distinct(ctx, "game", h.applyEnabledGames(bson.M{"category": "card"}))
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

		cards, err := h.topCardsForGame(ctx, gameStr, "card")
		if err != nil {
			fmt.Printf("Error retrieving cards for game %s: %v\n", gameStr, err)
			continue
		}

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
	games, err := collection.Distinct(ctx, "game", h.applyEnabledGames(bson.M{"category": "sealed"}))
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

		cards, err := h.topCardsForGame(ctx, gameStr, "sealed")
		if err != nil {
			fmt.Printf("Error retrieving sealed for game %s: %v\n", gameStr, err)
			continue
		}

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

// gameGroupSize is how many items each by-game row shows (2 rows of 6).
const gameGroupSize = 12

// topCardsForGame returns the items to feature for a game, preferring ones that
// already have scraped sales data; unpriced catalog entries only fill the gap.
func (h *Handlers) topCardsForGame(ctx context.Context, game, category string) ([]models.Card, error) {
	collection := h.db.Collection("cards")
	base := bson.M{"game": game, "category": category, "active": bson.M{"$ne": false}}

	fetch := func(filter bson.M, limit int) ([]models.Card, error) {
		opts := options.Find().SetSort(bson.D{{Key: "current_price", Value: -1}}).SetLimit(int64(limit))
		cursor, err := collection.Find(ctx, filter, opts)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)
		var out []models.Card
		if err := cursor.All(ctx, &out); err != nil {
			return nil, err
		}
		h.stripHiddenImages(out)
		return out, nil
	}

	priced := bson.M{"current_price": bson.M{"$gt": 0}}
	for k, v := range base {
		priced[k] = v
	}
	cards, err := fetch(priced, gameGroupSize)
	if err != nil {
		return nil, err
	}
	if len(cards) >= gameGroupSize {
		return cards, nil
	}

	unpriced := bson.M{"current_price": bson.M{"$not": bson.M{"$gt": 0}}}
	for k, v := range base {
		unpriced[k] = v
	}
	filler, err := fetch(unpriced, gameGroupSize-len(cards))
	if err != nil {
		return nil, err
	}
	return append(cards, filler...), nil
}
