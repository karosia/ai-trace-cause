package graph_test

import (
	"context"
	"testing"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

// newDiamond builds:
//
//	a --e1--> b --e3--> d
//	a --e2--> c --e4--> d
//	c --e5--> b            (cross edge between same-depth nodes)
//	d --e6--> d            (self-loop)
func newDiamond(t *testing.T) *graph.Graph {
	t.Helper()

	ctx := context.Background()

	g, err := graph.New(memory.New())
	if err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"a", "b", "c", "d"} {
		if err := g.AddNode(ctx, graph.Node{ID: id, Type: "N"}); err != nil {
			t.Fatal(err)
		}
	}

	for _, e := range []graph.Edge{
		{ID: "e1", From: "a", To: "b", Type: "R"},
		{ID: "e2", From: "a", To: "c", Type: "R"},
		{ID: "e3", From: "b", To: "d", Type: "R"},
		{ID: "e4", From: "c", To: "d", Type: "R"},
		{ID: "e5", From: "c", To: "b", Type: "R"},
		{ID: "e6", From: "d", To: "d", Type: "R"},
	} {
		if err := g.AddEdge(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	return g
}

func edgeIDs(edges []graph.Edge) []string {
	ids := make([]string, len(edges))
	for i, edge := range edges {
		ids[i] = edge.ID
	}

	return ids
}

func assertIDs(t *testing.T, got []string, want []string) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestSubgraphCollectsEveryEdgeBetweenVisitedNodes(t *testing.T) {
	g := newDiamond(t)

	sub, err := g.Subgraph(
		context.Background(),
		"d",
		graph.DirectionIncoming,
		5,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(sub.Visits) != 4 {
		t.Fatalf("len(Visits) = %d, want 4", len(sub.Visits))
	}

	// BFS from d (incoming): d, b, c, a. Edges grouped by the node
	// they were collected from, in that order, then by the visit
	// order of the opposite endpoint (b's causes: c before a).
	assertIDs(
		t,
		edgeIDs(sub.Edges),
		[]string{"e6", "e3", "e4", "e5", "e1", "e2"},
	)
}

func TestSubgraphOutgoing(t *testing.T) {
	g := newDiamond(t)

	sub, err := g.Subgraph(
		context.Background(),
		"a",
		graph.DirectionOutgoing,
		5,
	)
	if err != nil {
		t.Fatal(err)
	}

	// BFS from a (outgoing): a, b, c, d.
	assertIDs(
		t,
		edgeIDs(sub.Edges),
		[]string{"e1", "e2", "e3", "e5", "e4", "e6"},
	)
}

func TestSubgraphRespectsMaxDepth(t *testing.T) {
	g := newDiamond(t)

	sub, err := g.Subgraph(
		context.Background(),
		"d",
		graph.DirectionIncoming,
		1,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Visited: d, b, c. a is beyond depth 1, so e1 and e2 are
	// excluded; the cross edge c->b connects two visited nodes and is
	// kept.
	assertIDs(
		t,
		edgeIDs(sub.Edges),
		[]string{"e6", "e3", "e4", "e5"},
	)
}

func TestSubgraphAtFiltersEdgesByTime(t *testing.T) {
	ctx := context.Background()

	g, err := graph.New(memory.New())
	if err != nil {
		t.Fatal(err)
	}

	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 := t0.Add(time.Hour)

	for _, id := range []string{"a", "b"} {
		if err := g.AddNode(ctx, graph.Node{ID: id, Type: "N", RecordedAt: t0}); err != nil {
			t.Fatal(err)
		}
	}

	for _, e := range []graph.Edge{
		{ID: "early", From: "a", To: "b", Type: "R", RecordedAt: t0},
		{ID: "late", From: "a", To: "b", Type: "R2", RecordedAt: t1},
	} {
		if err := g.AddEdge(ctx, e); err != nil {
			t.Fatal(err)
		}
	}

	sub, err := g.SubgraphAt(ctx, "b", graph.DirectionIncoming, 3, t0)
	if err != nil {
		t.Fatal(err)
	}

	assertIDs(t, edgeIDs(sub.Edges), []string{"early"})

	sub, err = g.SubgraphAt(ctx, "b", graph.DirectionIncoming, 3, t1)
	if err != nil {
		t.Fatal(err)
	}

	assertIDs(t, edgeIDs(sub.Edges), []string{"early", "late"})
}

func TestSubgraphPropagatesTraversalErrors(t *testing.T) {
	g := newDiamond(t)

	if _, err := g.Subgraph(
		context.Background(),
		"missing",
		graph.DirectionIncoming,
		1,
	); err == nil {
		t.Fatal("Subgraph(missing) error = nil, want error")
	}

	if _, err := g.SubgraphAt(
		context.Background(),
		"d",
		graph.DirectionIncoming,
		-1,
		time.Now(),
	); err == nil {
		t.Fatal("SubgraphAt(maxDepth=-1) error = nil, want error")
	}
}
