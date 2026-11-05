package aitracecause_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	aitracecause "github.com/karosia/ai-trace-cause"
	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
)

func newTickingTrace(t *testing.T, options ...aitracecause.Option) *aitracecause.Trace {
	t.Helper()

	options = append(
		[]aitracecause.Option{
			aitracecause.WithClock(tickingClock(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))),
		},
		options...,
	)

	return must(aitracecause.New(options...))
}

// A two-step agent loop: the agent decides to search, the search
// result becomes evidence for the next decision. The trace from the
// final action must reach back through the earlier tool call.
func TestAgentLoopIsTraceableAcrossToolCalls(t *testing.T) {
	ctx := context.Background()
	trace := newTickingTrace(t)
	run := trace.Run("incident-4521")

	// Step 1: plan and call a tool.
	step1 := run.Step("llm-call-1")
	alert := step1.Observation(aitracecause.Observation{Name: "pager_alert", Value: "p99 latency high"})
	symptom := step1.Fact(aitracecause.Fact{Statement: "checkout is slow", Confidence: 0.9}, alert)
	plan := step1.Decision(aitracecause.Decision{Outcome: "investigate checkout latency"}, symptom)
	search := step1.Decision(aitracecause.Decision{Outcome: "query metrics"}, plan)
	query := step1.Action(aitracecause.Action{Name: "prometheus_query", Target: "checkout"}, search)
	must(step1.Commit(ctx))

	// Step 2: the tool result drives the remediation.
	step2 := run.Step("llm-call-2")
	result := step2.Observation(aitracecause.Observation{Name: "cpu_usage", Value: 97}, query)
	cpuHigh := step2.Fact(aitracecause.Fact{Statement: "CPU saturated", Confidence: 0.95}, result)
	rootCause := step2.Fact(aitracecause.Fact{Statement: "checkout is CPU-bound", Confidence: 0.8}, cpuHigh, symptom)
	scale := step2.Decision(aitracecause.Decision{Outcome: "scale checkout"}, rootCause, plan)
	act := step2.Action(aitracecause.Action{Name: "scale_service", Target: "checkout"}, scale)
	res2 := must(step2.Commit(ctx))

	if len(res2.NodeIDs) != 5 {
		t.Fatalf("step 2 recorded %d nodes, want 5", len(res2.NodeIDs))
	}

	causal := must(trace.TraceActionCauseGraph(ctx, act, 20))
	reached := map[string]bool{}

	for _, visit := range causal.Visits {
		reached[visit.Node.ID] = true
	}

	for name, id := range map[string]string{
		"previous tool call": query,
		"plan":               plan,
		"original alert":     alert,
	} {
		if !reached[id] {
			t.Errorf("trace from final action did not reach the %s", name)
		}
	}

	explanation := must(trace.Explain(ctx, act, 20))

	for _, sentence := range []string{
		`Observation "cpu_usage" (value=97) resulted from Action "prometheus_query" (target=checkout).`,
		`Fact "checkout is CPU-bound" was inferred from Fact "CPU saturated".`,
		`Fact "checkout is CPU-bound" was inferred from Fact "checkout is slow".`,
		`Decision "scale checkout" refines Decision "investigate checkout latency".`,
		`Decision "query metrics" refines Decision "investigate checkout latency".`,
	} {
		if !strings.Contains(explanation, sentence) {
			t.Errorf("explanation missing %q:\n%s", sentence, explanation)
		}
	}

	// Every node carries its run and step.
	nodes := must(run.FindNodes(ctx, aitracecause.NodeFilter{}))
	if len(nodes) != 10 {
		t.Fatalf("run has %d nodes, want 10", len(nodes))
	}

	steps := map[any]int{}
	for _, node := range nodes {
		steps[node.Properties[semantic.PropertyStepID]]++
	}

	if steps["llm-call-1"] != 5 || steps["llm-call-2"] != 5 {
		t.Errorf("nodes per step = %v, want 5 and 5", steps)
	}

	decisions := must(run.FindNodes(ctx, aitracecause.NodeFilter{Type: "Decision"}))
	if len(decisions) != 3 {
		t.Errorf("run has %d decisions, want 3", len(decisions))
	}
}

func TestRunFindNodesIsolatesRuns(t *testing.T) {
	ctx := context.Background()
	trace := newTickingTrace(t)

	for _, runID := range []string{"a", "b"} {
		step := trace.Run(runID).Step("")
		step.Source(aitracecause.Source{Kind: "api"})
		must(step.Commit(ctx))
	}

	// Entities recorded directly can also be tagged with a run.
	must(trace.RecordSource(ctx, aitracecause.Source{Kind: "api", RunID: "a"}))

	if got := len(must(trace.Run("a").FindNodes(ctx, aitracecause.NodeFilter{}))); got != 2 {
		t.Errorf("run a has %d nodes, want 2", got)
	}

	if got := len(must(trace.Run("b").FindNodes(ctx, aitracecause.NodeFilter{}))); got != 1 {
		t.Errorf("run b has %d nodes, want 1", got)
	}
}

