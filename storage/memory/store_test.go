package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

func TestStorePutAndGetNode(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	node := graph.Node{
		ID:   "fact-001",
		Type: "Fact",
		Properties: map[string]any{
			"statement":  "CPU usage is high",
			"confidence": 0.98,
		},
	}

	if err := store.PutNode(ctx, node); err != nil {
		t.Fatalf("PutNode() error = %v", err)
	}

	got, err := store.GetNode(ctx, node.ID)
	if err != nil {
		t.Fatalf("GetNode() error = %v", err)
	}

	if got.ID != node.ID {
		t.Errorf(
			"GetNode() ID = %q, want %q",
			got.ID,
			node.ID,
		)
	}

	if got.Type != node.Type {
		t.Errorf(
			"GetNode() Type = %q, want %q",
			got.Type,
			node.Type,
		)
	}

	if got.Properties["statement"] != "CPU usage is high" {
		t.Errorf(
			"GetNode() statement = %v, want %q",
			got.Properties["statement"],
			"CPU usage is high",
		)
	}
}

func TestStorePutNodeRejectsDuplicate(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	node := graph.Node{
		ID:   "fact-001",
		Type: "Fact",
	}

	if err := store.PutNode(ctx, node); err != nil {
		t.Fatalf("first PutNode() error = %v", err)
	}

	err := store.PutNode(ctx, node)

	if !errors.Is(err, graph.ErrNodeAlreadyExists) {
		t.Fatalf(
			"second PutNode() error = %v, want ErrNodeAlreadyExists",
			err,
		)
	}
}

func TestStorePutNodeRejectsEmptyID(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	err := store.PutNode(ctx, graph.Node{
		Type: "Fact",
	})

	if !errors.Is(err, graph.ErrEmptyNodeID) {
		t.Fatalf(
			"PutNode() error = %v, want ErrEmptyNodeID",
			err,
		)
	}
}

func TestStorePutNodeRejectsEmptyType(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	err := store.PutNode(ctx, graph.Node{
		ID: "fact-001",
	})

	if !errors.Is(err, graph.ErrEmptyNodeType) {
		t.Fatalf(
			"PutNode() error = %v, want ErrEmptyNodeType",
			err,
		)
	}
}

func TestStoreGetNodeReturnsNotFound(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	_, err := store.GetNode(ctx, "missing-node")

	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf(
			"GetNode() error = %v, want ErrNodeNotFound",
			err,
		)
	}
}

func TestStorePutAndGetEdge(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	from := graph.Node{
		ID:   "observation-001",
		Type: "Observation",
	}

	to := graph.Node{
		ID:   "fact-001",
		Type: "Fact",
	}

	if err := store.PutNode(ctx, from); err != nil {
		t.Fatalf("PutNode(from) error = %v", err)
	}

	if err := store.PutNode(ctx, to); err != nil {
		t.Fatalf("PutNode(to) error = %v", err)
	}

	edge := graph.Edge{
		ID:   "edge-001",
		From: from.ID,
		To:   to.ID,
		Type: "SUPPORTS",
		Properties: map[string]any{
			"confidence": 0.95,
		},
	}

	if err := store.PutEdge(ctx, edge); err != nil {
		t.Fatalf("PutEdge() error = %v", err)
	}

	got, err := store.GetEdge(ctx, edge.ID)
	if err != nil {
		t.Fatalf("GetEdge() error = %v", err)
	}

	if got.ID != edge.ID {
		t.Errorf(
			"GetEdge() ID = %q, want %q",
			got.ID,
			edge.ID,
		)
	}

	if got.From != edge.From {
		t.Errorf(
			"GetEdge() From = %q, want %q",
			got.From,
			edge.From,
		)
	}

	if got.To != edge.To {
		t.Errorf(
			"GetEdge() To = %q, want %q",
			got.To,
			edge.To,
		)
	}

	if got.Type != edge.Type {
		t.Errorf(
			"GetEdge() Type = %q, want %q",
			got.Type,
			edge.Type,
		)
	}
}

