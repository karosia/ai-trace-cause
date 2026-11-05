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

	schema Schema
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

// WithSchema configures which relationships the Service accepts. The
// default is DefaultSchema(); extend it with Schema.With to model
// application-specific relationships.
func WithSchema(
	schema Schema,
) Option {
	return func(service *Service) {
		service.schema = schema
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
		schema: DefaultSchema(),
	}

	for _, option := range options {
		option(service)
	}

	return service, nil
}

// RecordSource validates and records a Source node, generating an ID
// if source.ID is empty. It returns ErrEmptySourceKind if Kind is
// empty. The returned Source carries the assigned ID and RecordedAt.
func (s *Service) RecordSource(
	ctx context.Context,
	source Source,
) (Source, error) {
	node, err := source.toNode()
	if err != nil {
		return Source{}, err
	}

	node, err = s.record(ctx, node, nil)
	if err != nil {
		return Source{}, err
	}

	source.ID = node.ID
	source.RecordedAt = node.RecordedAt

	return source, nil
}

// RecordObservation validates and records an Observation node,
// generating an ID if observation.ID is empty. It returns
// ErrEmptyObservationName if Name is empty. Use WithCauses to link the
// Source(s) or Action(s) that produced it in the same call.
func (s *Service) RecordObservation(
	ctx context.Context,
	observation Observation,
	options ...RecordOption,
) (Observation, error) {
	node, err := observation.toNode()
	if err != nil {
		return Observation{}, err
	}

	node, err = s.record(ctx, node, options)
	if err != nil {
		return Observation{}, err
	}

	observation.ID = node.ID
	observation.RecordedAt = node.RecordedAt

	return observation, nil
}

// RecordFact validates and records a Fact node, generating an ID if
// fact.ID is empty. It returns ErrEmptyFactStatement if Statement is
// empty, or ErrInvalidConfidence if Confidence is outside [0, 1]. Use
// WithCauses to link the supporting Observation(s) or premise Fact(s)
// in the same call.
func (s *Service) RecordFact(
	ctx context.Context,
	fact Fact,
	options ...RecordOption,
) (Fact, error) {
	node, err := fact.toNode()
	if err != nil {
		return Fact{}, err
	}

	node, err = s.record(ctx, node, options)
	if err != nil {
		return Fact{}, err
	}

	fact.ID = node.ID
	fact.RecordedAt = node.RecordedAt

	return fact, nil
}

// RecordDecision validates and records a Decision node, generating an
// ID if decision.ID is empty. It returns ErrEmptyDecisionOutcome if
// Outcome is empty, or ErrInvalidConfidence if Confidence is outside
// [0, 1]. Use WithCauses to link the Fact(s) it was based on, or the
// Decision it refines, in the same call.
func (s *Service) RecordDecision(
	ctx context.Context,
	decision Decision,
	options ...RecordOption,
) (Decision, error) {
	node, err := decision.toNode()
	if err != nil {
		return Decision{}, err
	}

	node, err = s.record(ctx, node, options)
	if err != nil {
		return Decision{}, err
	}

	decision.ID = node.ID
	decision.RecordedAt = node.RecordedAt

	return decision, nil
}

// RecordAction validates and records an Action node, generating an ID
// if action.ID is empty. It returns ErrEmptyActionName if Name is
// empty. Use WithCauses to link the Decision(s) that caused it in the
// same call; this is required for TraceActionCauseAtActionTime to see
// those decisions.
func (s *Service) RecordAction(
	ctx context.Context,
	action Action,
	options ...RecordOption,
) (Action, error) {
	node, err := action.toNode()
	if err != nil {
		return Action{}, err
	}

	node, err = s.record(ctx, node, options)
	if err != nil {
		return Action{}, err
	}

	action.ID = node.ID
	action.RecordedAt = node.RecordedAt

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
	return s.typedLink(
		ctx,
		sourceID, NodeTypeSource,
		RelationProduced,
		observationID, NodeTypeObservation,
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
	return s.typedLink(
		ctx,
		observationID, NodeTypeObservation,
		RelationSupports,
		factID, NodeTypeFact,
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
	return s.typedLink(
		ctx,
		factID, NodeTypeFact,
		RelationBasisOf,
		decisionID, NodeTypeDecision,
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
	return s.typedLink(
		ctx,
		decisionID, NodeTypeDecision,
		RelationCaused,
		actionID, NodeTypeAction,
	)
}

// ResultedIn records an Action -> Observation RelationResultedIn
// relationship, such as a tool call and the result it returned. It
// returns UnexpectedNodeTypeError if either ID does not refer to a
// node of the expected type.
func (s *Service) ResultedIn(
	ctx context.Context,
	actionID string,
	observationID string,
) error {
	return s.typedLink(
		ctx,
		actionID, NodeTypeAction,
		RelationResultedIn,
		observationID, NodeTypeObservation,
	)
}

// Infers records a Fact -> Fact RelationInfers relationship from a
// premise to a conclusion drawn from it. It returns
// UnexpectedNodeTypeError if either ID does not refer to a Fact.
func (s *Service) Infers(
	ctx context.Context,
	premiseFactID string,
	conclusionFactID string,
) error {
	return s.typedLink(
		ctx,
		premiseFactID, NodeTypeFact,
		RelationInfers,
		conclusionFactID, NodeTypeFact,
	)
}

// Refines records a Decision -> Decision RelationRefines relationship
// from a broader decision (such as a plan) to a narrower one made
// within it. It returns UnexpectedNodeTypeError if either ID does not
// refer to a Decision.
func (s *Service) Refines(
	ctx context.Context,
	parentDecisionID string,
	childDecisionID string,
) error {
	return s.typedLink(
		ctx,
		parentDecisionID, NodeTypeDecision,
		RelationRefines,
		childDecisionID, NodeTypeDecision,
	)
}

func (s *Service) typedLink(
	ctx context.Context,
	fromID string,
	fromType NodeType,
	relation RelationType,
	toID string,
	toType NodeType,
) error {
	if err := s.requireNodeType(ctx, fromID, fromType); err != nil {
		return err
	}

	if err := s.requireNodeType(ctx, toID, toType); err != nil {
		return err
	}

	return s.Link(ctx, fromID, relation, toID)
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

// FindNodes lists nodes matching filter, for queries other than
// direct ID lookup or causal traversal, such as "every Decision
// recorded in the last hour." It returns graph.ErrStoreNotQueryable if
// the configured store does not implement graph.Queryable; the
// built-in in-memory store does.
func (s *Service) FindNodes(
	ctx context.Context,
	filter graph.NodeFilter,
) ([]graph.Node, error) {
	return s.graph.FindNodes(ctx, filter)
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

func (s *Service) applyTemporal(
	node *graph.Node,
) {
	node.RecordedAt = s.now().UTC()
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
