package semantic_test

import (
	"context"
	"errors"
	"testing"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

type fakeIDGenerator struct {
	ids   []string
	index int
}

func (g *fakeIDGenerator) NewID() (string, error) {
	if g.index >= len(g.ids) {
		return "", errors.New("no more fake IDs")
	}

	id := g.ids[g.index]
	g.index++

	return id, nil
}

func newTestService(
	t *testing.T,
	generatedIDs ...string,
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

	options := make(
		[]semantic.Option,
		0,
		1,
	)

	if len(generatedIDs) > 0 {
		options = append(
			options,
			semantic.WithIDGenerator(
				&fakeIDGenerator{
					ids: generatedIDs,
				},
			),
		)
	}

	service, err := semantic.NewService(
		g,
		options...,
	)
	if err != nil {
		t.Fatalf(
			"semantic.NewService() error = %v",
			err,
		)
	}

	return service, g
}

func TestRecordObservation(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(t)

	observation := semantic.Observation{
		ID:    "observation-001",
		Name:  "cpu_usage",
		Value: 94,
		Metadata: map[string]any{
			"unit": "%",
		},
	}

	recorded, err := service.RecordObservation(
		ctx,
		observation,
	)
	if err != nil {
		t.Fatalf(
			"RecordObservation() error = %v",
			err,
		)
	}

	if recorded.ID != observation.ID {
		t.Errorf(
			"recorded.ID = %q, want %q",
			recorded.ID,
			observation.ID,
		)
	}

	node, err := g.GetNode(
		ctx,
		recorded.ID,
	)
	if err != nil {
		t.Fatalf(
			"GetNode() error = %v",
			err,
		)
	}

	if node.Type != "Observation" {
		t.Errorf(
			"node.Type = %q, want Observation",
			node.Type,
		)
	}

	if node.Properties["name"] != "cpu_usage" {
		t.Errorf(
			"name = %v, want cpu_usage",
			node.Properties["name"],
		)
	}

	if node.Properties["value"] != 94 {
		t.Errorf(
			"value = %v, want 94",
			node.Properties["value"],
		)
	}
}

func TestRecordObservationGeneratesID(
	t *testing.T,
) {
	ctx := context.Background()

	service, g := newTestService(
		t,
		"generated-observation",
	)

	recorded, err := service.RecordObservation(
		ctx,
		semantic.Observation{
			Name:  "cpu_usage",
			Value: 94,
		},
	)
	if err != nil {
		t.Fatalf(
			"RecordObservation() error = %v",
			err,
		)
	}

	if recorded.ID != "generated-observation" {
		t.Errorf(
			"recorded.ID = %q, want generated-observation",
			recorded.ID,
		)
	}

	node, err := g.GetNode(
		ctx,
		recorded.ID,
	)
	if err != nil {
		t.Fatalf(
			"GetNode() error = %v",
			err,
		)
	}

	if node.ID != "generated-observation" {
		t.Errorf(
			"node.ID = %q, want generated-observation",
			node.ID,
		)
	}
}

func TestRecordObservationPreservesProvidedID(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(t)

	recorded, err := service.RecordObservation(
		ctx,
		semantic.Observation{
			ID:    "external-observation-id",
			Name:  "cpu_usage",
			Value: 94,
		},
	)
	if err != nil {
		t.Fatalf(
			"RecordObservation() error = %v",
			err,
		)
	}

	if recorded.ID != "external-observation-id" {
		t.Errorf(
			"recorded.ID = %q, want external-observation-id",
			recorded.ID,
		)
	}
}

func TestRecordObservationRejectsEmptyName(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(t)

	_, err := service.RecordObservation(
		ctx,
		semantic.Observation{
			ID: "observation-001",
		},
	)

	if !errors.Is(
		err,
		semantic.ErrEmptyObservationName,
	) {
		t.Fatalf(
			"RecordObservation() error = %v, want ErrEmptyObservationName",
			err,
		)
	}
}

func TestRecordFact(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(t)

	fact := semantic.Fact{
		ID:         "fact-001",
		Statement:  "CPU usage is high",
		Confidence: 0.98,
	}

	recorded, err := service.RecordFact(
		ctx,
		fact,
	)
	if err != nil {
		t.Fatalf(
			"RecordFact() error = %v",
			err,
		)
	}

	if recorded.ID != fact.ID {
		t.Errorf(
			"recorded.ID = %q, want %q",
			recorded.ID,
			fact.ID,
		)
	}

	node, err := g.GetNode(
		ctx,
		recorded.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if node.Type != "Fact" {
		t.Errorf(
			"node.Type = %q, want Fact",
			node.Type,
		)
	}

	if node.Properties["statement"] != fact.Statement {
		t.Errorf(
			"statement = %v, want %q",
			node.Properties["statement"],
			fact.Statement,
		)
	}

	if node.Properties["confidence"] != 0.98 {
		t.Errorf(
			"confidence = %v, want 0.98",
			node.Properties["confidence"],
		)
	}
}

func TestRecordFactGeneratesID(t *testing.T) {
	ctx := context.Background()

	service, _ := newTestService(
		t,
		"generated-fact",
	)

	recorded, err := service.RecordFact(
		ctx,
		semantic.Fact{
			Statement:  "CPU usage is high",
			Confidence: 0.98,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if recorded.ID != "generated-fact" {
		t.Errorf(
			"recorded.ID = %q, want generated-fact",
			recorded.ID,
		)
	}
}

func TestRecordFactRejectsInvalidConfidence(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(t)

	_, err := service.RecordFact(
		ctx,
		semantic.Fact{
			ID:         "fact-001",
			Statement:  "CPU usage is high",
			Confidence: 1.2,
		},
	)

	if !errors.Is(
		err,
		semantic.ErrInvalidConfidence,
	) {
		t.Fatalf(
			"RecordFact() error = %v, want ErrInvalidConfidence",
			err,
		)
	}
}

func TestRecordFactRejectsEmptyStatement(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(t)

	_, err := service.RecordFact(
		ctx,
		semantic.Fact{
			ID:         "fact-001",
			Confidence: 0.9,
		},
	)

	if !errors.Is(
		err,
		semantic.ErrEmptyFactStatement,
	) {
		t.Fatalf(
			"RecordFact() error = %v, want ErrEmptyFactStatement",
			err,
		)
	}
}

func TestRecordDecision(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(t)

	decision := semantic.Decision{
		ID:         "decision-001",
		Outcome:    "Scale the service",
		Rationale:  "CPU usage is consistently above 90%",
		Confidence: 0.92,
	}

	recorded, err := service.RecordDecision(
		ctx,
		decision,
	)
	if err != nil {
		t.Fatalf(
			"RecordDecision() error = %v",
			err,
		)
	}

	if recorded.ID != decision.ID {
		t.Errorf(
			"recorded.ID = %q, want %q",
			recorded.ID,
			decision.ID,
		)
	}

	node, err := g.GetNode(
		ctx,
		recorded.ID,
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

func TestRecordDecisionGeneratesID(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(
		t,
		"generated-decision",
	)

	recorded, err := service.RecordDecision(
		ctx,
		semantic.Decision{
			Outcome:    "Scale service",
			Rationale:  "CPU usage is high",
			Confidence: 0.92,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if recorded.ID != "generated-decision" {
		t.Errorf(
			"recorded.ID = %q, want generated-decision",
			recorded.ID,
		)
	}
}

func TestRecordDecisionRejectsInvalidConfidence(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(t)

	_, err := service.RecordDecision(
		ctx,
		semantic.Decision{
			ID:         "decision-001",
			Outcome:    "Scale service",
			Confidence: -0.1,
		},
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

	_, err := service.RecordDecision(
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

func TestRecordAction(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(t)

	action := semantic.Action{
		ID:     "action-001",
		Name:   "scale_service",
		Target: "payments-api",
		Parameters: map[string]any{
			"replicas": 5,
		},
	}

	recorded, err := service.RecordAction(
		ctx,
		action,
	)
	if err != nil {
		t.Fatalf(
			"RecordAction() error = %v",
			err,
		)
	}

	if recorded.ID != action.ID {
		t.Errorf(
			"recorded.ID = %q, want %q",
			recorded.ID,
			action.ID,
		)
	}

	node, err := g.GetNode(
		ctx,
		recorded.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if node.Type != "Action" {
		t.Errorf(
			"node.Type = %q, want Action",
			node.Type,
		)
	}

	if node.Properties["name"] != "scale_service" {
		t.Errorf(
			"name = %v, want scale_service",
			node.Properties["name"],
		)
	}

	if node.Properties["target"] != "payments-api" {
		t.Errorf(
			"target = %v, want payments-api",
			node.Properties["target"],
		)
	}
}

func TestRecordActionGeneratesID(t *testing.T) {
	ctx := context.Background()

	service, _ := newTestService(
		t,
		"generated-action",
	)

	recorded, err := service.RecordAction(
		ctx,
		semantic.Action{
			Name: "scale_service",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if recorded.ID != "generated-action" {
		t.Errorf(
			"recorded.ID = %q, want generated-action",
			recorded.ID,
		)
	}
}

func TestRecordActionRejectsEmptyName(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(t)

	_, err := service.RecordAction(
		ctx,
		semantic.Action{
			ID: "action-001",
		},
	)

	if !errors.Is(
		err,
		semantic.ErrEmptyActionName,
	) {
		t.Fatalf(
			"RecordAction() error = %v, want ErrEmptyActionName",
			err,
		)
	}
}

func TestRecordSource(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(t)

	source := semantic.Source{
		ID:   "source-001",
		Kind: "Prometheus",
		URI:  "prometheus://production/cpu_usage",
		Metadata: map[string]any{
			"environment": "production",
		},
	}

	recorded, err := service.RecordSource(
		ctx,
		source,
	)
	if err != nil {
		t.Fatalf(
			"RecordSource() error = %v",
			err,
		)
	}

	if recorded.ID != source.ID {
		t.Errorf(
			"recorded.ID = %q, want %q",
			recorded.ID,
			source.ID,
		)
	}

	node, err := g.GetNode(
		ctx,
		recorded.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if node.Type != "Source" {
		t.Errorf(
			"node.Type = %q, want Source",
			node.Type,
		)
	}

	if node.Properties["kind"] != source.Kind {
		t.Errorf(
			"kind = %v, want %q",
			node.Properties["kind"],
			source.Kind,
		)
	}

	if node.Properties["uri"] != source.URI {
		t.Errorf(
			"uri = %v, want %q",
			node.Properties["uri"],
			source.URI,
		)
	}
}

func TestRecordSourceGeneratesID(t *testing.T) {
	ctx := context.Background()

	service, _ := newTestService(
		t,
		"generated-source",
	)

	recorded, err := service.RecordSource(
		ctx,
		semantic.Source{
			Kind: "Prometheus",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if recorded.ID != "generated-source" {
		t.Errorf(
			"recorded.ID = %q, want generated-source",
			recorded.ID,
		)
	}
}

func TestRecordSourceRejectsEmptyKind(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(t)

	_, err := service.RecordSource(
		ctx,
		semantic.Source{
			ID: "source-001",
		},
	)

	if !errors.Is(
		err,
		semantic.ErrEmptySourceKind,
	) {
		t.Fatalf(
			"RecordSource() error = %v, want ErrEmptySourceKind",
			err,
		)
	}
}

func TestSupports(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(
		t,
		"edge-supports",
	)

	observation, err := service.RecordObservation(
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

	fact, err := service.RecordFact(
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

	if err := service.Supports(
		ctx,
		observation.ID,
		fact.ID,
	); err != nil {
		t.Fatalf(
			"Supports() error = %v",
			err,
		)
	}

	edge, err := g.GetEdge(
		ctx,
		"edge-supports",
	)
	if err != nil {
		t.Fatal(err)
	}

	if edge.From != observation.ID {
		t.Errorf(
			"edge.From = %q, want %q",
			edge.From,
			observation.ID,
		)
	}

	if edge.To != fact.ID {
		t.Errorf(
			"edge.To = %q, want %q",
			edge.To,
			fact.ID,
		)
	}

	if edge.Type != "SUPPORTS" {
		t.Errorf(
			"edge.Type = %q, want SUPPORTS",
			edge.Type,
		)
	}
}

func TestSupportsRejectsWrongNodeType(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(
		t,
		"edge-supports",
	)

	factA, err := service.RecordFact(
		ctx,
		semantic.Fact{
			ID:         "fact-a",
			Statement:  "Fact A",
			Confidence: 0.9,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	factB, err := service.RecordFact(
		ctx,
		semantic.Fact{
			ID:         "fact-b",
			Statement:  "Fact B",
			Confidence: 0.9,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = service.Supports(
		ctx,
		factA.ID,
		factB.ID,
	)

	if !errors.Is(
		err,
		semantic.ErrUnexpectedNodeType,
	) {
		t.Fatalf(
			"Supports() error = %v, want ErrUnexpectedNodeType",
			err,
		)
	}
}

func TestBasisOf(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(
		t,
		"edge-basis",
	)

	fact, err := service.RecordFact(
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

	decision, err := service.RecordDecision(
		ctx,
		semantic.Decision{
			ID:         "decision-001",
			Outcome:    "Scale the service",
			Rationale:  "High CPU load requires more capacity",
			Confidence: 0.92,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := service.BasisOf(
		ctx,
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
		"edge-basis",
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

	service, _ := newTestService(
		t,
		"edge-basis",
	)

	observation, err := service.RecordObservation(
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

	decision, err := service.RecordDecision(
		ctx,
		semantic.Decision{
			ID:         "decision-001",
			Outcome:    "Scale the service",
			Confidence: 0.9,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = service.BasisOf(
		ctx,
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

func TestCaused(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(
		t,
		"edge-caused",
	)

	decision, err := service.RecordDecision(
		ctx,
		semantic.Decision{
			ID:         "decision-001",
			Outcome:    "Scale the service",
			Rationale:  "CPU usage is too high",
			Confidence: 0.92,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	action, err := service.RecordAction(
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

	if err := service.Caused(
		ctx,
		decision.ID,
		action.ID,
	); err != nil {
		t.Fatalf(
			"Caused() error = %v",
			err,
		)
	}

	edge, err := g.GetEdge(
		ctx,
		"edge-caused",
	)
	if err != nil {
		t.Fatal(err)
	}

	if edge.From != decision.ID {
		t.Errorf(
			"edge.From = %q, want %q",
			edge.From,
			decision.ID,
		)
	}

	if edge.To != action.ID {
		t.Errorf(
			"edge.To = %q, want %q",
			edge.To,
			action.ID,
		)
	}

	if edge.Type != "CAUSED" {
		t.Errorf(
			"edge.Type = %q, want CAUSED",
			edge.Type,
		)
	}
}

func TestCausedRejectsWrongDecisionType(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(
		t,
		"edge-caused",
	)

	fact, err := service.RecordFact(
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

	action, err := service.RecordAction(
		ctx,
		semantic.Action{
			ID:   "action-001",
			Name: "scale_service",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	err = service.Caused(
		ctx,
		fact.ID,
		action.ID,
	)

	if !errors.Is(
		err,
		semantic.ErrUnexpectedNodeType,
	) {
		t.Fatalf(
			"Caused() error = %v, want ErrUnexpectedNodeType",
			err,
		)
	}
}

func TestProduced(t *testing.T) {
	ctx := context.Background()

	service, g := newTestService(
		t,
		"edge-produced",
	)

	source, err := service.RecordSource(
		ctx,
		semantic.Source{
			ID:   "source-001",
			Kind: "Prometheus",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	observation, err := service.RecordObservation(
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

	if err := service.Produced(
		ctx,
		source.ID,
		observation.ID,
	); err != nil {
		t.Fatalf(
			"Produced() error = %v",
			err,
		)
	}

	edge, err := g.GetEdge(
		ctx,
		"edge-produced",
	)
	if err != nil {
		t.Fatal(err)
	}

	if edge.From != source.ID {
		t.Errorf(
			"edge.From = %q, want %q",
			edge.From,
			source.ID,
		)
	}

	if edge.To != observation.ID {
		t.Errorf(
			"edge.To = %q, want %q",
			edge.To,
			observation.ID,
		)
	}

	if edge.Type != "PRODUCED" {
		t.Errorf(
			"edge.Type = %q, want PRODUCED",
			edge.Type,
		)
	}
}

func TestProducedRejectsWrongSourceType(
	t *testing.T,
) {
	ctx := context.Background()

	service, _ := newTestService(
		t,
		"edge-produced",
	)

	fact, err := service.RecordFact(
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

	observation, err := service.RecordObservation(
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

	err = service.Produced(
		ctx,
		fact.ID,
		observation.ID,
	)

	if !errors.Is(
		err,
		semantic.ErrUnexpectedNodeType,
	) {
		t.Fatalf(
			"Produced() error = %v, want ErrUnexpectedNodeType",
			err,
		)
	}
}

func TestTraceActionCause(t *testing.T) {
	ctx := context.Background()

	service, _ := newTestService(
		t,
		"edge-produced",
		"edge-supports",
		"edge-basis",
		"edge-caused",
	)

	source, err := service.RecordSource(
		ctx,
		semantic.Source{
			ID:   "source-001",
			Kind: "Prometheus",
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	observation, err := service.RecordObservation(
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

	fact, err := service.RecordFact(
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

	decision, err := service.RecordDecision(
		ctx,
		semantic.Decision{
			ID:         "decision-001",
			Outcome:    "Scale the service",
			Rationale:  "CPU load is above the expected threshold",
			Confidence: 0.92,
		},
	)
	if err != nil {
		t.Fatal(err)
	}

	action, err := service.RecordAction(
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

	if err := service.Produced(
		ctx,
		source.ID,
		observation.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := service.Supports(
		ctx,
		observation.ID,
		fact.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := service.BasisOf(
		ctx,
		fact.ID,
		decision.ID,
	); err != nil {
		t.Fatal(err)
	}

	if err := service.Caused(
		ctx,
		decision.ID,
		action.ID,
	); err != nil {
		t.Fatal(err)
	}

	results, err := service.TraceActionCause(
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