func TestStorePutEdgeRejectsDuplicate(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	if err := store.PutNode(ctx, graph.Node{
		ID:   "observation-001",
		Type: "Observation",
	}); err != nil {
		t.Fatal(err)
	}

	if err := store.PutNode(ctx, graph.Node{
		ID:   "fact-001",
		Type: "Fact",
	}); err != nil {
		t.Fatal(err)
	}

	edge := graph.Edge{
		ID:   "edge-001",
		From: "observation-001",
		To:   "fact-001",
		Type: "SUPPORTS",
	}

	if err := store.PutEdge(ctx, edge); err != nil {
		t.Fatalf("first PutEdge() error = %v", err)
	}

	err := store.PutEdge(ctx, edge)

	if !errors.Is(err, graph.ErrEdgeAlreadyExists) {
		t.Fatalf(
			"second PutEdge() error = %v, want ErrEdgeAlreadyExists",
			err,
		)
	}
}

func TestStorePutEdgeRejectsUnknownFromNode(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	if err := store.PutNode(ctx, graph.Node{
		ID:   "fact-001",
		Type: "Fact",
	}); err != nil {
		t.Fatal(err)
	}

	err := store.PutEdge(ctx, graph.Edge{
		ID:   "edge-001",
		From: "missing-observation",
		To:   "fact-001",
		Type: "SUPPORTS",
	})

	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf(
			"PutEdge() error = %v, want ErrNodeNotFound",
			err,
		)
	}
}

func TestStorePutEdgeRejectsUnknownToNode(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	if err := store.PutNode(ctx, graph.Node{
		ID:   "observation-001",
		Type: "Observation",
	}); err != nil {
		t.Fatal(err)
	}

	err := store.PutEdge(ctx, graph.Edge{
		ID:   "edge-001",
		From: "observation-001",
		To:   "missing-fact",
		Type: "SUPPORTS",
	})

	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf(
			"PutEdge() error = %v, want ErrNodeNotFound",
			err,
		)
	}
}

func TestStoreGetEdgeReturnsNotFound(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	_, err := store.GetEdge(ctx, "missing-edge")

	if !errors.Is(err, graph.ErrEdgeNotFound) {
		t.Fatalf(
			"GetEdge() error = %v, want ErrEdgeNotFound",
			err,
		)
	}
}

func TestStoreOutgoingEdges(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	nodes := []graph.Node{
		{
			ID:   "observation-001",
			Type: "Observation",
		},
		{
			ID:   "fact-001",
			Type: "Fact",
		},
		{
			ID:   "fact-002",
			Type: "Fact",
		},
	}

	for _, node := range nodes {
		if err := store.PutNode(ctx, node); err != nil {
			t.Fatalf(
				"PutNode(%q) error = %v",
				node.ID,
				err,
			)
		}
	}

	edges := []graph.Edge{
		{
			ID:   "edge-001",
			From: "observation-001",
			To:   "fact-001",
			Type: "SUPPORTS",
		},
		{
			ID:   "edge-002",
			From: "observation-001",
			To:   "fact-002",
			Type: "SUPPORTS",
		},
	}

	for _, edge := range edges {
		if err := store.PutEdge(ctx, edge); err != nil {
			t.Fatalf(
				"PutEdge(%q) error = %v",
				edge.ID,
				err,
			)
		}
	}

	got, err := store.OutgoingEdges(
		ctx,
		"observation-001",
	)
	if err != nil {
		t.Fatalf(
			"OutgoingEdges() error = %v",
			err,
		)
	}

	if len(got) != 2 {
		t.Fatalf(
			"len(OutgoingEdges()) = %d, want 2",
			len(got),
		)
	}

	if got[0].ID != "edge-001" {
		t.Errorf(
			"edges[0].ID = %q, want edge-001",
			got[0].ID,
		)
	}

	if got[1].ID != "edge-002" {
		t.Errorf(
			"edges[1].ID = %q, want edge-002",
			got[1].ID,
		)
	}
}

