package handlers

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jamesc159/monmetrics/configs"
)

func TestBuildSearchFilterAppliesCatalogFilters(t *testing.T) {
	h := &Handlers{}
	filter := h.buildSearchFilter(searchFilters{
		Game:        "Pokemon",
		Category:    "sealed",
		Set:         "Silver Tempest",
		ProductType: "booster_box",
		Rarity:      "Rare",
		PricedOnly:  true,
	})

	for _, key := range []string{"active", "game", "category", "set", "product_type", "rarity", "current_price"} {
		if _, ok := filter[key]; !ok {
			t.Fatalf("expected filter to constrain %q, got %v", key, filter)
		}
	}
	if filter["category"] != "sealed" {
		t.Fatalf("category should match exactly, got %v", filter["category"])
	}
	if got, ok := filter["current_price"].(bson.M); !ok || got["$gt"] != 0 {
		t.Fatalf("priced filter = %v, want $gt 0", filter["current_price"])
	}
}

func TestBuildSearchFilterEscapesQuery(t *testing.T) {
	h := &Handlers{}
	filter := h.buildSearchFilter(searchFilters{Query: "charizard.*"})

	or, ok := filter["$or"].([]bson.M)
	if !ok || len(or) == 0 {
		t.Fatalf("expected $or conditions, got %v", filter["$or"])
	}
	// A raw ".*" would match everything; QuoteMeta must neutralize it.
	name, ok := or[0]["name"].(bson.M)
	if !ok {
		t.Fatalf("expected name condition, got %v", or[0])
	}
	re, ok := name["$regex"].(primitive.Regex)
	if !ok {
		t.Fatalf("expected primitive.Regex, got %T", name["$regex"])
	}
	if re.Pattern != `charizard\.\*` {
		t.Fatalf("query pattern = %q, want escaped literal", re.Pattern)
	}
}

func TestBuildSearchFilterOmitsUnsetFields(t *testing.T) {
	h := &Handlers{}
	filter := h.buildSearchFilter(searchFilters{})

	if len(filter) != 1 {
		t.Fatalf("empty search should only filter on active, got %v", filter)
	}
	if _, ok := filter["active"]; !ok {
		t.Fatalf("expected active filter, got %v", filter)
	}
}

func TestApplyEnabledGamesIsInertWhenUnset(t *testing.T) {
	h := &Handlers{config: &configs.Config{}}
	filter := bson.M{"category": "card"}

	if got := h.applyEnabledGames(filter); len(got) != 1 || got["category"] != "card" {
		t.Fatalf("no enabled games should leave the filter untouched, got %v", got)
	}
}

// A client passing ?game=Pokemon must not be able to escape the flag.
func TestApplyEnabledGamesOverridesCallerGameFilter(t *testing.T) {
	h := &Handlers{config: &configs.Config{EnabledGames: []string{"Magic"}}}
	filter := h.buildSearchFilter(searchFilters{Game: "Pokemon"})

	clauses, ok := filter["$and"].([]bson.M)
	if !ok || len(clauses) != 2 {
		t.Fatalf("expected the flag to wrap the filter in $and, got %v", filter)
	}
	games, ok := clauses[1]["game"].(bson.M)
	if !ok {
		t.Fatalf("expected a game constraint clause, got %v", clauses[1])
	}
	patterns, ok := games["$in"].([]primitive.Regex)
	if !ok || len(patterns) != 1 || patterns[0].Pattern != "Magic" || patterns[0].Options != "i" {
		t.Fatalf("expected a case-insensitive Magic pattern, got %v", games["$in"])
	}
	// The caller's own game filter survives inside the first clause, so the two AND together
	// and a Pokemon query under a Magic-only deployment correctly returns nothing.
	if _, ok := clauses[0]["game"]; !ok {
		t.Fatalf("caller game filter should be preserved, got %v", clauses[0])
	}
}
