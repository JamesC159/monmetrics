package handlers

import (
	"context"
	"net/http"
	"sort"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/jamesc159/monmetrics/internal/features"
	"github.com/jamesc159/monmetrics/internal/models"
)

const (
	defaultSalesPerPage = 25
	maxSalesPerPage     = 100
	// gradeWindow is the widest rolling window the scraper stores, so it best answers "does this grade have data?".
	gradeWindow = "5y"
)

// CardSalesResponse is a page of individual scraped sales for one card.
type CardSalesResponse struct {
	CardID  string            `json:"card_id"`
	Grade   string            `json:"grade"`
	Source  string            `json:"source"`
	Total   int               `json:"total"`
	Page    int               `json:"page"`
	PerPage int               `json:"per_page"`
	Sales   []models.SoldSale `json:"sales"`
}

// CardGradeInfo describes one price series that has data for a card. Despite the name it
// covers both eBay grades (psa-10) and TCGplayer printings (reverse-holofoil).
type CardGradeInfo struct {
	Grade      string    `bson:"grade" json:"grade"`
	Source     string    `bson:"source" json:"source"`
	SaleCount  int       `bson:"sale_count" json:"sale_count"`
	Average    float64   `bson:"average" json:"average"`
	LastSaleAt time.Time `bson:"last_sale_at" json:"last_sale_at"`
}

// GetCardSales returns the raw scraped sold listings backing a card's price history.
func (h *Handlers) GetCardSales(w http.ResponseWriter, r *http.Request) {
	if !h.features.Enabled(features.KeyEbaySoldListings) {
		h.sendError(w, "Sold-listing data is not available", http.StatusNotFound, nil)
		return
	}

	cardID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid card ID", http.StatusBadRequest, nil)
		return
	}

	q := r.URL.Query()
	grade := q.Get("grade")
	if grade == "" {
		grade = "raw"
	}
	if !validGradePattern.MatchString(grade) {
		h.sendError(w, "Invalid grade", http.StatusBadRequest, nil)
		return
	}
	source := q.Get("source")
	if source == "" {
		source = "ebay"
	}
	if !validChartSources[source] {
		h.sendError(w, "Invalid source", http.StatusBadRequest, nil)
		return
	}

	page := 1
	if p, err := strconv.Atoi(q.Get("page")); err == nil && p > 0 {
		page = p
	}
	perPage := defaultSalesPerPage
	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 && l <= maxSalesPerPage {
		perPage = l
	}

	filter := bson.M{"card_id": cardID, "grade": grade}
	if source != "all" {
		filter["source"] = source
	}
	// Irrelevant matches (lots, proxies, wrong card) are stored but hidden unless explicitly requested.
	if q.Get("include_excluded") != "true" {
		filter["excluded"] = false
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	collection := h.db.Collection("sold_listings")
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		h.sendError(w, "Error counting sales", http.StatusInternalServerError, nil)
		return
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "sold_at", Value: -1}}).
		SetSkip(int64((page - 1) * perPage)).
		SetLimit(int64(perPage))

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		h.sendError(w, "Error retrieving sales", http.StatusInternalServerError, nil)
		return
	}
	defer cursor.Close(ctx)

	sales := make([]models.SoldSale, 0)
	if err := cursor.All(ctx, &sales); err != nil {
		h.sendError(w, "Error decoding sales", http.StatusInternalServerError, nil)
		return
	}

	sendJSON(w, http.StatusOK, CardSalesResponse{
		CardID:  cardID.Hex(),
		Grade:   grade,
		Source:  source,
		Total:   int(total),
		Page:    page,
		PerPage: perPage,
		Sales:   sales,
	})
}

// GetCardGrades lists the grades that have scraped data for a card, so the UI
// only offers grade options that will actually render a chart.
func (h *Handlers) GetCardGrades(w http.ResponseWriter, r *http.Request) {
	cardID, err := primitive.ObjectIDFromHex(r.PathValue("id"))
	if err != nil {
		h.sendError(w, "Invalid card ID", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Derived from the price points themselves rather than price_summaries, so it answers
	// "which series does this card actually have data for" regardless of the source that
	// produced them - eBay grades and TCGplayer printings alike.
	pipeline := []bson.M{
		{"$match": bson.M{"card_id": cardID, "grade": bson.M{"$nin": bson.A{nil, ""}}}},
		{"$group": bson.M{
			"_id":     bson.M{"grade": "$grade", "source": "$source"},
			"points":  bson.M{"$sum": 1},
			"latest":  bson.M{"$max": "$timestamp"},
			"average": bson.M{"$avg": "$price"},
		}},
		{"$project": bson.M{
			"_id":          0,
			"grade":        "$_id.grade",
			"source":       "$_id.source",
			"sale_count":   "$points",
			"last_sale_at": "$latest",
			"average":      bson.M{"$round": bson.A{"$average", 2}},
		}},
		{"$sort": bson.M{"sale_count": -1}},
	}

	cursor, err := h.db.Collection("prices").Aggregate(ctx, pipeline)
	if err != nil {
		h.sendError(w, "Error retrieving series", http.StatusInternalServerError, nil)
		return
	}
	defer cursor.Close(ctx)

	grades := make([]CardGradeInfo, 0)
	if err := cursor.All(ctx, &grades); err != nil {
		h.sendError(w, "Error decoding series", http.StatusInternalServerError, nil)
		return
	}

	// Default-worthy series first, then whichever has the most points behind it.
	sort.SliceStable(grades, func(i, j int) bool {
		li, lj := seriesPriority(grades[i].Grade), seriesPriority(grades[j].Grade)
		if li != lj {
			return li < lj
		}
		return grades[i].SaleCount > grades[j].SaleCount
	})

	sendJSON(w, http.StatusOK, grades)
}

// defaultSeriesOrder ranks which series a card should open on.
var defaultSeriesOrder = []string{"raw", "normal", "foil", "holofoil"}

func seriesPriority(series string) int {
	for i, name := range defaultSeriesOrder {
		if name == series {
			return i
		}
	}
	return len(defaultSeriesOrder)
}
