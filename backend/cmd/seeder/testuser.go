package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/jamesc159/monmetrics/configs"
	"github.com/jamesc159/monmetrics/internal/models"
	"github.com/jamesc159/monmetrics/internal/secure"
	"github.com/jamesc159/monmetrics/internal/valuation"
)

// Test user fixtures cover every portfolio, favorites, saved-chart, and marketplace
// state. They are rebuilt on every seed because card IDs change when cards are reinserted.

type portfolioFixture struct {
	card           string // catalog name; empty for custom items
	itemType       string
	condition      string
	finish         string
	language       string
	grading        *models.Grading
	quantity       int
	purchaseFactor float64 // purchase price as a fraction of current unit value
	purchasePrice  float64 // used when there is no market value
	monthsAgo      int
	source         string
	manualValue    *float64
	notes          string
	customName     string
	customGame     string
	customSet      string
}

func f64(v float64) *float64 { return &v }

func testUserFixtures() []portfolioFixture {
	return []portfolioFixture{
		// Raw cards: every condition, finishes, languages, quantity > 1
		{card: "Charizard VMAX", itemType: models.ItemTypeRawCard, condition: "NM", finish: "holo", language: "English", quantity: 1, purchaseFactor: 0.70, monthsAgo: 26, source: "Local game store", notes: "Pulled from a Champions Path ETB."},
		{card: "Umbreon VMAX", itemType: models.ItemTypeRawCard, condition: "LP", finish: "holo", language: "English", quantity: 1, purchaseFactor: 1.35, monthsAgo: 14, source: "eBay"},
		{card: "Pikachu VMAX", itemType: models.ItemTypeRawCard, condition: "MP", finish: "normal", language: "English", quantity: 4, purchaseFactor: 0.90, monthsAgo: 30, source: "TCGPlayer"},
		{card: "Lightning Bolt", itemType: models.ItemTypeRawCard, condition: "HP", finish: "foil", language: "English", quantity: 4, purchaseFactor: 1.10, monthsAgo: 40, source: "Trade"},
		{card: "Dark Magician", itemType: models.ItemTypeRawCard, condition: "DMG", finish: "1st_edition", language: "English", quantity: 1, purchaseFactor: 0.50, monthsAgo: 55, source: "Childhood collection"},
		{card: "Giratina VSTAR", itemType: models.ItemTypeRawCard, condition: "NM", finish: "reverse_holo", language: "Japanese", quantity: 2, purchaseFactor: 0.95, monthsAgo: 8, source: "Import shop"},
		{card: "Tarmogoyf", itemType: models.ItemTypeRawCard, condition: "NM", finish: "normal", language: "English", quantity: 1, purchasePrice: 75, manualValue: f64(60), monthsAgo: 20, source: "Card show", notes: "Manual value from recent in-person offer."},

		// Graded cards: every grading company
		{card: "Charizard VMAX", itemType: models.ItemTypeGradedCard, grading: &models.Grading{Company: "PSA", Grade: 10, CertNumber: "71234567"}, quantity: 1, purchaseFactor: 0.60, monthsAgo: 18, source: "PSA submission"},
		{card: "Umbreon VMAX", itemType: models.ItemTypeGradedCard, grading: &models.Grading{Company: "PSA", Grade: 9, CertNumber: "68800123"}, quantity: 1, purchaseFactor: 1.20, monthsAgo: 12, source: "eBay"},
		{card: "Blue-Eyes White Dragon", itemType: models.ItemTypeGradedCard, grading: &models.Grading{Company: "BGS", Grade: 9.5, CertNumber: "0011223344"}, quantity: 1, purchaseFactor: 0.80, monthsAgo: 36, source: "Auction"},
		{card: "The One Ring", itemType: models.ItemTypeGradedCard, grading: &models.Grading{Company: "CGC", Grade: 10, CertNumber: "4012345001"}, quantity: 1, purchaseFactor: 1.05, monthsAgo: 9, source: "CGC submission"},
		{card: "Mew VMAX", itemType: models.ItemTypeGradedCard, grading: &models.Grading{Company: "SGC", Grade: 8}, quantity: 1, purchaseFactor: 0.75, monthsAgo: 22, source: "Facebook group"},

		// Sealed products across games
		{card: "Pokemon 151 Booster Box", itemType: models.ItemTypeSealed, quantity: 3, purchaseFactor: 0.65, monthsAgo: 24, source: "Target"},
		{card: "Pokemon Evolving Skies Booster Box", itemType: models.ItemTypeSealed, quantity: 1, purchaseFactor: 0.40, monthsAgo: 48, source: "Pre-order"},
		{card: "Pokemon Crown Zenith Elite Trainer Box", itemType: models.ItemTypeSealed, quantity: 2, purchaseFactor: 1.15, monthsAgo: 16, source: "Walmart"},
		{card: "Modern Horizons 3 Play Booster Box", itemType: models.ItemTypeSealed, quantity: 1, purchaseFactor: 1.25, monthsAgo: 6, source: "LGS pre-release"},
		{card: "Yu-Gi-Oh 25th Anniversary Tin", itemType: models.ItemTypeSealed, quantity: 2, purchaseFactor: 0.85, monthsAgo: 10, source: "Amazon"},

		// Custom items: with and without a value
		{itemType: models.ItemTypeRawCard, condition: "NM", language: "English", quantity: 1, purchasePrice: 30, manualValue: f64(45), monthsAgo: 11, source: "Convention",
			customName: "Pikachu Staff Promo (Worlds 2023)", customGame: "Pokemon", customSet: "Worlds Promos", notes: "Not in catalog yet."},
		{itemType: models.ItemTypeSealed, quantity: 1, purchasePrice: 25, monthsAgo: 3, source: "Tournament prize",
			customName: "Regional Championship Playmat", customGame: "Magic The Gathering", customSet: "Organized Play"},
	}
}

