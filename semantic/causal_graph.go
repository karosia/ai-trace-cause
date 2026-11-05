package semantic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
)

// Retract records that the node identified by retraction.NodeID
// should no longer be relied upon. Nothing is modified or deleted: the
// retraction is stored as a RETRACTED annotation alongside the node,
// so a trace as of an earlier time still shows what was believed then,
// while a current trace marks the node as retracted.
//
// It returns ErrEmptyRetractionNodeID or ErrEmptyRetractionReason for
// missing fields, graph.ErrNodeNotFound if the node does not exist,
// and ErrAlreadyRetracted if the node already carries a different
// retraction. Repeating an identical retraction (or one with a zero
// EffectiveAt) returns the existing retraction and no error, so
// retries are safe.
func (s *Service) Retract(
	ctx context.Context,
	retraction Retraction,
) (Retraction, error) {
	if retraction.NodeID == "" {
		return Retraction{}, ErrEmptyRetractionNodeID
	}

	if retraction.Reason == "" {
		return Retraction{}, ErrEmptyRetractionReason
	}

	if _, err := s.graph.GetNode(
		ctx,
		retraction.NodeID,
	); err != nil {
		return Retraction{}, err
	}

	recordedAt := s.now().UTC()

	effectiveAt := retraction.EffectiveAt.UTC()
	if retraction.EffectiveAt.IsZero() {
		effectiveAt = recordedAt
	}

	edge := graph.Edge{
		ID: RelationID(
			RelationRetracted,
			retraction.NodeID,
			retraction.NodeID,
		),
		From: retraction.NodeID,
		To:   retraction.NodeID,
		Type: string(RelationRetracted),
		Properties: map[string]any{
			"reason": retraction.Reason,
		},
		RecordedAt: recordedAt,
		ValidFrom:  &effectiveAt,
	}

	s.applyEdgeTelemetry(ctx, &edge)

	err := s.graph.AddEdge(ctx, edge)
	if errors.Is(err, graph.ErrEdgeAlreadyExists) {
		existingEdge, getErr := s.graph.GetEdge(ctx, edge.ID)
		if getErr != nil {
			return Retraction{}, getErr
		}

		existing := retractionFromEdge(existingEdge)

		if existing.Reason == retraction.Reason &&
			(retraction.EffectiveAt.IsZero() ||
				existing.EffectiveAt.Equal(effectiveAt)) {
			return existing, nil
		}

		return existing, fmt.Errorf(
			"%w: %s",
			ErrAlreadyRetracted,
			retraction.NodeID,
		)
	}

	if err != nil {
		return Retraction{}, err
	}

	return Retraction{
		NodeID:      retraction.NodeID,
		Reason:      retraction.Reason,
		EffectiveAt: effectiveAt,
		RecordedAt:  recordedAt,
	}, nil
}

// TraceActionCauseGraph walks backward from the Action identified by
// actionID, like TraceActionCause, but returns every causal
// relationship between the reached nodes, plus all retractions
// recorded so far for those nodes. It returns UnexpectedNodeTypeError
// if actionID does not refer to an Action node.
func (s *Service) TraceActionCauseGraph(
	ctx context.Context,
	actionID string,
	maxDepth int,
) (CausalGraph, error) {
	if err := s.requireNodeType(
		ctx,
		actionID,
		NodeTypeAction,
	); err != nil {
		return CausalGraph{}, err
	}

	sub, err := s.graph.Subgraph(
		ctx,
		actionID,
		graph.DirectionIncoming,
		maxDepth,
	)
	if err != nil {
		return CausalGraph{}, err
	}

	return newCausalGraph(sub), nil
}

// TraceActionCauseGraphAt is like TraceActionCauseGraph but
// reconstructs the graph as it was known at at: only nodes and
// relationships recorded by at and valid at at are included, and a
// node is marked retracted only if its retraction was both recorded
// and in effect by at.
func (s *Service) TraceActionCauseGraphAt(
	ctx context.Context,
	actionID string,
	at time.Time,
	maxDepth int,
) (CausalGraph, error) {
	if err := s.requireNodeType(
		ctx,
		actionID,
		NodeTypeAction,
	); err != nil {
		return CausalGraph{}, err
	}

	return s.causalGraphAt(ctx, actionID, at, maxDepth)
}

// TraceActionCauseAtActionTime reconstructs what was known when the
// action was recorded, by calling TraceActionCauseGraphAt with the
// action's own RecordedAt. This answers "what did the agent know when
// it acted?" without the caller having to track timestamps.
//
// Only relationships recorded no later than the action are visible,
// so the action's decisions must be linked with WithCauses when the
// action is recorded. A Caused call made afterwards is, correctly,
// knowledge acquired after the action.
func (s *Service) TraceActionCauseAtActionTime(
	ctx context.Context,
	actionID string,
	maxDepth int,
) (CausalGraph, error) {
	action, err := s.graph.GetNode(ctx, actionID)
	if err != nil {
		return CausalGraph{}, err
	}

	if action.Type != string(NodeTypeAction) {
		return CausalGraph{}, &UnexpectedNodeTypeError{
			NodeID:   actionID,
			Expected: NodeTypeAction,
			Actual:   action.Type,
		}
	}

	return s.causalGraphAt(
		ctx,
		actionID,
		action.RecordedAt,
		maxDepth,
	)
}

func (s *Service) causalGraphAt(
	ctx context.Context,
	actionID string,
	at time.Time,
	maxDepth int,
) (CausalGraph, error) {
	sub, err := s.graph.SubgraphAt(
		ctx,
		actionID,
		graph.DirectionIncoming,
		maxDepth,
		at,
	)
	if err != nil {
		return CausalGraph{}, err
	}

	return newCausalGraph(sub), nil
}

// newCausalGraph separates retraction annotations from causal
// relationships.
func newCausalGraph(sub graph.Subgraph) CausalGraph {
	causal := CausalGraph{
		Visits: sub.Visits,
		Edges:  make([]graph.Edge, 0, len(sub.Edges)),
	}

	for _, edge := range sub.Edges {
		if edge.Type == string(RelationRetracted) &&
			edge.From == edge.To {
			if causal.Retractions == nil {
				causal.Retractions = make(map[string]Retraction)
			}

			causal.Retractions[edge.From] = retractionFromEdge(edge)

			continue
		}

		causal.Edges = append(causal.Edges, edge)
	}

	return causal
}

func retractionFromEdge(edge graph.Edge) Retraction {
	reason, _ := edge.Properties["reason"].(string)

	retraction := Retraction{
		NodeID:     edge.From,
		Reason:     reason,
		RecordedAt: edge.RecordedAt,
	}

	if edge.ValidFrom != nil {
		retraction.EffectiveAt = *edge.ValidFrom
	}

	return retraction
}

// TraceEffectsGraph walks forward from any node through everything it
// went on to influence, up to maxDepth hops, and returns the reached
// nodes, every causal relationship between them, and their
// retractions. It generalizes TraceSourceEffects to any starting
// point, answering for example "a fact was retracted: which decisions
// and actions relied on it?". It returns graph.ErrNodeNotFound if
// nodeID does not exist.
func (s *Service) TraceEffectsGraph(
	ctx context.Context,
	nodeID string,
	maxDepth int,
) (CausalGraph, error) {
	sub, err := s.graph.Subgraph(
		ctx,
		nodeID,
		graph.DirectionOutgoing,
		maxDepth,
	)
	if err != nil {
		return CausalGraph{}, err
	}

	return newCausalGraph(sub), nil
}
