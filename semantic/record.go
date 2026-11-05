package semantic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	"github.com/karosia/ai-trace-cause/graph"
)

// RecordOption configures a single Record* call.
type RecordOption func(*recordConfig)

type recordConfig struct {
	causes []causeLink
}

// causeLink is a pending relationship from an existing node to the
// node being recorded. An empty relation is inferred from the Schema.
type causeLink struct {
	id       string
	relation RelationType
}

// WithCauses links the entity being recorded to the given causes in
// the same call, inferring each relationship from the Service's
// Schema and the cause's type: with DefaultSchema, an Observation to
// its Source(s) or Action(s), a Fact to its Observation(s) or premise
// Fact(s), a Decision to its Fact(s) or parent Decision(s), an Action
// to its Decision(s). Use WithCausesVia when the schema allows more
// than one relation between the two types.
//
// Every cause is checked before anything is written, and the entity
// and its relationships are written in one batch, atomically when the
// Store implements graph.AtomicBatchStore (the built-in memory store
// does). The relationships share the entity's RecordedAt, which is
// what makes the entity's own causes visible to
// TraceActionCauseAtActionTime. Duplicate causes are ignored.
func WithCauses(causeIDs ...string) RecordOption {
	return WithCausesVia("", causeIDs...)
}

// WithCausesVia is like WithCauses but names the relation explicitly.
// The Schema must allow it between each cause's type and the recorded
// entity's type.
func WithCausesVia(
	relation RelationType,
	causeIDs ...string,
) RecordOption {
	return func(cfg *recordConfig) {
		for _, id := range causeIDs {
			cfg.causes = append(cfg.causes, causeLink{
				id:       id,
				relation: relation,
			})
		}
	}
}

// RelationID returns the ID of the relation edge from -> to. IDs are
// derived from the relationship itself rather than generated, so
// recording the same relationship twice yields the same edge and the
// second call is a no-op. The result is stable across processes and
// versions.
func RelationID(
	relation RelationType,
	from string,
	to string,
) string {
	sum := sha256.Sum256(
		[]byte(string(relation) + "\x00" + from + "\x00" + to),
	)

	return "rel_" + hex.EncodeToString(sum[:16])
}

// record is the shared implementation of the Record* methods: it
// resolves the node's ID, validates its causes against the schema,
// assigns RecordedAt and telemetry, and writes the node together with
// its cause relationships.
func (s *Service) record(
	ctx context.Context,
	node graph.Node,
	options []RecordOption,
) (graph.Node, error) {
	var cfg recordConfig

	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}

	id, err := s.resolveID(node.ID)
	if err != nil {
		return graph.Node{}, err
	}

	node.ID = id

	pending := make([]pendingEdge, 0, len(cfg.causes))
	for _, cause := range cfg.causes {
		pending = append(pending, pendingEdge{
			from:     cause.id,
			to:       node.ID,
			relation: cause.relation,
		})
	}

	known := map[string]NodeType{node.ID: NodeType(node.Type)}

	edges, err := s.resolveEdges(ctx, pending, known)
	if err != nil {
		return graph.Node{}, err
	}

	s.applyTemporal(&node)
	s.applyTelemetry(ctx, &node)

	for i := range edges {
		edges[i].RecordedAt = node.RecordedAt
		edges[i].Telemetry = node.Telemetry
	}

	if err := s.graph.AddBatch(
		ctx,
		[]graph.Node{node},
		edges,
	); err != nil {
		return graph.Node{}, err
	}

	return node, nil
}

// pendingEdge is a relationship to be validated and created. An empty
// relation is inferred from the Schema.
type pendingEdge struct {
	from     string
	to       string
	relation RelationType
}

// resolveEdges validates pending relationships against the Schema and
// returns the edges to write, without RecordedAt or Telemetry. Node
// types are taken from known (nodes not yet written) or else looked up
// in the graph. Duplicates collapse into one edge.
func (s *Service) resolveEdges(
	ctx context.Context,
	pending []pendingEdge,
	known map[string]NodeType,
) ([]graph.Edge, error) {
	typeOf := func(id string) (NodeType, error) {
		if nodeType, ok := known[id]; ok {
			return nodeType, nil
		}

		node, err := s.graph.GetNode(ctx, id)
		if err != nil {
			return "", err
		}

		known[id] = NodeType(node.Type)

		return NodeType(node.Type), nil
	}

	edges := make([]graph.Edge, 0, len(pending))
	seen := make(map[string]struct{}, len(pending))

	for _, link := range pending {
		fromType, err := typeOf(link.from)
		if err != nil {
			return nil, err
		}

		toType, err := typeOf(link.to)
		if err != nil {
			return nil, err
		}

		relation := link.relation
		if relation == "" {
			relation, err = s.schema.relationBetween(fromType, toType)
			if err != nil {
				return nil, err
			}
		} else if !s.schema.Allows(fromType, relation, toType) {
			return nil, &RelationNotAllowedError{
				From:     fromType,
				Relation: relation,
				To:       toType,
			}
		}

		edgeID := RelationID(relation, link.from, link.to)
		if _, ok := seen[edgeID]; ok {
			continue
		}

		seen[edgeID] = struct{}{}

		edges = append(edges, graph.Edge{
			ID:   edgeID,
			From: link.from,
			To:   link.to,
			Type: string(relation),
		})
	}

	return edges, nil
}

// Link records a relationship of type relation from the node fromID
// (the cause) to the node toID (the effect). The Service's Schema must
// allow relation between the two nodes' types; otherwise Link returns
// a *RelationNotAllowedError. Recording the same relationship again is
// a no-op.
//
// Produced, Supports, BasisOf, Caused, ResultedIn, Infers, and Refines
// are typed shorthands for Link.
func (s *Service) Link(
	ctx context.Context,
	fromID string,
	relation RelationType,
	toID string,
) error {
	edges, err := s.resolveEdges(
		ctx,
		[]pendingEdge{{from: fromID, to: toID, relation: relation}},
		map[string]NodeType{},
	)
	if err != nil {
		return err
	}

	edge := edges[0]
	edge.RecordedAt = s.now().UTC()

	s.applyEdgeTelemetry(ctx, &edge)

	err = s.graph.AddEdge(ctx, edge)
	if errors.Is(err, graph.ErrEdgeAlreadyExists) {
		// Same relationship recorded before (e.g. an agent retry).
		// The original edge, and its RecordedAt, are kept.
		return nil
	}

	return err
}