func seedTestUser(ctx context.Context, db *mongo.Database, config *configs.Config, cardIDs map[string]primitive.ObjectID) {
	envEmail := strings.ToLower(os.Getenv("TEST_USER_EMAIL"))
	envPassword := os.Getenv("TEST_USER_PASSWORD")

	// Prefer TEST_USER_EMAIL; otherwise reuse the existing "xx xx" account; otherwise create one.
	filter := bson.M{"first_name": "xx", "last_name": "xx"}
	if envEmail != "" {
		filter = bson.M{"email": envEmail}
	}

	now := time.Now().UTC()
	users := db.Collection("users")
	var user models.User
	err := users.FindOne(ctx, filter).Decode(&user)
	switch err {
	case nil:
		set := bson.M{"first_name": "xx", "last_name": "xx", "is_active": true, "updated_at": now}
		if envPassword != "" {
			hash, err := bcrypt.GenerateFromPassword([]byte(envPassword), 12)
			if err != nil {
				log.Printf("Warning: could not hash test user password: %v", err)
				return
			}
			set["password_hash"] = string(hash)
		}
		if _, err := users.UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": set}); err != nil {
			log.Printf("Warning: could not update test user: %v", err)
			return
		}
		fmt.Printf("👤 Seeding existing test user %s (password unchanged unless TEST_USER_PASSWORD is set)...\n", user.Email)
	case mongo.ErrNoDocuments:
		email := envEmail
		if email == "" {
			email = "xx.xx@monmetrics.test"
		}
		password := envPassword
		if password == "" {
			password = "TestUser123!"
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			log.Printf("Warning: could not hash test user password: %v", err)
			return
		}
		user = models.User{Email: email, PasswordHash: string(hash), FirstName: "xx", LastName: "xx",
			UserType: "free", IsActive: true, CreatedAt: now, UpdatedAt: now}
		res, err := users.InsertOne(ctx, user)
		if err != nil {
			log.Printf("Warning: could not insert test user: %v", err)
			return
		}
		user.ID = res.InsertedID.(primitive.ObjectID)
		fmt.Printf("👤 Created test user %s / %s\n", email, password)
	default:
		log.Printf("Warning: could not look up test user: %v", err)
		return
	}

	for _, coll := range []string{"portfolio_items", "favorites", "saved_charts", "linked_accounts", "marketplace_listings"} {
		if _, err := db.Collection(coll).DeleteMany(ctx, bson.M{"user_id": user.ID}); err != nil {
			log.Printf("Warning: could not clear %s for test user: %v", coll, err)
		}
	}

	cards := loadCardsByName(ctx, db, cardIDs)
	itemIDs := seedPortfolio(ctx, db, user.ID, cards, now)
	seedFavorites(ctx, db, user.ID, cardIDs, now)
	seedSavedCharts(ctx, db, user.ID, cardIDs, now)
	seedMarketplace(ctx, db, config, user.ID, cards, itemIDs, now)

	fmt.Printf("   ✅ Test user ready: %s (xx xx, %s plan)\n", user.Email, user.UserType)
}

func loadCardsByName(ctx context.Context, db *mongo.Database, ids map[string]primitive.ObjectID) map[string]*models.Card {
	out := map[string]*models.Card{}
	list := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		list = append(list, id)
	}
	cur, err := db.Collection("cards").Find(ctx, bson.M{"_id": bson.M{"$in": list}})
	if err != nil {
		log.Printf("Warning: could not load cards: %v", err)
		return out
	}
	defer cur.Close(ctx)
	var cards []models.Card
	if err := cur.All(ctx, &cards); err != nil {
		log.Printf("Warning: could not decode cards: %v", err)
		return out
	}
	for i := range cards {
		out[cards[i].Name] = &cards[i]
	}
	return out
}

