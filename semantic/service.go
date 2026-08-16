package semantic

import (
	"context"

	"github.com/karosia/ai-trace-cause/graph"
)

type Service struct {
	graph *graph.Graph
}

func NewService(g *graph.Graph) (*Service, error) {
	if g == nil {
		return nil, ErrNilGraph
	}

	return &Service{
		graph: g,
	}, nil
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

	return s.graph.AddEdge(
		ctx,
		graph.Edge{
			ID:   edgeID,
			From: sourceID,
			To:   observationID,
			Type: string(RelationProduced),
		},
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

	return s.graph.AddEdge(
		ctx,
		graph.Edge{
			ID:   edgeID,
			From: observationID,
			To:   factID,
			Type: string(RelationSupports),
		},
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

	return s.graph.AddEdge(
		ctx,
		graph.Edge{
			ID:   edgeID,
			From: factID,
			To:   decisionID,
			Type: string(RelationBasisOf),
		},
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

	return s.graph.AddEdge(
		ctx,
		graph.Edge{
			ID:   edgeID,
			From: decisionID,
			To:   actionID,
			Type: string(RelationCaused),
		},
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
