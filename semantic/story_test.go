package semantic_test

import (
	"context"
	"testing"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

func TestActionCauseStoryWithProvenance(
	t *testing.T,
) {
	ctx := context.Background()

	store := memory.New()

	g, err := graph.New(store)
	if err != nil {
		t.Fatal(err)
	}

	trace, err := semantic.NewService(g)
	if err != nil {
		t.Fatal(err)
	}

	source := semantic.Source{
		ID:   "source-001",
		Kind: "Prometheus",
		URI:  "prometheus://production/cpu_usage",
	}

	observation := semantic.Observation{
		ID:    "observation-001",
		Name:  "cpu_usage",
		Value: 94,
		Metadata: map[string]any{
			"unit": "%",
		},
	}

	fact := semantic.Fact{
		ID:         "fact-001",
		Statement:  "CPU usage is high",
		Confidence: 0.98,
	}

	decision := semantic.Decision{
		ID:         "decision-001",
		Outcome:    "Scale the service",
		Rationale:  "CPU utilization is consistently above 90%",
		Confidence: 0.92,
	}

	action := semantic.Action{
		ID:     "action-001",
		Name:   "scale_service",
		Target: "payments-api",
		Parameters: map[string]any{
			"replicas": 5,
		},
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
		"edge-produced",
		source.ID,
		observation.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.Supports(
		ctx,
		"edge-supports",
		observation.ID,
		fact.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.BasisOf(
		ctx,
		"edge-basis",
		fact.ID,
		decision.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := trace.Caused(
		ctx,
		"edge-caused",
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