func TestStepCommitsAllOrNothing(t *testing.T) {
	ctx := context.Background()
	trace := newTickingTrace(t)
	run := trace.Run("r")

	countNodes := func() int {
		return len(must(trace.FindNodes(ctx, aitracecause.NodeFilter{})))
	}

	// Schema violation discovered at commit: a Fact cannot cause an
	// Action.
	step := run.Step("")
	fact := step.Fact(aitracecause.Fact{Statement: "f"})
	step.Action(aitracecause.Action{Name: "a"}, fact)

	if _, err := step.Commit(ctx); !errors.Is(err, semantic.ErrRelationNotAllowed) {
		t.Fatalf("Commit() error = %v, want ErrRelationNotAllowed", err)
	}

	if n := countNodes(); n != 0 {
		t.Fatalf("%d nodes written by a rejected step", n)
	}

	// Validation error while building: kept, returned by Commit, and
	// later calls are ignored.
	step = run.Step("")
	step.Source(aitracecause.Source{Kind: "api"})
	if id := step.Fact(aitracecause.Fact{Statement: ""}); id != "" {
		t.Errorf("invalid Fact returned ID %q, want empty", id)
	}

	step.Source(aitracecause.Source{Kind: "ignored"})

	if !errors.Is(step.Err(), semantic.ErrEmptyFactStatement) {
		t.Errorf("Err() = %v", step.Err())
	}

	if _, err := step.Commit(ctx); !errors.Is(err, semantic.ErrEmptyFactStatement) {
		t.Fatalf("Commit() error = %v, want ErrEmptyFactStatement", err)
	}

	// Cause that does not exist.
	step = run.Step("")
	step.Observation(aitracecause.Observation{Name: "o"}, "missing-source")

	if _, err := step.Commit(ctx); !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf("Commit() error = %v, want ErrNodeNotFound", err)
	}

	if n := countNodes(); n != 0 {
		t.Fatalf("%d nodes written by rejected steps", n)
	}
}

func TestStepLifecycleErrors(t *testing.T) {
	ctx := context.Background()
	trace := newTickingTrace(t)

	if _, err := trace.Run("").Step("").Commit(ctx); !errors.Is(err, semantic.ErrEmptyRunID) {
		t.Errorf("empty run ID: err = %v", err)
	}

	step := trace.Run("r").Step("s")
	step.Source(aitracecause.Source{Kind: "api"})
	must(step.Commit(ctx))

	if _, err := step.Commit(ctx); !errors.Is(err, semantic.ErrStepCommitted) {
		t.Errorf("second Commit: err = %v", err)
	}

	if id := step.Source(aitracecause.Source{Kind: "api"}); id != "" {
		t.Errorf("add after commit returned %q", id)
	}

	dup := trace.Run("r").Step("")
	dup.Source(aitracecause.Source{ID: "same", Kind: "api"})
	dup.Source(aitracecause.Source{ID: "same", Kind: "api"})

	if _, err := dup.Commit(ctx); !errors.Is(err, graph.ErrNodeAlreadyExists) {
		t.Errorf("duplicate ID in step: err = %v", err)
	}
}

func TestStepSharesRecordedAtAndIsVisibleAtActionTime(t *testing.T) {
	ctx := context.Background()
	trace := newTickingTrace(t)

	step := trace.Run("r").Step("")
	obs := step.Observation(aitracecause.Observation{Name: "o"})
	fact := step.Fact(aitracecause.Fact{Statement: "f"}, obs)
	decision := step.Decision(aitracecause.Decision{Outcome: "d"}, fact)
	action := step.Action(aitracecause.Action{Name: "a"}, decision)
	result := must(step.Commit(ctx))

	causal := must(trace.TraceActionCauseAtActionTime(ctx, action, 10))
	if len(causal.Visits) != 4 || len(causal.Edges) != 3 {
		t.Fatalf("visits=%d edges=%d, want 4 and 3", len(causal.Visits), len(causal.Edges))
	}

	for _, visit := range causal.Visits {
		if !visit.Node.RecordedAt.Equal(result.RecordedAt) {
			t.Errorf("node %s RecordedAt = %v, want %v", visit.Node.ID, visit.Node.RecordedAt, result.RecordedAt)
		}
	}
}

