package aitracecause_test

import (
	"context"
	"testing"

	"github.com/karosia/ai-trace-cause"
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

func TestNewGeneratesIDsByDefault(
	t *testing.T,
) {
	ctx := context.Background()

	trace, err := aitracecause.New()
	if err != nil {
		t.Fatal(err)
	}

	observation, err := trace.RecordObservation(
		ctx,
		aitracecause.Observation{
			Name:  "cpu_usage",
			Value: 94,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if observation.ID == "" {
		t.Fatal(
			"observation.ID is empty",
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

	source, err := trace.RecordSource(
		ctx,
		aitracecause.Source{
			ID:   "source-001",
			Kind: "Prometheus",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	observation, err := trace.RecordObservation(
		ctx,
		aitracecause.Observation{
			ID:    "observation-001",
			Name:  "cpu_usage",
			Value: 94,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	fact, err := trace.RecordFact(
		ctx,
		aitracecause.Fact{
			ID:         "fact-001",
			Statement:  "CPU usage is high",
			Confidence: 0.98,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	decision, err := trace.RecordDecision(
		ctx,
		aitracecause.Decision{
			ID:         "decision-001",
			Outcome:    "Scale service",
			Rationale:  "CPU usage is above threshold",
			Confidence: 0.92,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	action, err := trace.RecordAction(
		ctx,
		aitracecause.Action{
			ID:   "action-001",
			Name: "scale_service",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := trace.Produced(
		ctx,
		source.ID,
		observation.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.Supports(
		ctx,
		observation.ID,
		fact.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.BasisOf(
		ctx,
		fact.ID,
		decision.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.Caused(
		ctx,
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

func TestAutomaticIDCausalStory(
	t *testing.T,
) {
	ctx := context.Background()

	trace, err := aitracecause.New()
	if err != nil {
		t.Fatal(err)
	}

	source, err := trace.RecordSource(
		ctx,
		aitracecause.Source{
			Kind: "Prometheus",
			URI:  "prometheus://production/cpu_usage",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	observation, err := trace.RecordObservation(
		ctx,
		aitracecause.Observation{
			Name:  "cpu_usage",
			Value: 94,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	fact, err := trace.RecordFact(
		ctx,
		aitracecause.Fact{
			Statement:  "CPU usage is high",
			Confidence: 0.98,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	decision, err := trace.RecordDecision(
		ctx,
		aitracecause.Decision{
			Outcome:    "Scale service",
			Rationale:  "CPU usage is high",
			Confidence: 0.92,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	action, err := trace.RecordAction(
		ctx,
		aitracecause.Action{
			Name:   "scale_service",
			Target: "payments-api",
			Parameters: map[string]any{
				"replicas": 5,
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if source.ID == "" {
		t.Fatal("source.ID is empty")
	}

	if observation.ID == "" {
		t.Fatal("observation.ID is empty")
	}

	if fact.ID == "" {
		t.Fatal("fact.ID is empty")
	}

	if decision.ID == "" {
		t.Fatal("decision.ID is empty")
	}

	if action.ID == "" {
		t.Fatal("action.ID is empty")
	}

	if err := trace.Produced(
		ctx,
		source.ID,
		observation.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.Supports(
		ctx,
		observation.ID,
		fact.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.BasisOf(
		ctx,
		fact.ID,
		decision.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.Caused(
		ctx,
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

	if len(results) != 5 {
		t.Fatalf(
			"len(results) = %d, want 5",
			len(results),
		)
	}

	if results[0].Node.ID != action.ID {
		t.Errorf(
			"results[0].Node.ID = %q, want %q",
			results[0].Node.ID,
			action.ID,
		)
	}

	if results[1].Node.ID != decision.ID {
		t.Errorf(
			"results[1].Node.ID = %q, want %q",
			results[1].Node.ID,
			decision.ID,
		)
	}

	if results[2].Node.ID != fact.ID {
		t.Errorf(
			"results[2].Node.ID = %q, want %q",
			results[2].Node.ID,
			fact.ID,
		)
	}

	if results[3].Node.ID != observation.ID {
		t.Errorf(
			"results[3].Node.ID = %q, want %q",
			results[3].Node.ID,
			observation.ID,
		)
	}

	if results[4].Node.ID != source.ID {
		t.Errorf(
			"results[4].Node.ID = %q, want %q",
			results[4].Node.ID,
			source.ID,
		)
	}
}
