#!/bin/bash

# MonMetrics Database Seeder Setup Script
# This script sets up the database with sample trading card data

echo "🚀 MonMetrics Database Seeder Setup"
echo "=================================="

# Check if we're in the correct directory
if [ ! -f "backend/go.mod" ]; then
    echo "❌ Error: Please run this script from the project root directory"
    echo "   Expected to find backend/go.mod"
    exit 1
fi

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed or not in PATH"
    echo "   Please install Go 1.21+ from https://golang.org/dl/"
    exit 1
fi

# Check if Docker is installed and running
if ! command -v docker &> /dev/null; then
    echo "❌ Error: Docker is not installed"
    echo "   Please install Docker from https://docker.com/get-started"
    exit 1
fi

if ! docker info &> /dev/null; then
    echo "❌ Error: Docker daemon is not running"
    echo "   Please start Docker and try again"
    exit 1
fi

echo "✅ Prerequisites check passed"
echo ""

# Create the seeder directory and file
echo "📁 Creating seeder directory..."
mkdir -p backend/cmd/seeder

# Copy the seeder code (this would be the seeder script we created)
cat > backend/cmd/seeder/main.go << 'EOF'
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jamesc159/monmetrics/configs"
	"github.com/jamesc159/monmetrics/internal/database"
	"github.com/jamesc159/monmetrics/internal/models"
)

