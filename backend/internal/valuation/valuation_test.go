package valuation

import (
	"testing"

	"github.com/jamesc159/monmetrics/internal/models"
)

func ptr(f float64) *float64 { return &f }

func TestValuePrecedence(t *testing.T) {
	card := &models.Card{CurrentPrice: 100}

	cases := []struct {
		name   string
		item   models.PortfolioItem
		card   *models.Card
		want   float64
		source string
	}{
		{"manual overrides market", models.PortfolioItem{ItemType: models.ItemTypeRawCard, Condition: "NM", ManualValue: ptr(42)}, card, 42, models.ValueSourceManual},
		{"raw NM", models.PortfolioItem{ItemType: models.ItemTypeRawCard, Condition: "NM"}, card, 100, models.ValueSourceMarket},
		{"raw LP", models.PortfolioItem{ItemType: models.ItemTypeRawCard, Condition: "LP"}, card, 85, models.ValueSourceMarket},
		{"raw DMG", models.PortfolioItem{ItemType: models.ItemTypeRawCard, Condition: "DMG"}, card, 30, models.ValueSourceMarket},
		{"PSA 10", models.PortfolioItem{ItemType: models.ItemTypeGradedCard, Grading: &models.Grading{Company: "PSA", Grade: 10}}, card, 400, models.ValueSourceGradedEstimate},
		{"BGS 9.5", models.PortfolioItem{ItemType: models.ItemTypeGradedCard, Grading: &models.Grading{Company: "BGS", Grade: 9.5}}, card, 250, models.ValueSourceGradedEstimate},
		{"sealed", models.PortfolioItem{ItemType: models.ItemTypeSealed}, card, 100, models.ValueSourceMarket},
		{"custom no value", models.PortfolioItem{ItemType: models.ItemTypeRawCard}, nil, 0, models.ValueSourceNone},
		{"custom manual", models.PortfolioItem{ItemType: models.ItemTypeSealed, ManualValue: ptr(12.5)}, nil, 12.5, models.ValueSourceManual},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Value(&tc.item, tc.card)
			if got.UnitValue != tc.want || got.ValueSource != tc.source {
				t.Fatalf("got %v/%s, want %v/%s", got.UnitValue, got.ValueSource, tc.want, tc.source)
			}
		})
	}
}

func TestGradeMultiplierFallback(t *testing.T) {
	if m := GradeMultiplier("BGS", 6.5); m != 0.5 {
		t.Fatalf("below table should floor to 0.5, got %v", m)
	}
	if m := GradeMultiplier("CGC", 9.0); m != 1.4 {
		t.Fatalf("exact grade, got %v", m)
	}
	if m := GradeMultiplier("UNKNOWN", 10); m != 1.0 {
		t.Fatalf("unknown company, got %v", m)
	}
}
