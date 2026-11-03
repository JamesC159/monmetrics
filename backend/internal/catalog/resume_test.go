package catalog

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

type stubSource struct {
	groups      []Group
	failOnGroup int
	emptyGroup  int // answers like TCGCSV does for a group with no products
	fetched     []int
}

func (s *stubSource) Name() string { return "stub" }

func (s *stubSource) Categories(ctx context.Context) ([]Category, error) {
	return []Category{{CategoryID: 3, Name: "Pokemon", DisplayName: "Pokemon"}}, nil
}

func (s *stubSource) Groups(ctx context.Context, categoryID int) ([]Group, error) {
	return s.groups, nil
}

func (s *stubSource) Products(ctx context.Context, categoryID, groupID int) ([]Product, error) {
	s.fetched = append(s.fetched, groupID)
	if groupID == s.failOnGroup {
		return nil, errors.New("source exploded")
	}
	if groupID == s.emptyGroup {
		return nil, fmt.Errorf("%w: HTTP 404", ErrNotFound)
	}
	return []Product{{ProductID: groupID * 10, GroupID: groupID, CategoryID: categoryID, Name: "Card"}}, nil
}

func TestBroadSyncMovesPastEmptyGroups(t *testing.T) {
	source := &stubSource{groups: []Group{{GroupID: 1}, {GroupID: 2}, {GroupID: 3}}, emptyGroup: 2}
	runner := &Runner{Source: source, Interval: time.Hour, DryRun: true}

	if err := runner.RunOnce(context.Background(), SyncOptions{}); err != nil {
		t.Fatalf("a group with no product file should not stop a broad sync: %v", err)
	}
	if len(source.fetched) != 3 {
		t.Errorf("expected every group to be fetched, got %v", source.fetched)
	}

	if err := runner.RunOnce(context.Background(), SyncOptions{GroupID: 2}); !errors.Is(err, ErrNotFound) {
		t.Errorf("a targeted sync should report the missing product file, got %v", err)
	}
}

func TestRunOnceStopsAtFailingGroup(t *testing.T) {
	source := &stubSource{
		groups:      []Group{{GroupID: 1}, {GroupID: 2}, {GroupID: 3}, {GroupID: 4}},
		failOnGroup: 3,
	}
	// DryRun keeps the runner off MongoDB; checkpoint persistence is covered live.
	runner := &Runner{Source: source, Interval: time.Hour, DryRun: true}

	if err := runner.RunOnce(context.Background(), SyncOptions{}); err == nil {
		t.Fatal("expected RunOnce to surface the group error")
	}
	if len(source.fetched) != 3 {
		t.Fatalf("expected the sync to stop at the failing group, fetched %v", source.fetched)
	}
	if source.fetched[2] != 3 {
		t.Errorf("expected group 3 to be the last attempted, got %v", source.fetched)
	}
}

func TestShouldSkipGroupUsesIngestionTimestamp(t *testing.T) {
	now := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	cutoff := now.Add(-24 * time.Hour)
	checkpoints := map[string]time.Time{
		checkpointKey(3, 1): now.Add(-2 * time.Hour),  // ingested recently
		checkpointKey(3, 2): now.Add(-48 * time.Hour), // stale, must re-ingest
	}

	if !shouldSkipGroup(checkpoints, 3, 1, cutoff) {
		t.Error("group ingested within the resume window should be skipped")
	}
	if shouldSkipGroup(checkpoints, 3, 2, cutoff) {
		t.Error("group ingested before the resume window should be re-ingested")
	}
	if shouldSkipGroup(checkpoints, 3, 99, cutoff) {
		t.Error("group without a checkpoint should be ingested")
	}
}

func TestIngestedAtRequiresPricesWhenTheSyncStoresThem(t *testing.T) {
	completed := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	earlier := completed.Add(-30 * time.Hour)

	unpriced := GroupCheckpoint{CompletedAt: completed}
	if got := unpriced.ingestedAt(false); !got.Equal(completed) {
		t.Errorf("a products-only sync should resume from completed_at, got %v", got)
	}
	if got := unpriced.ingestedAt(true); !got.IsZero() {
		t.Errorf("a group checkpointed without prices must be ingested again, got %v", got)
	}

	priced := GroupCheckpoint{CompletedAt: completed, PricesSyncedAt: &completed}
	if got := priced.ingestedAt(true); !got.Equal(completed) {
		t.Errorf("a priced group should resume from completed_at, got %v", got)
	}

	stalePrices := GroupCheckpoint{CompletedAt: completed, PricesSyncedAt: &earlier}
	if got := stalePrices.ingestedAt(true); !got.Equal(earlier) {
		t.Errorf("a later products-only pass must not hide stale prices, got %v", got)
	}
}

