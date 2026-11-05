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
// TraceActionCauseGraph does, and renders the resulting causal graph
// as a human-readable narrative (see ExplainGraph).
//
// It is intended for logging, debugging, and audit output where a
// plain-English answer to "why did this happen?" is more useful than
// a graph walked by hand. For programmatic access, use
// TraceActionCauseGraph. It returns the same errors as
// TraceActionCauseGraph.
func (s *Service) Explain(
	ctx context.Context,
	actionID string,
	maxDepth int,
) (string, error) {
	causal, err := s.TraceActionCauseGraph(
		ctx,
		actionID,
		maxDepth,
	)
	if err != nil {
		return "", err
	}

	return ExplainGraph(causal), nil
}

// ExplainGraph renders a CausalGraph as one sentence per causal
// relationship, joined by newlines. Every relationship in
// causal.Edges is rendered, so a node with several causes (or that
// causes several effects) appears in each relevant sentence. Retracted
// nodes are annotated with their retraction reason.
//
// Combine it with TraceActionCauseAtActionTime or
// TraceActionCauseGraphAt to explain an action as it was understood at
// a given time.
func ExplainGraph(causal CausalGraph) string {
	nodesByID := make(
		map[string]graph.Node,
		len(causal.Visits),
	)

	for _, visit := range causal.Visits {
		nodesByID[visit.Node.ID] = visit.Node
	}

	lines := make([]string, 0, len(causal.Edges))

	for _, edge := range causal.Edges {
		cause, ok := nodesByID[edge.From]
		if !ok {
			continue
		}

		effect, ok := nodesByID[edge.To]
		if !ok {
			continue
		}

		lines = append(
			lines,
			explainEdge(cause, effect, causal.Retractions),
		)
	}

	return strings.Join(lines, "\n")
}

// explainEdge renders a single causal step as a sentence, where cause
// is the node that occurred earlier in the chain and effect is the
// node it led to.
func explainEdge(
	cause graph.Node,
	effect graph.Node,
	retractions map[string]Retraction,
) string {
	verb, ok := causalVerbs[[2]NodeType{
		NodeType(cause.Type),
		NodeType(effect.Type),
	}]
	if !ok {
		verb = "led to"
	}

	return fmt.Sprintf(
		"%s %s %s.",
		describeWithRetraction(effect, retractions),
		verb,
		describeWithRetraction(cause, retractions),
	)
}

func describeWithRetraction(
	node graph.Node,
	retractions map[string]Retraction,
) string {
	description := describeNode(node)

	if retraction, ok := retractions[node.ID]; ok {
		description += fmt.Sprintf(
			" [retracted: %s]",
			retraction.Reason,
		)
	}

	return description
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
