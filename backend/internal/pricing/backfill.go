package pricing

import (
	"context"
	"sort"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/models"
)

const backfillBatch = 1000

// seriesRow is one stored price point, as the printings backfill reads it.
type seriesRow struct {
	Series    string    `bson:"grade"`
	Price     float64   `bson:"price"`
	Timestamp time.Time `bson:"timestamp"`
}

type cardSeriesRows struct {
	CardID primitive.ObjectID `bson:"_id"`
	Rows   []seriesRow        `bson:"rows"`
}

// BackfillPrintings fills in printings for cards priced before ingest recorded them,
// rebuilding each from the card's stored TCGplayer price points, so it makes no network
// requests. Cards that already have printings are left alone, which makes reruns safe.
// It reports how many cards it updated; a dry run reports how many priced cards it checked.
func BackfillPrintings(ctx context.Context, db *mongo.Database, dryRun bool) (int, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"source": SourceName, "grade": bson.M{"$nin": bson.A{nil, ""}}}},
		{"$group": bson.M{
			"_id":  "$card_id",
			"rows": bson.M{"$push": bson.M{"grade": "$grade", "price": "$price", "timestamp": "$timestamp"}},
		}},
	}
	cursor, err := db.Collection("prices").Aggregate(ctx, pipeline, options.Aggregate().SetAllowDiskUse(true))
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	updated := 0
	ops := make([]mongo.WriteModel, 0, backfillBatch)
	flush := func() error {
		if len(ops) == 0 {
			return nil
		}
		if !dryRun {
			result, err := db.Collection("cards").BulkWrite(ctx, ops, options.BulkWrite().SetOrdered(false))
			if err != nil {
				return err
			}
			updated += int(result.ModifiedCount)
		} else {
			updated += len(ops)
		}
		ops = ops[:0]
		return nil
	}

	for cursor.Next(ctx) {
		var card cardSeriesRows
		if err := cursor.Decode(&card); err != nil {
			return updated, err
		}
		printings := latestPrintings(card.Rows)
		if len(printings) == 0 {
			continue
		}
		ops = append(ops, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": card.CardID, "printings": bson.M{"$exists": false}}).
			SetUpdate(bson.M{"$set": bson.M{"printings": printings}}))
		if len(ops) == backfillBatch {
			if err := flush(); err != nil {
				return updated, err
			}
		}
	}
	if err := cursor.Err(); err != nil {
		return updated, err
	}
	return updated, flush()
}

// latestPrintings rebuilds a card's printings from its stored price points: each series
// priced on the card's most recent day, best first. Stored points only exist for priced
// printings, so unpriced ones appear after the next ingest.
func latestPrintings(rows []seriesRow) []models.Printing {
	var latest time.Time
	for _, row := range rows {
		if row.Timestamp.After(latest) {
			latest = row.Timestamp
		}
	}

	seen := map[string]bool{}
	printings := make([]models.Printing, 0, len(rows))
	for _, row := range rows {
		if !row.Timestamp.Equal(latest) || seen[row.Series] {
			continue
		}
		seen[row.Series] = true
		printings = append(printings, models.Printing{Series: row.Series, MarketPrice: round2(row.Price)})
	}
	// Rows arrive in no particular order, so settle ties by name to keep reruns stable.
	sort.Slice(printings, func(i, j int) bool { return printings[i].Series < printings[j].Series })
	sortPrintings(printings)
	return printings
}
