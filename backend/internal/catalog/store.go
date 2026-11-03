package catalog

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const upsertBatchSize = 500

// UpsertResult summarises a catalogue write batch.
type UpsertResult struct {
	Matched  int64
	Upserted int64
	Modified int64
}

func (r *UpsertResult) add(other *mongo.BulkWriteResult) {
	if other == nil {
		return
	}
	r.Matched += other.MatchedCount
	r.Upserted += other.UpsertedCount
	r.Modified += other.ModifiedCount
}

// UpsertCards writes catalogue metadata without touching scraper-owned fields
// (prices, ATH/ATL, sync timestamps, tier) on documents that already exist.
func UpsertCards(ctx context.Context, db *mongo.Database, cards []models.Card, now time.Time) (UpsertResult, error) {
	result := UpsertResult{}
	if len(cards) == 0 {
		return result, nil
	}
	collection := db.Collection("cards")
	for start := 0; start < len(cards); start += upsertBatchSize {
		end := start + upsertBatchSize
		if end > len(cards) {
			end = len(cards)
		}
		writes := make([]mongo.WriteModel, 0, end-start)
		for _, card := range cards[start:end] {
			writes = append(writes, mongo.NewUpdateOneModel().
				SetFilter(bson.M{"catalog_source": card.CatalogSource, "external_id": card.ExternalID}).
				SetUpdate(bson.M{"$set": catalogFields(card, now), "$setOnInsert": initialFields(now)}).
				SetUpsert(true))
		}
		written, err := collection.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false))
		if err != nil {
			return result, fmt.Errorf("upsert catalog cards: %w", err)
		}
		result.add(written)
	}
	return result, nil
}

func catalogFields(card models.Card, now time.Time) bson.M {
	fields := bson.M{
		"name":                 card.Name,
		"set":                  card.Set,
		"game":                 card.Game,
		"category":             card.Category,
		"image_url":            card.ImageURL,
		"search_terms":         card.SearchTerms,
		"catalog_source":       card.CatalogSource,
		"external_id":          card.ExternalID,
		"external_group_id":    card.ExternalGroupID,
		"external_category_id": card.ExternalCategory,
		"product_type":         card.ProductType,
		"language":             card.Language,
		"product_url":          card.ProductURL,
		"active":               true,
		"catalog_synced_at":    now.UTC(),
		"updated_at":           now.UTC(),
	}
	if card.Rarity != "" {
		fields["rarity"] = card.Rarity
	}
	if card.Number != "" {
		fields["number"] = card.Number
	}
	if card.Description != "" {
		fields["description"] = card.Description
	}
	return fields
}

// initialFields seed values that only apply to brand-new catalogue entries; the
// scraper owns them from then on. New items start in the long-tail tier, spread
// over the next 24h so a bulk import does not create one giant scrape burst.
func initialFields(now time.Time) bson.M {
	return bson.M{
		"created_at":      now.UTC(),
		"current_price":   0.0,
		"all_time_high":   0.0,
		"all_time_low":    0.0,
		"scrape_tier":     models.ScrapeTierColdTail,
		"next_scrape_at":  now.UTC().Add(time.Duration(rand.Int63n(int64(24 * time.Hour)))),
		"scrape_failures": 0,
	}
}
