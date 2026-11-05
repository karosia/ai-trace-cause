package aitracecause_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	aitracecause "github.com/karosia/ai-trace-cause"
	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
)

// tickingClock returns a clock that advances one second per call, so
// every recorded entity and relationship gets a distinct, ordered
// timestamp.
func tickingClock(start time.Time) func() time.Time {
	var mu sync.Mutex

	current := start

	return func() time.Time {
		mu.Lock()
		defer mu.Unlock()

		current = current.Add(time.Second)

		return current
	}
}

// must unwraps a (value, error) pair from a setup call. Setup errors
// are not what these tests exercise, so they panic, which still fails
// the test with a stack trace.
func must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}

	return value
}

func mustOK(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatal(err)
	}
}

// Two observations support two facts with a shared observation, both
// facts feed one decision. 8 causal relationships in total.
func TestExplainIncludesEveryRelationshipInDiamond(t *testing.T) {
	ctx := context.Background()
	trace := must(aitracecause.New(
		aitracecause.WithClock(tickingClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))),
	))

	src := must(trace.RecordSource(ctx, aitracecause.Source{Kind: "Prometheus"}))
	cpu := must(trace.RecordObservation(ctx, aitracecause.Observation{Name: "cpu", Value: 94}))
	lat := must(trace.RecordObservation(ctx, aitracecause.Observation{Name: "latency_p99", Value: 1200}))
	high := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "CPU high", Confidence: 0.9}))
	degraded := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "Service degraded", Confidence: 0.9}))
	decision := must(trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "scale", Confidence: 0.9}))
	action := must(trace.RecordAction(ctx, aitracecause.Action{Name: "scale_service"}))

	mustOK(t, trace.Produced(ctx, src.ID, cpu.ID))
	mustOK(t, trace.Produced(ctx, src.ID, lat.ID))
	mustOK(t, trace.Supports(ctx, cpu.ID, high.ID))
	mustOK(t, trace.Supports(ctx, cpu.ID, degraded.ID))
	mustOK(t, trace.Supports(ctx, lat.ID, degraded.ID))
	mustOK(t, trace.BasisOf(ctx, high.ID, decision.ID))
	mustOK(t, trace.BasisOf(ctx, degraded.ID, decision.ID))
	mustOK(t, trace.Caused(ctx, decision.ID, action.ID))

	causal := must(trace.TraceActionCauseGraph(ctx, action.ID, 10))
	if len(causal.Edges) != 8 {
		t.Fatalf("len(Edges) = %d, want 8", len(causal.Edges))
	}

	explanation := must(trace.Explain(ctx, action.ID, 10))

	want := []string{
		`Action "scale_service" was caused by Decision "scale".`,
		`Decision "scale" was based on Fact "CPU high".`,
		`Decision "scale" was based on Fact "Service degraded".`,
		`Fact "CPU high" was supported by Observation "cpu" (value=94).`,
		`Fact "Service degraded" was supported by Observation "cpu" (value=94).`,
		`Fact "Service degraded" was supported by Observation "latency_p99" (value=1200).`,
		`Observation "cpu" (value=94) was produced by Source "Prometheus".`,
		`Observation "latency_p99" (value=1200) was produced by Source "Prometheus".`,
	}

	lines := strings.Split(explanation, "\n")
	if len(lines) != len(want) {
		t.Fatalf("got %d sentences, want %d:\n%s", len(lines), len(want), explanation)
	}

	// Exact order: grouped by effect in BFS order, and within an
	// effect, causes appear in the order they were recorded.
	for i := range want {
		if lines[i] != want[i] {
			t.Errorf("sentence %d = %q, want %q", i, lines[i], want[i])
		}
	}
}

