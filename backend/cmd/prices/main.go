package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/jamesc159/monmetrics/configs"
	"github.com/jamesc159/monmetrics/internal/catalog"
	"github.com/jamesc159/monmetrics/internal/database"
	"github.com/jamesc159/monmetrics/internal/pricing"
)

func main() {
	maxGroups := flag.Int("max-groups", 0, "cap groups per run (0 = every group we hold cards for)")
	dryRun := flag.Bool("dry-run", false, "fetch and join but write nothing")
	timeout := flag.Duration("timeout", 2*time.Hour, "overall run timeout")
	backfill := flag.Bool("backfill-printings", false, "fill in card printings from stored prices, then exit (no network)")
	flag.Parse()

	config := configs.Load()
	db, err := database.Connect(config.MongoURI, config.DBName)
	if err != nil {
		log.Fatal("connect: ", err)
	}
	defer database.Disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if *backfill {
		start := time.Now()
		updated, err := pricing.BackfillPrintings(ctx, db, *dryRun)
		log.Printf("printings backfilled=%d dry_run=%t in %s", updated, *dryRun, time.Since(start).Round(time.Second))
		if err != nil {
			log.Fatal("printings backfill failed: ", err)
		}
		return
	}

	source := catalog.NewTCGCSVSource(
		config.CatalogBaseURL,
		config.CatalogUserAgent,
		config.PriceRequestDelay,
	)

	if stamp, err := source.LastUpdated(ctx); err == nil {
		log.Printf("tcgcsv last updated %s", stamp.Format(time.RFC3339))
	}

	runner := &pricing.Runner{
		DB:        db,
		Source:    source,
		MaxGroups: *maxGroups,
		DryRun:    *dryRun,
	}

	start := time.Now()
	result, err := runner.RunOnce(ctx)
	log.Printf("groups=%d products=%d price_points=%d cards_priced=%d skipped=%d in %s",
		result.Groups, result.Products, result.PricePoints, result.CardsPriced, result.Skipped,
		time.Since(start).Round(time.Second))
	if err != nil {
		log.Fatal("price ingest failed: ", err)
	}
}
