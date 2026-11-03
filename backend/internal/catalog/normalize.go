package catalog

import (
	"sort"
	"strconv"
	"strings"

	"github.com/jamesc159/monmetrics/internal/models"
)

// sealedKeywords map a product name fragment to a product_type for non-single items.
var sealedKeywords = []struct {
	match       string
	productType string
}{
	{"elite trainer box", "elite_trainer_box"},
	{"booster box case", "case"},
	{"booster case", "case"},
	{"booster box", "booster_box"},
	{"booster bundle", "bundle"},
	{"booster pack", "booster_pack"},
	{"build & battle", "build_and_battle"},
	{"collection box", "collection_box"},
	{"premium collection", "collection_box"},
	{"ultra premium", "collection_box"},
	{"gift set", "collection_box"},
	{"starter deck", "deck"},
	{"theme deck", "deck"},
	{"structure deck", "deck"},
	{"commander deck", "deck"},
	{"deck box", "deck"},
	{"bundle", "bundle"},
	{"blister", "blister"},
	{"tin", "tin"},
	{"collector box", "collection_box"},
	{"display", "booster_box"},
	{"case", "case"},
	{"pack", "booster_pack"},
	{"box", "box"},
}

var searchStopWords = map[string]bool{
	"the": true, "and": true, "of": true, "a": true, "an": true, "for": true, "with": true,
}

// NormalizeProduct converts a provider product into a Card document. Market and
// scheduling fields are left zero; the storage layer owns those.
func NormalizeProduct(source string, category Category, group Group, product Product) models.Card {
	game := firstNonEmpty(category.DisplayName, category.Name)
	set := strings.TrimSpace(group.Name)
	number := product.extended("Number")
	rarity := product.extended("Rarity")
	name := firstNonEmpty(product.CleanName, product.Name)

	productType, kind := classifyProduct(name, number, rarity)

	card := models.Card{
		Name:             name,
		Set:              set,
		Game:             game,
		Category:         kind,
		Rarity:           rarity,
		Number:           number,
		ImageURL:         strings.TrimSpace(product.ImageURL),
		Description:      product.extended("Card Text"),
		CatalogSource:    source,
		ExternalID:       strconv.Itoa(product.ProductID),
		ExternalGroupID:  strconv.Itoa(group.GroupID),
		ExternalCategory: strconv.Itoa(category.CategoryID),
		ProductType:      productType,
		Language:         "English",
		ProductURL:       strings.TrimSpace(product.URL),
		Active:           true,
	}
	card.SearchTerms = BuildSearchTerms(card)
	return card
}

func classifyProduct(name, number, rarity string) (productType, kind string) {
	lower := strings.ToLower(name)
	for _, keyword := range sealedKeywords {
		if strings.Contains(lower, keyword.match) {
			return keyword.productType, "sealed"
		}
	}
	if number == "" && rarity == "" {
		return "sealed_other", "sealed"
	}
	return "single", "card"
}

// BuildSearchTerms produces the lowercase token list the card search relies on.
func BuildSearchTerms(card models.Card) []string {
	unique := make(map[string]struct{})
	add := func(value string) {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			return
		}
		unique[value] = struct{}{}
	}
	for _, phrase := range []string{card.Name, card.Set, card.Game, card.Rarity, card.ProductType} {
		add(phrase)
		for _, token := range tokenize(phrase) {
			add(token)
		}
	}
	if number := strings.TrimSpace(card.Number); number != "" {
		add(number)
		if base, _, found := strings.Cut(number, "/"); found {
			add(base)
		}
	}
	terms := make([]string, 0, len(unique))
	for term := range unique {
		terms = append(terms, term)
	}
	sort.Strings(terms)
	return terms
}

func tokenize(value string) []string {
	fields := strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9')
	})
	tokens := make([]string, 0, len(fields))
	for _, field := range fields {
		if len(field) < 2 || searchStopWords[field] {
			continue
		}
		tokens = append(tokens, field)
	}
	return tokens
}
