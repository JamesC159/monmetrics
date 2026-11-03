package catalog

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const CheckpointCollection = "catalog_sync_state"

// GroupCheckpoint records when a set finished ingesting, so a sync that dies
// partway through resumes at the group it failed on instead of starting over.
type GroupCheckpoint struct {
	Source       string     `bson:"source" json:"source"`
	CategoryID   int        `bson:"category_id" json:"category_id"`
	GroupID      int        `bson:"group_id" json:"group_id"`
	CategoryName string     `bson:"category_name,omitempty" json:"category_name,omitempty"`
	GroupName    string     `bson:"group_name,omitempty" json:"group_name,omitempty"`
	Products     int        `bson:"products" json:"products"`
	CompletedAt  time.Time  `bson:"completed_at" json:"completed_at"`
	LastError    string     `bson:"last_error,omitempty" json:"last_error,omitempty"`
	FailedAt     *time.Time `bson:"failed_at,omitempty" json:"failed_at,omitempty"`
	// PricesSyncedAt is when a pass last stored the group's prices along with its products.
	// Groups checkpointed before prices joined the catalog pass have none.
	PricesSyncedAt *time.Time `bson:"prices_synced_at,omitempty" json:"prices_synced_at,omitempty"`
}

func checkpointKey(categoryID, groupID int) string {
	return fmt.Sprintf("%d/%d", categoryID, groupID)
}

// ingestedAt is when the group was last fully ingested. When the sync stores prices, a
// pass that left them out doesn't count, so a group with no stored prices returns zero.
func (c GroupCheckpoint) ingestedAt(withPrices bool) time.Time {
	if !withPrices {
		return c.CompletedAt
	}
	if c.PricesSyncedAt == nil {
		return time.Time{}
	}
	if c.PricesSyncedAt.Before(c.CompletedAt) {
		return *c.PricesSyncedAt
	}
	return c.CompletedAt
}

// LoadCheckpoints returns the ingestion time of every group already completed by
// this source, keyed by "categoryID/groupID". With withPrices, only groups whose
// prices were stored count as completed, as of that time.
func LoadCheckpoints(ctx context.Context, db *mongo.Database, source string, withPrices bool) (map[string]time.Time, error) {
	filter := bson.M{"source": source, "completed_at": bson.M{"$exists": true}}
	cursor, err := db.Collection(CheckpointCollection).Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("load catalog checkpoints: %w", err)
	}
	defer cursor.Close(ctx)
	var records []GroupCheckpoint
	if err := cursor.All(ctx, &records); err != nil {
		return nil, fmt.Errorf("decode catalog checkpoints: %w", err)
	}
	var priced map[string]time.Time
	if withPrices {
		if priced, err = lastPricedAt(ctx, db, source); err != nil {
			return nil, err
		}
	}
	return resumeTimes(records, priced, withPrices), nil
}

// resumeTimes keys each completed group's ingestion time by "categoryID/groupID".
// priced holds when each group's cards were last priced, from lastPricedAt.
func resumeTimes(records []GroupCheckpoint, priced map[string]time.Time, withPrices bool) map[string]time.Time {
	checkpoints := make(map[string]time.Time, len(records))
	for _, record := range records {
		key := checkpointKey(record.CategoryID, record.GroupID)
		if at, ok := priced[key]; ok && record.PricesSyncedAt == nil {
			// Checkpoints written before prices joined the catalog pass don't say whether
			// the group was priced. Its cards do, whether `make prices` priced them or not.
			record.PricesSyncedAt = &at
		}
		ingestedAt := record.ingestedAt(withPrices)
		if ingestedAt.IsZero() {
			continue
		}
		checkpoints[key] = ingestedAt
	}
	return checkpoints
}

// lastPricedAt returns when each group last had a card's price stored, keyed like
// checkpoints. Groups whose cards have never been priced are absent.
func lastPricedAt(ctx context.Context, db *mongo.Database, source string) (map[string]time.Time, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"catalog_source": source, "market_price_synced_at": bson.M{"$exists": true}}},
		{"$group": bson.M{
			"_id":       bson.M{"category": "$external_category_id", "group": "$external_group_id"},
			"priced_at": bson.M{"$max": "$market_price_synced_at"},
		}},
	}
	cursor, err := db.Collection("cards").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("load group price times: %w", err)
	}
	defer cursor.Close(ctx)
	var rows []struct {
		ID struct {
			Category string `bson:"category"`
			Group    string `bson:"group"`
		} `bson:"_id"`
		PricedAt time.Time `bson:"priced_at"`
	}
	if err := cursor.All(ctx, &rows); err != nil {
		return nil, fmt.Errorf("decode group price times: %w", err)
	}
	priced := make(map[string]time.Time, len(rows))
	for _, row := range rows {
		priced[row.ID.Category+"/"+row.ID.Group] = row.PricedAt
	}
	return priced, nil
}

// SaveCheckpoint marks a group as fully ingested and clears any previous failure. A
// checkpoint without PricesSyncedAt keeps the group's earlier prices time.
func SaveCheckpoint(ctx context.Context, db *mongo.Database, checkpoint GroupCheckpoint) error {
	set := bson.M{
		"category_name": checkpoint.CategoryName,
		"group_name":    checkpoint.GroupName,
		"products":      checkpoint.Products,
		"completed_at":  checkpoint.CompletedAt.UTC(),
	}
	if checkpoint.PricesSyncedAt != nil {
		set["prices_synced_at"] = checkpoint.PricesSyncedAt.UTC()
	}
	update := bson.M{
		"$set":   set,
		"$unset": bson.M{"last_error": "", "failed_at": ""},
	}
	return writeCheckpoint(ctx, db, checkpoint.Source, checkpoint.CategoryID, checkpoint.GroupID, update)
}

// SaveCheckpointFailure records why a group stopped without marking it complete,
// so the next run retries it.
func SaveCheckpointFailure(ctx context.Context, db *mongo.Database, source string, category Category, group Group, cause error, now time.Time) error {
	update := bson.M{"$set": bson.M{
		"category_name": category.DisplayName,
		"group_name":    group.Name,
		"last_error":    cause.Error(),
		"failed_at":     now.UTC(),
	}}
	return writeCheckpoint(ctx, db, source, category.CategoryID, group.GroupID, update)
}

func writeCheckpoint(ctx context.Context, db *mongo.Database, source string, categoryID, groupID int, update bson.M) error {
	filter := bson.M{"source": source, "category_id": categoryID, "group_id": groupID}
	opts := options.Update().SetUpsert(true)
	if _, err := db.Collection(CheckpointCollection).UpdateOne(ctx, filter, update, opts); err != nil {
		return fmt.Errorf("save catalog checkpoint: %w", err)
	}
	return nil
}
