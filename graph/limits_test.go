package graph_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

// newFanOut builds a hub node with n outgoing edges to leaf nodes.
func newFanOut(
	t *testing.T,
	n int,
	options ...graph.Option,
) *graph.Graph {
	t.Helper()

	ctx := context.Background()

	g, err := graph.New(memory.New(), options...)
	if err != nil {
		t.Fatal(err)
	}

	nodes := []graph.Node{{ID: "hub", Type: "N"}}
	edges := make([]graph.Edge, 0, n)

	for i := range n {
		id := fmt.Sprintf("leaf-%03d", i)
		nodes = append(nodes, graph.Node{ID: id, Type: "N"})
		edges = append(edges, graph.Edge{
			ID:   "e-" + id,
			From: "hub",
			To:   id,
			Type: "R",
		})
	}

	if err := g.AddBatch(ctx, nodes, edges); err != nil {
		t.Fatal(err)
	}

	return g
}

func TestMaxVisitsAppliesToEveryTraversal(t *testing.T) {
	ctx := context.Background()
	g := newFanOut(t, 20, graph.WithMaxVisits(10))

	traversals := map[string]func() error{
		"BFS": func() error {
			_, err := g.BFS(ctx, "hub", graph.DirectionOutgoing, 1)
			return err
		},
		"DFS": func() error {
			_, err := g.DFS(ctx, "hub", graph.DirectionOutgoing, 1)
			return err
		},
		"BFSAt": func() error {
			_, err := g.BFSAt(ctx, "hub", graph.DirectionOutgoing, 1, time.Now())
			return err
		},
		"Subgraph": func() error {
			_, err := g.Subgraph(ctx, "hub", graph.DirectionOutgoing, 1)
			return err
		},
	}

	for name, traverse := range traversals {
		if err := traverse(); !errors.Is(err, graph.ErrMaxVisitsExceeded) {
			t.Errorf("%s: err = %v, want ErrMaxVisitsExceeded", name, err)
		}
	}
}

func TestMaxVisitsAllowsTraversalsWithinLimit(t *testing.T) {
	ctx := context.Background()

	// 1 hub + 9 leaves = exactly 10 visits.
	g := newFanOut(t, 9, graph.WithMaxVisits(10))

	visits, err := g.BFS(ctx, "hub", graph.DirectionOutgoing, 1)
	if err != nil {
		t.Fatalf("BFS() error = %v", err)
	}

	if len(visits) != 10 {
		t.Fatalf("len(visits) = %d, want 10", len(visits))
	}

	visits, err = g.DFS(ctx, "hub", graph.DirectionOutgoing, 1)
	if err != nil || len(visits) != 10 {
		t.Fatalf("DFS() = %d visits, err = %v; want 10, nil", len(visits), err)
	}
}

func TestNoMaxVisitsByDefault(t *testing.T) {
	g := newFanOut(t, 50)

	visits, err := g.BFS(context.Background(), "hub", graph.DirectionOutgoing, 1)
	if err != nil || len(visits) != 51 {
		t.Fatalf("BFS() = %d visits, err = %v; want 51, nil", len(visits), err)
	}
}

func TestAddBatchIsAtomicOnMemoryStore(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	g, err := graph.New(store)
	if err != nil {
		t.Fatal(err)
	}

	err = g.AddBatch(
		ctx,
		[]graph.Node{{ID: "a", Type: "N"}, {ID: "b", Type: "N"}},
		[]graph.Edge{
			{ID: "ok", From: "a", To: "b", Type: "R"},
			{ID: "bad", From: "a", To: "missing", Type: "R"},
		},
	)
	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf("AddBatch() error = %v, want ErrNodeNotFound", err)
	}

	if _, err := g.GetNode(ctx, "a"); !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf("node from failed batch was written: err = %v", err)
	}

	// Duplicate node ID within the batch.
	err = g.AddBatch(
		ctx,
		[]graph.Node{{ID: "a", Type: "N"}, {ID: "a", Type: "N"}},
		nil,
	)
	if !errors.Is(err, graph.ErrNodeAlreadyExists) {
		t.Fatalf("duplicate node: err = %v", err)
	}

	// A valid batch whose edges reference nodes in the same batch.
	if err := g.AddBatch(
		ctx,
		[]graph.Node{{ID: "a", Type: "N"}, {ID: "b", Type: "N"}},
		[]graph.Edge{{ID: "ab", From: "a", To: "b", Type: "R"}},
	); err != nil {
		t.Fatalf("AddBatch() error = %v", err)
	}

	neighbors, err := g.OutgoingNeighbors(ctx, "a")
	if err != nil || len(neighbors) != 1 {
		t.Fatalf("OutgoingNeighbors() = %v, %v", neighbors, err)
	}

	// Duplicate edge ID against the store.
	err = g.AddBatch(
		ctx,
		[]graph.Node{{ID: "c", Type: "N"}},
		[]graph.Edge{{ID: "ab", From: "a", To: "c", Type: "R"}},
	)
	if !errors.Is(err, graph.ErrEdgeAlreadyExists) {
		t.Fatalf("duplicate edge: err = %v", err)
	}

	if _, err := g.GetNode(ctx, "c"); !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatal("node c from failed batch was written")
	}
}

