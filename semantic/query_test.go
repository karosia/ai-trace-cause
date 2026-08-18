package semantic_test

import (
	"context"
	"testing"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
)

func TestServiceFindNodes(t *testing.T) {
	ctx := context.Background()

	trace, _ := newTestService(t)

	if _, err := trace.RecordDecision(
		ctx,
		semantic.Decision{
			ID:      "decision-001",
			Outcome: "Scale service",
		},
	); err != nil {
		t.Fatal(err)
	}

	if _, err := trace.RecordDecision(
		ctx,
		semantic.Decision{
			ID:      "decision-002",
			Outcome: "Roll back deployment",
		},
	); err != nil {
		t.Fatal(err)
	}

	if _, err := trace.RecordAction(
		ctx,
		semantic.Action{
			ID:   "action-001",
			Name: "scale_service",
		},
	); err != nil {
		t.Fatal(err)
	}

	results, err := trace.FindNodes(
		ctx,
		graph.NodeFilter{Type: string(semantic.NodeTypeDecision)},
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
		if node.Type != string(semantic.NodeTypeDecision) {
			t.Errorf(
				"node.Type = %q, want %q",
				node.Type,
				semantic.NodeTypeDecision,
			)
		}
	}
}
