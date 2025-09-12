package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

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

	// Note: We don't call database.Disconnect() here since it's handled internally
	// The connection will be closed when the program exits

	fmt.Println("🌱 Starting database seeding...")

	// Clear existing data (optional - comment out to preserve existing data)
	ctx := context.Background()
	cardsCollection := db.Collection("cards")
	pricesCollection := db.Collection("prices")
	listingsCollection := db.Collection("listings")

	fmt.Println("🗑️  Clearing existing data...")
	// Remove existing data
	_, err = cardsCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear cards collection: %v", err)
	}

	_, err = pricesCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear prices collection: %v", err)
	}

	_, err = listingsCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear listings collection: %v", err)
	}

	// Seed cards with comprehensive sample data
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
		{
			Name:         "Arceus VSTAR",
			Set:          "Brilliant Stars",
			Game:         "Pokemon",
			Category:     "card",
			Rarity:       "VSTAR",
			Number:       "123/172",
			ImageURL:     "https://images.pokemontcg.io/swsh9/123_hires.png",
			Description:  "The Alpha Pokemon with ultimate versatility",
			CurrentPrice: 67.50,
			AllTimeHigh:  150.00,
			AllTimeLow:   35.00,
			ATHDate:      time.Now().AddDate(0, -5, 0),
			ATLDate:      time.Now().AddDate(0, -1, 0),
			SearchTerms:  []string{"arceus", "vstar", "brilliant", "stars", "colorless", "alpha", "pokemon"},
			Tags:         []string{"legendary", "versatile", "meta"},
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
		{
			Name:         "Mox Ruby",
			Set:          "Alpha",
			Game:         "Magic The Gathering",
			Category:     "card",
			Rarity:       "Rare",
			Number:       "",
			ImageURL:     "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=263&type=card",
			Description:  "Part of the iconic Power Nine, provides red mana",
			CurrentPrice: 8500.00,
			AllTimeHigh:  12000.00,
			AllTimeLow:   5500.00,
			ATHDate:      time.Now().AddDate(-1, -2, 0),
			ATLDate:      time.Now().AddDate(-2, -6, 0),
			SearchTerms:  []string{"mox", "ruby", "alpha", "power", "nine", "red", "mana", "magic"},
			Tags:         []string{"power-nine", "vintage", "mana", "red"},
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
		{
			Name:         "Yu-Gi-Oh LOB Booster Box",
			Set:          "Legend of Blue Eyes White Dragon",
			Game:         "Yu-Gi-Oh",
			Category:     "sealed",
			Rarity:       "",
			Number:       "",
			ImageURL:     "https://images.ygoprodeck.com/pics_artgame/55210709.jpg",
			Description:  "First edition LOB booster box - 24 packs",
			CurrentPrice: 15000.00,
			AllTimeHigh:  25000.00,
			AllTimeLow:   8500.00,
			ATHDate:      time.Now().AddDate(-1, -4, 0),
			ATLDate:      time.Now().AddDate(-2, -8, 0),
			SearchTerms:  []string{"yugioh", "lob", "legend", "booster", "box", "sealed", "first", "edition"},
			Tags:         []string{"sealed", "vintage", "first-edition", "investment"},
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	// Insert cards
	fmt.Printf("📦 Inserting %d sample cards...\n", len(cards))
	cardDocuments := make([]interface{}, len(cards))
	for i, card := range cards {
		cardDocuments[i] = card
	}

	result, err := cardsCollection.InsertMany(ctx, cardDocuments)
	if err != nil {
		log.Fatal("Failed to insert cards:", err)
	}

	fmt.Printf("✅ Successfully inserted %d cards\n", len(result.InsertedIDs))

	// Generate price history for each card
	fmt.Println("📈 Generating price history data...")

	for i, cardID := range result.InsertedIDs {
		objectID := cardID.(primitive.ObjectID)
		card := cards[i]

		fmt.Printf("   Processing: %s\n", card.Name)

		// Generate 5 years of daily price data
		startDate := time.Now().AddDate(-5, 0, 0)
		endDate := time.Now()
		currentDate := startDate

		var prices []interface{}
		currentPrice := card.AllTimeLow + (card.AllTimeHigh-card.AllTimeLow)*0.3 // Start at 30% of range

		dayCount := 0
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
			dayCount++
		}

		// Insert price history in batches to avoid memory issues
		batchSize := 1000
		totalBatches := (len(prices) + batchSize - 1) / batchSize

		for i := 0; i < len(prices); i += batchSize {
			end := i + batchSize
			if end > len(prices) {
				end = len(prices)
			}

			batch := prices[i:end]
			_, err := pricesCollection.InsertMany(ctx, batch)
			if err != nil {
				log.Printf("Warning: Failed to insert price batch for card %s: %v", card.Name, err)
				continue // Continue with next batch instead of failing completely
			}

			if totalBatches > 1 {
				fmt.Printf("     Batch %d/%d complete\n", (i/batchSize)+1, totalBatches)
			}
		}

		fmt.Printf("   ✅ %s: %d price points (%d days)\n", card.Name, len(prices), dayCount)
	}

	// Create sample listings for cards
	fmt.Println("🏪 Creating sample marketplace listings...")

	// Create listings for first 8 cards as examples
	totalListings := 0
	for i, cardID := range result.InsertedIDs {
		if i >= 8 { // Limit to first 8 cards
			break
		}

		objectID := cardID.(primitive.ObjectID)
		card := cards[i]

		// Generate 3-8 random listings per card
		numListings := rand.Intn(6) + 3
		var listings []interface{}

		for j := 0; j < numListings; j++ {
			// Price variation around current price
			priceVariation := (rand.Float64() - 0.5) * 0.4 // +/- 20%
			listingPrice := card.CurrentPrice * (1 + priceVariation)

			listing := models.Listing{
				CardID:    objectID,
				Title:     fmt.Sprintf("%s - %s", card.Name, getRandomCondition()),
				Price:     listingPrice,
				Quantity:  rand.Intn(3) + 1,
				Condition: getRandomCondition(),
				Seller:    getRandomSeller(),
				Source:    getRandomSource(),
				ImageURL:  card.ImageURL,
				CreatedAt: time.Now().AddDate(0, 0, -rand.Intn(30)),
				UpdatedAt: time.Now(),
			}

			listings = append(listings, listing)
		}

		// Insert all listings for this card at once
		_, err := listingsCollection.InsertMany(ctx, listings)
		if err != nil {
			log.Printf("Warning: Failed to insert listings for card %s: %v", card.Name, err)
		} else {
			fmt.Printf("   ✅ %s: %d listings created\n", card.Name, numListings)
			totalListings += numListings
		}
	}

	// Create text search indexes for better performance
	fmt.Println("🔍 Creating database indexes for optimal performance...")
	createSearchIndexes(ctx, db)

	fmt.Println("\n🎉 Database seeding completed successfully!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   • Cards created: %d\n", len(cards))
	fmt.Printf("   • Price data points: ~%d per card\n", 5*365*2) // 5 years * 365 days * 2 sources
	fmt.Printf("   • Total price records: ~%d\n", len(cards)*5*365*2)
	fmt.Printf("   • Sample listings: %d\n", totalListings)
	fmt.Println("\n🎯 Sample cards you can search for:")
	fmt.Println("   • Charizard VMAX")
	fmt.Println("   • Black Lotus")
	fmt.Println("   • Blue-Eyes White Dragon")
	fmt.Println("   • Pikachu VMAX")
	fmt.Println("   • Pokemon Base Set Booster Box")
	fmt.Println("   • Magic Alpha Starter Deck")
	fmt.Println("   • Mox Ruby")
	fmt.Println("   • Dark Magician")
	fmt.Println("\n✨ Ready to start the development server!")
	fmt.Println("   Backend: make dev (or cd backend && go run cmd/server/main.go)")
	fmt.Println("   Frontend: cd frontend && npm run dev")
	fmt.Println("   Health Check: http://localhost:8080/health")
	fmt.Println("   Frontend: http://localhost:3000")
}

// createSearchIndexes creates MongoDB text search indexes for better search performance
func createSearchIndexes(ctx context.Context, db *mongo.Database) {
	cardsCollection := db.Collection("cards")

	// Create a comprehensive text index for searching
	_, err := cardsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: map[string]interface{}{
			"name":         "text",
			"set":          "text",
			"game":         "text",
			"search_terms": "text",
			"tags":         "text",
		},
	})

	if err != nil {
		log.Printf("Warning: Failed to create text search index: %v", err)
	} else {
		fmt.Println("   ✅ Created text search index for cards")
	}

	// Create additional performance indexes
	indexes := []mongo.IndexModel{
		{Keys: map[string]interface{}{"game": 1, "category": 1}},
		{Keys: map[string]interface{}{"current_price": -1}},
		{Keys: map[string]interface{}{"updated_at": -1}},
	}

	_, err = cardsCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		log.Printf("Warning: Failed to create performance indexes: %v", err)
	} else {
		fmt.Println("   ✅ Created performance indexes for cards")
	}

	// Create price points indexes
	pricesCollection := db.Collection("prices")
	priceIndexes := []mongo.IndexModel{
		{Keys: map[string]interface{}{"card_id": 1, "timestamp": -1}},
		{Keys: map[string]interface{}{"card_id": 1, "source": 1, "timestamp": -1}},
	}

	_, err = pricesCollection.Indexes().CreateMany(ctx, priceIndexes)
	if err != nil {
		log.Printf("Warning: Failed to create price indexes: %v", err)
	} else {
		fmt.Println("   ✅ Created indexes for price history")
	}
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