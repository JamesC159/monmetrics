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

func TestResumeWindowDefaultsToInterval(t *testing.T) {
	if got := (&Runner{Interval: 6 * time.Hour}).resumeWindow(); got != 6*time.Hour {
		t.Errorf("resumeWindow = %v, want 6h", got)
	}
	if got := (&Runner{Interval: 6 * time.Hour, ResumeWindow: time.Hour}).resumeWindow(); got != time.Hour {
		t.Errorf("explicit resumeWindow = %v, want 1h", got)
	}
	if got := (&Runner{}).resumeWindow(); got != 24*time.Hour {
		t.Errorf("zero-value resumeWindow = %v, want 24h", got)
	}
}
