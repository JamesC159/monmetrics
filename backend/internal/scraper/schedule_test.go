package scraper

import (
	"strings"
	"testing"
	"time"

	"github.com/jamesc159/monmetrics/internal/models"
)

func TestSchedulerIntervalFallsBackPerTier(t *testing.T) {
	scheduler := &Scheduler{Intervals: map[int]time.Duration{models.ScrapeTierHot: 6 * time.Hour}}
	if got := scheduler.interval(models.ScrapeTierHot); got != 6*time.Hour {
		t.Errorf("configured tier interval = %v, want 6h", got)
	}
	if got := scheduler.interval(models.ScrapeTierWarm); got != 7*24*time.Hour {
		t.Errorf("warm default = %v, want 168h", got)
	}
	if got := scheduler.interval(models.ScrapeTierColdTail); got != 30*24*time.Hour {
		t.Errorf("cold default = %v, want 720h", got)
	}
}

func TestBackoffGrowsAndCaps(t *testing.T) {
	base := time.Hour
	if got := backoff(base, 0); got != base {
		t.Errorf("no failures = %v, want %v", got, base)
	}
	if got := backoff(base, 3); got != 4*base {
		t.Errorf("three failures = %v, want %v", got, 4*base)
	}
	capped := backoff(base, 50)
	if capped != base*time.Duration(1<<uint(maxFailureStep-1)) {
		t.Errorf("backoff not capped: %v", capped)
	}
}

func TestBuildSearchPlanSkipsGradeExclusionsForSealed(t *testing.T) {
	sealed := models.Card{Name: "Silver Tempest Elite Trainer Box", Set: "SWSH12", Category: "sealed"}
	plan := BuildSearchPlan(sealed)
	if plan.RawKeywords != plan.GradedKeywords {
		t.Errorf("sealed plan should use one query, got %q / %q", plan.RawKeywords, plan.GradedKeywords)
	}
	if strings.Contains(plan.RawKeywords, "-psa") {
		t.Errorf("sealed plan should not exclude grades: %q", plan.RawKeywords)
	}

	single := models.Card{Name: "Lugia VSTAR", Number: "211/195", Set: "Silver Tempest", Category: "card"}
	if !strings.Contains(BuildSearchPlan(single).RawKeywords, "-psa") {
		t.Error("single raw plan should exclude graded listings")
	}
}
