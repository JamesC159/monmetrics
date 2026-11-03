// Package pricing ingests daily market prices and accrues them into history.
//
// TCGplayer publishes only a current snapshot - no archive exists - so history is built
// forward, one dated row per run. Re-running on the same day overwrites that day rather
// than appending, which keeps the series honest and makes retries idempotent.
package pricing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jamesc159/monmetrics/internal/catalog"
)

// SourceName tags every row this package writes.
const SourceName = "tcgplayer"

// primaryPrintings ranks which printing represents a product's headline price, best first.
// Magic prints "Normal"/"Foil"; Pokemon adds "Holofoil"/"Reverse Holofoil".
var primaryPrintings = []string{"normal", "foil", "holofoil", "reverse-holofoil"}

// GroupRef identifies a catalog group we hold cards for.
type GroupRef struct {
	CategoryID int `bson:"category_id"`
	GroupID    int `bson:"group_id"`
}

// Result summarises one ingest cycle.
type Result struct {
	Groups      int
	Products    int
	PricePoints int
	CardsPriced int
	Skipped     int
}

// SeriesKey normalises a TCGplayer printing into a storage key.
// "Reverse Holofoil" -> "reverse-holofoil".
func SeriesKey(subTypeName string) string {
	key := strings.ToLower(strings.TrimSpace(subTypeName))
	key = strings.ReplaceAll(key, " ", "-")
	if key == "" {
		return "normal"
	}
	return key
}

// cardRef is the minimum card data the ingest needs.
type cardRef struct {
	ID         primitive.ObjectID `bson:"_id"`
	ExternalID string             `bson:"external_id"`
}

// GroupsWithCards returns the catalog groups we actually hold products for, so a cycle
// only requests price files it can use. Ingest deliberately covers every game, including
// ones hidden behind feature flags, so their history is accruing before they are shown.
func GroupsWithCards(ctx context.Context, db *mongo.Database) ([]GroupRef, error) {
	pipeline := []bson.M{
		{"$match": bson.M{
			"external_id":          bson.M{"$exists": true},
			"external_group_id":    bson.M{"$exists": true, "$ne": ""},
			"external_category_id": bson.M{"$exists": true, "$ne": ""},
		}},
		{"$group": bson.M{"_id": bson.M{
			"category_id": "$external_category_id",
			"group_id":    "$external_group_id",
		}}},
		{"$project": bson.M{
			"_id":         0,
			"category_id": bson.M{"$toInt": "$_id.category_id"},
			"group_id":    bson.M{"$toInt": "$_id.group_id"},
		}},
		{"$sort": bson.M{"category_id": 1, "group_id": 1}},
	}

	cursor, err := db.Collection("cards").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var refs []GroupRef
	if err := cursor.All(ctx, &refs); err != nil {
		return nil, err
	}
	return refs, nil
}

// Runner performs price ingest cycles.
type Runner struct {
	DB     *mongo.Database
	Source catalog.PriceSource
	// MaxGroups caps requests per cycle. TCGCSV asks for under 10k requests per 24h.
	MaxGroups int
	DryRun    bool
}

// RunOnce ingests today's prices for every group we hold cards for.
func (r *Runner) RunOnce(ctx context.Context) (Result, error) {
	var result Result

	groups, err := GroupsWithCards(ctx, r.DB)
	if err != nil {
		return result, fmt.Errorf("load groups: %w", err)
	}
	if r.MaxGroups > 0 && len(groups) > r.MaxGroups {
		result.Skipped = len(groups) - r.MaxGroups
		groups = groups[:r.MaxGroups]
	}

	// One UTC-midnight stamp for the whole cycle so a run spanning midnight stays coherent.
	day := time.Now().UTC().Truncate(24 * time.Hour)

	for _, group := range groups {
		prices, err := r.Source.Prices(ctx, group.CategoryID, group.GroupID)
		if err != nil {
			// A missing price file is normal for an empty group; throttling is not.
			if isFatal(err) {
				return result, fmt.Errorf("group %d/%d: %w", group.CategoryID, group.GroupID, err)
			}
			result.Skipped++
			continue
		}
		result.Groups++

		stats, err := r.persistGroup(ctx, group, prices, day)
		if err != nil {
			return result, fmt.Errorf("persist %d/%d: %w", group.CategoryID, group.GroupID, err)
		}
		result.Products += stats.Products
		result.PricePoints += stats.PricePoints
		result.CardsPriced += stats.CardsPriced
	}

	return result, nil
}

func isFatal(err error) bool {
	return err != nil && strings.Contains(err.Error(), catalog.ErrThrottled.Error())
}
