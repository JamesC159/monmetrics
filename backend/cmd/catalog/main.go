package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/jamesc159/monmetrics/configs"
	"github.com/jamesc159/monmetrics/internal/catalog"
	"github.com/jamesc159/monmetrics/internal/database"
)

func main() {
	categoryID := flag.Int("category", 0, "sync a single source category id (0 = all trading card categories)")
	groupID := flag.Int("group", 0, "sync a single group/set id within the category")
	limit := flag.Int("limit", 0, "stop after this many products (0 = unlimited)")
	dryRun := flag.Bool("dry-run", false, "print normalized cards without writing to MongoDB")
	resync := flag.Bool("resync", false, "ignore checkpoints and re-ingest every group")
	resumeWindow := flag.Duration("resume-window", 0, "skip groups ingested more recently than this (0 = CATALOG_RESUME_WINDOW)")
	timeout := flag.Duration("timeout", 2*time.Hour, "overall sync timeout")
	flag.Parse()

	config := configs.Load()
	db, err := database.Connect(config.MongoURI, config.DBName)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Disconnect()

	maxProducts := config.CatalogMaxProducts
	if *limit > 0 {
		maxProducts = *limit
	}
	window := config.CatalogResumeWindow
	if *resumeWindow > 0 {
		window = *resumeWindow
	}

	runner := &catalog.Runner{
		DB:           db,
		Source:       catalog.NewTCGCSVSource(config.CatalogBaseURL, config.CatalogUserAgent, config.CatalogRequestDelay),
		Interval:     config.CatalogInterval,
		Categories:   config.CatalogCategories,
		MaxProducts:  maxProducts,
		ResumeWindow: window,
		Resync:       *resync,
		DryRun:       *dryRun,
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	if err := runner.RunOnce(ctx, catalog.SyncOptions{CategoryID: *categoryID, GroupID: *groupID}); err != nil {
		log.Fatal("catalog sync failed:", err)
	}
}
