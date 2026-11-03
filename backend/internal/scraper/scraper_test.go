package scraper

import (
	"errors"
	"net/url"
	"testing"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
)

func TestParseSoldHTML(t *testing.T) {
	page := `<ul><li class="s-item"><a href="https://www.ebay.com/itm/Sabrinas-Gengar/123456789012"><h3 class="s-item__title">Sabrina's Gengar 14/132 Holo</h3></a><span class="s-item__price">$42.50</span><span class="s-item__shipping">+$4.00 shipping</span><span>Sold Sep 20, 2026</span></li></ul>`
	sales, err := ParseSoldHTML(page, time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(sales) != 1 {
		t.Fatalf("got %d sales, want 1; parsed page: %+v", len(sales), sales)
	}
	if sales[0].ItemID != "123456789012" || sales[0].Price != 42.5 || sales[0].Shipping != 4 || sales[0].Currency != "USD" {
		t.Fatalf("unexpected parsed sale: %+v", sales[0])
	}
	if sales[0].SoldAt.Format("2006-01-02") != "2026-09-20" {
		t.Fatalf("unexpected sold date: %s", sales[0].SoldAt)
	}
}

func TestParseSoldHTMLDetectsBlockPage(t *testing.T) {
	_, err := ParseSoldHTML(`<html><body>Robot check - please complete the CAPTCHA</body></html>`, time.Now())
	if !errors.Is(err, ErrBlocked) {
		t.Fatalf("got error %v, want ErrBlocked", err)
	}
}

func TestParseSoldHTMLSupportsNewCardLayoutAndRejectsPriceRanges(t *testing.T) {
	page := `<ul><li class="s-card"><a href="https://www.ebay.com/itm/123456789013"><span class="s-card__title">Sabrina's Gengar 14/132</span></a><span class="s-card__price">$10.00 to $20.00</span><span>Sold Sep 21, 2026</span></li></ul>`
	sales, err := ParseSoldHTML(page, time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if len(sales) != 0 {
		t.Fatalf("price range should not be treated as a sold price: %+v", sales)
	}
}

func TestBuildSearchURLUsesSoldAndCompletedFilters(t *testing.T) {
	searchURL, err := buildSearchURL("https://example.test", "Sabrina's Gengar 14/132", "183454", 2)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := url.Parse(searchURL)
	if err != nil {
		t.Fatal(err)
	}
	values := parsed.Query()
	if values.Get("LH_Sold") != "1" || values.Get("LH_Complete") != "1" || values.Get("_pgn") != "2" || values.Get("_ipg") != "240" {
		t.Fatalf("unexpected search parameters: %v", values)
	}
}

func TestDetectGrade(t *testing.T) {
	cases := map[string]string{
		"Sabrina's Gengar PSA 10 Gem Mint": "psa-10",
		"Gengar BGS 9.5 Pristine":          "bgs-9.5",
		"Gengar Beckett 9":                 "bgs-9",
		"Gengar CGC 8.5":                   "cgc-8.5",
		"Sabrina's Gengar NM":              "raw",
	}
	for title, want := range cases {
		if got := DetectGrade(title); got != want {
			t.Errorf("DetectGrade(%q) = %q, want %q", title, got, want)
		}
	}
}

func TestRelevant(t *testing.T) {
	card := models.Card{Name: "Sabrina's Gengar", Number: "014/132", EbayExclude: []string{"Japanese"}}
	cases := []struct {
		title string
		want  bool
	}{
		{"Sabrina's Gengar 14/132 Holo", true},
		{"Sabrina's Gengar 14/132 PSA 10", true},
		{"Sabrina's Gengar 14/132 proxy", false},
		{"Sabrina's Gengar 14/132 Japanese", false},
		{"Sabrina's Gengar 15/132 Holo", false},
		{"Gengar 14/132 Holo", false},
	}
	for _, test := range cases {
		got, _ := Relevant(card, test.title)
		if got != test.want {
			t.Errorf("Relevant(%q) = %v, want %v", test.title, got, test.want)
		}
	}
}

func TestBuildSearchPlanUsesOverrideAndIncludesGradedQuery(t *testing.T) {
	plan := BuildSearchPlan(models.Card{Name: "Sabrina's Gengar", Number: "14/132", Set: "Gym Challenge"})
	if plan.RawKeywords == plan.GradedKeywords || plan.GradedKeywords == "" {
		t.Fatalf("expected separate raw and graded queries, got %+v", plan)
	}
	plan = BuildSearchPlan(models.Card{EbayQuery: "custom query"})
	if plan.RawKeywords != "custom query" || plan.GradedKeywords != "custom query" {
		t.Fatalf("override not used: %+v", plan)
	}
}

func TestDailyBucketsAndDailyWindowStats(t *testing.T) {
	day := time.Date(2026, 9, 20, 10, 0, 0, 0, time.FixedZone("PDT", -7*60*60))
	sales := []models.SoldSale{
		{Price: 10, Currency: "USD", SoldAt: day, Grade: "raw"},
		{Price: 12, Currency: "USD", SoldAt: day.Add(time.Hour), Grade: "raw"},
		{Price: 11, Currency: "USD", SoldAt: day.Add(90 * time.Minute), Grade: "raw"},
		{Price: 1000, Currency: "USD", SoldAt: day.Add(2 * time.Hour), Grade: "raw"},
		{Price: 20, Currency: "USD", SoldAt: day.Add(3 * time.Hour), Grade: "raw", BestOffer: true},
	}
	buckets := DailyBuckets(sales)
	if len(buckets) != 1 || buckets[0].Volume != 4 || buckets[0].Average != 11 {
		t.Fatalf("unexpected daily aggregate: %+v", buckets)
	}
	stats := BuildWindowSummaries(buckets, time.Date(2026, 9, 21, 0, 0, 0, 0, time.UTC))
	if stats["1d"].Count != 4 || stats["7d"].Count != 4 {
		t.Fatalf("daily points should supply 1d and 7d summary stats: %+v", stats)
	}
}

func TestDailyWindowIncludesCalendarBucketOverlappingCutoff(t *testing.T) {
	day := time.Date(2026, 9, 19, 0, 0, 0, 0, time.UTC)
	buckets := []DailyAggregate{
		{Date: day, Grade: "raw", Average: 15, LowPrice: 15, HighPrice: 15, Volume: 1},
		{Date: day.Add(24 * time.Hour), Grade: "raw", Average: 20, LowPrice: 20, HighPrice: 20, Volume: 1},
	}
	stats := BuildWindowSummaries(buckets, time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC))
	if stats["1d"].Count != 1 {
		t.Fatalf("1d summary omitted the calendar bucket overlapping its cutoff: %+v", stats["1d"])
	}
	if stats["7d"].ChangePct != 33.33 {
		t.Fatalf("7d change should be chronological, got %+v", stats["7d"])
	}
}
