package handlers

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/jamesc159/monmetrics/internal/models"
)

// cardsByID batch-loads cards; missing IDs are simply absent from the map.
func (h *Handlers) cardsByID(ctx context.Context, ids []primitive.ObjectID) (map[primitive.ObjectID]*models.Card, error) {
	out := make(map[primitive.ObjectID]*models.Card, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	cur, err := h.db.Collection("cards").Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var cards []models.Card
	if err := cur.All(ctx, &cards); err != nil {
		return nil, err
	}
	for i := range cards {
		out[cards[i].ID] = &cards[i]
	}
	return out, nil
}

// cardExists reports whether a catalog card exists.
func (h *Handlers) cardExists(ctx context.Context, id primitive.ObjectID) (bool, error) {
	err := h.db.Collection("cards").FindOne(ctx, bson.M{"_id": id}).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	return err == nil, err
}
