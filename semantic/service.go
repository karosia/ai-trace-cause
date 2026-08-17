// Package semantic implements the AI-agent causal domain model on top
// of the graph package: the Source, Observation, Fact, Decision, and
// Action entity types, the typed relationships that connect them, and
// the Service that validates and records them.
package semantic

import (
	"context"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
)

// Service validates and records the semantic causal model into an
// underlying graph.Graph. It is the implementation behind the public
// aitracecause.Trace type.
type Service struct {
	graph *graph.Graph

	now func() time.Time

	telemetry TelemetryHook

	idGenerator IDGenerator
}

// Option configures a Service created with NewService.
type Option func(*Service)

// WithClock configures the Service to use now instead of time.Now for
// timestamping recorded entities and relationships. A nil now is
// ignored.
func WithClock(
	now func() time.Time,
) Option {
	return func(service *Service) {
		if now != nil {
			service.now = now
		}
	}
}

// WithTelemetryHook configures the Service to correlate recorded
// entities and relationships with the active OpenTelemetry trace and
// span found in the recording context, using hook.
func WithTelemetryHook(
	hook TelemetryHook,
) Option {
	return func(service *Service) {
		service.telemetry = hook
	}
}

// WithIDGenerator configures the Service to use generator for
// producing IDs when a caller does not supply one. A nil generator is
// ignored.
func WithIDGenerator(
	generator IDGenerator,
) Option {
	return func(service *Service) {
		if generator != nil {
			service.idGenerator = generator
		}
	}
}

// NewService creates a Service backed by graph g. It returns
// ErrNilGraph if g is nil.
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

// RecordSource validates and records a Source node, generating an ID
// if source.ID is empty. It returns ErrEmptySourceKind if Kind is
// empty.
func (s *Service) RecordSource(
	ctx context.Context,
	source Source,
) (Source, error) {
	if source.Kind == "" {
		return Source{}, ErrEmptySourceKind
	}

	id, err := s.resolveID(
		source.ID,
	)
	if err != nil {
		return Source{}, err
	}

	source.ID = id

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

	if err := s.graph.AddNode(
		ctx,
		node,
	); err != nil {
		return Source{}, err
	}

	return source, nil
}

// RecordObservation validates and records an Observation node,
// generating an ID if observation.ID is empty. It returns
// ErrEmptyObservationName if Name is empty.
func (s *Service) RecordObservation(
	ctx context.Context,
	observation Observation,
) (Observation, error) {
	if observation.Name == "" {
		return Observation{}, ErrEmptyObservationName
	}

	id, err := s.resolveID(
		observation.ID,
	)
	if err != nil {
		return Observation{}, err
	}

	observation.ID = id

	node := graph.Node{
		ID:   observation.ID,
		Type: string(NodeTypeObservation),
		Properties: map[string]any{
			"name":  observation.Name,
			"value": observation.Value,
		},
	}

	if observation.Metadata != nil {
		node.Properties["metadata"] = cloneMap(
			observation.Metadata,
		)
	}

	s.applyTemporal(
		&node,
		observation.Validity,
	)

	s.applyTelemetry(
		ctx,
		&node,
	)

	if err := s.graph.AddNode(
		ctx,
		node,
	); err != nil {
		return Observation{}, err
	}

	return observation, nil
}

// RecordFact validates and records a Fact node, generating an ID if
// fact.ID is empty. It returns ErrEmptyFactStatement if Statement is
// empty, or ErrInvalidConfidence if Confidence is outside [0, 1].
func (s *Service) RecordFact(
	ctx context.Context,
	fact Fact,
) (Fact, error) {
	if fact.Statement == "" {
		return Fact{}, ErrEmptyFactStatement
	}

	if fact.Confidence < 0 ||
		fact.Confidence > 1 {
		return Fact{}, ErrInvalidConfidence
	}

	id, err := s.resolveID(
		fact.ID,
	)
	if err != nil {
		return Fact{}, err
	}

	fact.ID = id

	node := graph.Node{
		ID:   fact.ID,
		Type: string(NodeTypeFact),
		Properties: map[string]any{
			"statement":  fact.Statement,
			"confidence": fact.Confidence,
		},
	}

	if fact.Metadata != nil {
		node.Properties["metadata"] = cloneMap(
			fact.Metadata,
		)
	}

	s.applyTemporal(
		&node,
		fact.Validity,
	)

	s.applyTelemetry(
		ctx,
		&node,
	)

	if err := s.graph.AddNode(
		ctx,
		node,
	); err != nil {
		return Fact{}, err
	}

	return fact, nil
}

