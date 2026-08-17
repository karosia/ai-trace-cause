package semantic

import (
	"context"
	"fmt"
	"strings"

	"github.com/karosia/ai-trace-cause/graph"
)

// causalVerbs describes, for each fixed (cause, effect) node type
// pair enforced by Produced, Supports, BasisOf, and Caused, how the
// effect relates back to its cause in a rendered sentence.
var causalVerbs = map[[2]NodeType]string{
	{NodeTypeSource, NodeTypeObservation}: "was produced by",
	{NodeTypeObservation, NodeTypeFact}:   "was supported by",
	{NodeTypeFact, NodeTypeDecision}:      "was based on",
	{NodeTypeDecision, NodeTypeAction}:    "was caused by",
}

// Explain walks backward from the Action identified by actionID, as
// TraceActionCause does, and renders the resulting causal chain as a
// human-readable narrative: one sentence per causal edge, ordered from
// the action back to its originating source(s), joined by newlines.
//
// It is intended for logging, debugging, and audit output where a
// plain-English answer to "why did this happen?" is more useful than
// a []graph.Visit walked by hand. For programmatic access to the
// chain itself, use TraceActionCause. It returns the same errors as
// TraceActionCause.
func (s *Service) Explain(
	ctx context.Context,
	actionID string,
	maxDepth int,
) (string, error) {
	visits, err := s.TraceActionCause(
		ctx,
		actionID,
		maxDepth,
	)
	if err != nil {
		return "", err
	}

	return renderExplanation(visits), nil
}

// renderExplanation turns a set of causally-linked visits into a
// narrative, one sentence per parent/child edge. The root visit (with
// no parent) contributes no sentence on its own; it appears as the
// effect side of its child's sentence instead.
func renderExplanation(visits []graph.Visit) string {
	nodesByID := make(
		map[string]graph.Node,
		len(visits),
	)

	for _, visit := range visits {
		nodesByID[visit.Node.ID] = visit.Node
	}

	lines := make([]string, 0, len(visits))

	for _, visit := range visits {
		if visit.ParentNodeID == "" {
			continue
		}

		parent, ok := nodesByID[visit.ParentNodeID]
		if !ok {
			continue
		}

		lines = append(
			lines,
			explainEdge(visit.Node, parent),
		)
	}

	return strings.Join(lines, "\n")
}

// explainEdge renders a single causal step as a sentence, where cause
// is the node that occurred earlier in the chain and effect is the
// node it led to.
func explainEdge(cause, effect graph.Node) string {
	verb, ok := causalVerbs[[2]NodeType{
		NodeType(cause.Type),
		NodeType(effect.Type),
	}]
	if !ok {
		verb = "led to"
	}

	return fmt.Sprintf(
		"%s %s %s.",
		describeNode(effect),
		verb,
		describeNode(cause),
	)
}

// describeNode renders a short, human-readable label for a node using
// the properties recorded for its NodeType.
func describeNode(node graph.Node) string {
	switch NodeType(node.Type) {

	case NodeTypeSource:
		kind, _ := node.Properties["kind"].(string)
		uri, _ := node.Properties["uri"].(string)

		if uri != "" {
			return fmt.Sprintf(
				"Source %q (%s)",
				kind,
				uri,
			)
		}

		return fmt.Sprintf("Source %q", kind)

	case NodeTypeObservation:
		name, _ := node.Properties["name"].(string)
		value := node.Properties["value"]

		return fmt.Sprintf(
			"Observation %q (value=%v)",
			name,
			value,
		)

	case NodeTypeFact:
		statement, _ := node.Properties["statement"].(string)

		return fmt.Sprintf("Fact %q", statement)

	case NodeTypeDecision:
		outcome, _ := node.Properties["outcome"].(string)

		return fmt.Sprintf("Decision %q", outcome)

	case NodeTypeAction:
		name, _ := node.Properties["name"].(string)
		target, _ := node.Properties["target"].(string)

		if target != "" {
			return fmt.Sprintf(
				"Action %q (target=%s)",
				name,
				target,
			)
		}

		return fmt.Sprintf("Action %q", name)

	default:
		return fmt.Sprintf(
			"%s %q",
			node.Type,
			node.ID,
		)
	}
}
