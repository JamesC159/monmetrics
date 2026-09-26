package handlers

import (
	"strings"
	"testing"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
)

func ema(period float64) models.ChartIndicator {
	return models.ChartIndicator{Type: "ema", Parameters: map[string]interface{}{"period": period}, Color: "#22c55e", Visible: true}
}

func TestValidateSavedChart(t *testing.T) {
	valid := func() models.SavedChartRequest {
		return models.SavedChartRequest{Name: " My chart ", TimeRange: "30d", Indicators: []models.ChartIndicator{ema(20)}}
	}

	req := valid()
	if err := validateSavedChart(&req, 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Name != "My chart" || req.Source != "all" || req.Indicators[0].Parameters["period"] != 20 {
		t.Fatalf("not normalized: %+v", req)
	}

	bad := []func(*models.SavedChartRequest){
		func(r *models.SavedChartRequest) { r.Name = "" },
		func(r *models.SavedChartRequest) { r.Name = strings.Repeat("a", 101) },
		func(r *models.SavedChartRequest) { r.TimeRange = "2w" },
		func(r *models.SavedChartRequest) { r.Source = "amazon" },
		func(r *models.SavedChartRequest) { r.Indicators = []models.ChartIndicator{ema(1)} },
		func(r *models.SavedChartRequest) { r.Indicators = []models.ChartIndicator{ema(20.5)} },
		func(r *models.SavedChartRequest) { r.Indicators = []models.ChartIndicator{{Type: "rsi"}} },
		func(r *models.SavedChartRequest) {
			i := ema(20)
			i.Color = "red;"
			r.Indicators = []models.ChartIndicator{i}
		},
		func(r *models.SavedChartRequest) {
			r.Indicators = []models.ChartIndicator{ema(5), ema(10), ema(20), ema(50)}
		},
	}
	for i, mutate := range bad {
		r := valid()
		mutate(&r)
		if err := validateSavedChart(&r, 3); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestValidateIndicatorParams(t *testing.T) {
	good := []models.ChartIndicator{
		{Type: "sma", Parameters: map[string]interface{}{"period": 20.0}},
		{Type: "bollinger", Parameters: map[string]interface{}{"period": 20.0, "stddev": 2.0}},
		{Type: "rsi", Parameters: map[string]interface{}{"period": 14.0, "extra": "x"}},
		{Type: "macd", Parameters: map[string]interface{}{"fast": 12.0, "slow": 26.0, "signal": 9.0}},
	}
	for _, ind := range good {
		p, err := validateIndicatorParams(ind.Type, ind.Parameters)
		if err != nil {
			t.Errorf("%s: unexpected error %v", ind.Type, err)
		}
		if _, ok := p["extra"]; ok {
			t.Errorf("%s: unknown params not dropped", ind.Type)
		}
	}
	bad := []models.ChartIndicator{
		{Type: "bollinger", Parameters: map[string]interface{}{"period": 20.0, "stddev": 9.0}},
		{Type: "rsi", Parameters: map[string]interface{}{"period": 150.0}},
		{Type: "macd", Parameters: map[string]interface{}{"fast": 26.0, "slow": 12.0, "signal": 9.0}},
		{Type: "vwap", Parameters: map[string]interface{}{}},
	}
	for _, ind := range bad {
		if _, err := validateIndicatorParams(ind.Type, ind.Parameters); err == nil {
			t.Errorf("%s %v: expected error", ind.Type, ind.Parameters)
		}
	}
}

func TestValidateDrawings(t *testing.T) {
	pt := func(t int64, p float64) models.DrawingPoint { return models.DrawingPoint{Time: t, Price: p} }
	valid := func() []models.ChartDrawing {
		return []models.ChartDrawing{
			{ID: "a1", Type: "trendline", Color: "#ffffff", Points: []models.DrawingPoint{pt(1700000000, 10), pt(1700086400, 12)}},
			{ID: "a2", Type: "text", Color: "#ffffff", Text: " hello\x00 ", Points: []models.DrawingPoint{pt(1700000000, 10)}},
		}
	}
	d := valid()
	if err := validateDrawings(d); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d[1].Text != "hello" || d[0].LineWidth != 2 {
		t.Fatalf("not normalized: %+v", d)
	}

	bad := []func([]models.ChartDrawing){
		func(d []models.ChartDrawing) { d[0].Type = "circle" },
		func(d []models.ChartDrawing) { d[0].Points = d[0].Points[:1] },
		func(d []models.ChartDrawing) { d[1].ID = d[0].ID },
		func(d []models.ChartDrawing) { d[0].ID = "<script>" },
		func(d []models.ChartDrawing) { d[0].Color = "red" },
		func(d []models.ChartDrawing) { d[0].Points[0].Price = -1 },
		func(d []models.ChartDrawing) { d[0].Points[0].Time = 0 },
		func(d []models.ChartDrawing) { d[1].Text = "" },
		func(d []models.ChartDrawing) { d[0].LineWidth = 20 },
		func(d []models.ChartDrawing) {
			d[0].Type = "freehand"
			d[0].Points = make([]models.DrawingPoint, 501)
			for i := range d[0].Points {
				d[0].Points[i] = pt(1700000000, 1)
			}
		},
	}
	for i, mutate := range bad {
		d := valid()
		mutate(d)
		if err := validateDrawings(d); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestValidateAlert(t *testing.T) {
	r := models.PriceAlertRequest{Condition: "above", TargetPrice: 99.999, Pct: 5, Note: " hi "}
	if err := validateAlert(&r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Source != "all" || r.Mode != "once" || r.TargetPrice != 100 || r.Pct != 0 || r.Note != "hi" {
		t.Fatalf("not normalized: %+v", r)
	}

	r = models.PriceAlertRequest{Condition: "pct_change", Pct: 10, Days: 7, Mode: "recurring"}
	if err := validateAlert(&r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Direction != "either" || r.CooldownHours != 24 {
		t.Fatalf("defaults not applied: %+v", r)
	}

	bad := []models.PriceAlertRequest{
		{Condition: "above"},
		{Condition: "below", TargetPrice: -5},
		{Condition: "pct_change", Pct: 0.1, Days: 7},
		{Condition: "pct_change", Pct: 10, Days: 0},
		{Condition: "ema_cross", EMAPeriod: 1},
		{Condition: "ema_cross", EMAPeriod: 20, Direction: "sideways"},
		{Condition: "above", TargetPrice: 10, Mode: "recurring", CooldownHours: 500},
		{Condition: "above", TargetPrice: 10, Mode: "forever"},
		{Condition: "above", TargetPrice: 10, Source: "amazon"},
		{Condition: "moon"},
	}
	for i, req := range bad {
		if err := validateAlert(&req); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestValidatePortfolioItem(t *testing.T) {
	raw := func() models.PortfolioItemRequest {
		return models.PortfolioItemRequest{CardID: "65f1c2a4b7e8d9f0a1b2c3d4", ItemType: "raw_card", Quantity: 1, Condition: "NM", PurchasePrice: 10}
	}
	r := raw()
	if _, err := validatePortfolioItem(&r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	graded := models.PortfolioItemRequest{CardID: "65f1c2a4b7e8d9f0a1b2c3d4", ItemType: "graded_card", Quantity: 1,
		Grading: &models.Grading{Company: "bgs", Grade: 9.5, CertNumber: "0012345"}}
	if _, err := validatePortfolioItem(&graded); err != nil || graded.Grading.Company != "BGS" {
		t.Fatalf("graded should be valid: %v", err)
	}

	custom := models.PortfolioItemRequest{ItemType: "sealed", CustomName: "Promo tin", Quantity: 2, CustomImageURL: "https://example.com/a.png"}
	if _, err := validatePortfolioItem(&custom); err != nil {
		t.Fatalf("custom should be valid: %v", err)
	}

	future := time.Now().AddDate(0, 0, 2).Format("2006-01-02")
	bad := []func(*models.PortfolioItemRequest){
		func(r *models.PortfolioItemRequest) { r.ItemType = "binder" },
		func(r *models.PortfolioItemRequest) { r.CardID = ""; r.CustomName = "" },
		func(r *models.PortfolioItemRequest) { r.CardID = "nope" },
		func(r *models.PortfolioItemRequest) { r.Quantity = 0 },
		func(r *models.PortfolioItemRequest) { r.Condition = "Mint" },
		func(r *models.PortfolioItemRequest) { r.PurchasePrice = -1 },
		func(r *models.PortfolioItemRequest) { r.PurchaseDate = future },
		func(r *models.PortfolioItemRequest) { r.PurchaseDate = "01/02/2024" },
		func(r *models.PortfolioItemRequest) { r.Grading = &models.Grading{Company: "PSA", Grade: 10} },
		func(r *models.PortfolioItemRequest) { r.CustomImageURL = "javascript:alert(1)" },
		func(r *models.PortfolioItemRequest) { r.Notes = strings.Repeat("x", 1001) },
		func(r *models.PortfolioItemRequest) {
			r.ItemType = "graded_card"
			r.Grading = &models.Grading{Company: "PSA", Grade: 9.5}
		},
		func(r *models.PortfolioItemRequest) {
			r.ItemType = "graded_card"
			r.Grading = &models.Grading{Company: "XYZ", Grade: 9}
		},
		func(r *models.PortfolioItemRequest) {
			r.ItemType = "graded_card"
			r.Grading = &models.Grading{Company: "CGC", Grade: 9, CertNumber: "12 34<script>"}
		},
	}
	for i, mutate := range bad {
		r := raw()
		mutate(&r)
		if _, err := validatePortfolioItem(&r); err == nil {
			t.Errorf("case %d: expected error", i)
		}
	}
}

func TestValidateListingDraft(t *testing.T) {
	d := models.ListingDraft{Provider: "ebay", Title: "Charizard VMAX PSA 10", Price: 100, Quantity: 1, Condition: "Graded"}
	if err := validateListingDraft(&d, 1); err != nil || d.Duration != "GTC" || d.Format != "fixed_price" {
		t.Fatalf("unexpected: %v %+v", err, d)
	}
	over := d
	over.Quantity = 2
	if err := validateListingDraft(&over, 1); err == nil {
		t.Fatal("quantity above owned should fail")
	}
	long := d
	long.Title = strings.Repeat("a", 81)
	if err := validateListingDraft(&long, 1); err == nil {
		t.Fatal("eBay title >80 should fail")
	}
	img := d
	img.ImageURLs = []string{"http://insecure.example.com/x.png"}
	if err := validateListingDraft(&img, 1); err == nil {
		t.Fatal("non-https image should fail")
	}
}
