package scraper

import (
	"context"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
)

type SoldQuery struct {
	Keywords string
	Category string
	StopAt   time.Time
	MaxPages int
}

type SoldSource interface {
	Name() string
	FetchSold(context.Context, SoldQuery) ([]models.SoldSale, error)
}
