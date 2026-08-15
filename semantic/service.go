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