func TestTraceAtActionTimeSeesCausesLinkedOnRecord(t *testing.T) {
	ctx := context.Background()
	trace := must(aitracecause.New(
		aitracecause.WithClock(tickingClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))),
	))

	fact := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "subscription expired", Confidence: 1}))
	decision := must(trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "suspend"}, aitracecause.WithCauses(fact.ID)))
	action := must(trace.RecordAction(ctx, aitracecause.Action{Name: "suspend_account"}, aitracecause.WithCauses(decision.ID)))

	if action.RecordedAt.IsZero() {
		t.Fatal("RecordAction() did not return RecordedAt")
	}

	causal := must(trace.TraceActionCauseAtActionTime(ctx, action.ID, 5))

	if len(causal.Visits) != 3 || len(causal.Edges) != 2 {
		t.Fatalf(
			"visits=%d edges=%d, want 3 and 2",
			len(causal.Visits),
			len(causal.Edges),
		)
	}
}

func TestTraceAtActionTimeExcludesLinksAddedAfterTheAction(t *testing.T) {
	ctx := context.Background()
	trace := must(aitracecause.New(
		aitracecause.WithClock(tickingClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))),
	))

	decision := must(trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "suspend"}))
	action := must(trace.RecordAction(ctx, aitracecause.Action{Name: "suspend_account"}))
	mustOK(t, trace.Caused(ctx, decision.ID, action.ID))

	causal := must(trace.TraceActionCauseAtActionTime(ctx, action.ID, 5))
	if len(causal.Visits) != 1 {
		t.Fatalf("len(Visits) = %d, want 1 (link was recorded after the action)", len(causal.Visits))
	}

	current := must(trace.TraceActionCauseGraph(ctx, action.ID, 5))
	if len(current.Visits) != 2 {
		t.Fatalf("current len(Visits) = %d, want 2", len(current.Visits))
	}
}

func TestRetractionIsVisibleNowButNotAtActionTime(t *testing.T) {
	ctx := context.Background()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	trace := must(aitracecause.New(aitracecause.WithClock(tickingClock(start))))

	obs := must(trace.RecordObservation(ctx, aitracecause.Observation{Name: "billing_status", Value: "unpaid"}))
	fact := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "subscription expired", Confidence: 1}, aitracecause.WithCauses(obs.ID)))
	decision := must(trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "suspend"}, aitracecause.WithCauses(fact.ID)))
	action := must(trace.RecordAction(ctx, aitracecause.Action{Name: "suspend_account"}, aitracecause.WithCauses(decision.ID)))

	// Retroactive correction, recorded after the action: the billing
	// feed was wrong since before the decision was made.
	retraction := must(trace.Retract(ctx, aitracecause.Retraction{
		NodeID:      fact.ID,
		Reason:      "billing feed was stale",
		EffectiveAt: start,
	}))

	if !retraction.RecordedAt.After(action.RecordedAt) {
		t.Fatalf("retraction recorded at %v, want after action %v", retraction.RecordedAt, action.RecordedAt)
	}

	atAction := must(trace.TraceActionCauseAtActionTime(ctx, action.ID, 5))
	if len(atAction.Retractions) != 0 {
		t.Errorf("at action time: retractions = %v, want none (not yet known)", atAction.Retractions)
	}

	if len(atAction.Edges) != 3 {
		t.Errorf("at action time: len(Edges) = %d, want 3 (RETRACTED is not causal)", len(atAction.Edges))
	}

	if strings.Contains(aitracecause.ExplainGraph(atAction), "retracted") {
		t.Errorf("at-action-time explanation mentions a retraction")
	}

	now := must(trace.TraceActionCauseGraph(ctx, action.ID, 5))
	if got := now.Retractions[fact.ID].Reason; got != "billing feed was stale" {
		t.Errorf("current retraction reason = %q", got)
	}

	explanation := must(trace.Explain(ctx, action.ID, 5))
	if !strings.Contains(explanation, `Fact "subscription expired" [retracted: billing feed was stale]`) {
		t.Errorf("current explanation does not mark the retraction:\n%s", explanation)
	}
}

