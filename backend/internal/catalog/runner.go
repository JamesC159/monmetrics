package catalog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
	"go.mongodb.org/mongo-driver/mongo"
)

// Runner periodically refreshes the card catalogue from a Source.
type Runner struct {
	DB           *mongo.Database
	Source       Source
	Interval     time.Duration
	Categories   []int         // empty = every trading-card category the source exposes
	MaxProducts  int           // 0 = unlimited; used to bound dev/test syncs
	ResumeWindow time.Duration // groups ingested more recently than this are skipped; defaults to Interval
	Resync       bool          // ignore checkpoints and re-ingest every group
	DryRun       bool
	// Prices, when set and Source is a ProductPriceSource, stores each group's prices from
	// the same request as its products, so no separate price ingest is needed that day.
	Prices PriceSink
}

// SyncOptions narrows a single sync, used by the CLI for targeted backfills.
type SyncOptions struct {
	CategoryID int
	GroupID    int
}

// resumeWindow is how recently a group must have been ingested to be skipped. It is
// half the interval, not the whole thing: a run that finished yesterday is about one
// interval old when the next daily run starts, so a full-interval window would skip
// every group and the daily sync would never re-ingest anything.
func (r *Runner) resumeWindow() time.Duration {
	if r.ResumeWindow > 0 {
		return r.ResumeWindow
	}
	if r.Interval > 0 {
		return r.Interval / 2
	}
	return 12 * time.Hour
}

func (r *Runner) Run(ctx context.Context) {
	log.Printf("catalog sync running every %s", r.Interval)
	if err := r.RunOnce(ctx, SyncOptions{}); err != nil {
		log.Printf("catalog sync cycle failed: %v", err)
	}
	ticker := time.NewTicker(r.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.RunOnce(ctx, SyncOptions{}); err != nil {
				log.Printf("catalog sync cycle failed: %v", err)
			}
		}
	}
}

func (r *Runner) RunOnce(ctx context.Context, opts SyncOptions) error {
	categories, err := r.selectCategories(ctx, opts)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	checkpoints, err := r.loadResumePoint(ctx)
	if err != nil {
		return err
	}
	resumeCutoff := now.Add(-r.resumeWindow())
	// One UTC-midnight price stamp for the whole run, as the price ingest does.
	day := now.Truncate(24 * time.Hour)

	total, inserted, skipped := 0, int64(0), 0
	for _, category := range categories {
		groups, err := r.Source.Groups(ctx, category.CategoryID)
		if err != nil {
			if opts.CategoryID == 0 && errors.Is(err, ErrNotFound) {
				log.Printf("catalog sync skipping %s (%d): %v", category.DisplayName, category.CategoryID, err)
				continue
			}
			return err
		}
		for _, group := range groups {
			if opts.GroupID > 0 && group.GroupID != opts.GroupID {
				continue
			}
			if shouldSkipGroup(checkpoints, category.CategoryID, group.GroupID, resumeCutoff) {
				skipped++
				continue
			}
			count, upserted, err := r.syncGroup(ctx, category, group, day)
			if err != nil && opts.GroupID == 0 && errors.Is(err, ErrNotFound) {
				// TCGCSV publishes no product file for an empty group. Checkpoint it as empty
				// so a broad sync moves past it, and retries it after the resume window.
				log.Printf("catalog sync %s / %s has no product file; recording it as empty", category.DisplayName, group.Name)
				count, upserted, err = 0, 0, nil
			}
			if err != nil {
				r.recordFailure(ctx, category, group, err, time.Now().UTC())
				log.Printf("catalog sync stopped at %s / %s (resumes here next run): %v", category.DisplayName, group.Name, err)
				return err
			}
			r.recordProgress(ctx, category, group, count, time.Now().UTC())
			total += count
			inserted += upserted
			if r.MaxProducts > 0 && total >= r.MaxProducts {
				log.Printf("catalog sync stopped at product cap %d", r.MaxProducts)
				return nil
			}
		}
	}
	log.Printf("catalog sync complete: categories=%d products=%d new=%d resumed_skips=%d", len(categories), total, inserted, skipped)
	return nil
}

// shouldSkipGroup reports whether a group was ingested recently enough that this
// run can resume past it.
func shouldSkipGroup(checkpoints map[string]time.Time, categoryID, groupID int, cutoff time.Time) bool {
	ingestedAt, ok := checkpoints[checkpointKey(categoryID, groupID)]
	return ok && ingestedAt.After(cutoff)
}

// loadResumePoint reads which groups were already ingested in a previous run.
// A checkpoint read failure is fatal: syncing without it would silently restart
// from the first category and re-fetch the whole catalog.
func (r *Runner) loadResumePoint(ctx context.Context) (map[string]time.Time, error) {
	if r.Resync || r.DryRun || r.DB == nil {
		return map[string]time.Time{}, nil
	}
	// When this sync stores prices, a group checkpointed without them is ingested again.
	checkpoints, err := LoadCheckpoints(ctx, r.DB, r.Source.Name(), r.storesPrices())
	if err != nil {
		return nil, err
	}
	if len(checkpoints) > 0 {
		what := "ingested"
		if r.storesPrices() {
			what = "ingested with prices"
		}
		log.Printf("catalog sync resuming: %d groups already %s within %s", len(checkpoints), what, r.resumeWindow())
	}
	return checkpoints, nil
}