func TestStoreIncomingEdges(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	nodes := []graph.Node{
		{
			ID:   "observation-001",
			Type: "Observation",
		},
		{
			ID:   "observation-002",
			Type: "Observation",
		},
		{
			ID:   "fact-001",
			Type: "Fact",
		},
	}

	for _, node := range nodes {
		if err := store.PutNode(ctx, node); err != nil {
			t.Fatalf(
				"PutNode(%q) error = %v",
				node.ID,
				err,
			)
		}
	}

	edges := []graph.Edge{
		{
			ID:   "edge-001",
			From: "observation-001",
			To:   "fact-001",
			Type: "SUPPORTS",
		},
		{
			ID:   "edge-002",
			From: "observation-002",
			To:   "fact-001",
			Type: "SUPPORTS",
		},
	}

	for _, edge := range edges {
		if err := store.PutEdge(ctx, edge); err != nil {
			t.Fatalf(
				"PutEdge(%q) error = %v",
				edge.ID,
				err,
			)
		}
	}

	got, err := store.IncomingEdges(
		ctx,
		"fact-001",
	)
	if err != nil {
		t.Fatalf(
			"IncomingEdges() error = %v",
			err,
		)
	}

	if len(got) != 2 {
		t.Fatalf(
			"len(IncomingEdges()) = %d, want 2",
			len(got),
		)
	}

	if got[0].ID != "edge-001" {
		t.Errorf(
			"edges[0].ID = %q, want edge-001",
			got[0].ID,
		)
	}

	if got[1].ID != "edge-002" {
		t.Errorf(
			"edges[1].ID = %q, want edge-002",
			got[1].ID,
		)
	}
}

func TestStoreOutgoingEdgesReturnsNodeNotFound(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	_, err := store.OutgoingEdges(
		ctx,
		"missing-node",
	)

	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf(
			"OutgoingEdges() error = %v, want ErrNodeNotFound",
			err,
		)
	}
}

func TestStoreIncomingEdgesReturnsNodeNotFound(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	_, err := store.IncomingEdges(
		ctx,
		"missing-node",
	)

	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf(
			"IncomingEdges() error = %v, want ErrNodeNotFound",
			err,
		)
	}
}

func TestStoreRespectsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	store := memory.New()

	err := store.PutNode(ctx, graph.Node{
		ID:   "fact-001",
		Type: "Fact",
	})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf(
			"PutNode() error = %v, want context.Canceled",
			err,
		)
	}
}

func TestStoreCopiesNodePropertiesOnWrite(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	properties := map[string]any{
		"statement": "CPU usage is high",
	}

	node := graph.Node{
		ID:         "fact-001",
		Type:       "Fact",
		Properties: properties,
	}

	if err := store.PutNode(ctx, node); err != nil {
		t.Fatal(err)
	}

	// Modify the original map after storing it.
	properties["statement"] = "modified externally"

	got, err := store.GetNode(ctx, node.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got.Properties["statement"] != "CPU usage is high" {
		t.Errorf(
			"stored property = %v, want %q",
			got.Properties["statement"],
			"CPU usage is high",
		)
	}
}

func TestStoreCopiesNodePropertiesOnRead(t *testing.T) {
	ctx := context.Background()

	store := memory.New()

	node := graph.Node{
		ID:   "fact-001",
		Type: "Fact",
		Properties: map[string]any{
			"statement": "CPU usage is high",
		},
	}

	if err := store.PutNode(ctx, node); err != nil {
		t.Fatal(err)
	}

	first, err := store.GetNode(ctx, node.ID)
	if err != nil {
		t.Fatal(err)
	}

	first.Properties["statement"] = "modified externally"

	second, err := store.GetNode(ctx, node.ID)
	if err != nil {
		t.Fatal(err)
	}

	if second.Properties["statement"] != "CPU usage is high" {
		t.Errorf(
			"stored property = %v, want %q",
			second.Properties["statement"],
			"CPU usage is high",
		)
	}
}