func TestResumeTimesFallBackToCardPriceTimes(t *testing.T) {
	completed := time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC)
	pricedLater := completed.Add(2 * time.Hour)
	records := []GroupCheckpoint{
		{CategoryID: 1, GroupID: 1, CompletedAt: completed},                             // cards priced since, e.g. by `make prices`
		{CategoryID: 1, GroupID: 2, CompletedAt: completed},                             // never priced
		{CategoryID: 1, GroupID: 3, CompletedAt: completed, PricesSyncedAt: &completed}, // priced by the catalog pass
	}
	priced := map[string]time.Time{checkpointKey(1, 1): pricedLater, checkpointKey(1, 3): pricedLater}

	got := resumeTimes(records, priced, true)
	if at, ok := got[checkpointKey(1, 1)]; !ok || !at.Equal(completed) {
		t.Errorf("a group whose cards were priced should resume from its checkpoint, got %v (present %v)", at, ok)
	}
	if _, ok := got[checkpointKey(1, 2)]; ok {
		t.Error("a group with no stored prices must be ingested again")
	}
	if at := got[checkpointKey(1, 3)]; !at.Equal(completed) {
		t.Errorf("the checkpoint's own prices time should win, got %v", at)
	}

	if all := resumeTimes(records, nil, false); len(all) != 3 {
		t.Errorf("a products-only sync should resume past every completed group, got %v", all)
	}
}

func TestStoresPricesNeedsCombinedSourceAndSink(t *testing.T) {
	combined := &productPriceStub{stubSource: &stubSource{}}
	for _, test := range []struct {
		name   string
		runner *Runner
		want   bool
	}{
		{"products-only source", &Runner{Source: &stubSource{}, Prices: &sinkStub{}}, false},
		{"no sink", &Runner{Source: combined}, false},
		{"combined source and sink", &Runner{Source: combined, Prices: &sinkStub{}}, true},
	} {
		if got := test.runner.storesPrices(); got != test.want {
			t.Errorf("%s: storesPrices = %v, want %v", test.name, got, test.want)
		}
	}
}

func TestResumeWindowDefaultsToHalfInterval(t *testing.T) {
	if got := (&Runner{Interval: 6 * time.Hour}).resumeWindow(); got != 3*time.Hour {
		t.Errorf("resumeWindow = %v, want 3h", got)
	}
	if got := (&Runner{Interval: 6 * time.Hour, ResumeWindow: time.Hour}).resumeWindow(); got != time.Hour {
		t.Errorf("explicit resumeWindow = %v, want 1h", got)
	}
	if got := (&Runner{}).resumeWindow(); got != 12*time.Hour {
		t.Errorf("zero-value resumeWindow = %v, want 12h", got)
	}
}

// A daily run starts one interval after the last one, so the groups it finished are
// about an interval old. They must not be skipped, or every daily run after the first
// does nothing.
func TestDailyRerunReingestsPreviousRunsGroups(t *testing.T) {
	runner := &Runner{Interval: 24 * time.Hour}
	now := time.Date(2026, 11, 4, 3, 0, 0, 0, time.UTC)
	previousRunGroup := now.Add(-24*time.Hour + 5*time.Minute) // finished 5 minutes after yesterday's 03:00 start
	checkpoints := map[string]time.Time{checkpointKey(3, 1): previousRunGroup}

	if shouldSkipGroup(checkpoints, 3, 1, now.Add(-runner.resumeWindow())) {
		t.Errorf("group ingested one interval ago should be re-ingested on the daily run")
	}
	// A run that died partway and is retried shortly after should still resume past it.
	retryGroup := now.Add(-time.Hour)
	checkpoints[checkpointKey(3, 2)] = retryGroup
	if !shouldSkipGroup(checkpoints, 3, 2, now.Add(-runner.resumeWindow())) {
		t.Errorf("group ingested an hour ago should be skipped when a retry resumes")
	}
}
