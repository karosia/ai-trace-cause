package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

func TestStorePutNodes(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	nodes := []graph.Node{
		{ID: "a", Type: "Source"},
		{ID: "b", Type: "Observation"},
	}

	if err := store.PutNodes(ctx, nodes); err != nil {
		t.Fatalf("PutNodes() error = %v", err)
	}

	for _, id := range []string{"a", "b"} {
		if _, err := store.GetNode(ctx, id); err != nil {
			t.Errorf("GetNode(%q) error = %v", id, err)
		}
	}
}

func TestStorePutNodesRejectsDuplicateWithinBatch(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	nodes := []graph.Node{
		{ID: "a", Type: "Source"},
		{ID: "a", Type: "Source"},
	}

	err := store.PutNodes(ctx, nodes)

	if !errors.Is(err, graph.ErrNodeAlreadyExists) {
		t.Fatalf(
			"PutNodes() error = %v, want ErrNodeAlreadyExists",
			err,
		)
	}

	if _, err := store.GetNode(ctx, "a"); !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf(
			"GetNode() error = %v, want ErrNodeNotFound (rejected batch must not partially apply)",
			err,
		)
	}
}

func TestStorePutNodesRejectsDuplicateAgainstStore(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	if err := store.PutNode(ctx, graph.Node{ID: "a", Type: "Source"}); err != nil {
		t.Fatalf("PutNode() error = %v", err)
	}

	err := store.PutNodes(
		ctx,
		[]graph.Node{{ID: "a", Type: "Source"}},
	)

	if !errors.Is(err, graph.ErrNodeAlreadyExists) {
		t.Fatalf(
			"PutNodes() error = %v, want ErrNodeAlreadyExists",
			err,
		)
	}
}

func TestStorePutEdges(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	if err := store.PutNodes(ctx, []graph.Node{
		{ID: "a", Type: "Source"},
		{ID: "b", Type: "Observation"},
	}); err != nil {
		t.Fatalf("PutNodes() error = %v", err)
	}

	edges := []graph.Edge{
		{ID: "e1", From: "a", To: "b", Type: "PRODUCED"},
	}

	if err := store.PutEdges(ctx, edges); err != nil {
		t.Fatalf("PutEdges() error = %v", err)
	}

	outgoing, err := store.OutgoingEdges(ctx, "a")
	if err != nil {
		t.Fatalf("OutgoingEdges() error = %v", err)
	}

	if len(outgoing) != 1 || outgoing[0].ID != "e1" {
		t.Fatalf(
			"OutgoingEdges() = %v, want [e1]",
			outgoing,
		)
	}

	incoming, err := store.IncomingEdges(ctx, "b")
	if err != nil {
		t.Fatalf("IncomingEdges() error = %v", err)
	}

	if len(incoming) != 1 || incoming[0].ID != "e1" {
		t.Fatalf(
			"IncomingEdges() = %v, want [e1]",
			incoming,
		)
	}
}

func TestStorePutNodesRejectsEmptyID(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	err := store.PutNodes(
		ctx,
		[]graph.Node{{Type: "Source"}},
	)

	if !errors.Is(err, graph.ErrEmptyNodeID) {
		t.Fatalf(
			"PutNodes() error = %v, want ErrEmptyNodeID",
			err,
		)
	}
}

func TestStorePutNodesRejectsEmptyType(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	err := store.PutNodes(
		ctx,
		[]graph.Node{{ID: "a"}},
	)

	if !errors.Is(err, graph.ErrEmptyNodeType) {
		t.Fatalf(
			"PutNodes() error = %v, want ErrEmptyNodeType",
			err,
		)
	}
}

func TestStorePutEdgesRejectsEmptyID(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	err := store.PutEdges(
		ctx,
		[]graph.Edge{{From: "a", To: "b", Type: "PRODUCED"}},
	)

	if !errors.Is(err, graph.ErrEmptyEdgeID) {
		t.Fatalf(
			"PutEdges() error = %v, want ErrEmptyEdgeID",
			err,
		)
	}
}

func TestStorePutEdgesRejectsEmptyType(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	err := store.PutEdges(
		ctx,
		[]graph.Edge{{ID: "e1", From: "a", To: "b"}},
	)

	if !errors.Is(err, graph.ErrEmptyEdgeType) {
		t.Fatalf(
			"PutEdges() error = %v, want ErrEmptyEdgeType",
			err,
		)
	}
}

func TestStorePutEdgesRejectsDuplicateWithinBatch(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	if err := store.PutNodes(ctx, []graph.Node{
		{ID: "a", Type: "Source"},
		{ID: "b", Type: "Observation"},
	}); err != nil {
		t.Fatalf("PutNodes() error = %v", err)
	}

	edges := []graph.Edge{
		{ID: "e1", From: "a", To: "b", Type: "PRODUCED"},
		{ID: "e1", From: "a", To: "b", Type: "PRODUCED"},
	}

	err := store.PutEdges(ctx, edges)

	if !errors.Is(err, graph.ErrEdgeAlreadyExists) {
		t.Fatalf(
			"PutEdges() error = %v, want ErrEdgeAlreadyExists",
			err,
		)
	}

	if _, err := store.GetEdge(ctx, "e1"); !errors.Is(err, graph.ErrEdgeNotFound) {
		t.Fatalf(
			"GetEdge() error = %v, want ErrEdgeNotFound (rejected batch must not partially apply)",
			err,
		)
	}
}

func TestStorePutEdgesRejectsMissingNode(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	err := store.PutEdges(
		ctx,
		[]graph.Edge{{ID: "e1", From: "a", To: "b", Type: "PRODUCED"}},
	)

	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf(
			"PutEdges() error = %v, want ErrNodeNotFound",
			err,
		)
	}
}