func TestAddBatchFallsBackWithoutAtomicStore(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()

	g, err := graph.New(store)
	if err != nil {
		t.Fatal(err)
	}

	if err := g.AddBatch(
		ctx,
		[]graph.Node{{ID: "a", Type: "N"}, {ID: "b", Type: "N"}},
		[]graph.Edge{{ID: "ab", From: "a", To: "b", Type: "R"}},
	); err != nil {
		t.Fatalf("AddBatch() error = %v", err)
	}

	if store.putNodeCalls != 2 || store.putEdgeCalls != 1 {
		t.Fatalf(
			"putNodeCalls = %d, putEdgeCalls = %d; want 2 and 1",
			store.putNodeCalls,
			store.putEdgeCalls,
		)
	}
}

func TestAddBatchValidatesBeforeWriting(t *testing.T) {
	ctx := context.Background()
	store := newFakeStore()

	g, err := graph.New(store)
	if err != nil {
		t.Fatal(err)
	}

	err = g.AddBatch(
		ctx,
		[]graph.Node{{ID: "a", Type: "N"}},
		[]graph.Edge{{ID: "", From: "a", To: "a", Type: "R"}},
	)
	if !errors.Is(err, graph.ErrEmptyEdgeID) {
		t.Fatalf("err = %v, want ErrEmptyEdgeID", err)
	}

	if store.putNodeCalls != 0 {
		t.Fatal("node written although the batch was invalid")
	}

	err = g.AddBatch(
		ctx,
		[]graph.Node{{ID: "a"}},
		nil,
	)
	if !errors.Is(err, graph.ErrEmptyNodeType) {
		t.Fatalf("err = %v, want ErrEmptyNodeType", err)
	}
}

func TestFindNodesByProperties(t *testing.T) {
	ctx := context.Background()

	g, err := graph.New(memory.New())
	if err != nil {
		t.Fatal(err)
	}

	for _, node := range []graph.Node{
		{ID: "a", Type: "N", Properties: map[string]any{"run": "r1", "n": 1}},
		{ID: "b", Type: "N", Properties: map[string]any{"run": "r1", "n": 2}},
		{ID: "c", Type: "N", Properties: map[string]any{"run": "r2", "n": 1}},
		{ID: "d", Type: "N"},
	} {
		if err := g.AddNode(ctx, node); err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		properties map[string]any
		want       []string
	}{
		{map[string]any{"run": "r1"}, []string{"a", "b"}},
		{map[string]any{"run": "r1", "n": 1}, []string{"a"}},
		{map[string]any{"n": int64(1)}, nil}, // dynamic types must match
		{map[string]any{"missing": "x"}, nil},
		{nil, []string{"a", "b", "c", "d"}},
	}

	for _, tc := range cases {
		nodes, err := g.FindNodes(ctx, graph.NodeFilter{Properties: tc.properties})
		if err != nil {
			t.Fatal(err)
		}

		got := make([]string, len(nodes))
		for i, node := range nodes {
			got[i] = node.ID
		}

		if fmt.Sprint(got) != fmt.Sprint(tc.want) && !(len(got) == 0 && len(tc.want) == 0) {
			t.Errorf("Properties %v: got %v, want %v", tc.properties, got, tc.want)
		}
	}
}