func TestRetractValidationAndIdempotence(t *testing.T) {
	ctx := context.Background()
	trace := must(aitracecause.New())
	fact := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "x"}))

	if _, err := trace.Retract(ctx, aitracecause.Retraction{NodeID: fact.ID}); !errors.Is(err, semantic.ErrEmptyRetractionReason) {
		t.Errorf("empty reason: err = %v", err)
	}

	if _, err := trace.Retract(ctx, aitracecause.Retraction{Reason: "r"}); !errors.Is(err, semantic.ErrEmptyRetractionNodeID) {
		t.Errorf("empty node ID: err = %v", err)
	}

	if _, err := trace.Retract(ctx, aitracecause.Retraction{NodeID: "missing", Reason: "r"}); !errors.Is(err, graph.ErrNodeNotFound) {
		t.Errorf("missing node: err = %v", err)
	}

	first := must(trace.Retract(ctx, aitracecause.Retraction{NodeID: fact.ID, Reason: "wrong"}))

	again, err := trace.Retract(ctx, aitracecause.Retraction{NodeID: fact.ID, Reason: "wrong"})
	if err != nil {
		t.Errorf("identical retry: err = %v, want nil", err)
	}

	if !again.RecordedAt.Equal(first.RecordedAt) {
		t.Errorf("retry returned a new retraction: %v vs %v", again.RecordedAt, first.RecordedAt)
	}

	if _, err := trace.Retract(ctx, aitracecause.Retraction{NodeID: fact.ID, Reason: "different"}); !errors.Is(err, semantic.ErrAlreadyRetracted) {
		t.Errorf("conflicting retraction: err = %v, want ErrAlreadyRetracted", err)
	}
}

func TestRecordedValuesAreImmuneToCallerMutation(t *testing.T) {
	ctx := context.Background()
	trace := must(aitracecause.New())

	params := map[string]any{
		"cfg":   map[string]any{"replicas": 5},
		"zones": []string{"a", "b"},
	}
	must(trace.RecordAction(ctx, aitracecause.Action{Name: "scale", Parameters: params}))

	params["cfg"].(map[string]any)["replicas"] = 999
	params["zones"].([]string)[0] = "mutated"

	nodes := must(trace.FindNodes(ctx, aitracecause.NodeFilter{Type: "Action"}))
	stored := nodes[0].Properties["parameters"].(map[string]any)

	if got := stored["cfg"].(map[string]any)["replicas"]; got != 5 {
		t.Errorf("replicas = %v, want 5", got)
	}

	if got := stored["zones"].([]string)[0]; got != "a" {
		t.Errorf("zones[0] = %q, want a", got)
	}

	// Mutating a value returned by a read must not reach the store
	// either.
	stored["cfg"].(map[string]any)["replicas"] = 111

	nodes = must(trace.FindNodes(ctx, aitracecause.NodeFilter{Type: "Action"}))
	if got := nodes[0].Properties["parameters"].(map[string]any)["cfg"].(map[string]any)["replicas"]; got != 5 {
		t.Errorf("after mutating a read result: replicas = %v, want 5", got)
	}
}

func TestUnsupportedPropertyValueIsRejected(t *testing.T) {
	ctx := context.Background()
	trace := must(aitracecause.New())

	_, err := trace.RecordAction(ctx, aitracecause.Action{
		Name:       "x",
		Parameters: map[string]any{"callback": func() {}},
	})
	if !errors.Is(err, graph.ErrUnsupportedPropertyValue) {
		t.Fatalf("err = %v, want ErrUnsupportedPropertyValue", err)
	}

	if !strings.Contains(err.Error(), "parameters.callback") {
		t.Errorf("error %q does not name the offending path", err)
	}

	nodes := must(trace.FindNodes(ctx, aitracecause.NodeFilter{}))
	if len(nodes) != 0 {
		t.Errorf("rejected action was stored: %d nodes", len(nodes))
	}
}

