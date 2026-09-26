package marketplace

import (
	"context"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jamesc159/monmetrics/internal/models"
)

func TestComputeStatsTrimsOutliers(t *testing.T) {
	prices := []float64{10, 11, 12, 13, 14, 500}
	sold := make([]models.SoldListing, len(prices))
	for i, p := range prices {
		sold[i] = models.SoldListing{Price: p}
	}
	s := ComputeStats(sold)
	if s.Count != 6 || s.Min != 10 || s.Max != 500 {
		t.Fatalf("unexpected stats %+v", s)
	}
	if s.Median != 12 {
		t.Fatalf("median should ignore outlier, got %v", s.Median)
	}
	if s.SuggestedPrice != s.Median {
		t.Fatalf("suggested should equal median")
	}
}

func TestComputeStatsEmpty(t *testing.T) {
	if s := ComputeStats(nil); s.Count != 0 || s.SuggestedPrice != 0 {
		t.Fatalf("expected zero stats, got %+v", s)
	}
}

func TestBuildTitleRespectsMax(t *testing.T) {
	title := BuildTitle([]string{"Charizard VMAX", "PSA 10", "020/073", "Champions Path", "VMAX", "Pokemon", "Extra words that overflow the limit"}, EbayTitleMax)
	if len(title) > EbayTitleMax {
		t.Fatalf("title too long: %d", len(title))
	}
	if !strings.HasPrefix(title, "Charizard VMAX PSA 10") {
		t.Fatalf("unexpected title %q", title)
	}
}

func TestBuildDraftGradedEbay(t *testing.T) {
	cardID := primitive.NewObjectID()
	view := &models.PortfolioItemView{
		PortfolioItem: models.PortfolioItem{
			ID:       primitive.NewObjectID(),
			CardID:   &cardID,
			ItemType: models.ItemTypeGradedCard,
			Quantity: 1,
			Grading:  &models.Grading{Company: "PSA", Grade: 10, CertNumber: "12345678"},
		},
		Card:        &models.Card{Number: "020/073", Rarity: "VMAX"},
		DisplayName: "Charizard VMAX",
		DisplaySet:  "Champions Path",
		DisplayGame: "Pokemon",
		UnitValue:   360,
	}
	d := BuildDraft(models.ProviderEbay, view, models.CompsStats{SuggestedPrice: 350})
	if d.Price != 350 || d.Condition != "Graded" || d.CategoryID == "" {
		t.Fatalf("unexpected draft %+v", d)
	}
	found := false
	for _, s := range d.ItemSpecifics {
		if s.Name == "Certification Number" && s.Value == "12345678" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing cert number specific")
	}
	d2 := BuildDraft(models.ProviderEbay, view, models.CompsStats{})
	if d2.Price != 360 {
		t.Fatalf("should fall back to unit value, got %v", d2.Price)
	}
}

func TestMockExchangeCode(t *testing.T) {
	p := newMockProvider(models.ProviderEbay, "eBay", "http://localhost:8080")
	if _, _, err := p.ExchangeCode(context.Background(), "bogus"); err == nil {
		t.Fatal("expected error for non-mock code")
	}
	code := NewMockCode()
	tok, acct, err := p.ExchangeCode(context.Background(), code)
	if err != nil || tok.AccessToken == "" || acct.Username == "" {
		t.Fatalf("exchange failed: %v", err)
	}
	_, acct2, _ := p.ExchangeCode(context.Background(), code)
	if acct.ExternalUserID != acct2.ExternalUserID {
		t.Fatal("account id should be deterministic per code")
	}
}
