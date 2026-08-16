package graph_test

import (
	"context"
	"testing"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

func TestBFSAtUsesTemporalValidity(
	t *testing.T,
) {
	ctx := context.Background()

	store := memory.New()

	g, err := graph.New(store)
	if err != nil {
		t.Fatal(err)
	}

	base := time.Date(
		2026,
		time.August,
		16,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	switchTime := base.Add(10 * time.Minute)

	oldFact := graph.Node{
		ID:         "fact-old",
		Type:       "Fact",
		RecordedAt: base.Add(-time.Minute),
		ValidUntil: &switchTime,
	}

	newFact := graph.Node{
		ID:         "fact-new",
		Type:       "Fact",
		RecordedAt: switchTime,
		ValidFrom:  &switchTime,
	}

	decision := graph.Node{
		ID:         "decision",
		Type:       "Decision",
		RecordedAt: base.Add(-time.Minute),
	}

	for _, node := range []graph.Node{
		oldFact,
		newFact,
		decision,
	} {
		if err := g.AddNode(ctx, node); err != nil {
			t.Fatal(err)
		}
	}

	oldEdge := graph.Edge{
		ID:         "edge-old",
		From:       "fact-old",
		To:         "decision",
		Type:       "BASIS_OF",
		RecordedAt: base.Add(-time.Minute),
		ValidUntil: &switchTime,
	}

	newEdge := graph.Edge{
		ID:         "edge-new",
		From:       "fact-new",
		To:         "decision",
		Type:       "BASIS_OF",
		RecordedAt: switchTime,
		ValidFrom:  &switchTime,
	}

	if err := g.AddEdge(ctx, oldEdge); err != nil {
		t.Fatal(err)
	}

	if err := g.AddEdge(ctx, newEdge); err != nil {
		t.Fatal(err)
	}

	beforeSwitch := base.Add(5 * time.Minute)

	results, err := g.BFSAt(
		ctx,
		"decision",
		graph.DirectionIncoming,
		1,
		beforeSwitch,
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 2 {
		t.Fatalf(
			"len(results) = %d, want 2",
			len(results),
		)
	}

	if results[1].Node.ID != "fact-old" {
		t.Errorf(
			"fact = %q, want fact-old",
			results[1].Node.ID,
		)
	}
}

func TestBFSAtExcludesKnowledgeRecordedLater(
	t *testing.T,
) {
	ctx := context.Background()

	store := memory.New()

	g, err := graph.New(store)
	if err != nil {
		t.Fatal(err)
	}

	base := time.Date(
		2026,
		time.August,
		16,
		10,
		0,
		0,
		0,
		time.UTC,
	)

	fact := graph.Node{
		ID:   "fact",
		Type: "Fact",

		RecordedAt: base.Add(
			10 * time.Minute,
		),

		ValidFrom: &base,
	}

	decision := graph.Node{
		ID:         "decision",
		Type:       "Decision",
		RecordedAt: base,
	}

	if err := g.AddNode(ctx, fact); err != nil {
		t.Fatal(err)
	}

	if err := g.AddNode(ctx, decision); err != nil {
		t.Fatal(err)
	}

	edge := graph.Edge{
		ID:         "edge",
		From:       fact.ID,
		To:         decision.ID,
		Type:       "BASIS_OF",
		RecordedAt: base.Add(10 * time.Minute),
	}

	if err := g.AddEdge(ctx, edge); err != nil {
		t.Fatal(err)
	}

	results, err := g.BFSAt(
		ctx,
		decision.ID,
		graph.DirectionIncoming,
		1,
		base.Add(5*time.Minute),
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf(
			"len(results) = %d, want 1",
			len(results),
		)
	}
}
