package catalog

import (
	"context"
	"strings"
	"time"
)

// Category is a game or product line (TCGplayer "category").
type Category struct {
	CategoryID  int    `json:"categoryId"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	SeoName     string `json:"seoCategoryName"`
}

// Group is a set/expansion within a category.
type Group struct {
	GroupID      int    `json:"groupId"`
	CategoryID   int    `json:"categoryId"`
	Name         string `json:"name"`
	Abbreviation string `json:"abbreviation"`
	PublishedOn  string `json:"publishedOn"`
}

// ExtendedField holds provider-specific attributes such as Number or Rarity.
type ExtendedField struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Value       string `json:"value"`
}

// Product is a single catalogue entry: a card, a booster box, a tin, etc.
type Product struct {
	ProductID    int             `json:"productId"`
	CategoryID   int             `json:"categoryId"`
	GroupID      int             `json:"groupId"`
	Name         string          `json:"name"`
	CleanName    string          `json:"cleanName"`
	ImageURL     string          `json:"imageUrl"`
	URL          string          `json:"url"`
	ModifiedOn   string          `json:"modifiedOn"`
	ExtendedData []ExtendedField `json:"extendedData"`
}

// ProductPrice is one printing's current market price. marketPrice is nullable: very low
// sales volume leaves TCGplayer without a figure.
type ProductPrice struct {
	ProductID      int      `json:"productId"`
	LowPrice       *float64 `json:"lowPrice"`
	MidPrice       *float64 `json:"midPrice"`
	HighPrice      *float64 `json:"highPrice"`
	MarketPrice    *float64 `json:"marketPrice"`
	DirectLowPrice *float64 `json:"directLowPrice"`
	SubTypeName    string   `json:"subTypeName"`
}

// Source is a provider of catalogue data. A second provider can be added without
// touching the normalisation or storage layers.
type Source interface {
	Name() string
	Categories(ctx context.Context) ([]Category, error)
	Groups(ctx context.Context, categoryID int) ([]Group, error)
	Products(ctx context.Context, categoryID, groupID int) ([]Product, error)
}

// PriceSource is the subset needed for daily price ingest.
type PriceSource interface {
	Name() string
	Prices(ctx context.Context, categoryID, groupID int) ([]ProductPrice, error)
	LastUpdated(ctx context.Context) (time.Time, error)
}

// ProductPriceSource is a Source that returns a group's products and prices from one
// request, so a single pass refreshes the catalogue and that day's prices.
type ProductPriceSource interface {
	ProductsAndPrices(ctx context.Context, categoryID, groupID int) ([]Product, []ProductPrice, error)
}

// PriceSink stores the prices fetched alongside a group's products and reports how many
// price points it wrote. The pricing package implements it; it imports this package, so
// the dependency has to point this way.
type PriceSink interface {
	IngestGroupPrices(ctx context.Context, categoryID, groupID int, prices []ProductPrice, day time.Time) (int, error)
}

// extended matches field against each item's name or display name, ignoring case and
// spaces: the JSON feed names Pokemon card text "CardText" (display "Card Text"), and
// the CSV feed carries only the name.
func (p Product) extended(field string) string {
	key := extendedKey(field)
	for _, item := range p.ExtendedData {
		if extendedKey(item.Name) == key || extendedKey(item.DisplayName) == key {
			return strings.TrimSpace(item.Value)
		}
	}
	return ""
}

func extendedKey(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", ""))
}

// nonTCGCategories are TCGplayer categories that are accessories, miniatures or
// board games rather than trading card products we can price from sold listings.
var nonTCGCategories = map[string]bool{
	"supplies":                     true,
	"card sleeves":                 true,
	"deck boxes":                   true,
	"card storage tins":            true,
	"life counters":                true,
	"playmats":                     true,
	"organizers & stores":          true,
	"tcgplayer":                    true,
	"funko":                        true,
	"boardgames":                   true,
	"heroclix":                     true,
	"axis & allies":                true,
	"d & d miniatures":             true,
	"monsterpocalypse":             true,
	"star wars miniatures":         true,
	"world of warcraft miniatures": true,
	"dice masters":                 true,
}

var nonTCGKeywords = []string{"sleeve", "playmat", "deck box", "storage", "supplies", "miniature", "accessor"}

// emptyCategories hold no products, and TCGCSV's FAQ asks clients to skip them. 69 and
// 70 are abandoned comic-book categories with thousands of groups that all 404.
var emptyCategories = map[int]bool{9: true, 10: true, 12: true, 14: true, 21: true, 55: true, 69: true, 70: true}

// IsTradingCardCategory reports whether a category should be catalogued.
func IsTradingCardCategory(c Category) bool {
	if emptyCategories[c.CategoryID] {
		return false
	}
	name := strings.ToLower(strings.TrimSpace(firstNonEmpty(c.DisplayName, c.Name)))
	if name == "" || nonTCGCategories[name] {
		return false
	}
	for _, keyword := range nonTCGKeywords {
		if strings.Contains(name, keyword) {
			return false
		}
	}
	return true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
