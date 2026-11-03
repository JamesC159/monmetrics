package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jamesc159/monmetrics/configs"
	"github.com/jamesc159/monmetrics/internal/database"
	"github.com/jamesc159/monmetrics/internal/models"
	"github.com/jamesc159/monmetrics/internal/scraper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type repeatedValues []string

func (values *repeatedValues) String() string { return fmt.Sprint([]string(*values)) }
func (values *repeatedValues) Set(value string) error {
	*values = append(*values, value)
	return nil
}

func main() {
	var cardValues repeatedValues
	var all, dryRun bool
	var fixture string
	flag.Var(&cardValues, "card", "Card ObjectID to scrape (repeatable)")
	flag.BoolVar(&all, "all", false, "Scrape all catalog cards")
	flag.BoolVar(&dryRun, "dry-run", false, "Fetch and calculate summaries without writing data")
	flag.StringVar(&fixture, "fixture", "", "Read sold listings from a saved HTML file instead of eBay")
	flag.Parse()
	if !all && len(cardValues) == 0 {
		log.Fatal("specify -all or at least one -card <id>")
	}

	cardIDs := make([]primitive.ObjectID, 0, len(cardValues))
	for _, value := range cardValues {
		id, err := scraper.ParseCardID(value)
		if err != nil {
			log.Fatal(err)
		}
		cardIDs = append(cardIDs, id)
	}
	config := configs.Load()
	db, err := database.Connect(config.MongoURI, config.DBName)
	if err != nil {
		log.Fatal("connect database: ", err)
	}
	defer database.Disconnect()

	var source scraper.SoldSource
	if fixture != "" {
		body, err := os.ReadFile(fixture)
		if err != nil {
			log.Fatal("read fixture: ", err)
		}
		source = scraper.FixtureSource{HTML: string(body)}
	} else {
		source = scraper.NewEbayHTMLSource(config.ScraperBaseURL, config.ScraperUserAgent, config.ScraperRequestDelay, config.ScraperMaxPages)
	}
	runner := &scraper.Runner{
		DB: db, Source: source, Interval: config.ScraperInterval, MaxPages: config.ScraperMaxPages,
		Scheduler:   &scraper.Scheduler{DB: db, Intervals: config.ScrapeTierIntervals},
		DailyBudget: config.ScraperDailyBudget,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	if !dryRun {
		if err := runner.RunOnce(ctx, cardIDs); err != nil {
			log.Fatal("scrape failed: ", err)
		}
		return
	}

	filter := bson.M{"category": bson.M{"$in": bson.A{"card", "sealed"}}}
	if !all {
		filter["_id"] = bson.M{"$in": cardIDs}
	}
	cursor, err := db.Collection("cards").Find(ctx, filter)
	if err != nil {
		log.Fatal("load cards: ", err)
	}
	defer cursor.Close(ctx)
	var cards []models.Card
	if err := cursor.All(ctx, &cards); err != nil {
		log.Fatal("decode cards: ", err)
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	for _, card := range cards {
		sales, err := runner.ScrapeCard(ctx, card, false)
		if err != nil {
			log.Printf("card %s failed: %v", card.ID.Hex(), err)
			continue
		}
		buckets := scraper.DailyBuckets(sales)
		byGrade := make(map[string][]scraper.DailyAggregate)
		for _, bucket := range buckets {
			byGrade[bucket.Grade] = append(byGrade[bucket.Grade], bucket)
		}
		output := struct {
			CardID   string                                        `json:"card_id"`
			Name     string                                        `json:"name"`
			Sales    int                                           `json:"sales"`
			Excluded int                                           `json:"excluded"`
			Windows  map[string]map[string]models.PriceWindowStats `json:"windows_by_grade"`
		}{CardID: card.ID.Hex(), Name: card.Name, Sales: len(sales), Windows: make(map[string]map[string]models.PriceWindowStats)}
		for _, sale := range sales {
			if sale.Excluded || sale.BestOffer {
				output.Excluded++
			}
		}
		for grade, gradeBuckets := range byGrade {
			output.Windows[grade] = scraper.BuildWindowSummaries(gradeBuckets, time.Now().UTC())
		}
		if err := encoder.Encode(output); err != nil {
			log.Fatal("encode dry-run output: ", err)
		}
	}
}
