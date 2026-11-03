package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// realDataMark is a trace a real data job leaves in MongoDB. The seeder writes none of
// these, so finding one means the database holds data that sample data must not replace.
type realDataMark struct {
	What       string
	Collection string
	Filter     bson.M
	Job        string
}

// hasGrade matches price rows from the real pipelines: TCGplayer rows carry a printing and
// eBay rows a grade, while seeded rows leave grade unset.
var hasGrade = bson.M{"grade": bson.M{"$nin": bson.A{nil, ""}}}

var realDataMarks = []realDataMark{
	{"catalog-imported cards", "cards", bson.M{"catalog_source": bson.M{"$exists": true}}, "make catalog-sync"},
	{"catalog sync checkpoints", "catalog_sync_state", bson.M{}, "make catalog-sync"},
	{"TCGCSV-priced cards", "cards", bson.M{"market_price_synced_at": bson.M{"$exists": true}}, "make catalog-sync or make prices"},
	{"scrape-scheduled cards", "cards", bson.M{"last_scrape_at": bson.M{"$exists": true}}, "make scrape"},
	{"eBay-synced cards", "cards", bson.M{"ebay_synced_at": bson.M{"$exists": true}}, "make scrape"},
	{"graded price rows", "prices", hasGrade, "make catalog-sync, make prices or make scrape"},
	{"graded market data", "market_data", hasGrade, "make catalog-sync, make prices or make scrape"},
	{"sold listings", "sold_listings", bson.M{}, "make scrape"},
	{"price summaries", "price_summaries", bson.M{}, "make scrape"},
}

// countFunc reports how many documents in collection match filter.
type countFunc func(ctx context.Context, collection string, filter bson.M) (int64, error)

func mongoCounter(db *mongo.Database) countFunc {
	return func(ctx context.Context, collection string, filter bson.M) (int64, error) {
		return db.Collection(collection).CountDocuments(ctx, filter, options.Count().SetLimit(1))
	}
}

// findRealData describes every mark present. A check that cannot run counts as found:
// the guard fails closed rather than seeding over data it could not rule out.
func findRealData(ctx context.Context, count countFunc) []string {
	var found []string
	for _, mark := range realDataMarks {
		n, err := count(ctx, mark.Collection, mark.Filter)
		switch {
		case err != nil:
			found = append(found, fmt.Sprintf("%s (%s): check failed: %v", mark.What, mark.Collection, err))
		case n > 0:
			found = append(found, fmt.Sprintf("%s (%s, from %s)", mark.What, mark.Collection, mark.Job))
		}
	}
	return found
}
