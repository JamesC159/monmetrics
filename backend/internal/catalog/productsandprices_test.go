package catalog

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

// productsAndPricesFixture is productsFixture as TCGCSV's ProductsAndPrices.csv: ext
// columns sit on both sides of the price columns, Lugia has two printings, and the Elite
// Trainer Box has no prices yet.
const productsAndPricesFixture = `productId,name,cleanName,imageUrl,categoryId,groupId,url,modifiedOn,imageCount,extCardText,lowPrice,midPrice,highPrice,marketPrice,directLowPrice,subTypeName,extRarity,extNumber
451396,Lugia VSTAR (Secret),Lugia VSTAR Secret,https://example.test/451396.jpg,3,3170,https://cpt.tcgcsv.com/Nmh3,2025-12-04T15:00:29.8,1,"Summoning Star, ""VSTAR""
<br>Once per game.",40.0,52.5,120.0,48.12,,Holofoil,Secret Rare,211/195
451396,Lugia VSTAR (Secret),Lugia VSTAR Secret,https://example.test/451396.jpg,3,3170,https://cpt.tcgcsv.com/Nmh3,2025-12-04T15:00:29.8,1,"Summoning Star, ""VSTAR""
<br>Once per game.",35.0,41.0,90.0,,38.5,Reverse Holofoil,Secret Rare,211/195
451400,SWSH12: Silver Tempest Elite Trainer Box,Silver Tempest Elite Trainer Box,https://example.test/451400.jpg,3,3170,https://cpt.tcgcsv.com/Pmh3,2025-11-01T00:00:00,1,,,,,,,,,
`

func csvServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/tcgplayer/3/3170/ProductsAndPrices.csv":
			w.Header().Set("Content-Type", "text/csv")
			w.Write([]byte(productsAndPricesFixture))
		case "/tcgplayer/3/3170/products":
			w.Write([]byte(productsFixture))
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func TestTCGCSVProductsAndPricesParsesRows(t *testing.T) {
	source := NewTCGCSVSource(csvServer(t).URL, "test-agent", 0)
	products, prices, err := source.ProductsAndPrices(context.Background(), 3, 3170)
	if err != nil {
		t.Fatalf("ProductsAndPrices returned error: %v", err)
	}

	if len(products) != 2 || products[0].ProductID != 451396 || products[1].ProductID != 451400 {
		t.Fatalf("expected one product per id in file order, got %+v", products)
	}
	lugia := products[0]
	if lugia.URL != "https://www.tcgplayer.com/product/451396" {
		t.Errorf("expected a TCGplayer product URL in place of the short link, got %q", lugia.URL)
	}
	if lugia.extended("Number") != "211/195" || lugia.extended("Rarity") != "Secret Rare" {
		t.Errorf("unexpected Number/Rarity: %+v", lugia.ExtendedData)
	}
	if got := lugia.extended("Card Text"); got != "Summoning Star, \"VSTAR\"\n<br>Once per game." {
		t.Errorf("extCardText should answer to \"Card Text\", got %q", got)
	}
	if len(products[1].ExtendedData) != 0 {
		t.Errorf("empty ext columns should be dropped, got %+v", products[1].ExtendedData)
	}

	if len(prices) != 2 {
		t.Fatalf("expected two printings and no row for the unpriced product, got %+v", prices)
	}
	holo, reverse := prices[0], prices[1]
	if holo.SubTypeName != "Holofoil" || holo.MarketPrice == nil || *holo.MarketPrice != 48.12 || holo.DirectLowPrice != nil {
		t.Errorf("unexpected Holofoil row: %+v", holo)
	}
	if reverse.SubTypeName != "Reverse Holofoil" || reverse.MarketPrice != nil || reverse.DirectLowPrice == nil || *reverse.DirectLowPrice != 38.5 {
		t.Errorf("an empty marketPrice must stay nil: %+v", reverse)
	}
}

