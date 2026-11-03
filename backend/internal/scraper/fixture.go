package scraper

import (
	"context"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
)

type FixtureSource struct {
	HTML string
}

func (s FixtureSource) Name() string { return models.ProviderEbay }

func (s FixtureSource) FetchSold(ctx context.Context, query SoldQuery) ([]models.SoldSale, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sales, err := ParseSoldListingsHTML(s.HTML, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	if query.StopAt.IsZero() {
		return sales, nil
	}
	filtered := sales[:0]
	for _, sale := range sales {
		if sale.SoldAt.After(query.StopAt) {
			filtered = append(filtered, sale)
		}
	}
	return filtered, nil
}
