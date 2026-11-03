package scraper

import (
	"math"
	"sort"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
)

type DailyAggregate struct {
	Date             time.Time
	Grade            string
	OpenPrice        float64
	ClosePrice       float64
	HighPrice        float64
	LowPrice         float64
	Average          float64
	Volume           int
	WeightedAvgPrice float64
}

func DailyBuckets(sales []models.SoldSale) []DailyAggregate {
	byKey := make(map[dailyKey][]models.SoldSale)
	for _, sale := range sales {
		if sale.Excluded || sale.BestOffer || sale.Price <= 0 || sale.Currency != "USD" || sale.SoldAt.IsZero() {
			continue
		}
		grade := sale.Grade
		if grade == "" {
			grade = DetectGrade(sale.Title)
		}
		date := sale.SoldAt.UTC().Truncate(24 * time.Hour)
		key := dailyKey{date: date, grade: grade}
		byKey[key] = append(byKey[key], sale)
	}

	out := make([]DailyAggregate, 0, len(byKey))
	for key, bucket := range byKey {
		sort.Slice(bucket, func(i, j int) bool { return bucket[i].SoldAt.Before(bucket[j].SoldAt) })
		prices := make([]float64, len(bucket))
		sum := 0.0
		low, high := bucket[0].Price, bucket[0].Price
		for i, sale := range bucket {
			prices[i] = sale.Price
			sum += sale.Price
			low = math.Min(low, sale.Price)
			high = math.Max(high, sale.Price)
		}
		trimmed := trimOutliers(prices)
		trimmedSum := 0.0
		for _, price := range trimmed {
			trimmedSum += price
		}
		average := trimmedSum / float64(len(trimmed))
		out = append(out, DailyAggregate{
			Date: key.date, Grade: key.grade, OpenPrice: bucket[0].Price,
			ClosePrice: bucket[len(bucket)-1].Price, HighPrice: high, LowPrice: low,
			Average: round2(average), Volume: len(bucket), WeightedAvgPrice: round2(sum / float64(len(bucket))),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Date.Equal(out[j].Date) {
			return out[i].Grade < out[j].Grade
		}
		return out[i].Date.Before(out[j].Date)
	})
	return out
}

type dailyKey struct {
	date  time.Time
	grade string
}

func BuildWindowSummaries(buckets []DailyAggregate, now time.Time) map[string]models.PriceWindowStats {
	windows := map[string]time.Time{
		"1d": now.UTC().AddDate(0, 0, -1), "7d": now.UTC().AddDate(0, 0, -7),
		"30d": now.UTC().AddDate(0, 0, -30), "90d": now.UTC().AddDate(0, 0, -90),
		"1y": now.UTC().AddDate(-1, 0, 0), "5y": now.UTC().AddDate(-5, 0, 0),
	}
	result := make(map[string]models.PriceWindowStats, len(windows))
	now = now.UTC()
	for name, from := range windows {
		fromDay := from.Truncate(24 * time.Hour)
		prices := make([]float64, 0)
		count := 0
		minValue, maxValue := math.Inf(1), math.Inf(-1)
		firstDate, lastDate := time.Time{}, time.Time{}
		firstPrice, lastPrice := 0.0, 0.0
		for _, bucket := range buckets {
			if bucket.Date.Before(fromDay) || bucket.Date.After(now) {
				continue
			}
			prices = append(prices, bucket.Average)
			count += bucket.Volume
			minValue = math.Min(minValue, bucket.LowPrice)
			maxValue = math.Max(maxValue, bucket.HighPrice)
			if firstDate.IsZero() || bucket.Date.Before(firstDate) {
				firstDate, firstPrice = bucket.Date, bucket.Average
			}
			if lastDate.IsZero() || bucket.Date.After(lastDate) {
				lastDate, lastPrice = bucket.Date, bucket.Average
			}
		}
		sort.Float64s(prices)
		if len(prices) == 0 {
			result[name] = models.PriceWindowStats{From: from, To: now}
			continue
		}
		sum := 0.0
		for _, price := range prices {
			sum += price
		}
		trimmed := trimOutliers(prices)
		trimmedSum := 0.0
		for _, price := range trimmed {
			trimmedSum += price
		}
		change := 0.0
		if firstPrice != 0 {
			change = (lastPrice - firstPrice) / firstPrice * 100
		}
		result[name] = models.PriceWindowStats{
			Count: count, Average: round2(sum / float64(len(prices))),
			TrimmedMean: round2(trimmedSum / float64(len(trimmed))),
			Median:      round2(percentile(prices, 0.5)), Min: round2(minValue), Max: round2(maxValue),
			ChangePct: round2(change), From: from, To: now,
		}
	}
	return result
}

func trimOutliers(sortedOrUnsorted []float64) []float64 {
	values := append([]float64(nil), sortedOrUnsorted...)
	sort.Float64s(values)
	if len(values) < 4 {
		return values
	}
	q1, q3 := percentile(values, 0.25), percentile(values, 0.75)
	low, high := q1-1.5*(q3-q1), q3+1.5*(q3-q1)
	trimmed := make([]float64, 0, len(values))
	for _, value := range values {
		if value >= low && value <= high {
			trimmed = append(trimmed, value)
		}
	}
	if len(trimmed) == 0 {
		return values
	}
	return trimmed
}

func percentile(sorted []float64, quantile float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	position := quantile * float64(len(sorted)-1)
	lower, upper := int(math.Floor(position)), int(math.Ceil(position))
	return sorted[lower] + (sorted[upper]-sorted[lower])*(position-float64(lower))
}

func round2(value float64) float64 { return math.Round(value*100) / 100 }
