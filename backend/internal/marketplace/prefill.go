package marketplace

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/jamesc159/monmetrics/internal/models"
	"github.com/jamesc159/monmetrics/internal/valuation"
)

// Title length limits per provider.
const (
	EbayTitleMax      = 80
	TCGPlayerTitleMax = 255
)

// Placeholder eBay category IDs; verify against the eBay Taxonomy API before going live.
const (
	ebayCategorySingles = "183454"
	ebayCategorySealed  = "183456"
)

// ComputeStats summarizes sold prices, trimming IQR outliers before the median.
func ComputeStats(sold []models.SoldListing) models.CompsStats {
	if len(sold) == 0 {
		return models.CompsStats{}
	}
	prices := make([]float64, len(sold))
	for i, s := range sold {
		prices[i] = s.Price
	}
	sort.Float64s(prices)

	stats := models.CompsStats{
		Count: len(prices),
		Min:   prices[0],
		Max:   prices[len(prices)-1],
	}
	sum := 0.0
	for _, p := range prices {
		sum += p
	}
	stats.Average = round2(sum / float64(len(prices)))

	trimmed := prices
	if len(prices) >= 4 {
		q1, q3 := percentile(prices, 0.25), percentile(prices, 0.75)
		iqr := q3 - q1
		lo, hi := q1-1.5*iqr, q3+1.5*iqr
		trimmed = trimmed[:0:0]
		for _, p := range prices {
			if p >= lo && p <= hi {
				trimmed = append(trimmed, p)
			}
		}
	}
	stats.Median = round2(percentile(trimmed, 0.5))
	stats.SuggestedPrice = stats.Median
	return stats
}

func percentile(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	pos := q * float64(len(sorted)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	return sorted[lo] + (sorted[hi]-sorted[lo])*(pos-float64(lo))
}

func round2(v float64) float64 {
	return math.Round(v*100) / 100
}

// BuildTitle joins non-empty parts, dropping trailing parts to fit max length.
func BuildTitle(parts []string, max int) string {
	title := ""
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		candidate := p
		if title != "" {
			candidate = title + " " + p
		}
		if len(candidate) > max {
			continue
		}
		title = candidate
	}
	if title == "" && len(parts) > 0 {
		title = strings.TrimSpace(parts[0])
		if len(title) > max {
			title = title[:max]
		}
	}
	return title
}

// ConditionFor maps an item's condition/grade to the provider's vocabulary.
func ConditionFor(provider string, item *models.PortfolioItem) string {
	switch item.ItemType {
	case models.ItemTypeSealed:
		if provider == models.ProviderEbay {
			return "New"
		}
		return "Sealed"
	case models.ItemTypeGradedCard:
		if provider == models.ProviderEbay {
			return "Graded"
		}
		return "Near Mint"
	}
	label := valuation.ConditionLabels[item.Condition]
	if label == "" {
		label = "Near Mint"
	}
	if provider == models.ProviderEbay {
		return "Ungraded - " + label
	}
	return label
}