// recordProgress checkpoints a group that synced without error. When the sync stores
// prices, that includes the group's prices, or confirms it has none.
func (r *Runner) recordProgress(ctx context.Context, category Category, group Group, products int, now time.Time) {
	if r.DryRun || r.DB == nil {
		return
	}
	checkpoint := GroupCheckpoint{
		Source: r.Source.Name(), CategoryID: category.CategoryID, GroupID: group.GroupID,
		CategoryName: category.DisplayName, GroupName: group.Name, Products: products, CompletedAt: now,
	}
	if r.storesPrices() {
		checkpoint.PricesSyncedAt = &now
	}
	if err := SaveCheckpoint(ctx, r.DB, checkpoint); err != nil {
		log.Printf("catalog checkpoint write failed for %s / %s: %v", category.DisplayName, group.Name, err)
	}
}

func (r *Runner) recordFailure(ctx context.Context, category Category, group Group, cause error, now time.Time) {
	if r.DryRun || r.DB == nil || ctx.Err() != nil {
		return
	}
	if err := SaveCheckpointFailure(ctx, r.DB, r.Source.Name(), category, group, cause, now); err != nil {
		log.Printf("catalog checkpoint failure write failed for %s / %s: %v", category.DisplayName, group.Name, err)
	}
}

func (r *Runner) syncGroup(ctx context.Context, category Category, group Group, day time.Time) (int, int64, error) {
	products, prices, withPrices, err := r.fetchGroup(ctx, category.CategoryID, group.GroupID)
	if err != nil {
		return 0, 0, err
	}
	if r.MaxProducts > 0 && len(products) > r.MaxProducts {
		products = products[:r.MaxProducts]
	}
	cards := make([]models.Card, 0, len(products))
	for _, product := range products {
		cards = append(cards, NormalizeProduct(r.Source.Name(), category, group, product))
	}
	if r.DryRun {
		for _, card := range cards {
			log.Printf("[dry-run] %s | %s | %s | %s | %s", card.Game, card.Set, card.Name, card.Category, card.ExternalID)
		}
		if withPrices {
			log.Printf("[dry-run] %s / %s: price_rows=%d", category.DisplayName, group.Name, len(prices))
		}
		return len(cards), 0, nil
	}
	result, err := UpsertCards(ctx, r.DB, cards, time.Now().UTC())
	if err != nil {
		return 0, 0, err
	}
	if !withPrices {
		log.Printf("catalog sync %s / %s: products=%d new=%d", category.DisplayName, group.Name, len(cards), result.Upserted)
		return len(cards), result.Upserted, nil
	}
	// Prices join onto cards by product id, so they go in after the upsert creates new cards.
	points, err := r.Prices.IngestGroupPrices(ctx, category.CategoryID, group.GroupID, prices, day)
	if err != nil {
		return 0, 0, fmt.Errorf("store prices: %w", err)
	}
	log.Printf("catalog sync %s / %s: products=%d new=%d price_points=%d", category.DisplayName, group.Name, len(cards), result.Upserted, points)
	return len(cards), result.Upserted, nil
}

// fetchGroup reads a group's products, and its prices too when the source serves both in
// one request and a sink is set to store them. withPrices reports which happened.
func (r *Runner) fetchGroup(ctx context.Context, categoryID, groupID int) (products []Product, prices []ProductPrice, withPrices bool, err error) {
	if r.storesPrices() {
		products, prices, err = r.Source.(ProductPriceSource).ProductsAndPrices(ctx, categoryID, groupID)
		return products, prices, true, err
	}
	products, err = r.Source.Products(ctx, categoryID, groupID)
	return products, nil, false, err
}

// storesPrices reports whether each group's prices are stored in the same pass as its
// products: the source serves both in one request and a sink is set.
func (r *Runner) storesPrices() bool {
	_, ok := r.Source.(ProductPriceSource)
	return ok && r.Prices != nil
}

func (r *Runner) selectCategories(ctx context.Context, opts SyncOptions) ([]Category, error) {
	categories, err := r.Source.Categories(ctx)
	if err != nil {
		return nil, err
	}
	allowed := make(map[int]bool, len(r.Categories))
	for _, id := range r.Categories {
		allowed[id] = true
	}
	selected := make([]Category, 0, len(categories))
	for _, category := range categories {
		if opts.CategoryID > 0 && category.CategoryID != opts.CategoryID {
			continue
		}
		if len(allowed) > 0 && !allowed[category.CategoryID] {
			continue
		}
		if opts.CategoryID == 0 && !IsTradingCardCategory(category) {
			continue
		}
		selected = append(selected, category)
	}
	return selected, nil
}