func TestParseProductsAndPricesWithoutPriceColumns(t *testing.T) {
	// An unreleased set: TCGCSV drops every price column, subTypeName included.
	const unpriced = `productId,name,cleanName,imageUrl,categoryId,groupId,url,modifiedOn,imageCount,extRarity,extNumber
726604,Gem Blade Ruby,Gem Blade Ruby,https://example.test/726604.jpg,17,24919,https://cpt.tcgcsv.com/6GQ7,2026-10-28T13:55:59.187,1,Marvel Rare,CPB-001 MR
`
	products, prices, err := parseProductsAndPrices(strings.NewReader(unpriced))
	if err != nil {
		t.Fatalf("a group with no price columns should parse: %v", err)
	}
	if len(products) != 1 || products[0].extended("Number") != "CPB-001 MR" {
		t.Errorf("unexpected products: %+v", products)
	}
	if len(prices) != 0 {
		t.Errorf("expected no price rows, got %+v", prices)
	}
}

func TestProductsAndPricesNormalizeLikeJSON(t *testing.T) {
	source := NewTCGCSVSource(csvServer(t).URL, "test-agent", 0)
	fromJSON, err := source.Products(context.Background(), 3, 3170)
	if err != nil {
		t.Fatalf("Products returned error: %v", err)
	}
	fromCSV, _, err := source.ProductsAndPrices(context.Background(), 3, 3170)
	if err != nil {
		t.Fatalf("ProductsAndPrices returned error: %v", err)
	}
	if len(fromJSON) != len(fromCSV) {
		t.Fatalf("JSON has %d products, CSV %d", len(fromJSON), len(fromCSV))
	}

	category := Category{CategoryID: 3, Name: "Pokemon", DisplayName: "Pokemon"}
	group := Group{GroupID: 3170, CategoryID: 3, Name: "SWSH12: Silver Tempest"}
	for i := range fromJSON {
		want := NormalizeProduct("tcgcsv", category, group, fromJSON[i])
		got := NormalizeProduct("tcgcsv", category, group, fromCSV[i])
		want.ProductURL, got.ProductURL = "", "" // the CSV builds its own; checked above
		if !reflect.DeepEqual(want, got) {
			t.Errorf("product %d normalises differently:\n json %+v\n csv  %+v", fromJSON[i].ProductID, want, got)
		}
	}
}

type productPriceStub struct {
	*stubSource
	combined int
}

func (s *productPriceStub) ProductsAndPrices(ctx context.Context, categoryID, groupID int) ([]Product, []ProductPrice, error) {
	s.combined++
	products, err := s.Products(ctx, categoryID, groupID)
	return products, []ProductPrice{{ProductID: groupID * 10, SubTypeName: "Normal"}}, err
}

type sinkStub struct{ calls int }

func (s *sinkStub) IngestGroupPrices(ctx context.Context, categoryID, groupID int, prices []ProductPrice, day time.Time) (int, error) {
	s.calls++
	return len(prices), nil
}

func TestRunnerReadsPricesOnlyWithASink(t *testing.T) {
	source := &productPriceStub{stubSource: &stubSource{}}

	_, prices, withPrices, err := (&Runner{Source: source}).fetchGroup(context.Background(), 3, 1)
	if err != nil || withPrices || prices != nil || source.combined != 0 {
		t.Errorf("without a sink the runner should fetch products only: withPrices=%v prices=%v combined=%d err=%v",
			withPrices, prices, source.combined, err)
	}

	_, prices, withPrices, err = (&Runner{Source: source, Prices: &sinkStub{}}).fetchGroup(context.Background(), 3, 1)
	if err != nil || !withPrices || len(prices) != 1 || source.combined != 1 {
		t.Errorf("with a sink the runner should use the combined file: withPrices=%v prices=%v combined=%d err=%v",
			withPrices, prices, source.combined, err)
	}
}

func TestDryRunStoresNoPrices(t *testing.T) {
	source := &productPriceStub{stubSource: &stubSource{groups: []Group{{GroupID: 1}, {GroupID: 2}}}}
	sink := &sinkStub{}
	runner := &Runner{Source: source, Interval: time.Hour, DryRun: true, Prices: sink}

	if err := runner.RunOnce(context.Background(), SyncOptions{}); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if source.combined != 2 {
		t.Errorf("dry run should still read both groups' combined files, read %d", source.combined)
	}
	if sink.calls != 0 {
		t.Errorf("dry run must not store prices, stored %d groups", sink.calls)
	}
}
