package semantic_test

import (
	"context"
	"errors"
	"testing"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

func newTestService(
	t *testing.T,
) (*semantic.Service, *graph.Graph) {
	t.Helper()

	store := memory.New()

	g, err := graph.New(store)
	if err != nil {
		t.Fatalf(
			"graph.New() error = %v",
			err,
		)
	}

	service, err := semantic.NewService(g)
	if err != nil {
		t.Fatalf(
			"semantic.NewService() error = %v",
			err,
		)
	}

	return service, g
}

func TestRecordDecision(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(t)

	decision := semantic.Decision{
		ID: "decision-001",

		Outcome: "Scale the service",

		Rationale: "CPU usage is consistently above 90%",

		Confidence: 0.92,
	}

	if err := service.RecordDecision(
		ctx,
		decision,
	); err != nil {
		t.Fatalf(
			"RecordDecision() error = %v",
			err,
		)
	}

	node, err := g.GetNode(
		ctx,
		decision.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if node.Type != "Decision" {
		t.Errorf(
			"node.Type = %q, want Decision",
			node.Type,
		)
	}

	if node.Properties["outcome"] != decision.Outcome {
		t.Errorf(
			"outcome = %v, want %q",
			node.Properties["outcome"],
			decision.Outcome,
		)
	}

	if node.Properties["rationale"] != decision.Rationale {
		t.Errorf(
			"rationale = %v, want %q",
			node.Properties["rationale"],
			decision.Rationale,
		)
	}

	if node.Properties["confidence"] != 0.92 {
		t.Errorf(
			"confidence = %v, want 0.92",
			node.Properties["confidence"],
		)
	}
}

func TestRecordDecisionRejectsInvalidConfidence(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(t)

	decision := semantic.Decision{
		ID:         "decision-001",
		Outcome:    "Scale service",
		Confidence: -0.1,
	}

	err := service.RecordDecision(
		ctx,
		decision,
	)

	if !errors.Is(
		err,
		semantic.ErrInvalidConfidence,
	) {
		t.Fatalf(
			"RecordDecision() error = %v, want ErrInvalidConfidence",
			err,
		)
	}
}

func TestRecordDecisionRejectsEmptyOutcome(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(t)

	err := service.RecordDecision(
		ctx,
		semantic.Decision{
			ID:         "decision-001",
			Confidence: 0.9,
		},
	)

	if !errors.Is(
		err,
		semantic.ErrEmptyDecisionOutcome,
	) {
		t.Fatalf(
			"RecordDecision() error = %v, want ErrEmptyDecisionOutcome",
			err,
		)
	}
}

func TestBasisOf(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(t)

	fact := semantic.Fact{
		ID:         "fact-001",
		Statement:  "CPU usage is high",
		Confidence: 0.98,
	}

	decision := semantic.Decision{
		ID:         "decision-001",
		Outcome:    "Scale the service",
		Rationale:  "High CPU load requires more capacity",
		Confidence: 0.92,
	}

	if err := service.RecordFact(
		ctx,
		fact,
	); err != nil {
		t.Fatal(err)
	}

	if err := service.RecordDecision(
		ctx,
		decision,
	); err != nil {
		t.Fatal(err)
	}

	if err := service.BasisOf(
		ctx,
		"edge-001",
		fact.ID,
		decision.ID,
	); err != nil {
		t.Fatalf(
			"BasisOf() error = %v",
			err,
		)
	}

	edge, err := g.GetEdge(
		ctx,
		"edge-001",
	)
	if err != nil {
		t.Fatal(err)
	}

	if edge.From != fact.ID {
		t.Errorf(
			"edge.From = %q, want %q",
			edge.From,
			fact.ID,
		)
	}

	if edge.To != decision.ID {
		t.Errorf(
			"edge.To = %q, want %q",
			edge.To,
			decision.ID,
		)
	}

	if edge.Type != "BASIS_OF" {
		t.Errorf(
			"edge.Type = %q, want BASIS_OF",
			edge.Type,
		)
	}
}

func TestBasisOfRejectsWrongFactType(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(t)

	observation := semantic.Observation{
		ID:    "observation-001",
		Name:  "cpu_usage",
		Value: 94,
	}

	decision := semantic.Decision{
		ID:         "decision-001",
		Outcome:    "Scale the service",
		Confidence: 0.9,
	}

	if err := service.RecordObservation(
		ctx,
		observation,
	); err != nil {
		t.Fatal(err)
	}

	if err := service.RecordDecision(
		ctx,
		decision,
	); err != nil {
		t.Fatal(err)
	}

	err := service.BasisOf(
		ctx,
		"edge-001",
		observation.ID,
		decision.ID,
	)

	if !errors.Is(
		err,
		semantic.ErrUnexpectedNodeType,
	) {
		t.Fatalf(
			"BasisOf() error = %v, want ErrUnexpectedNodeType",
			err,
		)
	}
}

func TestDecisionCauseTrace(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(t)

	observation := semantic.Observation{
		ID:    "observation-001",
		Name:  "cpu_usage",
		Value: 94,
	}

	fact := semantic.Fact{
		ID:         "fact-001",
		Statement:  "CPU usage is high",
		Confidence: 0.98,
	}

	decision := semantic.Decision{
		ID:         "decision-001",
		Outcome:    "Scale the service",
		Rationale:  "CPU load is above the expected threshold",
		Confidence: 0.92,
	}

	if err := service.RecordObservation(
		ctx,
		observation,
	); err != nil {
		t.Fatal(err)
	}

	if err := service.RecordFact(
		ctx,
		fact,
	); err != nil {
		t.Fatal(err)
	}

	if err := service.RecordDecision(
		ctx,
		decision,
	); err != nil {
		t.Fatal(err)
	}

	if err := service.Supports(
		ctx,
		"edge-supports",
		observation.ID,
		fact.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := service.BasisOf(
		ctx,
		"edge-basis",
		fact.ID,
		decision.ID,
	); err != nil {
		t.Fatal(err)
	}

	results, err := g.BFS(
		ctx,
		decision.ID,
		graph.DirectionIncoming,
		2,
	)
	if err != nil {
		t.Fatal(err)
	}

	want := []string{
		"decision-001",
		"fact-001",
		"observation-001",
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