func TestStepRelinkingExistingRelationshipIsNoop(t *testing.T) {
	ctx := context.Background()
	trace := newTickingTrace(t)

	fact := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "f"}))
	decision := must(trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "d"}, aitracecause.WithCauses(fact.ID)))

	step := trace.Run("r").Step("")
	step.Link(fact.ID, aitracecause.RelationBasisOf, decision.ID)
	result, err := step.Commit(ctx)
	if err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	if len(result.EdgeIDs) != 0 {
		t.Errorf("EdgeIDs = %v, want none (relationship already existed)", result.EdgeIDs)
	}
}

func TestStepIsSafeForConcurrentUse(t *testing.T) {
	ctx := context.Background()
	trace := newTickingTrace(t)
	step := trace.Run("r").Step("")
	source := step.Source(aitracecause.Source{Kind: "api"})

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			step.Observation(aitracecause.Observation{Name: fmt.Sprint("o", i)}, source)
		})
	}

	wg.Wait()

	result := must(step.Commit(ctx))
	if len(result.NodeIDs) != 51 || len(result.EdgeIDs) != 50 {
		t.Fatalf("nodes=%d edges=%d, want 51 and 50", len(result.NodeIDs), len(result.EdgeIDs))
	}
}

func TestTypedLoopRelationsValidateTypes(t *testing.T) {
	ctx := context.Background()
	trace := newTickingTrace(t)

	fact := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "f"}))
	decision := must(trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "d"}))
	action := must(trace.RecordAction(ctx, aitracecause.Action{Name: "a"}))
	obs := must(trace.RecordObservation(ctx, aitracecause.Observation{Name: "o"}))

	for name, err := range map[string]error{
		"Refines(fact, decision)":  trace.Refines(ctx, fact.ID, decision.ID),
		"Infers(fact, decision)":   trace.Infers(ctx, fact.ID, decision.ID),
		"ResultedIn(fact, obs)":    trace.ResultedIn(ctx, fact.ID, obs.ID),
		"ResultedIn(action, fact)": trace.ResultedIn(ctx, action.ID, fact.ID),
	} {
		if !errors.Is(err, semantic.ErrUnexpectedNodeType) {
			t.Errorf("%s: err = %v, want ErrUnexpectedNodeType", name, err)
		}
	}

	mustOK(t, trace.ResultedIn(ctx, action.ID, obs.ID))
	mustOK(t, trace.ResultedIn(ctx, action.ID, obs.ID)) // idempotent

	effects := must(trace.TraceEffectsGraph(ctx, action.ID, 2))
	if len(effects.Edges) != 1 {
		t.Errorf("len(Edges) = %d, want 1", len(effects.Edges))
	}

	if err := trace.Link(ctx, fact.ID, aitracecause.RelationCaused, action.ID); !errors.Is(err, semantic.ErrRelationNotAllowed) {
		t.Errorf("Link(Fact CAUSED Action): err = %v, want ErrRelationNotAllowed", err)
	}
}

func TestCustomSchema(t *testing.T) {
	ctx := context.Background()

	const challenges aitracecause.RelationType = "CHALLENGES"

	schema := must(aitracecause.DefaultSchema().With(aitracecause.RelationRule{
		From:     aitracecause.NodeTypeFact,
		Relation: challenges,
		To:       aitracecause.NodeTypeDecision,
	}))

	trace := newTickingTrace(t, aitracecause.WithSchema(schema))
	fact := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "budget frozen"}))

	// Fact -> Decision is now ambiguous (BASIS_OF or CHALLENGES).
	_, err := trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "d"}, aitracecause.WithCauses(fact.ID))
	if !errors.Is(err, semantic.ErrAmbiguousRelation) {
		t.Fatalf("err = %v, want ErrAmbiguousRelation", err)
	}

	decision := must(trace.RecordDecision(
		ctx,
		aitracecause.Decision{Outcome: "scale anyway"},
		aitracecause.WithCausesVia(challenges, fact.ID),
	))
	action := must(trace.RecordAction(ctx, aitracecause.Action{Name: "scale"}, aitracecause.WithCauses(decision.ID)))

	explanation := must(trace.Explain(ctx, action.ID, 5))
	if !strings.Contains(explanation, `Decision "scale anyway" <-[CHALLENGES]- Fact "budget frozen".`) {
		t.Errorf("custom relation not rendered:\n%s", explanation)
	}

	// The default schema is unchanged by With.
	if aitracecause.DefaultSchema().Allows(aitracecause.NodeTypeFact, challenges, aitracecause.NodeTypeDecision) {
		t.Error("Schema.With mutated the default schema")
	}
}