// BuildDraft creates a pre-filled listing draft for an owned item.
func BuildDraft(provider string, view *models.PortfolioItemView, stats models.CompsStats) models.ListingDraft {
	item := &view.PortfolioItem
	card := view.Card

	number, rarity := "", ""
	if card != nil {
		number, rarity = card.Number, card.Rarity
	}
	gradeLabel := ""
	if item.ItemType == models.ItemTypeGradedCard && item.Grading != nil {
		gradeLabel = fmt.Sprintf("%s %s", item.Grading.Company, formatGrade(item.Grading.Grade))
	}

	parts := []string{view.DisplayName}
	if gradeLabel != "" {
		parts = append(parts, gradeLabel)
	}
	parts = append(parts, number, view.DisplaySet, rarity)
	if finishLabel(item.Finish) != "" {
		parts = append(parts, finishLabel(item.Finish))
	}
	parts = append(parts, view.DisplayGame)
	if item.ItemType == models.ItemTypeSealed {
		parts = append(parts, "Factory Sealed")
	} else if gradeLabel == "" && item.Condition != "" {
		parts = append(parts, item.Condition)
	}

	max := TCGPlayerTitleMax
	if provider == models.ProviderEbay {
		max = EbayTitleMax
	}

	price := stats.SuggestedPrice
	if price <= 0 {
		price = view.UnitValue
	}

	draft := models.ListingDraft{
		PortfolioItemID: item.ID.Hex(),
		Provider:        provider,
		Title:           BuildTitle(parts, max),
		Description:     buildDescription(view, gradeLabel),
		Price:           round2(price),
		Quantity:        item.Quantity,
		MaxQuantity:     item.Quantity,
		Condition:       ConditionFor(provider, item),
		Format:          "fixed_price",
		ImageURLs:       []string{},
		ItemSpecifics:   buildSpecifics(view, gradeLabel),
	}
	if view.DisplayImageURL != "" {
		draft.ImageURLs = append(draft.ImageURLs, view.DisplayImageURL)
	}
	if provider == models.ProviderEbay {
		draft.Duration = "GTC"
		draft.ShippingPrice = 4.99
		draft.CategoryID = ebayCategorySingles
		if item.ItemType == models.ItemTypeSealed {
			draft.CategoryID = ebayCategorySealed
		}
	} else {
		draft.CategoryID = view.DisplayGame
	}
	return draft
}

func buildDescription(view *models.PortfolioItemView, gradeLabel string) string {
	item := &view.PortfolioItem
	var b strings.Builder
	fmt.Fprintf(&b, "%s", view.DisplayName)
	if view.DisplaySet != "" {
		fmt.Fprintf(&b, " from %s", view.DisplaySet)
	}
	if view.DisplayGame != "" {
		fmt.Fprintf(&b, " (%s)", view.DisplayGame)
	}
	b.WriteString(".\n\n")
	switch item.ItemType {
	case models.ItemTypeGradedCard:
		fmt.Fprintf(&b, "Professionally graded %s.", gradeLabel)
		if item.Grading != nil && item.Grading.CertNumber != "" {
			fmt.Fprintf(&b, " Cert #%s.", item.Grading.CertNumber)
		}
	case models.ItemTypeSealed:
		b.WriteString("Factory sealed, never opened.")
	default:
		fmt.Fprintf(&b, "Condition: %s.", valuation.ConditionLabels[item.Condition])
	}
	if item.Language != "" && item.Language != "English" {
		fmt.Fprintf(&b, " Language: %s.", item.Language)
	}
	b.WriteString("\n\nShips securely in a sleeve and top loader / protective packaging.")
	return b.String()
}

func buildSpecifics(view *models.PortfolioItemView, gradeLabel string) []models.ItemSpecific {
	item := &view.PortfolioItem
	specs := []models.ItemSpecific{}
	add := func(name, value string) {
		if strings.TrimSpace(value) != "" {
			specs = append(specs, models.ItemSpecific{Name: name, Value: value})
		}
	}
	add("Game", view.DisplayGame)
	add("Set", view.DisplaySet)
	add("Card Name", view.DisplayName)
	if view.Card != nil {
		add("Card Number", view.Card.Number)
		add("Rarity", view.Card.Rarity)
	}
	add("Finish", finishLabel(item.Finish))
	add("Language", item.Language)
	if item.ItemType == models.ItemTypeGradedCard && item.Grading != nil {
		add("Graded", "Yes")
		add("Professional Grader", item.Grading.Company)
		add("Grade", formatGrade(item.Grading.Grade))
		add("Certification Number", item.Grading.CertNumber)
	} else if item.ItemType == models.ItemTypeRawCard {
		add("Graded", "No")
		add("Card Condition", valuation.ConditionLabels[item.Condition])
	}
	return specs
}

func finishLabel(f string) string {
	switch f {
	case "holo":
		return "Holo"
	case "reverse_holo":
		return "Reverse Holo"
	case "foil":
		return "Foil"
	case "1st_edition":
		return "1st Edition"
	}
	return ""
}
