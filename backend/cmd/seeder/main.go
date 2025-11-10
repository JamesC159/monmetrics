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
	featuredCollection := db.Collection("featured_content")

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

	_, err = featuredCollection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Printf("Warning: Failed to clear featured_content collection: %v", err)
	}

	// Seed cards with comprehensive sample data
	cards := []models.Card{
		// Pokemon Cards
		{
			Name:           "Charizard VMAX",
			Set:            "Champions Path",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "VMAX",
			Number:         "020/073",
			ImageURL:       "https://images.pokemontcg.io/swsh35/20_hires.png",
			Description:    "A powerful Fire-type Pokemon VMAX card from Champions Path",
			CurrentPrice:   89.99,
			AllTimeHigh:    350.00,
			AllTimeLow:     45.00,
			ATHDate:        time.Now().AddDate(0, -8, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"charizard", "vmax", "champions", "path", "fire", "pokemon"},
			Tags:           []string{"popular", "valuable", "competitive"},
			PopularityRank: 1, // Most popular
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Giratina VSTAR",
			Set:            "Lost Origin",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "VSTAR",
			Number:         "131/196",
			ImageURL:       "https://images.pokemontcg.io/swsh11/131_hires.png",
			Description:    "Legendary Dragon Pokemon VSTAR with devastating Lost Zone mechanics",
			CurrentPrice:   55.00,
			AllTimeHigh:    120.00,
			AllTimeLow:     30.00,
			ATHDate:        time.Now().AddDate(0, -6, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"giratina", "vstar", "lost", "origin", "dragon", "ghost", "pokemon"},
			Tags:           []string{"legendary", "meta", "competitive"},
			PopularityRank: 5,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Miraidon ex",
			Set:            "Scarlet & Violet",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "Double Rare",
			Number:         "080/198",
			ImageURL:       "https://images.pokemontcg.io/sv1/80_hires.png",
			Description:    "Futuristic Paradox Pokemon with Electric-type acceleration",
			CurrentPrice:   42.00,
			AllTimeHigh:    85.00,
			AllTimeLow:     25.00,
			ATHDate:        time.Now().AddDate(0, -4, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"miraidon", "ex", "scarlet", "violet", "electric", "paradox", "pokemon"},
			Tags:           []string{"modern", "competitive", "paradox"},
			PopularityRank: 6,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Koraidon ex",
			Set:            "Scarlet & Violet",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "Double Rare",
			Number:         "087/198",
			ImageURL:       "https://images.pokemontcg.io/sv1/87_hires.png",
			Description:    "Ancient Paradox Pokemon with Fighting-type power",
			CurrentPrice:   38.00,
			AllTimeHigh:    75.00,
			AllTimeLow:     22.00,
			ATHDate:        time.Now().AddDate(0, -4, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"koraidon", "ex", "scarlet", "violet", "fighting", "paradox", "pokemon"},
			Tags:           []string{"modern", "competitive", "paradox"},
			PopularityRank: 7,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Umbreon VMAX",
			Set:            "Evolving Skies",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "VMAX",
			Number:         "215/203",
			ImageURL:       "https://images.pokemontcg.io/swsh7/215_hires.png",
			Description:    "Alternate Art Secret Rare of the beloved Dark-type Eeveelution",
			CurrentPrice:   425.00,
			AllTimeHigh:    650.00,
			AllTimeLow:     280.00,
			ATHDate:        time.Now().AddDate(0, -10, 0),
			ATLDate:        time.Now().AddDate(0, -3, 0),
			SearchTerms:    []string{"umbreon", "vmax", "evolving", "skies", "dark", "eeveelution", "alt", "art", "pokemon"},
			Tags:           []string{"secret-rare", "alt-art", "highly-sought", "eeveelution"},
			PopularityRank: 8,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Mew VMAX",
			Set:            "Fusion Strike",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "VMAX",
			Number:         "269/264",
			ImageURL:       "https://images.pokemontcg.io/swsh8/269_hires.png",
			Description:    "Psychic-type legendary Pokemon with Fusion Strike synergy",
			CurrentPrice:   68.00,
			AllTimeHigh:    145.00,
			AllTimeLow:     35.00,
			ATHDate:        time.Now().AddDate(0, -9, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"mew", "vmax", "fusion", "strike", "psychic", "legendary", "pokemon"},
			Tags:           []string{"legendary", "meta", "fusion-strike"},
			PopularityRank: 9,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Mewtwo ex",
			Set:            "151",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "Double Rare",
			Number:         "150/165",
			ImageURL:       "https://images.pokemontcg.io/sv3pt5/150_hires.png",
			Description:    "The iconic Psychic-type Pokemon from the special 151 set",
			CurrentPrice:   95.00,
			AllTimeHigh:    180.00,
			AllTimeLow:     60.00,
			ATHDate:        time.Now().AddDate(0, -3, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"mewtwo", "ex", "151", "psychic", "legendary", "pokemon"},
			Tags:           []string{"legendary", "151-set", "popular"},
			PopularityRank: 10,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Rayquaza VMAX",
			Set:            "Evolving Skies",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "VMAX",
			Number:         "218/203",
			ImageURL:       "https://images.pokemontcg.io/swsh7/218_hires.png",
			Description:    "Alternate Art Secret Rare of the Sky High Dragon Pokemon",
			CurrentPrice:   380.00,
			AllTimeHigh:    550.00,
			AllTimeLow:     250.00,
			ATHDate:        time.Now().AddDate(0, -8, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"rayquaza", "vmax", "evolving", "skies", "dragon", "alt", "art", "pokemon"},
			Tags:           []string{"secret-rare", "alt-art", "legendary", "dragon"},
			PopularityRank: 11,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Leafeon VSTAR",
			Set:            "Crown Zenith",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "VSTAR",
			Number:         "015/159",
			ImageURL:       "https://images.pokemontcg.io/swsh12pt5/15_hires.png",
			Description:    "Grass-type Eeveelution with powerful healing abilities",
			CurrentPrice:   28.00,
			AllTimeHigh:    65.00,
			AllTimeLow:     18.00,
			ATHDate:        time.Now().AddDate(0, -5, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"leafeon", "vstar", "crown", "zenith", "grass", "eeveelution", "pokemon"},
			Tags:           []string{"eeveelution", "grass", "healing"},
			PopularityRank: 12,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Charizard ex",
			Set:            "Obsidian Flames",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "Double Rare",
			Number:         "125/197",
			ImageURL:       "https://images.pokemontcg.io/sv3/125_hires.png",
			Description:    "Modern Charizard ex from the Scarlet & Violet era",
			CurrentPrice:   115.00,
			AllTimeHigh:    220.00,
			AllTimeLow:     75.00,
			ATHDate:        time.Now().AddDate(0, -2, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"charizard", "ex", "obsidian", "flames", "fire", "dragon", "pokemon"},
			Tags:           []string{"charizard", "modern", "popular"},
			PopularityRank: 13,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Roaring Moon ex",
			Set:            "Paradox Rift",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "Double Rare",
			Number:         "124/182",
			ImageURL:       "https://images.pokemontcg.io/sv4/124_hires.png",
			Description:    "Ancient Paradox Pokemon with Dark-type aggression",
			CurrentPrice:   32.00,
			AllTimeHigh:    68.00,
			AllTimeLow:     20.00,
			ATHDate:        time.Now().AddDate(0, -2, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"roaring", "moon", "ex", "paradox", "rift", "dark", "dragon", "pokemon"},
			Tags:           []string{"paradox", "modern", "dark"},
			PopularityRank: 14,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pikachu VMAX",
			Set:            "Vivid Voltage",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "VMAX",
			Number:         "044/185",
			ImageURL:       "https://images.pokemontcg.io/swsh4/44_hires.png",
			Description:    "Electric-type Pokemon VMAX card featuring the iconic Pikachu",
			CurrentPrice:   25.99,
			AllTimeHigh:    89.99,
			AllTimeLow:     15.00,
			ATHDate:        time.Now().AddDate(0, -6, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"pikachu", "vmax", "vivid", "voltage", "electric", "pokemon"},
			Tags:           []string{"iconic", "electric", "popular"},
			PopularityRank: 2,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Lugia VSTAR",
			Set:            "Silver Tempest",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "VSTAR",
			Number:         "139/195",
			ImageURL:       "https://images.pokemontcg.io/swsh12/139_hires.png",
			Description:    "Psychic-type legendary Pokemon VSTAR with incredible power",
			CurrentPrice:   45.50,
			AllTimeHigh:    125.00,
			AllTimeLow:     25.00,
			ATHDate:        time.Now().AddDate(0, -4, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"lugia", "vstar", "silver", "tempest", "psychic", "legendary", "pokemon"},
			Tags:           []string{"legendary", "powerful", "recent"},
			PopularityRank: 3,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Arceus VSTAR",
			Set:            "Brilliant Stars",
			Game:           "Pokemon",
			Category:       "card",
			Rarity:         "VSTAR",
			Number:         "123/172",
			ImageURL:       "https://images.pokemontcg.io/swsh9/123_hires.png",
			Description:    "The Alpha Pokemon with ultimate versatility",
			CurrentPrice:   67.50,
			AllTimeHigh:    150.00,
			AllTimeLow:     35.00,
			ATHDate:        time.Now().AddDate(0, -5, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"arceus", "vstar", "brilliant", "stars", "colorless", "alpha", "pokemon"},
			Tags:           []string{"legendary", "versatile", "meta"},
			PopularityRank: 4,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},

		// Magic The Gathering Cards
		{
			Name:           "Black Lotus",
			Set:            "Alpha",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Rare",
			Number:         "",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=3&type=card",
			Description:    "The most iconic and valuable Magic card ever printed",
			CurrentPrice:   65000.00,
			AllTimeHigh:    87000.00,
			AllTimeLow:     45000.00,
			ATHDate:        time.Now().AddDate(-1, 0, 0),
			ATLDate:        time.Now().AddDate(-2, 0, 0),
			SearchTerms:    []string{"black", "lotus", "alpha", "power", "nine", "vintage", "magic"},
			Tags:           []string{"power-nine", "vintage", "investment", "iconic"},
			PopularityRank: 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "The One Ring",
			Set:            "The Lord of the Rings: Tales of Middle-earth",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Mythic Rare",
			Number:         "246/281",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=633037&type=card",
			Description:    "Powerful card protection artifact from the LOTR crossover set",
			CurrentPrice:   95.00,
			AllTimeHigh:    120.00,
			AllTimeLow:     65.00,
			ATHDate:        time.Now().AddDate(0, -3, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"the", "one", "ring", "lotr", "lord", "rings", "artifact", "protection", "magic"},
			Tags:           []string{"modern", "commander", "artifact", "lotr"},
			PopularityRank: 5,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Ragavan, Nimble Pilferer",
			Set:            "Modern Horizons 2",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Mythic Rare",
			Number:         "138/303",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=522286&type=card",
			Description:    "The premier red one-drop creature dominating Modern and Legacy",
			CurrentPrice:   78.00,
			AllTimeHigh:    95.00,
			AllTimeLow:     55.00,
			ATHDate:        time.Now().AddDate(0, -6, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"ragavan", "nimble", "pilferer", "modern", "horizons", "red", "creature", "magic"},
			Tags:           []string{"modern-staple", "legacy", "red", "competitive"},
			PopularityRank: 6,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Force of Negation",
			Set:            "Modern Horizons",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Rare",
			Number:         "052/254",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=464039&type=card",
			Description:    "Free counterspell for noncreature spells, Modern and Legacy staple",
			CurrentPrice:   52.00,
			AllTimeHigh:    75.00,
			AllTimeLow:     35.00,
			ATHDate:        time.Now().AddDate(0, -8, 0),
			ATLDate:        time.Now().AddDate(0, -3, 0),
			SearchTerms:    []string{"force", "negation", "modern", "horizons", "blue", "counterspell", "magic"},
			Tags:           []string{"modern-staple", "blue", "counterspell", "competitive"},
			PopularityRank: 7,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Dockside Extortionist",
			Set:            "Commander 2019",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Rare",
			Number:         "024/302",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=470659&type=card",
			Description:    "Explosive mana-generating goblin, Commander format all-star",
			CurrentPrice:   68.00,
			AllTimeHigh:    95.00,
			AllTimeLow:     45.00,
			ATHDate:        time.Now().AddDate(0, -5, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"dockside", "extortionist", "commander", "red", "treasure", "goblin", "magic"},
			Tags:           []string{"commander-staple", "red", "mana", "cedh"},
			PopularityRank: 8,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Wrenn and Six",
			Set:            "Modern Horizons",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Mythic Rare",
			Number:         "217/254",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=464088&type=card",
			Description:    "Powerful two-mana planeswalker dominating multiple formats",
			CurrentPrice:   62.00,
			AllTimeHigh:    88.00,
			AllTimeLow:     42.00,
			ATHDate:        time.Now().AddDate(0, -7, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"wrenn", "six", "modern", "horizons", "planeswalker", "lands", "magic"},
			Tags:           []string{"modern-staple", "planeswalker", "lands-matter"},
			PopularityRank: 9,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Orcish Bowmasters",
			Set:            "The Lord of the Rings: Tales of Middle-earth",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Rare",
			Number:         "103/281",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=632992&type=card",
			Description:    "Format-warping creature punishing card draw across all formats",
			CurrentPrice:   58.00,
			AllTimeHigh:    82.00,
			AllTimeLow:     38.00,
			ATHDate:        time.Now().AddDate(0, -4, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"orcish", "bowmasters", "lotr", "lord", "rings", "creature", "draw", "magic"},
			Tags:           []string{"modern-staple", "legacy", "lotr", "competitive"},
			PopularityRank: 10,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Sheoldred, the Apocalypse",
			Set:            "Dominaria United",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Mythic Rare",
			Number:         "107/281",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=579033&type=card",
			Description:    "Powerful Phyrexian Praetor dominating Standard and beyond",
			CurrentPrice:   42.00,
			AllTimeHigh:    68.00,
			AllTimeLow:     28.00,
			ATHDate:        time.Now().AddDate(0, -5, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"sheoldred", "apocalypse", "dominaria", "phyrexian", "black", "creature", "magic"},
			Tags:           []string{"standard", "modern", "phyrexian", "competitive"},
			PopularityRank: 11,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Jeweled Lotus",
			Set:            "Commander Legends",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Mythic Rare",
			Number:         "319/361",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=497694&type=card",
			Description:    "Commander-only Black Lotus variant for explosive starts",
			CurrentPrice:   85.00,
			AllTimeHigh:    135.00,
			AllTimeLow:     55.00,
			ATHDate:        time.Now().AddDate(0, -8, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"jeweled", "lotus", "commander", "legends", "mana", "artifact", "magic"},
			Tags:           []string{"commander-staple", "mana", "artifact", "cedh"},
			PopularityRank: 12,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Fierce Guardianship",
			Set:            "Commander 2020",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Rare",
			Number:         "035/322",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=479702&type=card",
			Description:    "Free counterspell in Commander with your commander out",
			CurrentPrice:   48.00,
			AllTimeHigh:    72.00,
			AllTimeLow:     32.00,
			ATHDate:        time.Now().AddDate(0, -6, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"fierce", "guardianship", "commander", "blue", "counterspell", "free", "magic"},
			Tags:           []string{"commander-staple", "blue", "counterspell"},
			PopularityRank: 13,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Deflecting Swat",
			Set:            "Commander 2020",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Rare",
			Number:         "050/322",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=479714&type=card",
			Description:    "Free redirect spell protecting your board in Commander",
			CurrentPrice:   55.00,
			AllTimeHigh:    82.00,
			AllTimeLow:     38.00,
			ATHDate:        time.Now().AddDate(0, -7, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"deflecting", "swat", "commander", "red", "redirect", "free", "magic"},
			Tags:           []string{"commander-staple", "red", "protection"},
			PopularityRank: 14,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Tarmogoyf",
			Set:            "Future Sight",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Rare",
			Number:         "153/180",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=136142&type=card",
			Description:    "A powerful green creature that defines Modern format",
			CurrentPrice:   89.99,
			AllTimeHigh:    199.99,
			AllTimeLow:     45.00,
			ATHDate:        time.Now().AddDate(0, -18, 0),
			ATLDate:        time.Now().AddDate(0, -3, 0),
			SearchTerms:    []string{"tarmogoyf", "future", "sight", "green", "creature", "modern", "magic"},
			Tags:           []string{"modern-staple", "competitive", "green"},
			PopularityRank: 2,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Lightning Bolt",
			Set:            "Alpha",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Common",
			Number:         "",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=209&type=card",
			Description:    "Deal 3 damage to any target - a classic red instant",
			CurrentPrice:   125.00,
			AllTimeHigh:    250.00,
			AllTimeLow:     85.00,
			ATHDate:        time.Now().AddDate(0, -12, 0),
			ATLDate:        time.Now().AddDate(0, -6, 0),
			SearchTerms:    []string{"lightning", "bolt", "alpha", "red", "instant", "damage", "magic"},
			Tags:           []string{"classic", "red", "vintage"},
			PopularityRank: 3,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Mox Ruby",
			Set:            "Alpha",
			Game:           "Magic The Gathering",
			Category:       "card",
			Rarity:         "Rare",
			Number:         "",
			ImageURL:       "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=263&type=card",
			Description:    "Part of the iconic Power Nine, provides red mana",
			CurrentPrice:   8500.00,
			AllTimeHigh:    12000.00,
			AllTimeLow:     5500.00,
			ATHDate:        time.Now().AddDate(-1, -2, 0),
			ATLDate:        time.Now().AddDate(-2, -6, 0),
			SearchTerms:    []string{"mox", "ruby", "alpha", "power", "nine", "red", "mana", "magic"},
			Tags:           []string{"power-nine", "vintage", "mana", "red"},
			PopularityRank: 4,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},

		// Yu-Gi-Oh Cards
		{
			Name:           "Blue-Eyes White Dragon",
			Set:            "Legend of Blue Eyes White Dragon",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Ultra Rare",
			Number:         "LOB-001",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/89631139.jpg",
			Description:    "This legendary dragon is a powerful engine of destruction",
			CurrentPrice:   2500.00,
			AllTimeHigh:    5500.00,
			AllTimeLow:     1200.00,
			ATHDate:        time.Now().AddDate(-1, -6, 0),
			ATLDate:        time.Now().AddDate(0, -4, 0),
			SearchTerms:    []string{"blue", "eyes", "white", "dragon", "legend", "lob", "kaiba", "yugioh"},
			Tags:           []string{"iconic", "dragon", "kaiba", "nostalgic"},
			PopularityRank: 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Kashtira Fenrir",
			Set:            "Photon Hypernova",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Ultra Rare",
			Number:         "PHHY-EN016",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/98701400.jpg",
			Description:    "Powerful Kashtira monster dominating the current meta",
			CurrentPrice:   85.00,
			AllTimeHigh:    165.00,
			AllTimeLow:     55.00,
			ATHDate:        time.Now().AddDate(0, -4, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"kashtira", "fenrir", "photon", "hypernova", "meta", "competitive", "yugioh"},
			Tags:           []string{"meta", "competitive", "modern", "kashtira"},
			PopularityRank: 4,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Diabellstar the Black Witch",
			Set:            "Legacy of Destruction",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Ultra Rare",
			Number:         "LEDE-EN001",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/19748583.jpg",
			Description:    "Core engine piece for Fiendsmith strategies",
			CurrentPrice:   125.00,
			AllTimeHigh:    220.00,
			AllTimeLow:     85.00,
			ATHDate:        time.Now().AddDate(0, -2, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"diabellstar", "black", "witch", "legacy", "destruction", "fiendsmith", "yugioh"},
			Tags:           []string{"meta", "competitive", "modern", "engine"},
			PopularityRank: 5,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Snake-Eye Ash",
			Set:            "Age of Overlord",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Ultra Rare",
			Number:         "AGOV-EN015",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/46710683.jpg",
			Description:    "Key Snake-Eye combo piece dominating tier 1",
			CurrentPrice:   95.00,
			AllTimeHigh:    175.00,
			AllTimeLow:     62.00,
			ATHDate:        time.Now().AddDate(0, -3, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"snake", "eye", "ash", "age", "overlord", "fire", "meta", "yugioh"},
			Tags:           []string{"meta", "competitive", "modern", "snake-eye"},
			PopularityRank: 6,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Red-Eyes Black Dragon",
			Set:            "Legend of Blue Eyes White Dragon",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Ultra Rare",
			Number:         "LOB-070",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/74677422.jpg",
			Description:    "Joey's signature dragon from the original set",
			CurrentPrice:   1200.00,
			AllTimeHigh:    2800.00,
			AllTimeLow:     750.00,
			ATHDate:        time.Now().AddDate(-1, -2, 0),
			ATLDate:        time.Now().AddDate(0, -5, 0),
			SearchTerms:    []string{"red", "eyes", "black", "dragon", "legend", "lob", "joey", "yugioh"},
			Tags:           []string{"iconic", "dragon", "joey", "nostalgic"},
			PopularityRank: 7,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Slifer the Sky Dragon",
			Set:            "Battle Pack 2: War of the Giants",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Secret Rare",
			Number:         "BP02-EN126",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/10000020.jpg",
			Description:    "Egyptian God Card with unlimited ATK potential",
			CurrentPrice:   75.00,
			AllTimeHigh:    145.00,
			AllTimeLow:     48.00,
			ATHDate:        time.Now().AddDate(0, -8, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"slifer", "sky", "dragon", "egyptian", "god", "divine", "yugioh"},
			Tags:           []string{"god-card", "iconic", "divine-beast"},
			PopularityRank: 8,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Obelisk the Tormentor",
			Set:            "Battle Pack 2: War of the Giants",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Secret Rare",
			Number:         "BP02-EN127",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/10000000.jpg",
			Description:    "Egyptian God Card with devastating offensive power",
			CurrentPrice:   72.00,
			AllTimeHigh:    138.00,
			AllTimeLow:     45.00,
			ATHDate:        time.Now().AddDate(0, -8, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"obelisk", "tormentor", "egyptian", "god", "divine", "yugioh"},
			Tags:           []string{"god-card", "iconic", "divine-beast"},
			PopularityRank: 9,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "The Winged Dragon of Ra",
			Set:            "Battle Pack 2: War of the Giants",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Secret Rare",
			Number:         "BP02-EN128",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/10000010.jpg",
			Description:    "Egyptian God Card with flexible power levels",
			CurrentPrice:   68.00,
			AllTimeHigh:    132.00,
			AllTimeLow:     42.00,
			ATHDate:        time.Now().AddDate(0, -8, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"winged", "dragon", "ra", "egyptian", "god", "divine", "yugioh"},
			Tags:           []string{"god-card", "iconic", "divine-beast"},
			PopularityRank: 10,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pot of Greed",
			Set:            "Legend of Blue Eyes White Dragon",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Super Rare",
			Number:         "LOB-119",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/55144522.jpg",
			Description:    "Iconic banned card that draws 2 cards",
			CurrentPrice:   125.00,
			AllTimeHigh:    250.00,
			AllTimeLow:     85.00,
			ATHDate:        time.Now().AddDate(-1, 0, 0),
			ATLDate:        time.Now().AddDate(0, -6, 0),
			SearchTerms:    []string{"pot", "greed", "legend", "lob", "spell", "draw", "banned", "yugioh"},
			Tags:           []string{"iconic", "banned", "spell", "nostalgic"},
			PopularityRank: 11,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Ash Blossom & Joyous Spring",
			Set:            "Maximum Crisis",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Secret Rare",
			Number:         "MACR-EN036",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/14558127.jpg",
			Description:    "Premier hand trap staple in every competitive deck",
			CurrentPrice:   45.00,
			AllTimeHigh:    88.00,
			AllTimeLow:     28.00,
			ATHDate:        time.Now().AddDate(0, -12, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"ash", "blossom", "joyous", "spring", "hand", "trap", "staple", "yugioh"},
			Tags:           []string{"staple", "hand-trap", "competitive"},
			PopularityRank: 12,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Infinite Impermanence",
			Set:            "Flames of Destruction",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Secret Rare",
			Number:         "FLOD-EN065",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/10045474.jpg",
			Description:    "Versatile trap negation staple in every format",
			CurrentPrice:   42.00,
			AllTimeHigh:    75.00,
			AllTimeLow:     28.00,
			ATHDate:        time.Now().AddDate(0, -10, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"infinite", "impermanence", "flames", "destruction", "trap", "negation", "yugioh"},
			Tags:           []string{"staple", "trap", "competitive"},
			PopularityRank: 13,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Dark Magician",
			Set:            "Legend of Blue Eyes White Dragon",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Ultra Rare",
			Number:         "LOB-005",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/46986414.jpg",
			Description:    "The ultimate wizard in terms of attack and defense",
			CurrentPrice:   1800.00,
			AllTimeHigh:    3200.00,
			AllTimeLow:     900.00,
			ATHDate:        time.Now().AddDate(-1, -3, 0),
			ATLDate:        time.Now().AddDate(0, -5, 0),
			SearchTerms:    []string{"dark", "magician", "legend", "lob", "spellcaster", "yugi", "yugioh"},
			Tags:           []string{"iconic", "spellcaster", "yugi", "nostalgic"},
			PopularityRank: 2,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Exodia the Forbidden One",
			Set:            "Legend of Blue Eyes White Dragon",
			Game:           "Yu-Gi-Oh",
			Category:       "card",
			Rarity:         "Ultra Rare",
			Number:         "LOB-124",
			ImageURL:       "https://images.ygoprodeck.com/images/cards/33396948.jpg",
			Description:    "If you have all 5 pieces, you automatically win the duel",
			CurrentPrice:   450.00,
			AllTimeHigh:    850.00,
			AllTimeLow:     250.00,
			ATHDate:        time.Now().AddDate(0, -9, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"exodia", "forbidden", "one", "legend", "lob", "win", "condition", "yugioh"},
			Tags:           []string{"win-condition", "rare", "nostalgic"},
			PopularityRank: 3,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},

		// Sealed Products - Pokemon
		{
			Name:           "Pokemon Base Set Booster Box",
			Set:            "Base Set",
			Game:           "Pokemon",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://52f4e29a8321344e30ae-0f55c9129972ac85d6b1f4e703468e6b.ssl.cf2.rackcdn.com/products/pictures/1085368.jpg",
			Description:    "Factory sealed Pokemon Base Set booster box - 36 packs",
			CurrentPrice:   45000.00,
			AllTimeHigh:    75000.00,
			AllTimeLow:     25000.00,
			ATHDate:        time.Now().AddDate(-1, 0, 0),
			ATLDate:        time.Now().AddDate(-3, 0, 0),
			SearchTerms:    []string{"pokemon", "base", "set", "booster", "box", "sealed", "vintage"},
			Tags:           []string{"sealed", "investment", "vintage", "rare"},
			PopularityRank: 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pokemon Jungle Booster Box",
			Set:            "Jungle",
			Game:           "Pokemon",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://52f4e29a8321344e30ae-0f55c9129972ac85d6b1f4e703468e6b.ssl.cf2.rackcdn.com/products/pictures/147884.jpg",
			Description:    "First edition Jungle booster box - 36 packs",
			CurrentPrice:   18500.00,
			AllTimeHigh:    28000.00,
			AllTimeLow:     12000.00,
			ATHDate:        time.Now().AddDate(-1, -2, 0),
			ATLDate:        time.Now().AddDate(-2, -6, 0),
			SearchTerms:    []string{"pokemon", "jungle", "booster", "box", "sealed", "vintage", "first", "edition"},
			Tags:           []string{"sealed", "investment", "vintage", "first-edition"},
			PopularityRank: 2,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pokemon Fossil Booster Box",
			Set:            "Fossil",
			Game:           "Pokemon",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://52f4e29a8321344e30ae-0f55c9129972ac85d6b1f4e703468e6b.ssl.cf2.rackcdn.com/products/pictures/147885.jpg",
			Description:    "First edition Fossil booster box - 36 packs",
			CurrentPrice:   16500.00,
			AllTimeHigh:    25000.00,
			AllTimeLow:     11000.00,
			ATHDate:        time.Now().AddDate(-1, -3, 0),
			ATLDate:        time.Now().AddDate(-2, -4, 0),
			SearchTerms:    []string{"pokemon", "fossil", "booster", "box", "sealed", "vintage", "first", "edition"},
			Tags:           []string{"sealed", "investment", "vintage", "first-edition"},
			PopularityRank: 3,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pokemon Paradox Rift Booster Box",
			Set:            "Paradox Rift",
			Game:           "Pokemon",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://images.pokemontcg.io/sv4/logo.png",
			Description:    "Scarlet & Violet Paradox Rift booster box - 36 packs",
			CurrentPrice:   125.00,
			AllTimeHigh:    165.00,
			AllTimeLow:     95.00,
			ATHDate:        time.Now().AddDate(0, -2, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"pokemon", "paradox", "rift", "booster", "box", "sealed", "scarlet", "violet"},
			Tags:           []string{"sealed", "modern", "scarlet-violet"},
			PopularityRank: 4,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pokemon Obsidian Flames Booster Box",
			Set:            "Obsidian Flames",
			Game:           "Pokemon",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://images.pokemontcg.io/sv3/logo.png",
			Description:    "Scarlet & Violet Obsidian Flames booster box - 36 packs",
			CurrentPrice:   110.00,
			AllTimeHigh:    145.00,
			AllTimeLow:     88.00,
			ATHDate:        time.Now().AddDate(0, -3, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"pokemon", "obsidian", "flames", "booster", "box", "sealed", "scarlet", "violet"},
			Tags:           []string{"sealed", "modern", "scarlet-violet"},
			PopularityRank: 5,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pokemon 151 Booster Box",
			Set:            "151",
			Game:           "Pokemon",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://images.pokemontcg.io/sv3pt5/logo.png",
			Description:    "Special set featuring all original 151 Pokemon - 36 packs",
			CurrentPrice:   175.00,
			AllTimeHigh:    285.00,
			AllTimeLow:     135.00,
			ATHDate:        time.Now().AddDate(0, -3, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"pokemon", "151", "booster", "box", "sealed", "special", "kanto"},
			Tags:           []string{"sealed", "modern", "special-set", "nostalgic"},
			PopularityRank: 6,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pokemon Paldea Evolved Booster Box",
			Set:            "Paldea Evolved",
			Game:           "Pokemon",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://images.pokemontcg.io/sv2/logo.png",
			Description:    "Scarlet & Violet Paldea Evolved booster box - 36 packs",
			CurrentPrice:   105.00,
			AllTimeHigh:    138.00,
			AllTimeLow:     85.00,
			ATHDate:        time.Now().AddDate(0, -4, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"pokemon", "paldea", "evolved", "booster", "box", "sealed", "scarlet", "violet"},
			Tags:           []string{"sealed", "modern", "scarlet-violet"},
			PopularityRank: 7,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pokemon Crown Zenith Elite Trainer Box",
			Set:            "Crown Zenith",
			Game:           "Pokemon",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://images.pokemontcg.io/swsh12pt5/logo.png",
			Description:    "Crown Zenith Elite Trainer Box with 10 packs and accessories",
			CurrentPrice:   68.00,
			AllTimeHigh:    95.00,
			AllTimeLow:     52.00,
			ATHDate:        time.Now().AddDate(0, -5, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"pokemon", "crown", "zenith", "elite", "trainer", "box", "etb", "sealed"},
			Tags:           []string{"sealed", "etb", "modern"},
			PopularityRank: 8,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pokemon Evolving Skies Booster Box",
			Set:            "Evolving Skies",
			Game:           "Pokemon",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://images.pokemontcg.io/swsh7/logo.png",
			Description:    "Highly sought Evolving Skies booster box with Eeveelutions - 36 packs",
			CurrentPrice:   245.00,
			AllTimeHigh:    385.00,
			AllTimeLow:     165.00,
			ATHDate:        time.Now().AddDate(0, -8, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"pokemon", "evolving", "skies", "booster", "box", "sealed", "eeveelution"},
			Tags:           []string{"sealed", "investment", "eeveelution", "highly-sought"},
			PopularityRank: 9,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Pokemon Team Rocket Booster Box",
			Set:            "Team Rocket",
			Game:           "Pokemon",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://52f4e29a8321344e30ae-0f55c9129972ac85d6b1f4e703468e6b.ssl.cf2.rackcdn.com/products/pictures/147886.jpg",
			Description:    "First edition Team Rocket booster box - 36 packs",
			CurrentPrice:   14500.00,
			AllTimeHigh:    22000.00,
			AllTimeLow:     9500.00,
			ATHDate:        time.Now().AddDate(-1, -4, 0),
			ATLDate:        time.Now().AddDate(-2, -2, 0),
			SearchTerms:    []string{"pokemon", "team", "rocket", "booster", "box", "sealed", "vintage", "first", "edition"},
			Tags:           []string{"sealed", "investment", "vintage", "first-edition"},
			PopularityRank: 10,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},

		// Sealed Products - Magic The Gathering
		// Sealed Products - Magic The Gathering
		{
			Name:           "Magic Alpha Starter Deck",
			Set:            "Alpha",
			Game:           "Magic The Gathering",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://crystal-cdn4.crystalcommerce.com/photos/6213849/large/en_alpha_starterdecktype1.jpg",
			Description:    "Factory sealed Magic Alpha starter deck - extremely rare",
			CurrentPrice:   125000.00,
			AllTimeHigh:    200000.00,
			AllTimeLow:     85000.00,
			ATHDate:        time.Now().AddDate(-2, 0, 0),
			ATLDate:        time.Now().AddDate(-4, 0, 0),
			SearchTerms:    []string{"magic", "alpha", "starter", "deck", "sealed", "vintage", "93"},
			Tags:           []string{"sealed", "alpha", "investment", "museum"},
			PopularityRank: 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Magic Beta Booster Box",
			Set:            "Beta",
			Game:           "Magic The Gathering",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://crystal-cdn4.crystalcommerce.com/photos/6213850/large/en_beta_boosterbox.jpg",
			Description:    "Sealed Beta booster box - 60 packs of pure history",
			CurrentPrice:   185000.00,
			AllTimeHigh:    280000.00,
			AllTimeLow:     125000.00,
			ATHDate:        time.Now().AddDate(-2, -3, 0),
			ATLDate:        time.Now().AddDate(-4, -6, 0),
			SearchTerms:    []string{"magic", "beta", "booster", "box", "sealed", "vintage", "93"},
			Tags:           []string{"sealed", "beta", "investment", "museum"},
			PopularityRank: 2,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Magic Revised Booster Box",
			Set:            "Revised Edition",
			Game:           "Magic The Gathering",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://crystal-cdn4.crystalcommerce.com/photos/6213852/large/en_revised_boosterbox.jpg",
			Description:    "Sealed Revised booster box - 60 packs from 1994",
			CurrentPrice:   8500.00,
			AllTimeHigh:    14000.00,
			AllTimeLow:     5500.00,
			ATHDate:        time.Now().AddDate(-1, -8, 0),
			ATLDate:        time.Now().AddDate(-3, -2, 0),
			SearchTerms:    []string{"magic", "revised", "booster", "box", "sealed", "vintage", "94"},
			Tags:           []string{"sealed", "revised", "vintage", "investment"},
			PopularityRank: 3,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Wilds of Eldraine Collector Booster Box",
			Set:            "Wilds of Eldraine",
			Game:           "Magic The Gathering",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://product-images.tcgplayer.com/fit-in/400x558/509933.jpg",
			Description:    "Collector booster box with premium treatments - 12 packs",
			CurrentPrice:   265.00,
			AllTimeHigh:    320.00,
			AllTimeLow:     215.00,
			ATHDate:        time.Now().AddDate(0, -3, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"magic", "wilds", "eldraine", "collector", "booster", "box", "sealed", "modern"},
			Tags:           []string{"sealed", "modern", "collector", "premium"},
			PopularityRank: 4,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Lost Caverns of Ixalan Set Booster Box",
			Set:            "Lost Caverns of Ixalan",
			Game:           "Magic The Gathering",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://product-images.tcgplayer.com/fit-in/400x558/521412.jpg",
			Description:    "Set booster box featuring dinosaurs and caves - 30 packs",
			CurrentPrice:   115.00,
			AllTimeHigh:    145.00,
			AllTimeLow:     95.00,
			ATHDate:        time.Now().AddDate(0, -2, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"magic", "lost", "caverns", "ixalan", "set", "booster", "box", "sealed", "modern"},
			Tags:           []string{"sealed", "modern", "set-booster"},
			PopularityRank: 5,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Murders at Karlov Manor Play Booster Box",
			Set:            "Murders at Karlov Manor",
			Game:           "Magic The Gathering",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://product-images.tcgplayer.com/fit-in/400x558/528765.jpg",
			Description:    "Mystery-themed play booster box - 36 packs",
			CurrentPrice:   125.00,
			AllTimeHigh:    158.00,
			AllTimeLow:     105.00,
			ATHDate:        time.Now().AddDate(0, -2, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"magic", "murders", "karlov", "manor", "play", "booster", "box", "sealed", "modern"},
			Tags:           []string{"sealed", "modern", "play-booster"},
			PopularityRank: 6,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "The Lord of the Rings Set Booster Box",
			Set:            "The Lord of the Rings: Tales of Middle-earth",
			Game:           "Magic The Gathering",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://product-images.tcgplayer.com/fit-in/400x558/506321.jpg",
			Description:    "LOTR crossover set booster box - 30 packs",
			CurrentPrice:   185.00,
			AllTimeHigh:    265.00,
			AllTimeLow:     145.00,
			ATHDate:        time.Now().AddDate(0, -4, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"magic", "lord", "rings", "lotr", "middle", "earth", "set", "booster", "box", "sealed"},
			Tags:           []string{"sealed", "lotr", "crossover", "special"},
			PopularityRank: 7,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Commander Legends: Battle for Baldur's Gate Draft Box",
			Set:            "Commander Legends: Battle for Baldur's Gate",
			Game:           "Magic The Gathering",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://product-images.tcgplayer.com/fit-in/400x558/274987.jpg",
			Description:    "Commander draft booster box - 24 packs",
			CurrentPrice:   145.00,
			AllTimeHigh:    195.00,
			AllTimeLow:     115.00,
			ATHDate:        time.Now().AddDate(0, -6, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"magic", "commander", "legends", "baldurs", "gate", "draft", "booster", "box", "sealed"},
			Tags:           []string{"sealed", "commander", "draft", "dnd"},
			PopularityRank: 8,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Double Masters 2022 Draft Booster Box",
			Set:            "Double Masters 2022",
			Game:           "Magic The Gathering",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://product-images.tcgplayer.com/fit-in/400x558/280452.jpg",
			Description:    "Premium reprint set with double rares - 24 packs",
			CurrentPrice:   285.00,
			AllTimeHigh:    385.00,
			AllTimeLow:     225.00,
			ATHDate:        time.Now().AddDate(0, -8, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"magic", "double", "masters", "2022", "draft", "booster", "box", "sealed", "reprint"},
			Tags:           []string{"sealed", "masters", "reprint", "premium"},
			PopularityRank: 9,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Modern Horizons 3 Play Booster Box",
			Set:            "Modern Horizons 3",
			Game:           "Magic The Gathering",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://product-images.tcgplayer.com/fit-in/400x558/541238.jpg",
			Description:    "Latest Modern Horizons set with powerful new cards - 36 packs",
			CurrentPrice:   195.00,
			AllTimeHigh:    245.00,
			AllTimeLow:     165.00,
			ATHDate:        time.Now().AddDate(0, -1, 0),
			ATLDate:        time.Now().AddDate(0, 0, -15),
			SearchTerms:    []string{"magic", "modern", "horizons", "3", "play", "booster", "box", "sealed"},
			Tags:           []string{"sealed", "modern", "horizons", "new"},
			PopularityRank: 10,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},

		// Sealed Products - Yu-Gi-Oh
		// Sealed Products - Yu-Gi-Oh
		{
			Name:           "Yu-Gi-Oh LOB Booster Box",
			Set:            "Legend of Blue Eyes White Dragon",
			Game:           "Yu-Gi-Oh",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://images.ygoprodeck.com/pics_artgame/55210709.jpg",
			Description:    "First edition LOB booster box - 24 packs",
			CurrentPrice:   15000.00,
			AllTimeHigh:    25000.00,
			AllTimeLow:     8500.00,
			ATHDate:        time.Now().AddDate(-1, -4, 0),
			ATLDate:        time.Now().AddDate(-2, -8, 0),
			SearchTerms:    []string{"yugioh", "lob", "legend", "booster", "box", "sealed", "first", "edition"},
			Tags:           []string{"sealed", "vintage", "first-edition", "investment"},
			PopularityRank: 1,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Yu-Gi-Oh Metal Raiders Booster Box",
			Set:            "Metal Raiders",
			Game:           "Yu-Gi-Oh",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://ms.yugipedia.com//3/32/MRD-BoosterEN.png",
			Description:    "First edition Metal Raiders booster box - 24 packs",
			CurrentPrice:   12000.00,
			AllTimeHigh:    18500.00,
			AllTimeLow:     7500.00,
			ATHDate:        time.Now().AddDate(-1, -6, 0),
			ATLDate:        time.Now().AddDate(-2, -4, 0),
			SearchTerms:    []string{"yugioh", "metal", "raiders", "booster", "box", "sealed", "first", "edition", "vintage"},
			Tags:           []string{"sealed", "vintage", "first-edition", "investment"},
			PopularityRank: 2,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Yu-Gi-Oh Pharaoh's Servant Booster Box",
			Set:            "Pharaoh's Servant",
			Game:           "Yu-Gi-Oh",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://ms.yugipedia.com//5/5a/PSV-BoosterEN.png",
			Description:    "First edition Pharaoh's Servant booster box - 24 packs",
			CurrentPrice:   9500.00,
			AllTimeHigh:    15000.00,
			AllTimeLow:     6000.00,
			ATHDate:        time.Now().AddDate(-1, -5, 0),
			ATLDate:        time.Now().AddDate(-2, -6, 0),
			SearchTerms:    []string{"yugioh", "pharaoh", "servant", "booster", "box", "sealed", "first", "edition", "vintage"},
			Tags:           []string{"sealed", "vintage", "first-edition", "investment"},
			PopularityRank: 3,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Yu-Gi-Oh Duelist Nexus Booster Box",
			Set:            "Duelist Nexus",
			Game:           "Yu-Gi-Oh",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://ms.yugipedia.com//e/e3/DUPO-BoosterEN.png",
			Description:    "Modern booster box featuring Snake-Eye archetype - 24 packs",
			CurrentPrice:   95.00,
			AllTimeHigh:    135.00,
			AllTimeLow:     75.00,
			ATHDate:        time.Now().AddDate(0, -3, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"yugioh", "duelist", "nexus", "booster", "box", "sealed", "modern", "snake-eye"},
			Tags:           []string{"sealed", "modern", "meta"},
			PopularityRank: 4,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Yu-Gi-Oh Age of Overlord Booster Box",
			Set:            "Age of Overlord",
			Game:           "Yu-Gi-Oh",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://ms.yugipedia.com//thumb/c/c6/AGOV-BoosterEN.png/300px-AGOV-BoosterEN.png",
			Description:    "Core set with powerful meta cards - 24 packs",
			CurrentPrice:   88.00,
			AllTimeHigh:    125.00,
			AllTimeLow:     68.00,
			ATHDate:        time.Now().AddDate(0, -3, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"yugioh", "age", "overlord", "booster", "box", "sealed", "modern", "meta"},
			Tags:           []string{"sealed", "modern", "meta"},
			PopularityRank: 5,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Yu-Gi-Oh Maze of Millenia Booster Box",
			Set:            "Maze of Millenia",
			Game:           "Yu-Gi-Oh",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://ms.yugipedia.com//thumb/7/7e/MAZE-BoosterEN.png/300px-MAZE-BoosterEN.png",
			Description:    "Latest core booster set - 24 packs",
			CurrentPrice:   82.00,
			AllTimeHigh:    105.00,
			AllTimeLow:     72.00,
			ATHDate:        time.Now().AddDate(0, -1, 0),
			ATLDate:        time.Now().AddDate(0, 0, -15),
			SearchTerms:    []string{"yugioh", "maze", "millenia", "booster", "box", "sealed", "modern", "new"},
			Tags:           []string{"sealed", "modern", "new"},
			PopularityRank: 6,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Yu-Gi-Oh Premium Gold Set",
			Set:            "Premium Gold",
			Game:           "Yu-Gi-Oh",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://ms.yugipedia.com//thumb/3/3e/PGLD-PackEN.png/300px-PGLD-PackEN.png",
			Description:    "Premium Gold mini box with gold rare reprints - 3 packs",
			CurrentPrice:   45.00,
			AllTimeHigh:    75.00,
			AllTimeLow:     32.00,
			ATHDate:        time.Now().AddDate(0, -12, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"yugioh", "premium", "gold", "mini", "box", "sealed", "gold", "rare"},
			Tags:           []string{"sealed", "premium", "gold-rare"},
			PopularityRank: 7,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Yu-Gi-Oh Legendary Collection Kaiba Box",
			Set:            "Legendary Collection Kaiba",
			Game:           "Yu-Gi-Oh",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://ms.yugipedia.com//thumb/b/b7/LCKC-SetEN.png/300px-LCKC-SetEN.png",
			Description:    "Special collection featuring Kaiba's iconic cards",
			CurrentPrice:   125.00,
			AllTimeHigh:    185.00,
			AllTimeLow:     85.00,
			ATHDate:        time.Now().AddDate(0, -18, 0),
			ATLDate:        time.Now().AddDate(0, -4, 0),
			SearchTerms:    []string{"yugioh", "legendary", "collection", "kaiba", "box", "sealed", "blue-eyes"},
			Tags:           []string{"sealed", "special", "legendary", "kaiba"},
			PopularityRank: 8,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Yu-Gi-Oh Albaz Strike Structure Deck",
			Set:            "Albaz Strike",
			Game:           "Yu-Gi-Oh",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://ms.yugipedia.com//thumb/2/2f/SDAZ-DeckEN.png/300px-SDAZ-DeckEN.png",
			Description:    "Structure deck with Albaz Fusion support",
			CurrentPrice:   35.00,
			AllTimeHigh:    55.00,
			AllTimeLow:     25.00,
			ATHDate:        time.Now().AddDate(0, -10, 0),
			ATLDate:        time.Now().AddDate(0, -2, 0),
			SearchTerms:    []string{"yugioh", "albaz", "strike", "structure", "deck", "sealed", "fusion"},
			Tags:           []string{"sealed", "structure-deck", "fusion"},
			PopularityRank: 9,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Name:           "Yu-Gi-Oh 25th Anniversary Tin",
			Set:            "25th Anniversary Tin",
			Game:           "Yu-Gi-Oh",
			Category:       "sealed",
			Rarity:         "",
			Number:         "",
			ImageURL:       "https://ms.yugipedia.com//thumb/8/8f/TIN23-SetEN.png/300px-TIN23-SetEN.png",
			Description:    "Special anniversary tin with reprint packs",
			CurrentPrice:   28.00,
			AllTimeHigh:    42.00,
			AllTimeLow:     22.00,
			ATHDate:        time.Now().AddDate(0, -5, 0),
			ATLDate:        time.Now().AddDate(0, -1, 0),
			SearchTerms:    []string{"yugioh", "25th", "anniversary", "tin", "sealed", "reprint"},
			Tags:           []string{"sealed", "tin", "anniversary", "reprint"},
			PopularityRank: 10,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
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

	// Create featured content
	fmt.Println("🎪 Creating featured carousel content...")

	// Get some card IDs for featured content
	charizardID := result.InsertedIDs[0].(primitive.ObjectID)
	blackLotusID := result.InsertedIDs[4].(primitive.ObjectID)
	blueEyesID := result.InsertedIDs[8].(primitive.ObjectID)

	featuredItems := []models.FeaturedContent{
		{
			Type:             "market_mover",
			Title:            "Charizard VMAX Surges 25% This Week!",
			Description:      "The iconic Charizard VMAX from Champions Path has seen massive price increases following tournament success",
			ImageURL:         "https://images.pokemontcg.io/swsh35/20_hires.png",
			CardID:           &charizardID,
			Priority:         100,
			Active:           true,
			CreatedAt:        time.Now(),
			PriceChange:      25.5,
			PriceChangeValue: 18.30,
		},
		{
			Type:        "product",
			Title:       "Featured: Black Lotus - The Ultimate Investment",
			Description: "Own a piece of Magic: The Gathering history. Alpha Black Lotus continues to appreciate in value",
			ImageURL:    "https://gatherer.wizards.com/Handlers/Image.ashx?multiverseid=3&type=card",
			CardID:      &blackLotusID,
			Priority:    90,
			Active:      true,
			CreatedAt:   time.Now(),
		},
		{
			Type:        "news",
			Title:       "Yu-Gi-Oh! Market Report: Vintage Cards Heat Up",
			Description: "First edition LOB cards including Blue-Eyes White Dragon see record sales at recent auctions",
			ImageURL:    "https://images.ygoprodeck.com/images/cards/89631139.jpg",
			CardID:      &blueEyesID,
			Link:        "https://www.tcgplayer.com/news",
			Priority:    80,
			Active:      true,
			CreatedAt:   time.Now(),
		},
		{
			Type:        "pickup",
			Title:       "Hot Pickup: Pokemon VSTAR Cards",
			Description: "Smart collectors are scooping up VSTAR cards before the next format rotation",
			ImageURL:    "https://images.pokemontcg.io/swsh9/123_hires.png",
			Priority:    70,
			Active:      true,
			CreatedAt:   time.Now(),
		},
		{
			Type:        "sponsored",
			Title:       "Premium Card Grading Services - 20% Off",
			Description: "Get your valuable cards professionally graded. Limited time offer on bulk submissions",
			ImageURL:    "https://via.placeholder.com/1200x450/6366f1/ffffff?text=Premium+Grading+Services",
			Link:        "https://www.psacard.com",
			Priority:    60,
			Active:      true,
			CreatedAt:   time.Now(),
			ExpiresAt:   timePtr(time.Now().AddDate(0, 1, 0)), // Expires in 1 month
		},
	}

	featuredDocuments := make([]interface{}, len(featuredItems))
	for i, item := range featuredItems {
		featuredDocuments[i] = item
	}

	featResult, err := featuredCollection.InsertMany(ctx, featuredDocuments)
	if err != nil {
		log.Printf("Warning: Failed to insert featured content: %v", err)
	} else {
		fmt.Printf("✅ Successfully inserted %d featured carousel items\n", len(featResult.InsertedIDs))
	}

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
			if rand.Float64() < 0.02 {             // 2% chance of large movement
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

// timePtr returns a pointer to a time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}