// RecordDecision validates and records a Decision node, generating an
// ID if decision.ID is empty. It returns ErrEmptyDecisionOutcome if
// Outcome is empty, or ErrInvalidConfidence if Confidence is outside
// [0, 1].
func (s *Service) RecordDecision(
	ctx context.Context,
	decision Decision,
) (Decision, error) {
	if decision.Outcome == "" {
		return Decision{}, ErrEmptyDecisionOutcome
	}

	if decision.Confidence < 0 ||
		decision.Confidence > 1 {
		return Decision{}, ErrInvalidConfidence
	}

	id, err := s.resolveID(
		decision.ID,
	)
	if err != nil {
		return Decision{}, err
	}

	decision.ID = id

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

	if err := s.graph.AddNode(
		ctx,
		node,
	); err != nil {
		return Decision{}, err
	}

	return decision, nil
}

// RecordAction validates and records an Action node, generating an ID
// if action.ID is empty. It returns ErrEmptyActionName if Name is
// empty.
func (s *Service) RecordAction(
	ctx context.Context,
	action Action,
) (Action, error) {
	if action.Name == "" {
		return Action{}, ErrEmptyActionName
	}

	id, err := s.resolveID(
		action.ID,
	)
	if err != nil {
		return Action{}, err
	}

	action.ID = id

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

	if err := s.graph.AddNode(
		ctx,
		node,
	); err != nil {
		return Action{}, err
	}

	return action, nil
}

// Produced records a Source -> Observation RelationProduced
// relationship. It returns UnexpectedNodeTypeError if either ID does
// not refer to a node of the expected type.
func (s *Service) Produced(
	ctx context.Context,
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

	return s.addRelation(
		ctx,
		sourceID,
		observationID,
		RelationProduced,
	)
}

// Supports records an Observation -> Fact RelationSupports
// relationship. It returns UnexpectedNodeTypeError if either ID does
// not refer to a node of the expected type.
func (s *Service) Supports(
	ctx context.Context,
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

	return s.addRelation(
		ctx,
		observationID,
		factID,
		RelationSupports,
	)
}

// BasisOf records a Fact -> Decision RelationBasisOf relationship. It
// returns UnexpectedNodeTypeError if either ID does not refer to a
// node of the expected type.
func (s *Service) BasisOf(
	ctx context.Context,
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

	return s.addRelation(
		ctx,
		factID,
		decisionID,
		RelationBasisOf,
	)
}

// Caused records a Decision -> Action RelationCaused relationship. It
// returns UnexpectedNodeTypeError if either ID does not refer to a
// node of the expected type.
func (s *Service) Caused(
	ctx context.Context,
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

	return s.addRelation(
		ctx,
		decisionID,
		actionID,
		RelationCaused,
	)
}

// TraceActionCause walks backward from the Action identified by
// actionID through its causal chain (Decision, Fact, Observation,
// Source), up to maxDepth hops, and returns the visited nodes in
// breadth-first order. It returns UnexpectedNodeTypeError if actionID
// does not refer to an Action node.
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

// TraceActionCauseAt is like TraceActionCause but reconstructs the
// causal chain as it existed at the given point in time, including
// only entities and relationships that were recorded and valid at at.
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

// TraceSourceEffects walks forward from the Source identified by
// sourceID through its downstream effects (Observation, Fact,
// Decision, Action), up to maxDepth hops, and returns the visited
// nodes in breadth-first order. It is the forward-tracing counterpart
// to TraceActionCause: instead of asking "why did this action
// happen?", it answers "what did this source go on to influence?". It
// returns UnexpectedNodeTypeError if sourceID does not refer to a
// Source node.
func (s *Service) TraceSourceEffects(
	ctx context.Context,
	sourceID string,
	maxDepth int,
) ([]graph.Visit, error) {
	if err := s.requireNodeType(
		ctx,
		sourceID,
		NodeTypeSource,
	); err != nil {
		return nil, err
	}

	return s.graph.BFS(
		ctx,
		sourceID,
		graph.DirectionOutgoing,
		maxDepth,
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

	return new(*value)
}

func (s *Service) resolveID(
	id string,
) (string, error) {
	if id != "" {
		return id, nil
	}

	if s.idGenerator == nil {
		return "", ErrIDGeneratorUnavailable
	}

	return s.idGenerator.NewID()
}

func (s *Service) addRelation(
	ctx context.Context,
	from string,
	to string,
	relation RelationType,
) error {
	edgeID, err := s.resolveID("")
	if err != nil {
		return err
	}

	edge := graph.Edge{
		ID:         edgeID,
		From:       from,
		To:         to,
		Type:       string(relation),
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
