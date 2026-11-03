package pricing

import (
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jamesc159/monmetrics/internal/catalog"
	"github.com/jamesc159/monmetrics/internal/models"
)

func price(v float64) *float64 { return &v }

func TestPrintingsByCard(t *testing.T) {
	charizard, pikachu := primitive.NewObjectID(), primitive.NewObjectID()
	cards := map[string]primitive.ObjectID{"100": charizard, "200": pikachu}

	prices := []catalog.ProductPrice{
		{ProductID: 100, SubTypeName: "Reverse Holofoil", MarketPrice: price(4.5)},
		{ProductID: 100, SubTypeName: "1st Edition", MarketPrice: price(80)},
		{ProductID: 100, SubTypeName: "Holofoil", MarketPrice: price(12.345)},
		{ProductID: 100, SubTypeName: "Unlimited", MarketPrice: price(20)},
		{ProductID: 200, SubTypeName: "Normal"}, // listed, but TCGplayer has no figure
		{ProductID: 200, SubTypeName: "Foil", MarketPrice: price(1.1)},
		{ProductID: 999, SubTypeName: "Normal", MarketPrice: price(3)}, // not a card we carry
	}

	got := printingsByCard(cards, prices)
	want := map[primitive.ObjectID][]models.Printing{
		charizard: {
			{Series: "holofoil", MarketPrice: 12.35},
			{Series: "reverse-holofoil", MarketPrice: 4.5},
			// Neither is a primary printing, so they keep the price file's order.
			{Series: "1st-edition", MarketPrice: 80},
			{Series: "unlimited", MarketPrice: 20},
		},
		pikachu: {
			{Series: "normal"},
			{Series: "foil", MarketPrice: 1.1},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("printingsByCard =\n%v\nwant\n%v", got, want)
	}

	// The headline skips an unpriced printing rather than reporting $0.
	if h := headlinePrice(got[pikachu]); h != 1.1 {
		t.Errorf("pikachu headline = %v, want 1.1", h)
	}
	if h := headlinePrice(got[charizard]); h != 12.35 {
		t.Errorf("charizard headline = %v, want 12.35", h)
	}
	if h := headlinePrice([]models.Printing{{Series: "normal"}}); h != 0 {
		t.Errorf("unpriced headline = %v, want 0", h)
	}
}

func TestLatestPrintings(t *testing.T) {
	older := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 11, 3, 0, 0, 0, 0, time.UTC)

	rows := []seriesRow{
		{Series: "foil", Price: 2, Timestamp: older},
		{Series: "foil", Price: 2.499, Timestamp: newer},
		{Series: "unlimited", Price: 9, Timestamp: newer},
		{Series: "1st-edition", Price: 30, Timestamp: newer},
		{Series: "normal", Price: 0.5, Timestamp: newer},
		{Series: "etched", Price: 7, Timestamp: older}, // not priced on the latest day
	}

	got := latestPrintings(rows)
	want := []models.Printing{
		{Series: "normal", MarketPrice: 0.5},
		{Series: "foil", MarketPrice: 2.5},
		{Series: "1st-edition", MarketPrice: 30},
		{Series: "unlimited", MarketPrice: 9},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("latestPrintings = %v, want %v", got, want)
	}

	if got := latestPrintings(nil); len(got) != 0 {
		t.Errorf("latestPrintings(nil) = %v, want none", got)
	}
}