// seedPortfolio inserts fixtures and returns item IDs keyed by "card|type" for listing fixtures.
func seedPortfolio(ctx context.Context, db *mongo.Database, userID primitive.ObjectID, cards map[string]*models.Card, now time.Time) map[string]primitive.ObjectID {
	ids := map[string]primitive.ObjectID{}
	docs := []interface{}{}
	for i, f := range testUserFixtures() {
		item := models.PortfolioItem{
			ID: primitive.NewObjectID(), UserID: userID, ItemType: f.itemType, Quantity: f.quantity,
			Condition: f.condition, Finish: f.finish, Language: f.language, Grading: f.grading,
			PurchaseSource: f.source, ManualValue: f.manualValue, Notes: f.notes,
			CustomName: f.customName, CustomGame: f.customGame, CustomSet: f.customSet,
			CreatedAt: now.Add(-time.Duration(i) * time.Minute), UpdatedAt: now,
		}
		purchase := now.AddDate(0, -f.monthsAgo, 0).Truncate(24 * time.Hour)
		item.PurchaseDate = &purchase

		var card *models.Card
		if f.card != "" {
			card = cards[f.card]
			if card == nil {
				log.Printf("Warning: fixture card %q not found; skipping", f.card)
				continue
			}
			item.CardID = &card.ID
		}
		item.PurchasePrice = f.purchasePrice
		if f.purchaseFactor > 0 {
			unit := valuation.Value(&item, card).UnitValue
			item.PurchasePrice = math.Round(unit*f.purchaseFactor*100) / 100
		}
		ids[f.card+"|"+f.itemType] = item.ID
		docs = append(docs, item)
	}
	if len(docs) == 0 {
		return ids
	}
	if _, err := db.Collection("portfolio_items").InsertMany(ctx, docs); err != nil {
		log.Printf("Warning: could not insert portfolio fixtures: %v", err)
		return ids
	}
	fmt.Printf("   ✅ Portfolio: %d items (raw, graded, sealed, custom)\n", len(docs))
	return ids
}

func seedFavorites(ctx context.Context, db *mongo.Database, userID primitive.ObjectID, cardIDs map[string]primitive.ObjectID, now time.Time) {
	names := []string{"Charizard ex", "Lugia VSTAR", "Sheoldred, the Apocalypse", "Ash Blossom & Joyous Spring", "Pokemon Base Set Booster Box", "Magic Beta Booster Box"}
	docs := []interface{}{}
	for i, name := range names {
		id, ok := cardIDs[name]
		if !ok {
			log.Printf("Warning: favorite card %q not found; skipping", name)
			continue
		}
		docs = append(docs, models.Favorite{UserID: userID, CardID: id, CreatedAt: now.Add(-time.Duration(i) * time.Hour)})
	}
	if len(docs) == 0 {
		return
	}
	if _, err := db.Collection("favorites").InsertMany(ctx, docs); err != nil {
		log.Printf("Warning: could not insert favorites: %v", err)
		return
	}
	fmt.Printf("   ✅ Favorites: %d\n", len(docs))
}

func emaIndicator(period int, color string) models.ChartIndicator {
	return models.ChartIndicator{Type: "ema", Parameters: map[string]interface{}{"period": period}, Color: color, Visible: true}
}

func seedSavedCharts(ctx context.Context, db *mongo.Database, userID primitive.ObjectID, cardIDs map[string]primitive.ObjectID, now time.Time) {
	fixtures := []struct {
		card, name, desc, timeRange, source string
		indicators                          []models.ChartIndicator
	}{
		{"Charizard VMAX", "Charizard VMAX baseline", "Plain 90 day view, no indicators.", "90d", "all", []models.ChartIndicator{}},
		{"Umbreon VMAX", "Umbreon EMA 20/50 crossover", "Watching for the 20 to cross the 50 on eBay sales.", "1y", "ebay",
			[]models.ChartIndicator{emaIndicator(20, "#22c55e"), emaIndicator(50, "#f59e0b")}},
		{"Black Lotus", "Black Lotus EMA stack", "Free plan maximum of three indicators.", "5y", "tcgplayer",
			[]models.ChartIndicator{emaIndicator(9, "#06b6d4"), emaIndicator(21, "#a855f7"), {Type: "ema", Parameters: map[string]interface{}{"period": 50}, Color: "#f43f5e", Visible: false}}},
	}
	docs := []interface{}{}
	for i, f := range fixtures {
		id, ok := cardIDs[f.card]
		if !ok {
			continue
		}
		ts := now.Add(-time.Duration(i) * 24 * time.Hour)
		docs = append(docs, models.SavedChart{UserID: userID, CardID: id, Name: f.name, Description: f.desc,
			Indicators: f.indicators, TimeRange: f.timeRange, Source: f.source, CreatedAt: ts, UpdatedAt: ts})
	}
	if len(docs) == 0 {
		return
	}
	if _, err := db.Collection("saved_charts").InsertMany(ctx, docs); err != nil {
		log.Printf("Warning: could not insert saved charts: %v", err)
		return
	}
	fmt.Printf("   ✅ Saved charts: %d\n", len(docs))
}