func TestRestrictedSchemaRejectsBuiltInRelations(t *testing.T) {
	ctx := context.Background()

	schema := must(aitracecause.NewSchema(aitracecause.RelationRule{
		From:     aitracecause.NodeTypeFact,
		Relation: aitracecause.RelationBasisOf,
		To:       aitracecause.NodeTypeDecision,
	}))

	trace := newTickingTrace(t, aitracecause.WithSchema(schema))
	decision := must(trace.RecordDecision(ctx, aitracecause.Decision{Outcome: "d"}))
	action := must(trace.RecordAction(ctx, aitracecause.Action{Name: "a"}))

	var relErr *semantic.RelationNotAllowedError
	if err := trace.Caused(ctx, decision.ID, action.ID); !errors.As(err, &relErr) || relErr.Relation != aitracecause.RelationCaused {
		t.Fatalf("Caused() error = %v, want RelationNotAllowedError(CAUSED)", err)
	}
}

func TestSchemaRejectsInvalidRules(t *testing.T) {
	for name, rule := range map[string]aitracecause.RelationRule{
		"empty relation": {From: "Fact", To: "Fact"},
		"reserved":       {From: "Fact", Relation: semantic.RelationRetracted, To: "Fact"},
		"duplicate":      {From: "Source", Relation: aitracecause.RelationProduced, To: "Observation"},
	} {
		if _, err := aitracecause.DefaultSchema().With(rule); !errors.Is(err, semantic.ErrInvalidRelationRule) {
			t.Errorf("%s: err = %v, want ErrInvalidRelationRule", name, err)
		}
	}

	if got := len(aitracecause.DefaultSchema().Rules()); got != 7 {
		t.Errorf("DefaultSchema has %d rules, want 7", got)
	}
}

func TestInferenceCyclesTerminate(t *testing.T) {
	ctx := context.Background()
	trace := newTickingTrace(t)

	a := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "A"}))
	b := must(trace.RecordFact(ctx, aitracecause.Fact{Statement: "B"}, aitracecause.WithCauses(a.ID)))
	mustOK(t, trace.Infers(ctx, b.ID, a.ID))

	effects := must(trace.TraceEffectsGraph(ctx, a.ID, 100))
	if len(effects.Visits) != 2 || len(effects.Edges) != 2 {
		t.Fatalf("visits=%d edges=%d, want 2 and 2", len(effects.Visits), len(effects.Edges))
	}
}

func TestMaxVisitsBoundsLargeTraces(t *testing.T) {
	ctx := context.Background()

	build := func(trace *aitracecause.Trace) string {
		step := trace.Run("r").Step("")
		source := step.Source(aitracecause.Source{Kind: "firehose"})

		for i := range 100 {
			step.Observation(aitracecause.Observation{Name: fmt.Sprint("o", i)}, source)
		}

		must(step.Commit(ctx))

		return source
	}

	limited := newTickingTrace(t, aitracecause.WithMaxVisits(50))
	if _, err := limited.TraceSourceEffects(ctx, build(limited), 1); !errors.Is(err, graph.ErrMaxVisitsExceeded) {
		t.Errorf("limited: err = %v, want ErrMaxVisitsExceeded", err)
	}

	unlimited := newTickingTrace(t, aitracecause.WithMaxVisits(0))
	visits, err := unlimited.TraceSourceEffects(ctx, build(unlimited), 1)
	if err != nil || len(visits) != 101 {
		t.Errorf("unlimited: %d visits, err = %v; want 101, nil", len(visits), err)
	}
}

func TestRetractedFactBlastRadius(t *testing.T) {
	ctx := context.Background()
	trace := newTickingTrace(t)

	step := trace.Run("r").Step("")
	fact := step.Fact(aitracecause.Fact{Statement: "customer is fraudulent"})
	block := step.Decision(aitracecause.Decision{Outcome: "block account"}, fact)
	step.Action(aitracecause.Action{Name: "block"}, block)
	step.Action(aitracecause.Action{Name: "notify_customer"}, block)
	must(step.Commit(ctx))

	must(trace.Retract(ctx, aitracecause.Retraction{NodeID: fact, Reason: "false positive"}))

	effects := must(trace.TraceEffectsGraph(ctx, fact, 5))

	actions := 0
	for _, visit := range effects.Visits {
		if visit.Node.Type == "Action" {
			actions++
		}
	}

	if actions != 2 {
		t.Errorf("affected actions = %d, want 2", actions)
	}

	if effects.Retractions[fact].Reason != "false positive" {
		t.Errorf("retraction missing from effects graph: %v", effects.Retractions)
	}
}
