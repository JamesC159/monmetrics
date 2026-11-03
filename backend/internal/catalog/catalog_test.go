package catalog

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const productsFixture = `{"success":true,"errors":[],"results":[
	{"productId":451396,"categoryId":3,"groupId":3170,"name":"Lugia VSTAR (Secret)","cleanName":"Lugia VSTAR Secret",
	 "imageUrl":"https://example.test/451396.jpg","url":"https://example.test/lugia",
	 "extendedData":[{"name":"Number","displayName":"Number","value":"211/195"},{"name":"Rarity","displayName":"Rarity","value":"Secret Rare"},
	  {"name":"CardText","displayName":"Card Text","value":"Summoning Star, \"VSTAR\"\n<br>Once per game."}]},
	{"productId":451400,"categoryId":3,"groupId":3170,"name":"SWSH12: Silver Tempest Elite Trainer Box","cleanName":"Silver Tempest Elite Trainer Box",
	 "imageUrl":"https://example.test/451400.jpg","url":"https://example.test/etb","extendedData":[]}
]}`

func TestTCGCSVProductsDecodesEnvelope(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tcgplayer/3/3170/products" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(productsFixture))
	}))
	defer server.Close()

	source := NewTCGCSVSource(server.URL, "test-agent", 0)
	products, err := source.Products(context.Background(), 3, 3170)
	if err != nil {
		t.Fatalf("Products returned error: %v", err)
	}
	if len(products) != 2 {
		t.Fatalf("expected 2 products, got %d", len(products))
	}
	if products[0].extended("Number") != "211/195" {
		t.Errorf("expected extended Number 211/195, got %q", products[0].extended("Number"))
	}
}

func TestTCGCSVPropagatesHTTPErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	source := NewTCGCSVSource(server.URL, "test-agent", 0)
	if _, err := source.Categories(context.Background()); err == nil {
		t.Fatal("expected throttle error, got nil")
	}
}

func TestRunnerHandlesMissingGroupFeed(t *testing.T) {
	for _, test := range []struct {
		name       string
		status     int
		categoryID int
		wantError  bool
		wantNext   bool
	}{
		{name: "broad sync skips missing feed", status: http.StatusNotFound, wantNext: true},
		{name: "targeted sync reports missing feed", status: http.StatusNotFound, categoryID: 21, wantError: true},
		{name: "broad sync reports server errors", status: http.StatusInternalServerError, wantError: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			requestedNext := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/tcgplayer/categories":
					w.Write([]byte(`{"success":true,"results":[{"categoryId":21,"displayName":"My Little Pony"},{"categoryId":3,"displayName":"Pokemon"}]}`))
				case "/tcgplayer/21/groups":
					w.WriteHeader(test.status)
				case "/tcgplayer/3/groups":
					requestedNext = true
					w.Write([]byte(`{"success":true,"results":[]}`))
				default:
					t.Errorf("unexpected path %s", r.URL.Path)
				}
			}))
			defer server.Close()

			runner := &Runner{Source: NewTCGCSVSource(server.URL, "test-agent", 0), DryRun: true}
			err := runner.RunOnce(context.Background(), SyncOptions{CategoryID: test.categoryID})
			if (err != nil) != test.wantError {
				t.Fatalf("RunOnce error = %v, wantError %v", err, test.wantError)
			}
			if test.status == http.StatusNotFound && test.wantError && !errors.Is(err, ErrNotFound) {
				t.Errorf("expected ErrNotFound, got %v", err)
			}
			if requestedNext != test.wantNext {
				t.Errorf("requested next category = %v, want %v", requestedNext, test.wantNext)
			}
		})
	}
}

func TestNormalizeProductClassifiesSingleAndSealed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(productsFixture))
	}))
	defer server.Close()
	source := NewTCGCSVSource(server.URL, "test-agent", 0)
	products, err := source.Products(context.Background(), 3, 3170)
	if err != nil {
		t.Fatalf("Products returned error: %v", err)
	}

	category := Category{CategoryID: 3, Name: "Pokemon", DisplayName: "Pokemon"}
	group := Group{GroupID: 3170, CategoryID: 3, Name: "SWSH12: Silver Tempest"}

	single := NormalizeProduct("tcgcsv", category, group, products[0])
	if single.Category != "card" || single.ProductType != "single" {
		t.Errorf("expected card/single, got %s/%s", single.Category, single.ProductType)
	}
	if single.ExternalID != "451396" || single.Set != "SWSH12: Silver Tempest" || single.Game != "Pokemon" {
		t.Errorf("unexpected identity fields: %+v", single)
	}
	if !single.Active {
		t.Error("expected new catalog cards to be active")
	}

	sealed := NormalizeProduct("tcgcsv", category, group, products[1])
	if sealed.Category != "sealed" || sealed.ProductType != "elite_trainer_box" {
		t.Errorf("expected sealed/elite_trainer_box, got %s/%s", sealed.Category, sealed.ProductType)
	}
}

func TestBuildSearchTermsAreLowercaseTokens(t *testing.T) {
	category := Category{CategoryID: 3, DisplayName: "Pokemon"}
	group := Group{GroupID: 3170, Name: "SWSH12: Silver Tempest"}
	product := Product{ProductID: 1, Name: "Lugia VSTAR (Secret)", ExtendedData: []ExtendedField{
		{Name: "Number", Value: "211/195"},
	}}

	terms := NormalizeProduct("tcgcsv", category, group, product).SearchTerms
	want := []string{"lugia", "vstar", "secret", "silver", "tempest", "pokemon", "211/195", "211"}
	for _, expected := range want {
		if !contains(terms, expected) {
			t.Errorf("expected search term %q in %v", expected, terms)
		}
	}
	if contains(terms, "Lugia") {
		t.Error("search terms must be lowercase")
	}
}

func TestIsTradingCardCategoryFiltersAccessories(t *testing.T) {
	cases := map[string]bool{
		"Pokemon":      true,
		"One Piece":    true,
		"Card Sleeves": false,
		"Playmats":     false,
		"Supplies":     false,
		"Boardgames":   false,
	}
	for name, expected := range cases {
		if got := IsTradingCardCategory(Category{DisplayName: name}); got != expected {
			t.Errorf("IsTradingCardCategory(%q) = %v, want %v", name, got, expected)
		}
	}
}

func TestInitialFieldsScheduleWithinADay(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	fields := initialFields(now)
	next, ok := fields["next_scrape_at"].(time.Time)
	if !ok {
		t.Fatalf("next_scrape_at missing or wrong type: %T", fields["next_scrape_at"])
	}
	if next.Before(now) || next.After(now.Add(24*time.Hour)) {
		t.Errorf("next_scrape_at %v not within 24h of %v", next, now)
	}
	if _, exists := fields["ebay_synced_at"]; exists {
		t.Error("catalog must not seed scraper-owned ebay_synced_at")
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
