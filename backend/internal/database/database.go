package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// Connect establishes connection to MongoDB
func Connect(uri, dbName string) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMaxPoolSize(100)
	clientOptions.SetMaxConnIdleTime(30 * time.Second)
	clientOptions.SetConnectTimeout(10 * time.Second)

	// Connect to MongoDB
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	db := client.Database(dbName)

	// Create indexes (but don't fail if it errors - let app start anyway)
	if err := createIndexes(db); err != nil {
		// Log the error but don't fail - indexes can be created later
		fmt.Printf("Warning: Failed to create indexes: %v\n", err)
	}

	return db, nil
}

// Disconnect closes the MongoDB connection
func Disconnect() error {
	if client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return client.Disconnect(ctx)
}

// createIndexes creates necessary database indexes for performance
func createIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Increased timeout
	defer cancel()

	// Users collection indexes
	usersCollection := db.Collection("users")
	_, err := usersCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "created_at", Value: 1}},
		},
	})
	if err != nil {
		fmt.Printf("Warning: Failed to create users indexes: %v\n", err)
	}

	// Cards collection indexes
	cardsCollection := db.Collection("cards")

	// Check if collection has documents before creating text index
	count, err := cardsCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		fmt.Printf("Warning: Could not count cards documents: %v\n", err)
	}

	if count > 0 {
		// Try to create text search index
		textIndex := mongo.IndexModel{
			Keys: bson.D{
				{Key: "name", Value: "text"},
				{Key: "set", Value: "text"},
				{Key: "game", Value: "text"},
				{Key: "search_terms", Value: "text"},
			},
			Options: options.Index().SetName("search_index"),
		}

		_, err = cardsCollection.Indexes().CreateOne(ctx, textIndex)
		if err != nil {
			fmt.Printf("Warning: Failed to create text search index: %v\n", err)
		}
	} else {
		fmt.Println("Info: Skipping text index creation - no cards in database yet")
	}

	// Create other indexes for cards
	otherIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "game", Value: 1}, {Key: "set", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "category", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}},
		},
	}

	_, err = cardsCollection.Indexes().CreateMany(ctx, otherIndexes)
	if err != nil {
		fmt.Printf("Warning: Failed to create some card indexes: %v\n", err)
	}

	// Price points collection indexes
	pricesCollection := db.Collection("prices")
	_, err = pricesCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "card_id", Value: 1}, {Key: "timestamp", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "card_id", Value: 1}, {Key: "source", Value: 1}, {Key: "timestamp", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "timestamp", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(int32((5 * 365 * 24 * time.Hour).Seconds())), // 5 years TTL
		},
	})
	if err != nil {
		fmt.Printf("Warning: Failed to create price indexes: %v\n", err)
	}

	// Market data collection indexes
	marketDataCollection := db.Collection("market_data")
	_, err = marketDataCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "card_id", Value: 1}, {Key: "date", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "card_id", Value: 1}, {Key: "date", Value: -1}},
		},
	})
	if err != nil {
		fmt.Printf("Warning: Failed to create market data indexes: %v\n", err)
	}

	// Saved charts collection indexes
	chartsCollection := db.Collection("saved_charts")
	_, err = chartsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "card_id", Value: 1}},
		},
	})
	if err != nil {
		fmt.Printf("Warning: Failed to create chart indexes: %v\n", err)
	}

	// Listings collection indexes
	listingsCollection := db.Collection("listings")
	_, err = listingsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "card_id", Value: 1}, {Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "card_id", Value: 1}, {Key: "source", Value: 1}},
		},
	})
	if err != nil {
		fmt.Printf("Warning: Failed to create listing indexes: %v\n", err)
	}

	return nil
}

// RecreateTextIndex creates the text search index after data has been seeded
func RecreateTextIndex(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cardsCollection := db.Collection("cards")

	// Drop existing text index if it exists
	indexes, err := cardsCollection.Indexes().List(ctx)
	if err == nil {
		defer indexes.Close(ctx)
		for indexes.Next(ctx) {
			var index bson.M
			if err := indexes.Decode(&index); err == nil {
				if name, ok := index["name"].(string); ok && strings.Contains(name, "search") {
					cardsCollection.Indexes().DropOne(ctx, name)
					fmt.Printf("Dropped existing search index: %s\n", name)
				}
			}
		}
	}

	// Create new text search index
	textIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: "text"},
			{Key: "set", Value: "text"},
			{Key: "game", Value: "text"},
			{Key: "search_terms", Value: "text"},
		},
		Options: options.Index().SetName("search_index"),
	}

	indexName, err := cardsCollection.Indexes().CreateOne(ctx, textIndex)
	if err != nil {
		return fmt.Errorf("failed to create text search index: %v", err)
	}

	fmt.Printf("Successfully created text search index: %s\n", indexName)
	return nil
}