package semantic

import (
	"context"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
)

type Service struct {
	graph *graph.Graph

	now func() time.Time

	telemetry TelemetryHook
}

type Option func(*Service)

func WithClock(
	now func() time.Time,
) Option {
	return func(service *Service) {
		if now != nil {
			service.now = now
		}
	}
}

func WithTelemetryHook(
	hook TelemetryHook,
) Option {
	return func(service *Service) {
		service.telemetry = hook
	}
}

func NewService(
	g *graph.Graph,
	options ...Option,
) (*Service, error) {
	if g == nil {
		return nil, ErrNilGraph
	}

	service := &Service{
		graph: g,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}

	for _, option := range options {
		option(service)
	}

	return service, nil
}

func (s *Service) RecordSource(
	ctx context.Context,
	source Source,
) error {
	if source.ID == "" {
		return ErrEmptySourceID
	}

	if source.Kind == "" {
		return ErrEmptySourceKind
	}

	node := graph.Node{
		ID:   source.ID,
		Type: string(NodeTypeSource),
		Properties: map[string]any{
			"kind": source.Kind,
			"uri":  source.URI,
		},
	}

	if source.Metadata != nil {
		node.Properties["metadata"] = cloneMap(
			source.Metadata,
		)
	}

	s.applyTemporal(
		&node,
		source.Validity,
	)

	s.applyTelemetry(
		ctx,
		&node,
	)

	return s.graph.AddNode(
		ctx,
		node,
	)
}

func (s *Service) RecordObservation(ctx context.Context, observation Observation) error {
	if observation.ID == "" {
		return ErrEmptyObservationID
	}

	if observation.Name == "" {
		return ErrEmptyObservationName
	}

	node := graph.Node{
		ID:   observation.ID,
		Type: string(NodeTypeObservation),
		Properties: map[string]any{
			"name":  observation.Name,
			"value": observation.Value,
		},
	}

	if observation.Metadata != nil {
		node.Properties["metadata"] = cloneMap(observation.Metadata)
	}

	s.applyTemporal(
		&node,
		observation.Validity,
	)

	s.applyTelemetry(
		ctx,
		&node,
	)

	return s.graph.AddNode(ctx, node)
}

func (s *Service) RecordFact(ctx context.Context, fact Fact) error {
	if fact.ID == "" {
		return ErrEmptyFactID
	}

	if fact.Statement == "" {
		return ErrEmptyFactStatement
	}

	if fact.Confidence < 0 || fact.Confidence > 1 {
		return ErrInvalidConfidence
	}

	node := graph.Node{
		ID:   fact.ID,
		Type: string(NodeTypeFact),
		Properties: map[string]any{
			"statement":  fact.Statement,
			"confidence": fact.Confidence,
		},
	}

	if fact.Metadata != nil {
		node.Properties["metadata"] = cloneMap(fact.Metadata)
	}

	s.applyTemporal(
		&node,
		fact.Validity,
	)

	s.applyTelemetry(
		ctx,
		&node,
	)

	return s.graph.AddNode(ctx, node)
}

func (s *Service) RecordDecision(
	ctx context.Context,
	decision Decision,
) error {
	if decision.ID == "" {
		return ErrEmptyDecisionID
	}

	if decision.Outcome == "" {
		return ErrEmptyDecisionOutcome
	}

	if decision.Confidence < 0 ||
		decision.Confidence > 1 {
		return ErrInvalidConfidence
	}

	node := graph.Node{
		ID:   decision.ID,
		Type: string(NodeTypeDecision),
		Properties: map[string]any{
			"outcome":    decision.Outcome,
			"rationale":  decision.Rationale,
			"confidence": decision.Confidence,
		},
	}

	if decision.Metadata != nil {
		node.Properties["metadata"] = cloneMap(
			decision.Metadata,
		)
	}

	s.applyTemporal(
		&node,
		decision.Validity,
	)

	s.applyTelemetry(
		ctx,
		&node,
	)

	return s.graph.AddNode(
		ctx,
		node,
	)
}

func (s *Service) RecordAction(
	ctx context.Context,
	action Action,
) error {
	if action.ID == "" {
		return ErrEmptyActionID
	}

	if action.Name == "" {
		return ErrEmptyActionName
	}

	node := graph.Node{
		ID:   action.ID,
		Type: string(NodeTypeAction),
		Properties: map[string]any{
			"name":   action.Name,
			"target": action.Target,
		},
	}

	if action.Parameters != nil {
		node.Properties["parameters"] = cloneMap(
			action.Parameters,
		)
	}

	if action.Metadata != nil {
		node.Properties["metadata"] = cloneMap(
			action.Metadata,
		)
	}

	s.applyTemporal(
		&node,
		action.Validity,
	)

	s.applyTelemetry(
		ctx,
		&node,
	)

	return s.graph.AddNode(
		ctx,
		node,
	)
}

func (s *Service) Produced(
	ctx context.Context,
	edgeID string,
	sourceID string,
	observationID string,
) error {
	if err := s.requireNodeType(
		ctx,
		sourceID,
		NodeTypeSource,
	); err != nil {
		return err
	}

	if err := s.requireNodeType(
		ctx,
		observationID,
		NodeTypeObservation,
	); err != nil {
		return err
	}

	edge := graph.Edge{
		ID:         edgeID,
		From:       sourceID,
		To:         observationID,
		Type:       string(RelationProduced),
		RecordedAt: s.now().UTC(),
	}

	s.applyEdgeTelemetry(
		ctx,
		&edge,
	)

	return s.graph.AddEdge(
		ctx,
		edge,
	)
}

