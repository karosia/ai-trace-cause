package semantic_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/karosia/ai-trace-cause/semantic"
)

// buildLinearChain records the canonical Source -> Observation ->
// Fact -> Decision -> Action chain used across explain/effects tests.
func buildLinearChain(
	t *testing.T,
	ctx context.Context,
	trace *semantic.Service,
) (
	source semantic.Source,
	observation semantic.Observation,
	fact semantic.Fact,
	decision semantic.Decision,
	action semantic.Action,
) {
	t.Helper()

	var err error

	source, err = trace.RecordSource(
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

	observation, err = trace.RecordObservation(
		ctx,
		semantic.Observation{
			ID:    "observation-001",
			Name:  "cpu_usage",
			Value: 94,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	fact, err = trace.RecordFact(
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

	decision, err = trace.RecordDecision(
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

	action, err = trace.RecordAction(
		ctx,
		semantic.Action{
			ID:     "action-001",
			Name:   "scale_service",
			Target: "payments-api",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := trace.Produced(ctx, source.ID, observation.ID); err != nil {
		t.Fatal(err)
	}

	if err := trace.Supports(ctx, observation.ID, fact.ID); err != nil {
		t.Fatal(err)
	}

	if err := trace.BasisOf(ctx, fact.ID, decision.ID); err != nil {
		t.Fatal(err)
	}

	if err := trace.Caused(ctx, decision.ID, action.ID); err != nil {
		t.Fatal(err)
	}

	return source, observation, fact, decision, action
}

func TestTraceSourceEffects(t *testing.T) {
	ctx := context.Background()

	trace, _ := newTestService(
		t,
		"edge-produced",
		"edge-supports",
		"edge-basis",
		"edge-caused",
	)

	source, observation, fact, decision, action := buildLinearChain(
		t,
		ctx,
		trace,
	)

	results, err := trace.TraceSourceEffects(
		ctx,
		source.ID,
		4,
	)
	if err != nil {
		t.Fatalf(
			"TraceSourceEffects() error = %v",
			err,
		)
	}

	want := []string{
		source.ID,
		observation.ID,
		fact.ID,
		decision.ID,
		action.ID,
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

		if results[i].Depth != i {
			t.Errorf(
				"results[%d].Depth = %d, want %d",
				i,
				results[i].Depth,
				i,
			)
		}
	}
}

func TestTraceSourceEffectsRejectsNonSource(t *testing.T) {
	ctx := context.Background()

	trace, _ := newTestService(t)

	action, err := trace.RecordAction(
		ctx,
		semantic.Action{
			ID:   "action-001",
			Name: "scale_service",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = trace.TraceSourceEffects(
		ctx,
		action.ID,
		4,
	)

	var typeErr *semantic.UnexpectedNodeTypeError

	if !errors.As(err, &typeErr) {
		t.Fatalf(
			"TraceSourceEffects() error = %v, want *UnexpectedNodeTypeError",
			err,
		)
	}
}

func TestExplain(t *testing.T) {
	ctx := context.Background()

	trace, _ := newTestService(
		t,
		"edge-produced",
		"edge-supports",
		"edge-basis",
		"edge-caused",
	)

	buildLinearChain(t, ctx, trace)

	explanation, err := trace.Explain(
		ctx,
		"action-001",
		4,
	)
	if err != nil {
		t.Fatalf(
			"Explain() error = %v",
			err,
		)
	}

	want := strings.Join(
		[]string{
			`Action "scale_service" (target=payments-api) was caused by Decision "Scale the service".`,
			`Decision "Scale the service" was based on Fact "CPU usage is high".`,
			`Fact "CPU usage is high" was supported by Observation "cpu_usage" (value=94).`,
			`Observation "cpu_usage" (value=94) was produced by Source "Prometheus" (prometheus://production/cpu_usage).`,
		},
		"\n",
	)

	if explanation != want {
		t.Errorf(
			"Explain() = %q, want %q",
			explanation,
			want,
		)
	}
}

func TestExplainRejectsNonAction(t *testing.T) {
	ctx := context.Background()

	trace, _ := newTestService(t)

	source, err := trace.RecordSource(
		ctx,
		semantic.Source{
			ID:   "source-001",
			Kind: "Prometheus",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = trace.Explain(
		ctx,
		source.ID,
		4,
	)

	var typeErr *semantic.UnexpectedNodeTypeError

	if !errors.As(err, &typeErr) {
		t.Fatalf(
			"Explain() error = %v, want *UnexpectedNodeTypeError",
			err,
		)
	}
}
