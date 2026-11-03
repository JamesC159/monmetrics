package catalog

import (
	"context"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// lineBreakTag, markupTag and extraBlank cover what TCGplayer puts in Card Text: <br>
// line breaks and inline tags such as <em>. markupTag needs a letter after '<' so text
// like "10 < 20" is left alone.
var (
	lineBreakTag = regexp.MustCompile(`(?i)<br\s*/?>`)
	markupTag    = regexp.MustCompile(`</?[a-zA-Z][^>]*>`)
	extraBlank   = regexp.MustCompile(`\n{3,}`)
)

// markupPattern finds stored descriptions that still hold HTML or entities.
const markupPattern = `<|&[A-Za-z#0-9]+;`

// cleanCardText converts TCGplayer's HTML card text to plain text: <br> becomes a newline,
// other tags are dropped and entities are decoded. Descriptions are rendered as text, so
// any markup left here shows up literally on the card page.
func cleanCardText(text string) string {
	text = lineBreakTag.ReplaceAllString(text, "\n")
	text = markupTag.ReplaceAllString(text, "")
	text = html.UnescapeString(text)
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	text = strings.Join(lines, "\n")
	text = extraBlank.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

// RepairDescriptions rewrites stored card descriptions that predate cleanCardText. It
// reads and writes MongoDB only and never contacts the catalog source. With dryRun set it
// counts the rows that would change and writes nothing.
func RepairDescriptions(ctx context.Context, db *mongo.Database, dryRun bool) (scanned, changed int, err error) {
	collection := db.Collection("cards")
	cursor, err := collection.Find(ctx, bson.M{"description": bson.M{"$regex": markupPattern}},
		options.Find().SetProjection(bson.M{"description": 1}))
	if err != nil {
		return 0, 0, fmt.Errorf("find descriptions: %w", err)
	}
	defer cursor.Close(ctx)

	now := time.Now().UTC()
	writes := make([]mongo.WriteModel, 0, upsertBatchSize)
	flush := func() error {
		if dryRun || len(writes) == 0 {
			writes = writes[:0]
			return nil
		}
		if _, err := collection.BulkWrite(ctx, writes, options.BulkWrite().SetOrdered(false)); err != nil {
			return fmt.Errorf("repair descriptions: %w", err)
		}
		writes = writes[:0]
		return nil
	}

	for cursor.Next(ctx) {
		var doc struct {
			ID          any    `bson:"_id"`
			Description string `bson:"description"`
		}
		if err := cursor.Decode(&doc); err != nil {
			return scanned, changed, fmt.Errorf("decode description: %w", err)
		}
		scanned++
		cleaned := cleanCardText(doc.Description)
		if cleaned == doc.Description {
			continue
		}
		changed++
		writes = append(writes, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": doc.ID}).
			SetUpdate(bson.M{"$set": bson.M{"description": cleaned, "updated_at": now}}))
		if len(writes) == upsertBatchSize {
			if err := flush(); err != nil {
				return scanned, changed, err
			}
		}
	}
	if err := cursor.Err(); err != nil {
		return scanned, changed, fmt.Errorf("read descriptions: %w", err)
	}
	return scanned, changed, flush()
}
