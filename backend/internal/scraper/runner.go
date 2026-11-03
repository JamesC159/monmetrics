package scraper

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Runner struct {
	DB          *mongo.Database
	Source      SoldSource
	Interval    time.Duration
	MaxPages    int
	Scheduler   *Scheduler
	DailyBudget int
}

func (r *Runner) Run(ctx context.Context) {
	log.Printf("eBay sold scraper running every %s", r.Interval)
	if err := r.RunOnce(ctx, nil); err != nil {
		log.Printf("eBay sold scraper cycle failed: %v", err)
	}
	ticker := time.NewTicker(r.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := r.RunOnce(ctx, nil); err != nil {
				log.Printf("eBay sold scraper cycle failed: %v", err)
			}
		}
	}
}

func (r *Runner) RunOnce(ctx context.Context, cardIDs []primitive.ObjectID) error {
	now := time.Now().UTC()
	scheduler := r.Scheduler
	if scheduler == nil {
		scheduler = &Scheduler{DB: r.DB}
	}

	var entries []models.Card
	if len(cardIDs) > 0 {
		cursor, err := r.DB.Collection("cards").Find(ctx, bson.M{"_id": bson.M{"$in": cardIDs}}, nil)
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)
		if err := cursor.All(ctx, &entries); err != nil {
			return err
		}
	} else {
		if err := scheduler.AssignTiers(ctx, now); err != nil {
			log.Printf("eBay scraper tier assignment failed: %v", err)
		}
		due, err := scheduler.DueCards(ctx, now, r.budget())
		if err != nil {
			return err
		}
		entries = due
	}

	log.Printf("eBay scraper cycle: %d cards due (budget %d)", len(entries), r.budget())
	for _, card := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		sales, err := r.ScrapeCard(ctx, card, true)
		if err != nil {
			if errors.Is(err, ErrBlocked) {
				return err
			}
			log.Printf("eBay scraper card %s (%s): %v", card.ID.Hex(), card.Name, err)
			if markErr := scheduler.MarkFailed(ctx, card, time.Now().UTC()); markErr != nil {
				log.Printf("eBay scraper reschedule failed for %s: %v", card.ID.Hex(), markErr)
			}
			continue
		}
		if markErr := scheduler.MarkScraped(ctx, card, len(sales), time.Now().UTC()); markErr != nil {
			log.Printf("eBay scraper reschedule failed for %s: %v", card.ID.Hex(), markErr)
		}
	}
	return nil
}

func (r *Runner) budget() int {
	if r.DailyBudget > 0 {
		return r.DailyBudget
	}
	return 500
}

func (r *Runner) ScrapeCard(ctx context.Context, card models.Card, persist bool) ([]models.SoldSale, error) {
	if r.Source == nil {
		return nil, errors.New("sold source is required")
	}
	plan := BuildSearchPlan(card)
	stopAt := time.Now().UTC().AddDate(0, 0, -90)
	if card.EbaySyncedAt != nil {
		stopAt = card.EbaySyncedAt.UTC().AddDate(0, 0, -2)
	}
	queries := []string{plan.RawKeywords}
	if plan.GradedKeywords != plan.RawKeywords {
		queries = append(queries, plan.GradedKeywords)
	}
	seen := make(map[string]struct{})
	sales := make([]models.SoldSale, 0)
	for _, keywords := range queries {
		querySales, err := r.Source.FetchSold(ctx, SoldQuery{
			Keywords: keywords, Category: "183454", StopAt: stopAt, MaxPages: r.MaxPages,
		})
		if err != nil {
			return nil, err
		}
		for _, sale := range querySales {
			if _, exists := seen[sale.ItemID]; exists {
				continue
			}
			seen[sale.ItemID] = struct{}{}
			sale.CardID = card.ID
			sale.Source = r.Source.Name()
			sale.Grade = DetectGrade(sale.Title)
			if relevant, reason := Relevant(card, sale.Title); !relevant {
				sale.Excluded, sale.ExcludeReason = true, reason
			}
			sale.ScrapedAt = time.Now().UTC()
			sales = append(sales, sale)
		}
	}
	if persist {
		if err := PersistCard(ctx, r.DB, card, sales, time.Now().UTC()); err != nil {
			return nil, err
		}
	}
	included, excluded := 0, 0
	for _, sale := range sales {
		if sale.Excluded || sale.BestOffer {
			excluded++
		} else {
			included++
		}
	}
	log.Printf("eBay scraper card %s (%s): fetched=%d included=%d excluded=%d", card.ID.Hex(), card.Name, len(sales), included, excluded)
	return sales, nil
}
