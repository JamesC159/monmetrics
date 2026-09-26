package alerts

import (
	"math"
	"testing"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
)

func f(v float64) *float64 { return &v }

func TestEMAMatchesFrontend(t *testing.T) {
	got := EMA([]float64{1, 2, 3, 4, 5, 6}, 3)
	want := []*float64{nil, nil, f(2), f(3), f(4), f(5)}
	for i := range want {
		if (got[i] == nil) != (want[i] == nil) || (got[i] != nil && math.Abs(*got[i]-*want[i]) > 1e-9) {
			t.Fatalf("index %d: got %v want %v", i, got[i], want[i])
		}
	}
	if out := EMA([]float64{1, 2}, 3); out[1] != nil {
		t.Fatal("expected nil when series shorter than period")
	}
}

func TestShouldFire(t *testing.T) {
	now := time.Now().UTC()
	recent := now.Add(-time.Hour)
	old := now.Add(-48 * time.Hour)

	cases := []struct {
		name  string
		alert models.PriceAlert
		snap  Snapshot
		want  bool
	}{
		{"above crosses", models.PriceAlert{Condition: models.AlertAbove, TargetPrice: 100, LastPrice: f(99)}, Snapshot{Price: 100}, true},
		{"above already above", models.PriceAlert{Condition: models.AlertAbove, TargetPrice: 100, LastPrice: f(101)}, Snapshot{Price: 105}, false},
		{"above no baseline", models.PriceAlert{Condition: models.AlertAbove, TargetPrice: 100}, Snapshot{Price: 105}, false},
		{"below crosses", models.PriceAlert{Condition: models.AlertBelow, TargetPrice: 50, LastPrice: f(51)}, Snapshot{Price: 49}, true},
		{"below stays", models.PriceAlert{Condition: models.AlertBelow, TargetPrice: 50, LastPrice: f(55)}, Snapshot{Price: 51}, false},
		{"pct up", models.PriceAlert{Condition: models.AlertPctChange, Pct: 10, Direction: "up"}, Snapshot{Price: 111, PastPrice: f(100)}, true},
		{"pct up wrong direction", models.PriceAlert{Condition: models.AlertPctChange, Pct: 10, Direction: "up"}, Snapshot{Price: 80, PastPrice: f(100)}, false},
		{"pct either down", models.PriceAlert{Condition: models.AlertPctChange, Pct: 10, Direction: "either"}, Snapshot{Price: 89, PastPrice: f(100)}, true},
		{"pct no history", models.PriceAlert{Condition: models.AlertPctChange, Pct: 10}, Snapshot{Price: 200}, false},
		{"ema cross up", models.PriceAlert{Condition: models.AlertEMACross, Direction: "up", LastPrice: f(9), LastEMA: f(10)}, Snapshot{Price: 11, EMA: f(10)}, true},
		{"ema cross down filtered", models.PriceAlert{Condition: models.AlertEMACross, Direction: "up", LastPrice: f(11), LastEMA: f(10)}, Snapshot{Price: 9, EMA: f(10)}, false},
		{"ema no cross", models.PriceAlert{Condition: models.AlertEMACross, LastPrice: f(11), LastEMA: f(10)}, Snapshot{Price: 12, EMA: f(10)}, false},
		{"recurring in cooldown", models.PriceAlert{Condition: models.AlertAbove, TargetPrice: 100, LastPrice: f(99), Mode: models.AlertModeRecurring, CooldownHours: 24, LastTriggeredAt: &recent}, Snapshot{Price: 101}, false},
		{"recurring after cooldown", models.PriceAlert{Condition: models.AlertAbove, TargetPrice: 100, LastPrice: f(99), Mode: models.AlertModeRecurring, CooldownHours: 24, LastTriggeredAt: &old}, Snapshot{Price: 101}, true},
	}
	for _, c := range cases {
		c.alert.Active = true
		if got := ShouldFire(&c.alert, c.snap, now); got != c.want {
			t.Errorf("%s: got %v want %v", c.name, got, c.want)
		}
	}

	inactive := models.PriceAlert{Condition: models.AlertAbove, TargetPrice: 100, LastPrice: f(99)}
	if ShouldFire(&inactive, Snapshot{Price: 101}, now) {
		t.Error("inactive alert fired")
	}
}

func TestBuildSnapshotPastPrice(t *testing.T) {
	now := time.Date(2026, 1, 31, 12, 0, 0, 0, time.UTC)
	series := []DayPrice{
		{Day: time.Date(2026, 1, 20, 0, 0, 0, 0, time.UTC), Price: 80},
		{Day: time.Date(2026, 1, 24, 0, 0, 0, 0, time.UTC), Price: 100},
		{Day: time.Date(2026, 1, 25, 0, 0, 0, 0, time.UTC), Price: 105},
		{Day: time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC), Price: 120},
	}
	a := models.PriceAlert{Condition: models.AlertPctChange, Days: 7}
	s, ok := BuildSnapshot(&a, series, now)
	if !ok || s.Price != 120 || s.PastPrice == nil || *s.PastPrice != 100 {
		t.Fatalf("unexpected snapshot: %+v past=%v", s, s.PastPrice)
	}
	if _, ok := BuildSnapshot(&a, nil, now); ok {
		t.Fatal("expected no snapshot for empty series")
	}
}
