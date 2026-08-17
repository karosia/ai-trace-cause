package semantic_test

import (
	"context"
	"testing"

	"github.com/karosia/ai-trace-cause/semantic"
)

func TestActionCauseStoryWithProvenance(
	t *testing.T,
) {
	ctx := context.Background()

	trace, _ := newTestService(
		t,
		"edge-produced",
		"edge-supports",
		"edge-basis",
		"edge-caused",
	)

	source, err := trace.RecordSource(
		ctx,
		semantic.Source{
			ID:   "source-001",
			Kind: "Prometheus",
			URI:  "prometheus://production/cpu_usage",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	observation, err := trace.RecordObservation(
		ctx,
		semantic.Observation{
			ID:    "observation-001",
			Name:  "cpu_usage",
			Value: 94,
			Metadata: map[string]any{
				"unit": "%",
			},
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	fact, err := trace.RecordFact(
		ctx,
		semantic.Fact{
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
		semantic.Decision{
			ID:         "decision-001",
			Outcome:    "Scale the service",
			Rationale:  "CPU utilization is consistently above 90%",
			Confidence: 0.92,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	action, err := trace.RecordAction(
		ctx,
		semantic.Action{
			ID:     "action-001",
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
		t.Fatalf(
			"TraceActionCause() error = %v",
			err,
		)
	}

	want := []struct {
		id    string
		typ   string
		depth int
	}{
		{
			id:    "action-001",
			typ:   "Action",
			depth: 0,
		},
		{
			id:    "decision-001",
			typ:   "Decision",
			depth: 1,
		},
		{
			id:    "fact-001",
			typ:   "Fact",
			depth: 2,
		},
		{
			id:    "observation-001",
			typ:   "Observation",
			depth: 3,
		},
		{
			id:    "source-001",
			typ:   "Source",
			depth: 4,
		},
	}

	if len(results) != len(want) {
		t.Fatalf(
			"len(results) = %d, want %d",
			len(results),
			len(want),
		)
	}

	for i := range want {
		if results[i].Node.ID != want[i].id {
			t.Errorf(
				"results[%d].Node.ID = %q, want %q",
				i,
				results[i].Node.ID,
				want[i].id,
			)
		}

		if results[i].Node.Type != want[i].typ {
			t.Errorf(
				"results[%d].Node.Type = %q, want %q",
				i,
				results[i].Node.Type,
				want[i].typ,
			)
		}

		if results[i].Depth != want[i].depth {
			t.Errorf(
				"results[%d].Depth = %d, want %d",
				i,
				results[i].Depth,
				want[i].depth,
			)
		}
	}
}
