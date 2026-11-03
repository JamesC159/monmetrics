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
}

func checkpointKey(categoryID, groupID int) string {
	return fmt.Sprintf("%d/%d", categoryID, groupID)
}

// LoadCheckpoints returns the ingestion time of every group already completed by
// this source, keyed by "categoryID/groupID".
func LoadCheckpoints(ctx context.Context, db *mongo.Database, source string) (map[string]time.Time, error) {
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
	checkpoints := make(map[string]time.Time, len(records))
	for _, record := range records {
		if record.CompletedAt.IsZero() {
			continue
		}
		checkpoints[checkpointKey(record.CategoryID, record.GroupID)] = record.CompletedAt
	}
	return checkpoints, nil
}

// SaveCheckpoint marks a group as fully ingested and clears any previous failure.
func SaveCheckpoint(ctx context.Context, db *mongo.Database, checkpoint GroupCheckpoint) error {
	update := bson.M{
		"$set": bson.M{
			"category_name": checkpoint.CategoryName,
			"group_name":    checkpoint.GroupName,
			"products":      checkpoint.Products,
			"completed_at":  checkpoint.CompletedAt.UTC(),
		},
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
