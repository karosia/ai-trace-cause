package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

func TestStoreFindNodesByType(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	if err := store.PutNodes(ctx, []graph.Node{
		{ID: "source-1", Type: "Source"},
		{ID: "decision-1", Type: "Decision"},
	}); err != nil {
		t.Fatalf("PutNodes() error = %v", err)
	}

	results, err := store.FindNodes(
		ctx,
		graph.NodeFilter{Type: "Decision"},
	)
	if err != nil {
		t.Fatalf("FindNodes() error = %v", err)
	}

	if len(results) != 1 || results[0].ID != "decision-1" {
		t.Fatalf(
			"FindNodes() = %v, want [decision-1]",
			results,
		)
	}
}

func TestStoreFindNodesEmptyFilterMatchesAll(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	if err := store.PutNodes(ctx, []graph.Node{
		{ID: "a", Type: "Source"},
		{ID: "b", Type: "Decision"},
	}); err != nil {
		t.Fatalf("PutNodes() error = %v", err)
	}

	results, err := store.FindNodes(ctx, graph.NodeFilter{})
	if err != nil {
		t.Fatalf("FindNodes() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf(
			"len(results) = %d, want 2",
			len(results),
		)
	}
}

func TestStoreFindNodesFiltersByRecordedAt(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	early := time.Date(
		2026, time.August, 1, 0, 0, 0, 0, time.UTC,
	)
	late := time.Date(
		2026, time.August, 17, 0, 0, 0, 0, time.UTC,
	)

	if err := store.PutNodes(ctx, []graph.Node{
		{ID: "old", Type: "Fact", RecordedAt: early},
		{ID: "new", Type: "Fact", RecordedAt: late},
	}); err != nil {
		t.Fatalf("PutNodes() error = %v", err)
	}

	cutoff := time.Date(
		2026, time.August, 10, 0, 0, 0, 0, time.UTC,
	)

	after, err := store.FindNodes(
		ctx,
		graph.NodeFilter{RecordedAfter: &cutoff},
	)
	if err != nil {
		t.Fatalf("FindNodes() error = %v", err)
	}

	if len(after) != 1 || after[0].ID != "new" {
		t.Fatalf(
			"FindNodes(RecordedAfter) = %v, want [new]",
			after,
		)
	}

	before, err := store.FindNodes(
		ctx,
		graph.NodeFilter{RecordedBefore: &cutoff},
	)
	if err != nil {
		t.Fatalf("FindNodes() error = %v", err)
	}

	if len(before) != 1 || before[0].ID != "old" {
		t.Fatalf(
			"FindNodes(RecordedBefore) = %v, want [old]",
			before,
		)
	}
}
