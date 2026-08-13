package graph_test

import (
	"context"
	"testing"

	"github.com/yourname/ai-trace-cause/graph"
	"github.com/yourname/ai-trace-cause/storage/memory"
)

func TestGraphWithMemoryStore(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	g, err := graph.New(store)
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	observation := graph.Node{
		ID:   "observation-001",
		Type: "Observation",
		Properties: map[string]any{
			"metric": "cpu_usage",
			"value":  94,
		},
	}

	fact := graph.Node{
		ID:   "fact-001",
		Type: "Fact",
		Properties: map[string]any{
			"statement": "CPU usage is high",
		},
	}

	if err := g.AddNode(ctx, observation); err != nil {
		t.Fatalf(
			"AddNode(observation) error = %v",
			err,
		)
	}

	if err := g.AddNode(ctx, fact); err != nil {
		t.Fatalf(
			"AddNode(fact) error = %v",
			err,
		)
	}

	edge := graph.Edge{
		ID:   "edge-001",
		From: observation.ID,
		To:   fact.ID,
		Type: "SUPPORTS",
	}

	if err := g.AddEdge(ctx, edge); err != nil {
		t.Fatalf(
			"AddEdge() error = %v",
			err,
		)
	}

	neighbors, err := g.OutgoingNeighbors(
		ctx,
		observation.ID,
	)
	if err != nil {
		t.Fatalf(
			"OutgoingNeighbors() error = %v",
			err,
		)
	}

	if len(neighbors) != 1 {
		t.Fatalf(
			"len(neighbors) = %d, want 1",
			len(neighbors),
		)
	}

	if neighbors[0].ID != fact.ID {
		t.Errorf(
			"neighbor ID = %q, want %q",
			neighbors[0].ID,
			fact.ID,
		)
	}
}
