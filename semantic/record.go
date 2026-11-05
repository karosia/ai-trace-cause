package semantic

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/karosia/ai-trace-cause/graph"
)

// causeTypeOf maps each node type to the node type that may directly
// cause it, and relationInto to the relationship that connects them.
// Together they encode the chain enforced by Produced, Supports,
// BasisOf, and Caused.
var (
	causeTypeOf = map[NodeType]NodeType{
		NodeTypeObservation: NodeTypeSource,
		NodeTypeFact:        NodeTypeObservation,
		NodeTypeDecision:    NodeTypeFact,
		NodeTypeAction:      NodeTypeDecision,
	}

	relationInto = map[NodeType]RelationType{
		NodeTypeObservation: RelationProduced,
		NodeTypeFact:        RelationSupports,
		NodeTypeDecision:    RelationBasisOf,
		NodeTypeAction:      RelationCaused,
	}
)

// RecordOption configures a single Record* call.
type RecordOption func(*recordConfig)

type recordConfig struct {
	causes []string
}

// WithCauses links the entity being recorded to the given causes in
// the same call: an Observation to its Source(s), a Fact to its
// Observation(s), a Decision to its Fact(s), or an Action to its
// Decision(s). The relationship type is implied by the entity type.
//
// Every cause is checked before anything is written, so a missing or
// wrongly-typed cause leaves the graph unchanged. The relationships
// share the entity's RecordedAt, which is what makes the entity's own
// causes visible to TraceActionCauseAtActionTime; relationships added
// later with Produced, Supports, BasisOf, or Caused are, by
// definition, recorded after the entity. Duplicate IDs are ignored.
func WithCauses(causeIDs ...string) RecordOption {
	return func(cfg *recordConfig) {
		cfg.causes = append(cfg.causes, causeIDs...)
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
// validates causes, assigns ID, timestamps, and telemetry, stores the
// node, and links its causes.
func (s *Service) record(
	ctx context.Context,
	id string,
	nodeType NodeType,
	properties map[string]any,
	validity Validity,
	options []RecordOption,
) (graph.Node, error) {
	var cfg recordConfig

	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}

	causes := uniqueIDs(cfg.causes)

	if len(causes) > 0 {
		causeType, ok := causeTypeOf[nodeType]
		if !ok {
			return graph.Node{}, fmt.Errorf(
				"%s nodes cannot have causes",
				nodeType,
			)
		}

		for _, causeID := range causes {
			if err := s.requireNodeType(
				ctx,
				causeID,
				causeType,
			); err != nil {
				return graph.Node{}, err
			}
		}
	}

	resolvedID, err := s.resolveID(id)
	if err != nil {
		return graph.Node{}, err
	}

	node := graph.Node{
		ID:         resolvedID,
		Type:       string(nodeType),
		Properties: properties,
	}

	s.applyTemporal(&node, validity)
	s.applyTelemetry(ctx, &node)

	if err := s.graph.AddNode(ctx, node); err != nil {
		return graph.Node{}, err
	}

	if len(causes) == 0 {
		return node, nil
	}

	relation := relationInto[nodeType]
	edges := make([]graph.Edge, 0, len(causes))

	for _, causeID := range causes {
		edges = append(edges, graph.Edge{
			ID:         RelationID(relation, causeID, node.ID),
			From:       causeID,
			To:         node.ID,
			Type:       string(relation),
			RecordedAt: node.RecordedAt,
			Telemetry:  node.Telemetry,
		})
	}

	if err := s.graph.AddEdges(ctx, edges); err != nil {
		// Causes were validated above, so this means the store
		// itself failed (or a cause was concurrently removed by a
		// custom store). The node is already recorded; say so
		// rather than leave the caller guessing.
		return graph.Node{}, fmt.Errorf(
			"%s %q was recorded but linking its causes failed: %w",
			nodeType,
			node.ID,
			err,
		)
	}

	return node, nil
}

func uniqueIDs(ids []string) []string {
	if len(ids) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(ids))
	unique := make([]string, 0, len(ids))

	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}

		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	return unique
}