func seedMarketplace(ctx context.Context, db *mongo.Database, config *configs.Config, userID primitive.ObjectID, cards map[string]*models.Card, itemIDs map[string]primitive.ObjectID, now time.Time) {
	cipher, err := secure.NewCipher(config.TokenEncryptionKey)
	if err != nil {
		log.Printf("Warning: could not init token cipher: %v", err)
		return
	}
	access, err := cipher.Encrypt("mock-access-seeded")
	if err != nil {
		log.Printf("Warning: could not encrypt token: %v", err)
		return
	}
	refresh, err := cipher.Encrypt("mock-refresh-seeded")
	if err != nil {
		log.Printf("Warning: could not encrypt token: %v", err)
		return
	}
	// eBay linked with an expired access token to exercise refresh; TCGPlayer left unlinked.
	_, err = db.Collection("linked_accounts").InsertOne(ctx, models.LinkedAccount{
		UserID: userID, Provider: models.ProviderEbay, ExternalUserID: "ebay-seeded01", ExternalUsername: "mock_ebay_seller_xx",
		AccessTokenEnc: access, RefreshTokenEnc: refresh, Scopes: []string{"sell.inventory", "sell.account"},
		ExpiresAt: now.Add(-24 * time.Hour), Status: "active", LinkedAt: now.AddDate(0, -2, 0), UpdatedAt: now,
	})
	if err != nil {
		log.Printf("Warning: could not insert linked account: %v", err)
		return
	}

	published := now.AddDate(0, 0, -5)
	fixtures := []struct {
		key, card, provider, status, title, condition, errMsg string
		price                                                 float64
		publishedAt                                           *time.Time
		externalID                                            string
	}{
		{models.ItemTypeGradedCard, "Charizard VMAX", models.ProviderEbay, "published", "Charizard VMAX PSA 10 020/073 Champions Path VMAX Pokemon", "Graded", "", 0, &published, "EBAY-MOCK-SEEDED1"},
		{models.ItemTypeSealed, "Pokemon 151 Booster Box", models.ProviderTCGPlayer, "draft", "Pokemon 151 Booster Box Factory Sealed", "Sealed", "", 0, nil, ""},
		{models.ItemTypeRawCard, "Umbreon VMAX", models.ProviderEbay, "failed", "Umbreon VMAX 215/203 Evolving Skies VMAX Holo Pokemon LP", "Ungraded - Lightly Played", "Marketplace rejected the listing", 0, nil, ""},
	}
	docs := []interface{}{}
	for _, f := range fixtures {
		itemID, ok := itemIDs[f.card+"|"+f.key]
		card := cards[f.card]
		if !ok || card == nil {
			continue
		}
		price := card.CurrentPrice
		if f.key == models.ItemTypeGradedCard {
			price = card.CurrentPrice * valuation.GradeMultiplier("PSA", 10)
		}
		docs = append(docs, models.MarketplaceListing{
			UserID: userID, PortfolioItemID: itemID, CardID: &card.ID, Provider: f.provider,
			ExternalListingID: f.externalID, Status: f.status, Title: f.title,
			Description: "Seeded test listing.", Price: math.Round(price*100) / 100, Quantity: 1,
			Condition: f.condition, Format: "fixed_price", ShippingPrice: 4.99,
			ImageURLs: []string{card.ImageURL}, ItemSpecifics: []models.ItemSpecific{{Name: "Game", Value: card.Game}},
			ErrorMessage: f.errMsg, CreatedAt: now.AddDate(0, 0, -6), PublishedAt: f.publishedAt, UpdatedAt: now,
		})
	}
	if len(docs) == 0 {
		return
	}
	if _, err := db.Collection("marketplace_listings").InsertMany(ctx, docs); err != nil {
		log.Printf("Warning: could not insert marketplace listings: %v", err)
		return
	}
	fmt.Printf("   ✅ Marketplace: eBay linked (expired token -> refresh), TCGPlayer unlinked, %d listings\n", len(docs))
}
