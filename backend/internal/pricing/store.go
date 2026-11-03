package pricing

import (
	"context"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/catalog"
	"github.com/jamesc159/monmetrics/internal/models"
)

type groupStats struct {
	Products    int
	PricePoints int
	CardsPriced int
}

// persistGroup joins a group's price rows onto our cards and writes one dated point per
// printing, plus the card's headline price.
func (r *Runner) persistGroup(
	ctx context.Context,
	group GroupRef,
	prices []catalog.ProductPrice,
	day time.Time,
) (groupStats, error) {
	var stats groupStats
	if len(prices) == 0 {
		return stats, nil
	}

	cards, err := r.cardsByExternalID(ctx, group)
	if err != nil {
		return stats, err
	}
	if len(cards) == 0 {
		return stats, nil
	}

	now := time.Now().UTC()
	priceOps := make([]mongo.WriteModel, 0, len(prices))
	marketOps := make([]mongo.WriteModel, 0, len(prices))
	cardOps := make([]mongo.WriteModel, 0)

	// Best printing seen per card, used for the card's headline price.
	type headline struct {
		rank  int
		value float64
	}
	headlines := map[primitive.ObjectID]headline{}

	for _, price := range prices {
		cardID, ok := cards[externalKey(price.ProductID)]
		if !ok {
			continue // priced product we do not carry
		}
		// A null marketPrice means TCGplayer has no figure; recording 0 would be a lie.
		if price.MarketPrice == nil || *price.MarketPrice <= 0 {
			continue
		}

		market := *price.MarketPrice
		series := SeriesKey(price.SubTypeName)
		stats.Products++

		filter := bson.M{
			"card_id":   cardID,
			"source":    SourceName,
			"grade":     series,
			"timestamp": day,
		}
		point := models.PricePoint{
			CardID:    cardID,
			Price:     round2(market),
			Source:    SourceName,
			Grade:     series,
			Timestamp: day,
			CreatedAt: now,
		}
		priceOps = append(priceOps, mongo.NewReplaceOneModel().
			SetFilter(filter).SetReplacement(point).SetUpsert(true))

		// A daily snapshot has no intraday movement, so open == close. Low/high come from
		// the listing spread, and weighted_avg carries the median listing price.
		bar := models.MarketData{
			CardID:           cardID,
			Source:           SourceName,
			Grade:            series,
			Date:             day,
			OpenPrice:        round2(market),
			ClosePrice:       round2(market),
			HighPrice:        round2(orZero(price.HighPrice, market)),
			LowPrice:         round2(orZero(price.LowPrice, market)),
			Volume:           0,
			WeightedAvgPrice: round2(orZero(price.MidPrice, market)),
		}
		marketOps = append(marketOps, mongo.NewReplaceOneModel().
			SetFilter(bson.M{
				"card_id": cardID,
				"source":  SourceName,
				"grade":   series,
				"date":    day,
			}).SetReplacement(bar).SetUpsert(true))

		if rank := printingRank(series); rank < headlines[cardID].rank || headlines[cardID].value == 0 {
			headlines[cardID] = headline{rank: rank, value: market}
		}
	}

	for cardID, h := range headlines {
		cardOps = append(cardOps, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": cardID}).
			SetUpdate(bson.M{"$set": bson.M{
				"current_price":          round2(h.value),
				"market_price":           round2(h.value),
				"market_price_synced_at": now,
				"updated_at":             now,
			}}))
	}

	stats.PricePoints = len(priceOps)
	stats.CardsPriced = len(cardOps)

	if r.DryRun {
		return stats, nil
	}

	for _, batch := range []struct {
		collection string
		ops        []mongo.WriteModel
	}{
		{"prices", priceOps},
		{"market_data", marketOps},
		{"cards", cardOps},
	} {
		if len(batch.ops) == 0 {
			continue
		}
		_, err := r.DB.Collection(batch.collection).
			BulkWrite(ctx, batch.ops, options.BulkWrite().SetOrdered(false))
		if err != nil {
			return stats, err
		}
	}

	return stats, nil
}

// cardsByExternalID indexes a group's cards by their TCGplayer product id.
func (r *Runner) cardsByExternalID(ctx context.Context, group GroupRef) (map[string]primitive.ObjectID, error) {
	filter := bson.M{
		"external_category_id": externalKey(group.CategoryID),
		"external_group_id":    externalKey(group.GroupID),
		"external_id":          bson.M{"$exists": true, "$ne": ""},
	}
	cursor, err := r.DB.Collection("cards").
		Find(ctx, filter, options.Find().SetProjection(bson.M{"_id": 1, "external_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var refs []cardRef
	if err := cursor.All(ctx, &refs); err != nil {
		return nil, err
	}

	index := make(map[string]primitive.ObjectID, len(refs))
	for _, ref := range refs {
		index[ref.ExternalID] = ref.ID
	}
	return index, nil
}

func externalKey(id int) string {
	return strconv.Itoa(id)
}

func printingRank(series string) int {
	for i, name := range primaryPrintings {
		if name == series {
			return i
		}
	}
	return len(primaryPrintings)
}

func orZero(v *float64, fallback float64) float64 {
	if v == nil || *v <= 0 {
		return fallback
	}
	return *v
}

func round2(v float64) float64 {
	return float64(int64(v*100+0.5)) / 100
}
