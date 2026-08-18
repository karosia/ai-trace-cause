package graph_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

// fakeStore is a minimal graph.Store implementation with none of the
// optional capabilities (NodeBatchStore, EdgeBatchStore, Queryable),
// used to exercise Graph's fallback behavior and the
// ErrStoreNotQueryable path.
type fakeStore struct {
	nodes map[string]graph.Node
	edges map[string]graph.Edge

	putNodeCalls int
	putEdgeCalls int
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		nodes: make(map[string]graph.Node),
		edges: make(map[string]graph.Edge),
	}
}

func (f *fakeStore) PutNode(_ context.Context, node graph.Node) error {
	f.putNodeCalls++
	f.nodes[node.ID] = node
	return nil
}

func (f *fakeStore) GetNode(_ context.Context, id string) (graph.Node, error) {
	node, ok := f.nodes[id]
	if !ok {
		return graph.Node{}, fmt.Errorf("%w: %s", graph.ErrNodeNotFound, id)
	}
	return node, nil
}

func (f *fakeStore) PutEdge(_ context.Context, edge graph.Edge) error {
	f.putEdgeCalls++
	f.edges[edge.ID] = edge
	return nil
}

func (f *fakeStore) GetEdge(_ context.Context, id string) (graph.Edge, error) {
	edge, ok := f.edges[id]
	if !ok {
		return graph.Edge{}, fmt.Errorf("%w: %s", graph.ErrEdgeNotFound, id)
	}
	return edge, nil
}

func (f *fakeStore) OutgoingEdges(_ context.Context, nodeID string) ([]graph.Edge, error) {
	var out []graph.Edge
	for _, edge := range f.edges {
		if edge.From == nodeID {
			out = append(out, edge)
		}
	}
	return out, nil
}

func (f *fakeStore) IncomingEdges(_ context.Context, nodeID string) ([]graph.Edge, error) {
	var in []graph.Edge
	for _, edge := range f.edges {
		if edge.To == nodeID {
			in = append(in, edge)
		}
	}
	return in, nil
}

var _ graph.Store = (*fakeStore)(nil)

func TestGraphAddNodesUsesBatchStoreWhenAvailable(t *testing.T) {
	ctx := context.Background()

	g, err := graph.New(memory.New())
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	nodes := []graph.Node{
		{ID: "a", Type: "Source"},
		{ID: "b", Type: "Observation"},
	}

	if err := g.AddNodes(ctx, nodes); err != nil {
		t.Fatalf("AddNodes() error = %v", err)
	}

	for _, id := range []string{"a", "b"} {
		if _, err := g.GetNode(ctx, id); err != nil {
			t.Errorf("GetNode(%q) error = %v", id, err)
		}
	}
}

func TestGraphAddNodesFallsBackWithoutBatchStore(t *testing.T) {
	ctx := context.Background()

	store := newFakeStore()

	g, err := graph.New(store)
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	nodes := []graph.Node{
		{ID: "a", Type: "Source"},
		{ID: "b", Type: "Observation"},
	}

	if err := g.AddNodes(ctx, nodes); err != nil {
		t.Fatalf("AddNodes() error = %v", err)
	}

	if store.putNodeCalls != 2 {
		t.Errorf(
			"putNodeCalls = %d, want 2",
			store.putNodeCalls,
		)
	}
}

func TestGraphAddNodesRejectsEmptyID(t *testing.T) {
	ctx := context.Background()

	g, err := graph.New(memory.New())
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	err = g.AddNodes(
		ctx,
		[]graph.Node{{Type: "Source"}},
	)

	if !errors.Is(err, graph.ErrEmptyNodeID) {
		t.Fatalf(
			"AddNodes() error = %v, want ErrEmptyNodeID",
			err,
		)
	}
}

func TestGraphAddEdgesUsesBatchStoreWhenAvailable(t *testing.T) {
	ctx := context.Background()

	g, err := graph.New(memory.New())
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	nodes := []graph.Node{
		{ID: "a", Type: "Source"},
		{ID: "b", Type: "Observation"},
	}

	if err := g.AddNodes(ctx, nodes); err != nil {
		t.Fatalf("AddNodes() error = %v", err)
	}

	edges := []graph.Edge{
		{ID: "e1", From: "a", To: "b", Type: "PRODUCED"},
	}

	if err := g.AddEdges(ctx, edges); err != nil {
		t.Fatalf("AddEdges() error = %v", err)
	}

	neighbors, err := g.OutgoingNeighbors(ctx, "a")
	if err != nil {
		t.Fatalf(
			"OutgoingNeighbors() error = %v",
			err,
		)
	}

	if len(neighbors) != 1 || neighbors[0].ID != "b" {
		t.Fatalf(
			"neighbors = %v, want [b]",
			neighbors,
		)
	}
}

func TestGraphAddEdgesFallsBackWithoutBatchStore(t *testing.T) {
	ctx := context.Background()

	store := newFakeStore()

	g, err := graph.New(store)
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	edges := []graph.Edge{
		{ID: "e1", From: "a", To: "b", Type: "PRODUCED"},
	}

	if err := g.AddEdges(ctx, edges); err != nil {
		t.Fatalf("AddEdges() error = %v", err)
	}

	if store.putEdgeCalls != 1 {
		t.Errorf(
			"putEdgeCalls = %d, want 1",
			store.putEdgeCalls,
		)
	}
}

func TestGraphAddEdgesRejectsEmptyType(t *testing.T) {
	ctx := context.Background()

	g, err := graph.New(memory.New())
	if err != nil {
		t.Fatalf("graph.New() error = %v", err)
	}

	err = g.AddEdges(
		ctx,
		[]graph.Edge{{ID: "e1", From: "a", To: "b"}},
	)

	if !errors.Is(err, graph.ErrEmptyEdgeType) {
		t.Fatalf(
			"AddEdges() error = %v, want ErrEmptyEdgeType",
			err,
		)
	}
}