func TestRelationshipsAreIdempotentUnderConcurrentRetries(t *testing.T) {
	ctx := context.Background()
	trace := must(aitracecause.New())

	decision := must(trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "x"}))
	action := must(trace.RecordAction(ctx, aitracecause.Action{Name: "y"}))

	var wg sync.WaitGroup

	errs := make(chan error, 32)

	for range 32 {
		wg.Go(func() {
			errs <- trace.Caused(ctx, decision.ID, action.ID)
		})
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("Caused() retry error = %v", err)
		}
	}

	causal := must(trace.TraceActionCauseGraph(ctx, action.ID, 2))
	if len(causal.Edges) != 1 {
		t.Fatalf("len(Edges) = %d, want 1", len(causal.Edges))
	}

	if want := semantic.RelationID(semantic.RelationCaused, decision.ID, action.ID); causal.Edges[0].ID != want {
		t.Errorf("edge ID = %q, want %q", causal.Edges[0].ID, want)
	}
}

func TestWithCausesValidatesBeforeWriting(t *testing.T) {
	ctx := context.Background()
	trace := must(aitracecause.New())

	fact := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "f"}))

	// A Fact is not a valid direct cause of an Action.
	_, err := trace.RecordAction(ctx, aitracecause.Action{Name: "a"}, aitracecause.WithCauses(fact.ID))

	var typeErr *semantic.UnexpectedNodeTypeError
	if !errors.As(err, &typeErr) || typeErr.Expected != semantic.NodeTypeDecision {
		t.Fatalf("wrong cause type: err = %v, want UnexpectedNodeTypeError(expected Decision)", err)
	}

	_, err = trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "d"}, aitracecause.WithCauses(fact.ID, "missing"))
	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf("missing cause: err = %v, want ErrNodeNotFound", err)
	}

	nodes := must(trace.FindNodes(ctx, aitracecause.NodeFilter{}))
	if len(nodes) != 1 {
		t.Fatalf("len(nodes) = %d, want 1: rejected records must not be stored", len(nodes))
	}

	// Duplicate cause IDs collapse to a single relationship.
	decision := must(trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "d"}, aitracecause.WithCauses(fact.ID, fact.ID)))
	action := must(trace.RecordAction(ctx, aitracecause.Action{Name: "a"}, aitracecause.WithCauses(decision.ID)))

	causal := must(trace.TraceActionCauseGraph(ctx, action.ID, 5))
	if len(causal.Edges) != 2 {
		t.Fatalf("len(Edges) = %d, want 2", len(causal.Edges))
	}

	for _, edge := range causal.Edges {
		if !edge.RecordedAt.Equal(nodeRecordedAt(causal, edge.To)) {
			t.Errorf("edge %s RecordedAt = %v, want the effect's RecordedAt", edge.Type, edge.RecordedAt)
		}
	}
}

func nodeRecordedAt(causal aitracecause.CausalGraph, id string) time.Time {
	for _, visit := range causal.Visits {
		if visit.Node.ID == id {
			return visit.Node.RecordedAt
		}
	}

	return time.Time{}
}

func TestTraceAtActionTimeRejectsNonAction(t *testing.T) {
	ctx := context.Background()
	trace := must(aitracecause.New())
	fact := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "f"}))

	for name, call := range map[string]func() error{
		"AtActionTime": func() error {
			_, err := trace.TraceActionCauseAtActionTime(ctx, fact.ID, 3)
			return err
		},
		"GraphAt": func() error {
			_, err := trace.TraceActionCauseGraphAt(ctx, fact.ID, time.Now(), 3)
			return err
		},
		"Graph": func() error {
			_, err := trace.TraceActionCauseGraph(ctx, fact.ID, 3)
			return err
		},
	} {
		if err := call(); !errors.Is(err, semantic.ErrUnexpectedNodeType) {
			t.Errorf("%s: err = %v, want ErrUnexpectedNodeType", name, err)
		}
	}

	if _, err := trace.TraceActionCauseAtActionTime(ctx, "missing", 3); !errors.Is(err, graph.ErrNodeNotFound) {
		t.Errorf("missing: err = %v, want ErrNodeNotFound", err)
	}
}
