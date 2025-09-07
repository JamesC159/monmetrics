package database

import (
	"context"
	"time"

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
		// In production, you might want to handle this differently
		// For now, just print a warning
		println("Warning: Failed to create indexes:", err.Error())
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Users collection indexes
	usersCollection := db.Collection("users")
	_, err := usersCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"email": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"created_at": 1},
		},
	})
	if err != nil {
		return err
	}

	// Cards collection indexes
	cardsCollection := db.Collection("cards")
	_, err = cardsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: map[string]interface{}{
				"name":         "text",
				"set":          "text",
				"game":         "text",
				"search_terms": "text",
			},
			Options: options.Index().SetName("search_index"),
		},
		{
			Keys: map[string]interface{}{"game": 1, "set": 1},
		},
		{
			Keys: map[string]interface{}{"category": 1},
		},
		{
			Keys: map[string]interface{}{"updated_at": 1},
		},
	})
	if err != nil {
		return err
	}

	// Price points collection indexes
	pricesCollection := db.Collection("price_points")
	_, err = pricesCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"card_id": 1, "timestamp": -1},
		},
		{
			Keys: map[string]interface{}{"card_id": 1, "source": 1, "timestamp": -1},
		},
		{
			Keys: map[string]interface{}{"timestamp": 1},
			Options: options.Index().SetExpireAfterSeconds(int32((5 * 365 * 24 * time.Hour).Seconds())), // 5 years TTL
		},
	})
	if err != nil {
		return err
	}

	// Market data collection indexes
	marketDataCollection := db.Collection("market_data")
	_, err = marketDataCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    map[string]interface{}{"card_id": 1, "date": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: map[string]interface{}{"card_id": 1, "date": -1},
		},
	})
	if err != nil {
		return err
	}

	// Saved charts collection indexes
	chartsCollection := db.Collection("saved_charts")
	_, err = chartsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: map[string]interface{}{"user_id": 1, "created_at": -1},
		},
		{
			Keys: map[string]interface{}{"user_id": 1, "card_id": 1},
		},
	})
	if err != nil {
		return err
	}

	return nil
}