package scraper

import (
	"fmt"
	"strings"

	"github.com/jamesc159/monmetrics/internal/models"
)

type SearchPlan struct {
	RawKeywords    string
	GradedKeywords string
}

func BuildSearchPlan(card models.Card) SearchPlan {
	if override := strings.TrimSpace(card.EbayQuery); override != "" {
		return SearchPlan{RawKeywords: override, GradedKeywords: override}
	}
	parts := []string{strings.TrimSpace(card.Name), strings.TrimSpace(card.Number)}
	if card.Set != "" && !strings.Contains(strings.ToLower(card.Name), strings.ToLower(card.Set)) {
		parts = append(parts, strings.TrimSpace(card.Set))
	}
	base := strings.Join(nonEmpty(parts), " ")
	if card.Category == "sealed" {
		// Sealed product listings are never graded, so one query is enough.
		return SearchPlan{RawKeywords: base, GradedKeywords: base}
	}
	return SearchPlan{
		RawKeywords:    fmt.Sprintf("%s -psa -bgs -beckett -cgc -sgc -tag -ace -graded", base),
		GradedKeywords: base,
	}
}

func nonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			out = append(out, value)
		}
	}
	return out
}
