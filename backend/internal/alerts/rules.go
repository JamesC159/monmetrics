package alerts

import (
	"math"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
)

// DayPrice is one day's average price
type DayPrice struct {
	Day   time.Time
	Price float64
}

// Snapshot is the market state an alert is evaluated against
type Snapshot struct {
	Price     float64
	EMA       *float64
	PastPrice *float64
}

// EMA seeds with the SMA of the first period values, matching the frontend calculateEMA.
func EMA(values []float64, period int) []*float64 {
	out := make([]*float64, len(values))
	if period < 1 || len(values) < period {
		return out
	}
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	prev := sum / float64(period)
	out[period-1] = ptr(prev)
	alpha := 2.0 / float64(period+1)
	for i := period; i < len(values); i++ {
		prev = values[i]*alpha + prev*(1-alpha)
		out[i] = ptr(prev)
	}
	return out
}

// BuildSnapshot derives the evaluation snapshot for an alert from a chronological daily series.
func BuildSnapshot(a *models.PriceAlert, series []DayPrice, now time.Time) (Snapshot, bool) {
	if len(series) == 0 {
		return Snapshot{}, false
	}
	s := Snapshot{Price: series[len(series)-1].Price}
	switch a.Condition {
	case models.AlertEMACross:
		values := make([]float64, len(series))
		for i, d := range series {
			values[i] = d.Price
		}
		s.EMA = EMA(values, a.EMAPeriod)[len(values)-1]
	case models.AlertPctChange:
		cutoff := now.AddDate(0, 0, -a.Days)
		for i := len(series) - 1; i >= 0; i-- {
			if !series[i].Day.After(cutoff) {
				s.PastPrice = ptr(series[i].Price)
				break
			}
		}
	}
	return s, true
}

// ShouldFire reports whether the alert's condition is met by moving from its stored baseline to s.
func ShouldFire(a *models.PriceAlert, s Snapshot, now time.Time) bool {
	if !a.Active {
		return false
	}
	if a.Mode == models.AlertModeRecurring && a.LastTriggeredAt != nil &&
		now.Before(a.LastTriggeredAt.Add(time.Duration(a.CooldownHours)*time.Hour)) {
		return false
	}
	switch a.Condition {
	case models.AlertAbove:
		return a.LastPrice != nil && *a.LastPrice < a.TargetPrice && s.Price >= a.TargetPrice
	case models.AlertBelow:
		return a.LastPrice != nil && *a.LastPrice > a.TargetPrice && s.Price <= a.TargetPrice
	case models.AlertPctChange:
		if s.PastPrice == nil || *s.PastPrice <= 0 {
			return false
		}
		change := (s.Price / *s.PastPrice - 1) * 100
		return directional(a.Direction, change >= a.Pct, change <= -a.Pct)
	case models.AlertEMACross:
		if a.LastPrice == nil || a.LastEMA == nil || s.EMA == nil {
			return false
		}
		prev := *a.LastPrice - *a.LastEMA
		cur := s.Price - *s.EMA
		return directional(a.Direction, prev <= 0 && cur > 0, prev >= 0 && cur < 0)
	}
	return false
}

// WindowDays is how many days of history an alert needs for evaluation.
func WindowDays(a *models.PriceAlert) int {
	switch a.Condition {
	case models.AlertPctChange:
		return a.Days + 3
	case models.AlertEMACross:
		return int(math.Min(float64(a.EMAPeriod*3), 600))
	}
	return 3
}

func directional(dir string, up, down bool) bool {
	switch dir {
	case "up":
		return up
	case "down":
		return down
	}
	return up || down
}

func ptr(v float64) *float64 { return &v }
