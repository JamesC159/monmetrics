// Package valuation estimates the market value of portfolio items.
package valuation

import (
	"math"

	"github.com/jamesc159/monmetrics/internal/models"
)

// ConditionMultipliers scale a raw card's market price by condition.
var ConditionMultipliers = map[string]float64{
	"NM":  1.00,
	"LP":  0.85,
	"MP":  0.70,
	"HP":  0.50,
	"DMG": 0.30,
}

// ConditionLabels maps condition codes to display names.
var ConditionLabels = map[string]string{
	"NM":  "Near Mint",
	"LP":  "Lightly Played",
	"MP":  "Moderately Played",
	"HP":  "Heavily Played",
	"DMG": "Damaged",
}

// Grade multipliers relative to the raw near-mint price. Estimates only; there is
// no graded price feed yet.
var gradeMultipliers = map[string]map[float64]float64{
	"PSA": {10: 4.0, 9: 1.8, 8: 1.3, 7: 1.1, 6: 0.95, 5: 0.85, 4: 0.75, 3: 0.65, 2: 0.55, 1: 0.45},
	"BGS": {10: 6.0, 9.5: 2.5, 9: 1.6, 8.5: 1.3, 8: 1.15, 7.5: 1.05, 7: 1.0},
	"CGC": {10: 3.0, 9.5: 1.9, 9: 1.4, 8.5: 1.2, 8: 1.1, 7.5: 1.0, 7: 0.95},
	"SGC": {10: 3.2, 9.5: 2.0, 9: 1.5, 8.5: 1.25, 8: 1.1, 7.5: 1.0, 7: 0.95},
}

// GradeMultiplier returns the multiplier for a company/grade, falling back to the
// nearest lower listed grade and finally a conservative floor.
func GradeMultiplier(company string, grade float64) float64 {
	table, ok := gradeMultipliers[company]
	if !ok {
		return 1.0
	}
	if m, ok := table[grade]; ok {
		return m
	}
	best, bestGrade := 0.0, -1.0
	for g, m := range table {
		if g <= grade && g > bestGrade {
			best, bestGrade = m, g
		}
	}
	if bestGrade < 0 {
		return 0.5
	}
	return best
}

// Result is the computed valuation for one portfolio item.
type Result struct {
	UnitValue   float64
	Multiplier  float64
	ValueSource string
}

// Value computes the per-unit value. card may be nil for custom or orphaned items.
// Precedence: manual value > graded estimate > market (condition-adjusted) > none.
func Value(item *models.PortfolioItem, card *models.Card) Result {
	if item.ManualValue != nil {
		return Result{UnitValue: *item.ManualValue, Multiplier: 1, ValueSource: models.ValueSourceManual}
	}
	if card == nil || card.CurrentPrice <= 0 {
		return Result{ValueSource: models.ValueSourceNone}
	}
	switch item.ItemType {
	case models.ItemTypeGradedCard:
		if item.Grading == nil {
			return Result{ValueSource: models.ValueSourceNone}
		}
		m := GradeMultiplier(item.Grading.Company, item.Grading.Grade)
		return Result{UnitValue: round2(card.CurrentPrice * m), Multiplier: m, ValueSource: models.ValueSourceGradedEstimate}
	case models.ItemTypeRawCard:
		m, ok := ConditionMultipliers[item.Condition]
		if !ok {
			m = 1
		}
		return Result{UnitValue: round2(card.CurrentPrice * m), Multiplier: m, ValueSource: models.ValueSourceMarket}
	default:
		return Result{UnitValue: round2(card.CurrentPrice), Multiplier: 1, ValueSource: models.ValueSourceMarket}
	}
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
