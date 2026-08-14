package graph_test

import (
	"context"
	"testing"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

func TestBFS(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	g, err := graph.New(store)
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	nodes := []graph.Node{
		{ID: "a", Type: "Test"},
		{ID: "b", Type: "Test"},
		{ID: "c", Type: "Test"},
		{ID: "d", Type: "Test"},
		{ID: "e", Type: "Test"},
	}

	for _, node := range nodes {
		if err := g.AddNode(ctx, node); err != nil {
			t.Fatalf(
				"AddNode(%s) error = %v",
				node.ID,
				err,
			)
		}
	}

	edges := []graph.Edge{
		{
			ID:   "edge-01",
			From: "a",
			To:   "b",
			Type: "LINK",
		},
		{
			ID:   "edge-02",
			From: "a",
			To:   "c",
			Type: "LINK",
		},
		{
			ID:   "edge-03",
			From: "b",
			To:   "d",
			Type: "LINK",
		},
		{
			ID:   "edge-04",
			From: "c",
			To:   "e",
			Type: "LINK",
		},
	}

	for _, edge := range edges {
		if err := g.AddEdge(ctx, edge); err != nil {
			t.Fatalf(
				"AddEdge(%s) error = %v",
				edge.ID,
				err,
			)
		}
	}

	results, err := g.BFS(
		ctx,
		"a",
		graph.DirectionOutgoing,
		2,
	)
	if err != nil {
		t.Fatalf("BFS() error = %v", err)
	}

	want := []string{
		"a",
		"b",
		"c",
		"d",
		"e",
	}

	if len(results) != len(want) {
		t.Fatalf(
			"len(results) = %d, want %d",
			len(results),
			len(want),
		)
	}

	for i := range want {
		if results[i].Node.ID != want[i] {
			t.Errorf(
				"results[%d] = %q, want %q",
				i,
				results[i].Node.ID,
				want[i],
			)
		}
	}
}
