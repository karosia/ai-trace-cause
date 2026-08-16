package aitracecause_test

import (
	"context"
	"testing"

	aitracecause "github.com/karosia/ai-trace-cause"
)

func TestNewUsesMemoryStoreByDefault(
	t *testing.T,
) {
	trace, err := aitracecause.New()
	if err != nil {
		t.Fatalf(
			"aitracecause.New() error = %v",
			err,
		)
	}

	if trace == nil {
		t.Fatal(
			"aitracecause.New() returned nil",
		)
	}
}

func TestTraceActionCauseThroughSDK(
	t *testing.T,
) {
	ctx := context.Background()

	trace, err := aitracecause.New()
	if err != nil {
		t.Fatal(err)
	}

	source := aitracecause.Source{
		ID:   "source-001",
		Kind: "Prometheus",
	}

	observation := aitracecause.Observation{
		ID:    "observation-001",
		Name:  "cpu_usage",
		Value: 94,
	}

	fact := aitracecause.Fact{
		ID:         "fact-001",
		Statement:  "CPU usage is high",
		Confidence: 0.98,
	}

	decision := aitracecause.Decision{
		ID:         "decision-001",
		Outcome:    "Scale service",
		Rationale:  "CPU usage is above threshold",
		Confidence: 0.92,
	}

	action := aitracecause.Action{
		ID:   "action-001",
		Name: "scale_service",
	}

	if err := trace.RecordSource(
		ctx,
		source,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.RecordObservation(
		ctx,
		observation,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.RecordFact(
		ctx,
		fact,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.RecordDecision(
		ctx,
		decision,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.RecordAction(
		ctx,
		action,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.Produced(
		ctx,
		"edge-1",
		source.ID,
		observation.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.Supports(
		ctx,
		"edge-2",
		observation.ID,
		fact.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.BasisOf(
		ctx,
		"edge-3",
		fact.ID,
		decision.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.Caused(
		ctx,
		"edge-4",
		decision.ID,
		action.ID,
	); err != nil {
		t.Fatal(err)
	}

	results, err := trace.TraceActionCause(
		ctx,
		action.ID,
		4,
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"action-001",
		"decision-001",
		"fact-001",
		"observation-001",
		"source-001",
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
				"results[%d].Node.ID = %q, want %q",
				i,
				results[i].Node.ID,
				want[i],
			)
		}
	}
}
