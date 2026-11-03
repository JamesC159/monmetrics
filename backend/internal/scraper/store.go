package scraper

import (
	"context"
	"fmt"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func PersistCard(ctx context.Context, db *mongo.Database, card models.Card, sales []models.SoldSale, now time.Time) error {
	if len(sales) == 0 {
		return nil
	}
	writes := make([]mongo.WriteModel, 0, len(sales))
	for _, sale := range sales {
		sale.CardID = card.ID
		sale.Source = models.ProviderEbay
		if sale.Grade == "" {
			sale.Grade = DetectGrade(sale.Title)
		}
		if sale.ScrapedAt.IsZero() {
			sale.ScrapedAt = now.UTC()
		}
		filter := bson.M{"source": sale.Source, "item_id": sale.ItemID}
		writes = append(writes, mongo.NewReplaceOneModel().SetFilter(filter).SetReplacement(sale).SetUpsert(true))
	}
	if _, err := db.Collection("sold_listings").BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false)); err != nil {
		return fmt.Errorf("upsert sold listings: %w", err)
	}

	cutoff := now.UTC().AddDate(-5, 0, 0)
	cursor, err := db.Collection("sold_listings").Find(ctx, bson.M{
		"card_id": card.ID, "source": models.ProviderEbay, "sold_at": bson.M{"$gte": cutoff, "$lte": now.UTC()},
	})
	if err != nil {
		return fmt.Errorf("load sold listings: %w", err)
	}
	defer cursor.Close(ctx)
	var stored []models.SoldSale
	if err := cursor.All(ctx, &stored); err != nil {
		return fmt.Errorf("decode sold listings: %w", err)
	}

	buckets := DailyBuckets(stored)
	byGrade := make(map[string][]DailyAggregate)
	for _, bucket := range buckets {
		byGrade[bucket.Grade] = append(byGrade[bucket.Grade], bucket)
	}
	if _, exists := byGrade["raw"]; !exists {
		byGrade["raw"] = nil
	}
	prices := db.Collection("prices")
	market := db.Collection("market_data")
	summaries := db.Collection("price_summaries")
	var priceWrites, marketWrites, summaryWrites []mongo.WriteModel
	var rawBuckets []DailyAggregate
	for grade, gradeBuckets := range byGrade {
		if grade == "raw" {
			rawBuckets = gradeBuckets
		}
		for _, bucket := range gradeBuckets {
			filter := bson.M{"card_id": card.ID, "source": models.ProviderEbay, "grade": grade, "timestamp": bucket.Date}
			point := models.PricePoint{CardID: card.ID, Price: bucket.Average, Volume: bucket.Volume, Source: models.ProviderEbay, Grade: grade, Timestamp: bucket.Date, CreatedAt: now.UTC()}
			priceWrites = append(priceWrites, mongo.NewReplaceOneModel().SetFilter(filter).SetReplacement(point).SetUpsert(true))
			marketFilter := bson.M{"card_id": card.ID, "source": models.ProviderEbay, "grade": grade, "date": bucket.Date}
			marketPoint := models.MarketData{CardID: card.ID, Source: models.ProviderEbay, Grade: grade, Date: bucket.Date, OpenPrice: bucket.OpenPrice, ClosePrice: bucket.ClosePrice, HighPrice: bucket.HighPrice, LowPrice: bucket.LowPrice, Volume: bucket.Volume, WeightedAvgPrice: bucket.WeightedAvgPrice}
			marketWrites = append(marketWrites, mongo.NewReplaceOneModel().SetFilter(marketFilter).SetReplacement(marketPoint).SetUpsert(true))
		}
		summary := models.PriceSummary{CardID: card.ID, Source: models.ProviderEbay, Grade: grade, UpdatedAt: now.UTC(), Windows: BuildWindowSummaries(gradeBuckets, now)}
		summaryFilter := bson.M{"card_id": card.ID, "source": models.ProviderEbay, "grade": grade}
		summaryWrites = append(summaryWrites, mongo.NewReplaceOneModel().SetFilter(summaryFilter).SetReplacement(summary).SetUpsert(true))
	}
	if len(priceWrites) > 0 {
		if _, err := prices.BulkWrite(ctx, priceWrites, options.BulkWrite().SetOrdered(false)); err != nil {
			return fmt.Errorf("write daily prices: %w", err)
		}
	}
	if len(marketWrites) > 0 {
		if _, err := market.BulkWrite(ctx, marketWrites, options.BulkWrite().SetOrdered(false)); err != nil {
			return fmt.Errorf("write daily market data: %w", err)
		}
	}
	if _, err := summaries.BulkWrite(ctx, summaryWrites, options.BulkWrite().SetOrdered(false)); err != nil {
		return fmt.Errorf("write price summaries: %w", err)
	}
	if card.EbaySyncedAt == nil {
		if _, err := prices.DeleteMany(ctx, bson.M{
			"card_id": card.ID, "source": models.ProviderEbay,
			"$or": []bson.M{{"grade": bson.M{"$exists": false}}, {"grade": ""}},
		}); err != nil {
			return fmt.Errorf("delete mock eBay prices: %w", err)
		}
	}

	cardUpdate := bson.M{"ebay_synced_at": now.UTC(), "updated_at": now.UTC()}
	if len(rawBuckets) > 0 {
		latest := rawBuckets[len(rawBuckets)-1]
		cardUpdate["current_price"] = latest.Average
		cardUpdate["all_time_high"] = rawBuckets[0].HighPrice
		cardUpdate["all_time_low"] = rawBuckets[0].LowPrice
		high, low := rawBuckets[0], rawBuckets[0]
		for _, bucket := range rawBuckets[1:] {
			if bucket.HighPrice > high.HighPrice {
				high = bucket
			}
			if bucket.LowPrice < low.LowPrice {
				low = bucket
			}
		}
		cardUpdate["all_time_high"], cardUpdate["ath_date"] = high.HighPrice, high.Date
		cardUpdate["all_time_low"], cardUpdate["atl_date"] = low.LowPrice, low.Date
	}
	if _, err := db.Collection("cards").UpdateOne(ctx, bson.M{"_id": card.ID}, bson.M{"$set": cardUpdate}); err != nil {
		return fmt.Errorf("update card summary: %w", err)
	}
	return nil
}

func ParseCardID(value string) (primitive.ObjectID, error) {
	id, err := primitive.ObjectIDFromHex(value)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid card id %q: %w", value, err)
	}
	return id, nil
}