func (s *Service) Supports(
	ctx context.Context,
	edgeID string,
	observationID string,
	factID string,
) error {
	if err := s.requireNodeType(
		ctx,
		observationID,
		NodeTypeObservation,
	); err != nil {
		return err
	}

	if err := s.requireNodeType(
		ctx,
		factID,
		NodeTypeFact,
	); err != nil {
		return err
	}

	edge := graph.Edge{
		ID:         edgeID,
		From:       observationID,
		To:         factID,
		Type:       string(RelationSupports),
		RecordedAt: s.now().UTC(),
	}

	s.applyEdgeTelemetry(
		ctx,
		&edge,
	)

	return s.graph.AddEdge(
		ctx,
		edge,
	)
}

func (s *Service) BasisOf(
	ctx context.Context,
	edgeID string,
	factID string,
	decisionID string,
) error {
	if err := s.requireNodeType(
		ctx,
		factID,
		NodeTypeFact,
	); err != nil {
		return err
	}

	if err := s.requireNodeType(
		ctx,
		decisionID,
		NodeTypeDecision,
	); err != nil {
		return err
	}

	edge := graph.Edge{
		ID:         edgeID,
		From:       factID,
		To:         decisionID,
		Type:       string(RelationBasisOf),
		RecordedAt: s.now().UTC(),
	}

	s.applyEdgeTelemetry(
		ctx,
		&edge,
	)

	return s.graph.AddEdge(
		ctx,
		edge,
	)
}

func (s *Service) Caused(
	ctx context.Context,
	edgeID string,
	decisionID string,
	actionID string,
) error {
	if err := s.requireNodeType(
		ctx,
		decisionID,
		NodeTypeDecision,
	); err != nil {
		return err
	}

	if err := s.requireNodeType(
		ctx,
		actionID,
		NodeTypeAction,
	); err != nil {
		return err
	}

	edge := graph.Edge{
		ID:         edgeID,
		From:       decisionID,
		To:         actionID,
		Type:       string(RelationCaused),
		RecordedAt: s.now().UTC(),
	}

	s.applyEdgeTelemetry(
		ctx,
		&edge,
	)

	return s.graph.AddEdge(
		ctx,
		edge,
	)
}

func (s *Service) TraceActionCause(
	ctx context.Context,
	actionID string,
	maxDepth int,
) ([]graph.Visit, error) {
	if err := s.requireNodeType(
		ctx,
		actionID,
		NodeTypeAction,
	); err != nil {
		return nil, err
	}

	return s.graph.BFS(
		ctx,
		actionID,
		graph.DirectionIncoming,
		maxDepth,
	)
}

func (s *Service) TraceActionCauseAt(
	ctx context.Context,
	actionID string,
	at time.Time,
	maxDepth int,
) ([]graph.Visit, error) {
	if err := s.requireNodeType(
		ctx,
		actionID,
		NodeTypeAction,
	); err != nil {
		return nil, err
	}

	return s.graph.BFSAt(
		ctx,
		actionID,
		graph.DirectionIncoming,
		maxDepth,
		at,
	)
}

func (s *Service) requireNodeType(ctx context.Context, nodeID string, expected NodeType) error {
	node, err := s.graph.GetNode(ctx, nodeID)
	if err != nil {
		return err
	}
	if node.Type != string(expected) {
		return &UnexpectedNodeTypeError{
			NodeID:   nodeID,
			Expected: expected,
			Actual:   node.Type,
		}
	}
	return nil
}

func cloneMap(values map[string]any) map[string]any {
	if values == nil {
		return nil
	}

	cloned := make(map[string]any, len(values))

	for key, value := range values {
		cloned[key] = value
	}

	return cloned
}

func (s *Service) applyTemporal(
	node *graph.Node,
	validity Validity,
) {
	node.RecordedAt = s.now().UTC()

	node.ValidFrom = cloneTime(
		validity.ValidFrom,
	)

	node.ValidUntil = cloneTime(
		validity.ValidUntil,
	)
}

func (s *Service) applyTelemetry(
	ctx context.Context,
	node *graph.Node,
) {
	if s.telemetry == nil {
		return
	}

	ref, ok := s.telemetry.CorrelationFromContext(
		ctx,
	)
	if !ok {
		return
	}

	node.Telemetry = &graph.TelemetryRef{
		TraceID: ref.TraceID,
		SpanID:  ref.SpanID,
	}
}

func (s *Service) applyEdgeTelemetry(
	ctx context.Context,
	edge *graph.Edge,
) {
	if s.telemetry == nil {
		return
	}

	ref, ok := s.telemetry.CorrelationFromContext(
		ctx,
	)
	if !ok {
		return
	}

	edge.Telemetry = &graph.TelemetryRef{
		TraceID: ref.TraceID,
		SpanID:  ref.SpanID,
	}
}

func cloneTime(
	value *time.Time,
) *time.Time {
	if value == nil {
		return nil
	}

	cloned := *value

	return &cloned
}