func main() {
	// Load configuration
	config := configs.Load()

	// Initialize database
	db, err := database.Connect(config.MongoURI, config.DBName)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Disconnect()

	fmt.Println("Starting database seeding...")

	// Clear existing cards (optional - comment out to preserve existing data)
	ctx := context.Background()
	cardsCollection := db.Collection("cards")
	pricesCollection := db.Collection("prices")

	// Remove existing data
	_, err = cardsCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear cards collection: %v", err)
	}

	_, err = pricesCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear prices collection: %v", err)
	}

	// Seed cards
	cards := []models.Card{
		// Pokemon Cards
		{
			Name:         "Charizard VMAX",
			Set:          "Champions Path",
			Game:         "Pokemon",
			Category:     "card",
			Rarity:       "VMAX",
			Number:       "020/073",
			ImageURL:     "https://images.pokemontcg.io/swsh35/20_hires.png",
			Description:  "A powerful Fire-type Pokemon VMAX card from Champions Path",
			CurrentPrice: 89.99,
			AllTimeHigh:  350.00,
			AllTimeLow:   45.00,
			ATHDate:      time.Now().AddDate(0, -8, 0),
			ATLDate:      time.Now().AddDate(0, -2, 0),
			SearchTerms:  []string{"charizard", "vmax", "champions", "path", "fire", "pokemon"},
			Tags:         []string{"popular", "valuable", "competitive"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Name:         "Pikachu VMAX",
			Set:          "Vivid Voltage",
			Game:         "Pokemon",
			Category:     "card",
			Rarity:       "VMAX",
			Number:       "044/185",
			ImageURL:     "https://images.pokemontcg.io/swsh4/44_hires.png",
			Description:  "Electric-type Pokemon VMAX card featuring the iconic Pikachu",
			CurrentPrice: 25.99,
			AllTimeHigh:  89.99,
			AllTimeLow:   15.00,
			ATHDate:      time.Now().AddDate(0, -6, 0),
			ATLDate:      time.Now().AddDate(0, -1, 0),
			SearchTerms:  []string{"pikachu", "vmax", "vivid", "voltage", "electric", "pokemon"},
			Tags:         []string{"iconic", "electric", "popular"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Name:         "Lugia VSTAR",
			Set:          "Silver Tempest",
			Game:         "Pokemon",
			Category:     "card",
			Rarity:       "VSTAR",
			Number:       "139/195",
			ImageURL:     "https://images.pokemontcg.io/swsh12/139_hires.png",
			Description:  "Psychic-type legendary Pokemon VSTAR with incredible power",
			CurrentPrice: 45.50,
			AllTimeHigh:  125.00,
			AllTimeLow:   25.00,
			ATHDate:      time.Now().AddDate(0, -4, 0),
			ATLDate:      time.Now().AddDate(0, -1, 0),
			SearchTerms:  []string{"lugia", "vstar", "silver", "tempest", "psychic", "legendary", "pokemon"},
			Tags:         []string{"legendary", "powerful", "recent"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},

		// Magic The Gathering Cards
		{
			Name:         "Black Lotus",
			Set:          "Alpha",
			Game:         "Magic The Gathering",
			Category:     "card",
			Rarity:       "Rare",
			Number:       "",
			ImageURL:     "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=3&type=card",
			Description:  "The most iconic and valuable Magic card ever printed",
			CurrentPrice: 65000.00,
			AllTimeHigh:  87000.00,
			AllTimeLow:   45000.00,
			ATHDate:      time.Now().AddDate(-1, 0, 0),
			ATLDate:      time.Now().AddDate(-2, 0, 0),
			SearchTerms:  []string{"black", "lotus", "alpha", "power", "nine", "vintage", "magic"},
			Tags:         []string{"power-nine", "vintage", "investment", "iconic"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Name:         "Tarmogoyf",
			Set:          "Future Sight",
			Game:         "Magic The Gathering",
			Category:     "card",
			Rarity:       "Rare",
			Number:       "153/180",
			ImageURL:     "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=136142&type=card",
			Description:  "A powerful green creature that defines Modern format",
			CurrentPrice: 89.99,
			AllTimeHigh:  199.99,
			AllTimeLow:   45.00,
			ATHDate:      time.Now().AddDate(0, -18, 0),
			ATLDate:      time.Now().AddDate(0, -3, 0),
			SearchTerms:  []string{"tarmogoyf", "future", "sight", "green", "creature", "modern", "magic"},
			Tags:         []string{"modern-staple", "competitive", "green"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Name:         "Lightning Bolt",
			Set:          "Alpha",
			Game:         "Magic The Gathering",
			Category:     "card",
			Rarity:       "Common",
			Number:       "",
			ImageURL:     "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=209&type=card",
			Description:  "Deal 3 damage to any target - a classic red instant",
			CurrentPrice: 125.00,
			AllTimeHigh:  250.00,
			AllTimeLow:   85.00,
			ATHDate:      time.Now().AddDate(0, -12, 0),
			ATLDate:      time.Now().AddDate(0, -6, 0),
			SearchTerms:  []string{"lightning", "bolt", "alpha", "red", "instant", "damage", "magic"},
			Tags:         []string{"classic", "red", "vintage"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},

		// Yu-Gi-Oh Cards
		{
			Name:         "Blue-Eyes White Dragon",
			Set:          "Legend of Blue Eyes White Dragon",
			Game:         "Yu-Gi-Oh",
			Category:     "card",
			Rarity:       "Ultra Rare",
			Number:       "LOB-001",
			ImageURL:     "https://images.ygoprodeck.com/images/cards/89631139.jpg",
			Description:  "This legendary dragon is a powerful engine of destruction",
			CurrentPrice: 2500.00,
			AllTimeHigh:  5500.00,
			AllTimeLow:   1200.00,
			ATHDate:      time.Now().AddDate(-1, -6, 0),
			ATLDate:      time.Now().AddDate(0, -4, 0),
			SearchTerms:  []string{"blue", "eyes", "white", "dragon", "legend", "lob", "kaiba", "yugioh"},
			Tags:         []string{"iconic", "dragon", "kaiba", "nostalgic"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Name:         "Dark Magician",
			Set:          "Legend of Blue Eyes White Dragon",
			Game:         "Yu-Gi-Oh",
			Category:     "card",
			Rarity:       "Ultra Rare",
			Number:       "LOB-005",
			ImageURL:     "https://images.ygoprodeck.com/images/cards/46986414.jpg",
			Description:  "The ultimate wizard in terms of attack and defense",
			CurrentPrice: 1800.00,
			AllTimeHigh:  3200.00,
			AllTimeLow:   900.00,
			ATHDate:      time.Now().AddDate(-1, -3, 0),
			ATLDate:      time.Now().AddDate(0, -5, 0),
			SearchTerms:  []string{"dark", "magician", "legend", "lob", "spellcaster", "yugi", "yugioh"},
			Tags:         []string{"iconic", "spellcaster", "yugi", "nostalgic"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Name:         "Exodia the Forbidden One",
			Set:          "Legend of Blue Eyes White Dragon",
			Game:         "Yu-Gi-Oh",
			Category:     "card",
			Rarity:       "Ultra Rare",
			Number:       "LOB-124",
			ImageURL:     "https://images.ygoprodeck.com/images/cards/33396948.jpg",
			Description:  "If you have all 5 pieces, you automatically win the duel",
			CurrentPrice: 450.00,
			AllTimeHigh:  850.00,
			AllTimeLow:   250.00,
			ATHDate:      time.Now().AddDate(0, -9, 0),
			ATLDate:      time.Now().AddDate(0, -2, 0),
			SearchTerms:  []string{"exodia", "forbidden", "one", "legend", "lob", "win", "condition", "yugioh"},
			Tags:         []string{"win-condition", "rare", "nostalgic"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},

		// Sealed Products
		{
			Name:         "Pokemon Base Set Booster Box",
			Set:          "Base Set",
			Game:         "Pokemon",
			Category:     "sealed",
			Rarity:       "",
			Number:       "",
			ImageURL:     "https://52f4e29a8321344e30ae-0f55c9129972ac85d6b1f4e703468e6b.ssl.cf2.rackcdn.com/products/pictures/1085368.jpg",
			Description:  "Factory sealed Pokemon Base Set booster box - 36 packs",
			CurrentPrice: 45000.00,
			AllTimeHigh:  75000.00,
			AllTimeLow:   25000.00,
			ATHDate:      time.Now().AddDate(-1, 0, 0),
			ATLDate:      time.Now().AddDate(-3, 0, 0),
			SearchTerms:  []string{"pokemon", "base", "set", "booster", "box", "sealed", "vintage"},
			Tags:         []string{"sealed", "investment", "vintage", "rare"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			Name:         "Magic Alpha Starter Deck",
			Set:          "Alpha",
			Game:         "Magic The Gathering",
			Category:     "sealed",
			Rarity:       "",
			Number:       "",
			ImageURL:     "https://crystal-cdn4.crystalcommerce.com/photos/6213849/large/en_alpha_starterdecktype1.jpg",
			Description:  "Factory sealed Magic Alpha starter deck - extremely rare",
			CurrentPrice: 125000.00,
			AllTimeHigh:  200000.00,
			AllTimeLow:   85000.00,
			ATHDate:      time.Now().AddDate(-2, 0, 0),
			ATLDate:      time.Now().AddDate(-4, 0, 0),
			SearchTerms:  []string{"magic", "alpha", "starter", "deck", "sealed", "vintage", "93"},
			Tags:         []string{"sealed", "alpha", "investment", "museum"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Insert cards
	cardDocuments := make([]interface{}, len(cards))
	for i, card := range cards {
		cardDocuments[i] = card
	}

	result, err := cardsCollection.InsertMany(ctx, cardDocuments)
	if err != nil {
		log.Fatal("Failed to insert cards:", err)
	}

	fmt.Printf("Successfully inserted %d cards\n", len(result.InsertedIDs))

	// Generate price history for each card
	fmt.Println("Generating price history data...")

	for i, cardID := range result.InsertedIDs {
		objectID := cardID.(primitive.ObjectID)
		card := cards[i]

		// Generate 5 years of daily price data
		startDate := time.Now().AddDate(-5, 0, 0)
		endDate := time.Now()
		currentDate := startDate

		var prices []interface{}
		currentPrice := card.AllTimeLow + (card.AllTimeHigh-card.AllTimeLow)*0.3 // Start at 30% of range

		for currentDate.Before(endDate) {
			// Add some randomness to price movements
			change := (rand.Float64() - 0.5) * 0.1 // +/- 5% change
			if rand.Float64() < 0.02 { // 2% chance of large movement
				change = (rand.Float64() - 0.5) * 0.3 // +/- 15% change
			}

			currentPrice *= (1 + change)

			// Ensure price stays within reasonable bounds
			if currentPrice < card.AllTimeLow {
				currentPrice = card.AllTimeLow
			}
			if currentPrice > card.AllTimeHigh {
				currentPrice = card.AllTimeHigh
			}

			// Generate volume (random between 1-50 sales per day)
			volume := rand.Intn(50) + 1

			// Create price points for both eBay and TCGPlayer
			sources := []string{"ebay", "tcgplayer"}
			for _, source := range sources {
				// Add slight variation for different sources
				sourcePrice := currentPrice * (1 + (rand.Float64()-0.5)*0.05)

				pricePoint := models.PricePoint{
					CardID:    objectID,
					Price:     sourcePrice,
					Volume:    volume,
					Source:    source,
					Timestamp: currentDate,
					CreatedAt: currentDate,
				}
				prices = append(prices, pricePoint)
			}

			currentDate = currentDate.AddDate(0, 0, 1)
		}

		// Insert price history in batches to avoid memory issues
		batchSize := 1000
		for i := 0; i < len(prices); i += batchSize {
			end := i + batchSize
			if end > len(prices) {
				end = len(prices)
			}

			batch := prices[i:end]
			_, err := pricesCollection.InsertMany(ctx, batch)
			if err != nil {
				log.Printf("Warning: Failed to insert price batch for card %s: %v", card.Name, err)
			}
		}

		fmt.Printf("Generated price history for: %s (%d data points)\n", card.Name, len(prices))
	}

	// Create sample listings for cards
	fmt.Println("Creating sample listings...")
	listingsCollection := db.Collection("listings")

	// Clear existing listings
	_, err = listingsCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear listings collection: %v", err)
	}

	// Create listings for first 5 cards as examples
	for i, cardID := range result.InsertedIDs[:5] {
		objectID := cardID.(primitive.ObjectID)
		card := cards[i]

		// Generate 3-8 random listings per card
		numListings := rand.Intn(6) + 3
		for j := 0; j < numListings; j++ {
			// Price variation around current price
			priceVariation := (rand.Float64() - 0.5) * 0.4 // +/- 20%
			listingPrice := card.CurrentPrice * (1 + priceVariation)

			listing := bson.M{
				"card_id":     objectID,
				"title":       fmt.Sprintf("%s - %s", card.Name, getRandomCondition()),
				"price":       listingPrice,
				"quantity":    rand.Intn(3) + 1,
				"condition":   getRandomCondition(),
				"seller":      getRandomSeller(),
				"source":      getRandomSource(),
				"image_url":   card.ImageURL,
				"created_at":  time.Now().AddDate(0, 0, -rand.Intn(30)),
				"updated_at":  time.Now(),
			}

			_, err := listingsCollection.InsertOne(ctx, listing)
			if err != nil {
				log.Printf("Warning: Failed to insert listing: %v", err)
			}
		}

		fmt.Printf("Created %d listings for: %s\n", numListings, card.Name)
	}

	fmt.Println("\n✅ Database seeding completed successfully!")
	fmt.Printf("Total cards created: %d\n", len(cards))
	fmt.Printf("Price history: ~%d data points per card\n", 5*365*2) // 5 years * 365 days * 2 sources
	fmt.Printf("Sample listings created for first 5 cards\n")
	fmt.Println("\nYou can now:")
	fmt.Println("1. Start the development server: make dev")
	fmt.Println("2. Visit http://localhost:3000/search to test the search functionality")
	fmt.Println("3. Search for cards like 'Charizard', 'Black Lotus', or 'Blue-Eyes'")
}

func getRandomCondition() string {
	conditions := []string{"Near Mint", "Lightly Played", "Moderately Played", "Heavily Played", "Damaged"}
	return conditions[rand.Intn(len(conditions))]
}

func getRandomSeller() string {
	sellers := []string{"CardMaster2024", "TradingCardPro", "CollectorHaven", "CardVault", "GamingCards123", "MintConditionOnly", "CardLegends", "TopDeckGaming"}
	return sellers[rand.Intn(len(sellers))]
}

func getRandomSource() string {
	sources := []string{"ebay", "tcgplayer"}
	return sources[rand.Intn(len(sources))]
}
EOF

echo "✅ Seeder script created"
echo ""

# Start MongoDB
echo "🐳 Starting MongoDB container..."
docker-compose up -d mongodb

# Wait for MongoDB to be ready
echo "⏳ Waiting for MongoDB to be ready..."
sleep 5

# Test MongoDB connection
echo "🔍 Testing MongoDB connection..."
if ! docker exec monmetrics_mongo mongosh --eval "db.adminCommand('ping')" --quiet > /dev/null 2>&1; then
    echo "❌ Error: MongoDB is not responding"
    echo "   Try: docker-compose down && docker-compose up -d mongodb"
    exit 1
fi

echo "✅ MongoDB is ready"
echo ""

# Build and run the seeder
echo "🌱 Running database seeder..."
cd backend

# Ensure dependencies are up to date
echo "📦 Updating Go dependencies..."
go mod tidy

# Build the seeder
echo "🔨 Building seeder..."
if ! go build -o bin/seeder cmd/seeder/main.go; then
    echo "❌ Error: Failed to build seeder"
    exit 1
fi

# Run the seeder
echo "🚀 Executing seeder..."
if ! ./bin/seeder; then
    echo "❌ Error: Seeder execution failed"
    exit 1
fi

cd ..

echo ""
echo "🎉 Setup completed successfully!"
echo ""
echo "Next steps:"
echo "1. Start the development servers:"
echo "   make dev"
echo ""
echo "2. Open your browser to:"
echo "   http://localhost:3000"
echo ""
echo "3. Try searching for these cards:"
echo "   • Charizard"
echo "   • Black Lotus"
echo "   • Blue-Eyes White Dragon"
echo "   • Pikachu"
echo ""
echo "📊 Database contains:"
echo "   • 11 sample cards across Pokemon, Magic, and Yu-Gi-Oh"
echo "   • 5 years of price history for each card"
echo "   • Sample marketplace listings"
echo ""
echo "Happy trading! 🎯"