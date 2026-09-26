package handlers

import (
	"context"
	"math"
	"net/http"
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/models"
	"github.com/jamesc159/monmetrics/internal/valuation"
)

const maxPortfolioItems = 5000

type portfolioFilter struct {
	Game     string
	ItemType string
	Sort     string
}

// toPortfolioView attaches the card snapshot and computes valuation for an item.
func toPortfolioView(item models.PortfolioItem, card *models.Card) models.PortfolioItemView {
	v := models.PortfolioItemView{PortfolioItem: item, Card: card}
	if card != nil {
		v.DisplayName, v.DisplayGame, v.DisplaySet, v.DisplayImageURL = card.Name, card.Game, card.Set, card.ImageURL
	} else {
		v.CardUnavailable = item.CardID != nil
		v.DisplayName, v.DisplayGame, v.DisplaySet, v.DisplayImageURL = item.CustomName, item.CustomGame, item.CustomSet, item.CustomImageURL
		if v.DisplayName == "" {
			v.DisplayName = "Unavailable catalog item"
		}
	}
	if v.DisplayGame == "" {
		v.DisplayGame = "Other"
	}

	res := valuation.Value(&item, card)
	qty := float64(item.Quantity)
	v.UnitValue = res.UnitValue
	v.Multiplier = res.Multiplier
	v.ValueSource = res.ValueSource
	v.TotalValue = round2(res.UnitValue * qty)
	v.CostBasis = round2(item.PurchasePrice * qty)
	v.GainLoss = round2(v.TotalValue - v.CostBasis)
	if v.CostBasis > 0 && res.ValueSource != models.ValueSourceNone {
		v.GainLossPct = round2(v.GainLoss / v.CostBasis * 100)
	}
	if res.ValueSource == models.ValueSourceNone {
		v.GainLoss = 0
	}
	return v
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// buildPortfolio loads, values, filters, and summarizes a user's portfolio.
func (h *Handlers) buildPortfolio(ctx context.Context, userID primitive.ObjectID, f portfolioFilter) (*models.PortfolioResponse, error) {
	query := bson.M{"user_id": userID}
	if f.ItemType != "" {
		query["item_type"] = f.ItemType
	}
	cur, err := h.db.Collection("portfolio_items").Find(ctx, query,
		options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(maxPortfolioItems))
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	items := make([]models.PortfolioItem, 0)
	if err := cur.All(ctx, &items); err != nil {
		return nil, err
	}

	ids := make([]primitive.ObjectID, 0, len(items))
	for _, it := range items {
		if it.CardID != nil {
			ids = append(ids, *it.CardID)
		}
	}
	cards, err := h.cardsByID(ctx, ids)
	if err != nil {
		return nil, err
	}

	views := make([]models.PortfolioItemView, 0, len(items))
	for _, it := range items {
		var card *models.Card
		if it.CardID != nil {
			card = cards[*it.CardID]
		}
		v := toPortfolioView(it, card)
		if f.Game != "" && !strings.EqualFold(v.DisplayGame, f.Game) {
			continue
		}
		views = append(views, v)
	}

	sortPortfolio(views, f.Sort)
	return &models.PortfolioResponse{Items: views, Summary: summarizePortfolio(views)}, nil
}

func sortPortfolio(views []models.PortfolioItemView, key string) {
	var less func(a, b models.PortfolioItemView) bool
	switch key {
	case "value":
		less = func(a, b models.PortfolioItemView) bool { return a.TotalValue > b.TotalValue }
	case "gain":
		less = func(a, b models.PortfolioItemView) bool { return a.GainLoss > b.GainLoss }
	case "name":
		less = func(a, b models.PortfolioItemView) bool {
			return strings.ToLower(a.DisplayName) < strings.ToLower(b.DisplayName)
		}
	default:
		return // already newest first
	}
	sort.SliceStable(views, func(i, j int) bool { return less(views[i], views[j]) })
}

func summarizePortfolio(views []models.PortfolioItemView) models.PortfolioSummary {
	s := models.PortfolioSummary{
		ByGame:     []models.AllocationSlice{},
		ByItemType: []models.AllocationSlice{},
		TopGainers: []models.PortfolioMover{},
		TopLosers:  []models.PortfolioMover{},
	}
	byGame := map[string]*models.AllocationSlice{}
	byType := map[string]*models.AllocationSlice{}
	valued := make([]models.PortfolioItemView, 0, len(views))

	for _, v := range views {
		s.ItemCount++
		s.UnitCount += v.Quantity
		s.TotalValue += v.TotalValue
		if v.ValueSource == models.ValueSourceNone {
			s.UnvaluedCount++
		} else {
			s.CostBasis += v.CostBasis
			valued = append(valued, v)
		}
		addSlice(byGame, v.DisplayGame, v.TotalValue)
		addSlice(byType, v.ItemType, v.TotalValue)
	}
	s.TotalValue = round2(s.TotalValue)
	s.CostBasis = round2(s.CostBasis)
	valuedTotal := 0.0
	for _, v := range valued {
		valuedTotal += v.TotalValue
	}
	s.GainLoss = round2(valuedTotal - s.CostBasis)
	if s.CostBasis > 0 {
		s.GainLossPct = round2(s.GainLoss / s.CostBasis * 100)
	}

	for _, m := range []struct {
		src map[string]*models.AllocationSlice
		dst *[]models.AllocationSlice
	}{{byGame, &s.ByGame}, {byType, &s.ByItemType}} {
		for _, a := range m.src {
			a.Value = round2(a.Value)
			*m.dst = append(*m.dst, *a)
		}
		sort.Slice(*m.dst, func(i, j int) bool { return (*m.dst)[i].Value > (*m.dst)[j].Value })
	}

	sort.SliceStable(valued, func(i, j int) bool { return valued[i].GainLoss > valued[j].GainLoss })
	for i := 0; i < len(valued) && len(s.TopGainers) < 3; i++ {
		if valued[i].GainLoss > 0 {
			s.TopGainers = append(s.TopGainers, mover(valued[i]))
		}
	}
	for i := len(valued) - 1; i >= 0 && len(s.TopLosers) < 3; i-- {
		if valued[i].GainLoss < 0 {
			s.TopLosers = append(s.TopLosers, mover(valued[i]))
		}
	}
	return s
}

func addSlice(m map[string]*models.AllocationSlice, key string, value float64) {
	if m[key] == nil {
		m[key] = &models.AllocationSlice{Key: key}
	}
	m[key].Value += value
	m[key].Count++
}

func mover(v models.PortfolioItemView) models.PortfolioMover {
	return models.PortfolioMover{ID: v.ID, DisplayName: v.DisplayName, GainLoss: v.GainLoss, GainLossPct: v.GainLossPct}
}

// GetPortfolio lists the user's portfolio with valuation and summary
func (h *Handlers) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	q := r.URL.Query()
	f := portfolioFilter{Game: strings.TrimSpace(q.Get("game")), ItemType: q.Get("item_type"), Sort: q.Get("sort")}
	if f.ItemType != "" && !validItemTypes[f.ItemType] {
		h.sendError(w, "Invalid item_type", http.StatusBadRequest, nil)
		return
	}
	if len(f.Game) > 60 {
		h.sendError(w, "Invalid game", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.buildPortfolio(ctx, userID, f)
	if err != nil {
		h.sendError(w, "Error retrieving portfolio", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, resp)
}

// loadPortfolioItemView fetches one owned item with valuation; ok=false means not found.
func (h *Handlers) loadPortfolioItemView(ctx context.Context, userID, itemID primitive.ObjectID) (*models.PortfolioItemView, bool, error) {
	var item models.PortfolioItem
	err := h.db.Collection("portfolio_items").FindOne(ctx, bson.M{"_id": itemID, "user_id": userID}).Decode(&item)
	if err == mongo.ErrNoDocuments {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var card *models.Card
	if item.CardID != nil {
		cards, err := h.cardsByID(ctx, []primitive.ObjectID{*item.CardID})
		if err != nil {
			return nil, false, err
		}
		card = cards[*item.CardID]
	}
	v := toPortfolioView(item, card)
	return &v, true, nil
}

// GetPortfolioItem returns one owned item
func (h *Handlers) GetPortfolioItem(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	itemID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid item ID", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	v, ok, err := h.loadPortfolioItemView(ctx, userID, itemID)
	if err != nil {
		h.sendError(w, "Error retrieving item", http.StatusInternalServerError, nil)
		return
	}
	if !ok {
		h.sendError(w, "Item not found", http.StatusNotFound, nil)
		return
	}
	sendJSON(w, http.StatusOK, v)
}

// parsePortfolioRequest decodes and validates an item payload, resolving the card.
func (h *Handlers) parsePortfolioRequest(w http.ResponseWriter, r *http.Request, ctx context.Context) (*models.PortfolioItem, bool) {
	var req models.PortfolioItemRequest
	if err := decodeJSON(w, r, &req); err != nil {
		h.sendError(w, "Invalid request body", http.StatusBadRequest, nil)
		return nil, false
	}
	purchaseDate, err := validatePortfolioItem(&req)
	if err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest, nil)
		return nil, false
	}

	item := &models.PortfolioItem{
		ItemType:       req.ItemType,
		Quantity:       req.Quantity,
		Condition:      req.Condition,
		Finish:         req.Finish,
		Language:       req.Language,
		Grading:        req.Grading,
		PurchasePrice:  req.PurchasePrice,
		PurchaseDate:   purchaseDate,
		PurchaseSource: req.PurchaseSource,
		ManualValue:    req.ManualValue,
		Notes:          req.Notes,
	}

	if req.CardID != "" {
		cardID, err := primitive.ObjectIDFromHex(req.CardID)
		if err != nil {
			h.sendError(w, "Invalid card_id", http.StatusBadRequest, nil)
			return nil, false
		}
		var card models.Card
		err = h.db.Collection("cards").FindOne(ctx, bson.M{"_id": cardID}).Decode(&card)
		if err == mongo.ErrNoDocuments {
			h.sendError(w, "Card not found", http.StatusBadRequest, nil)
			return nil, false
		}
		if err != nil {
			h.sendError(w, "Error validating card", http.StatusInternalServerError, nil)
			return nil, false
		}
		if (card.Category == "sealed") != (req.ItemType == models.ItemTypeSealed) {
			h.sendError(w, "item_type does not match the selected product", http.StatusBadRequest, nil)
			return nil, false
		}
		item.CardID = &cardID
	} else {
		item.CustomName = req.CustomName
		item.CustomGame = req.CustomGame
		item.CustomSet = req.CustomSet
		item.CustomImageURL = req.CustomImageURL
	}
	return item, true
}

// CreatePortfolioItem adds an item to the user's portfolio
func (h *Handlers) CreatePortfolioItem(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, ok := h.parsePortfolioRequest(w, r, ctx)
	if !ok {
		return
	}

	coll := h.db.Collection("portfolio_items")
	count, err := coll.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		h.sendError(w, "Error checking portfolio", http.StatusInternalServerError, nil)
		return
	}
	if count >= maxPortfolioItems {
		h.sendError(w, "Portfolio item limit reached", http.StatusBadRequest, nil)
		return
	}

	now := time.Now().UTC()
	item.UserID = userID
	item.CreatedAt = now
	item.UpdatedAt = now
	res, err := coll.InsertOne(ctx, item)
	if err != nil {
		h.sendError(w, "Error saving item", http.StatusInternalServerError, nil)
		return
	}
	item.ID = res.InsertedID.(primitive.ObjectID)

	v, _, err := h.loadPortfolioItemView(ctx, userID, item.ID)
	if err != nil || v == nil {
		h.sendError(w, "Item saved but could not be loaded", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusCreated, v)
}

// UpdatePortfolioItem replaces an owned item's editable fields
func (h *Handlers) UpdatePortfolioItem(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	itemID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid item ID", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, ok := h.parsePortfolioRequest(w, r, ctx)
	if !ok {
		return
	}

	set := bson.M{
		"item_type":        item.ItemType,
		"quantity":         item.Quantity,
		"purchase_price":   item.PurchasePrice,
		"custom_name":      item.CustomName,
		"custom_game":      item.CustomGame,
		"custom_set":       item.CustomSet,
		"custom_image_url": item.CustomImageURL,
		"condition":        item.Condition,
		"finish":           item.Finish,
		"language":         item.Language,
		"purchase_source":  item.PurchaseSource,
		"notes":            item.Notes,
		"updated_at":       time.Now().UTC(),
	}
	unset := bson.M{}
	optional := map[string]interface{}{
		"card_id":       item.CardID,
		"grading":       item.Grading,
		"purchase_date": item.PurchaseDate,
		"manual_value":  item.ManualValue,
	}
	for key, val := range optional {
		switch v := val.(type) {
		case *primitive.ObjectID:
			if v == nil {
				unset[key] = ""
			} else {
				set[key] = *v
			}
		case *models.Grading:
			if v == nil {
				unset[key] = ""
			} else {
				set[key] = v
			}
		case *time.Time:
			if v == nil {
				unset[key] = ""
			} else {
				set[key] = *v
			}
		case *float64:
			if v == nil {
				unset[key] = ""
			} else {
				set[key] = *v
			}
		}
	}
	update := bson.M{"$set": set}
	if len(unset) > 0 {
		update["$unset"] = unset
	}

	res, err := h.db.Collection("portfolio_items").UpdateOne(ctx, bson.M{"_id": itemID, "user_id": userID}, update)
	if err != nil {
		h.sendError(w, "Error updating item", http.StatusInternalServerError, nil)
		return
	}
	if res.MatchedCount == 0 {
		h.sendError(w, "Item not found", http.StatusNotFound, nil)
		return
	}

	v, _, err := h.loadPortfolioItemView(ctx, userID, itemID)
	if err != nil || v == nil {
		h.sendError(w, "Item updated but could not be loaded", http.StatusInternalServerError, nil)
		return
	}
	sendJSON(w, http.StatusOK, v)
}

// DeletePortfolioItem removes an owned item
func (h *Handlers) DeletePortfolioItem(w http.ResponseWriter, r *http.Request) {
	userID, _, err := userIDFromRequest(r)
	if err != nil {
		h.sendError(w, "Unauthorized", http.StatusUnauthorized, nil)
		return
	}
	itemID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid item ID", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := h.db.Collection("portfolio_items").DeleteOne(ctx, bson.M{"_id": itemID, "user_id": userID})
	if err != nil {
		h.sendError(w, "Error deleting item", http.StatusInternalServerError, nil)
		return
	}
	if res.DeletedCount == 0 {
		h.sendError(w, "Item not found", http.StatusNotFound, nil)
		return
	}
	sendJSON(w, http.StatusOK, map[string]string{"message": "Item deleted"})
}
