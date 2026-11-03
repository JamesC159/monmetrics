package main

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

// countOnly reports one document for the given mark and none for anything else.
func countOnly(target realDataMark) countFunc {
	return func(_ context.Context, collection string, filter bson.M) (int64, error) {
		if collection == target.Collection && reflect.DeepEqual(filter, target.Filter) {
			return 1, nil
		}
		return 0, nil
	}
}

func TestFindRealDataAllowsEmptyDatabase(t *testing.T) {
	empty := func(context.Context, string, bson.M) (int64, error) { return 0, nil }
	if found := findRealData(context.Background(), empty); len(found) != 0 {
		t.Fatalf("empty database reported real data: %v", found)
	}
}

func TestFindRealDataReportsEachMark(t *testing.T) {
	for _, mark := range realDataMarks {
		found := findRealData(context.Background(), countOnly(mark))
		if len(found) != 1 {
			t.Fatalf("%s: got %d findings, want 1: %v", mark.What, len(found), found)
		}
		if !strings.Contains(found[0], mark.What) || !strings.Contains(found[0], mark.Job) {
			t.Errorf("%s: finding %q should name the mark and the job", mark.What, found[0])
		}
	}
}

func TestFindRealDataFailsClosed(t *testing.T) {
	broken := func(context.Context, string, bson.M) (int64, error) { return 0, errors.New("connection reset") }
	found := findRealData(context.Background(), broken)
	if len(found) != len(realDataMarks) {
		t.Fatalf("got %d findings, want one per mark (%d) when checks fail", len(found), len(realDataMarks))
	}
	if !strings.Contains(found[0], "check failed") {
		t.Errorf("finding %q should say the check failed", found[0])
	}
}

func TestEveryDataJobHasAMark(t *testing.T) {
	for _, job := range []string{"make catalog-sync", "make prices", "make scrape"} {
		covered := false
		for _, mark := range realDataMarks {
			if strings.Contains(mark.Job, job) {
				covered = true
				break
			}
		}
		if !covered {
			t.Errorf("no real-data mark covers %s", job)
		}
	}
}
