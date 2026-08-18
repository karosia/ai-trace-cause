package graph_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

func TestGraphFindNodesFiltersByType(t *testing.T) {
	ctx := context.Background()

	g, err := graph.New(memory.New())
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	nodes := []graph.Node{
		{ID: "source-1", Type: "Source"},
		{ID: "decision-1", Type: "Decision"},
		{ID: "decision-2", Type: "Decision"},
	}

	if err := g.AddNodes(ctx, nodes); err != nil {
		t.Fatalf("AddNodes() error = %v", err)
	}

	results, err := g.FindNodes(
		ctx,
		graph.NodeFilter{Type: "Decision"},
	)
	if err != nil {
		t.Fatalf("FindNodes() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf(
			"len(results) = %d, want 2",
			len(results),
		)
	}

	for _, node := range results {
		if node.Type != "Decision" {
			t.Errorf(
				"node.Type = %q, want Decision",
				node.Type,
			)
		}
	}
}

func TestGraphFindNodesFiltersByRecordedAt(t *testing.T) {
	ctx := context.Background()

	g, err := graph.New(memory.New())
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	early := time.Date(
		2026, time.August, 1, 0, 0, 0, 0, time.UTC,
	)
	late := time.Date(
		2026, time.August, 17, 0, 0, 0, 0, time.UTC,
	)

	nodes := []graph.Node{
		{ID: "old", Type: "Fact", RecordedAt: early},
		{ID: "new", Type: "Fact", RecordedAt: late},
	}

	if err := g.AddNodes(ctx, nodes); err != nil {
		t.Fatalf("AddNodes() error = %v", err)
	}

	cutoff := time.Date(
		2026, time.August, 10, 0, 0, 0, 0, time.UTC,
	)

	results, err := g.FindNodes(
		ctx,
		graph.NodeFilter{RecordedAfter: &cutoff},
	)
	if err != nil {
		t.Fatalf("FindNodes() error = %v", err)
	}

	if len(results) != 1 || results[0].ID != "new" {
		t.Fatalf(
			"results = %v, want [new]",
			results,
		)
	}

	results, err = g.FindNodes(
		ctx,
		graph.NodeFilter{RecordedBefore: &cutoff},
	)
	if err != nil {
		t.Fatalf("FindNodes() error = %v", err)
	}

	if len(results) != 1 || results[0].ID != "old" {
		t.Fatalf(
			"results = %v, want [old]",
			results,
		)
	}
}

func TestGraphFindNodesReturnsErrStoreNotQueryable(t *testing.T) {
	ctx := context.Background()

	g, err := graph.New(newFakeStore())
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	_, err = g.FindNodes(ctx, graph.NodeFilter{})

	if !errors.Is(err, graph.ErrStoreNotQueryable) {
		t.Fatalf(
			"FindNodes() error = %v, want ErrStoreNotQueryable",
			err,
		)
	}
}
