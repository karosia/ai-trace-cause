package semantic

import (
	"github.com/karosia/ai-trace-cause/graph"
)

// The toNode methods validate an entity and convert it to an unsaved
// graph.Node. The ID may still be empty (to be generated), and
// RecordedAt and Telemetry are assigned when the node is written. They
// are shared by the Record* methods and Step, so both paths enforce
// identical rules.

func (source Source) toNode() (graph.Node, error) {
	if source.Kind == "" {
		return graph.Node{}, ErrEmptySourceKind
	}

	properties := map[string]any{
		"kind": source.Kind,
		"uri":  source.URI,
	}
	setIfNotNil(properties, "metadata", source.Metadata)

	return newNode(
		source.ID,
		NodeTypeSource,
		properties,
		source.RunID,
		source.StepID,
		source.Validity,
	), nil
}

func (observation Observation) toNode() (graph.Node, error) {
	if observation.Name == "" {
		return graph.Node{}, ErrEmptyObservationName
	}

	properties := map[string]any{
		"name":  observation.Name,
		"value": observation.Value,
	}
	setIfNotNil(properties, "metadata", observation.Metadata)

	return newNode(
		observation.ID,
		NodeTypeObservation,
		properties,
		observation.RunID,
		observation.StepID,
		observation.Validity,
	), nil
}

func (fact Fact) toNode() (graph.Node, error) {
	if fact.Statement == "" {
		return graph.Node{}, ErrEmptyFactStatement
	}

	if fact.Confidence < 0 || fact.Confidence > 1 {
		return graph.Node{}, ErrInvalidConfidence
	}

	properties := map[string]any{
		"statement":  fact.Statement,
		"confidence": fact.Confidence,
	}
	setIfNotNil(properties, "metadata", fact.Metadata)

	return newNode(
		fact.ID,
		NodeTypeFact,
		properties,
		fact.RunID,
		fact.StepID,
		fact.Validity,
	), nil
}

func (decision Decision) toNode() (graph.Node, error) {
	if decision.Outcome == "" {
		return graph.Node{}, ErrEmptyDecisionOutcome
	}

	if decision.Confidence < 0 || decision.Confidence > 1 {
		return graph.Node{}, ErrInvalidConfidence
	}

	properties := map[string]any{
		"outcome":    decision.Outcome,
		"rationale":  decision.Rationale,
		"confidence": decision.Confidence,
	}
	setIfNotNil(properties, "metadata", decision.Metadata)

	return newNode(
		decision.ID,
		NodeTypeDecision,
		properties,
		decision.RunID,
		decision.StepID,
		decision.Validity,
	), nil
}

func (action Action) toNode() (graph.Node, error) {
	if action.Name == "" {
		return graph.Node{}, ErrEmptyActionName
	}

	properties := map[string]any{
		"name":   action.Name,
		"target": action.Target,
	}
	setIfNotNil(properties, "parameters", action.Parameters)
	setIfNotNil(properties, "metadata", action.Metadata)

	return newNode(
		action.ID,
		NodeTypeAction,
		properties,
		action.RunID,
		action.StepID,
		action.Validity,
	), nil
}

func newNode(
	id string,
	nodeType NodeType,
	properties map[string]any,
	runID string,
	stepID string,
	validity Validity,
) graph.Node {
	if runID != "" {
		properties[PropertyRunID] = runID
	}

	if stepID != "" {
		properties[PropertyStepID] = stepID
	}

	return graph.Node{
		ID:         id,
		Type:       string(nodeType),
		Properties: properties,
		ValidFrom:  cloneTime(validity.ValidFrom),
		ValidUntil: cloneTime(validity.ValidUntil),
	}
}

func setIfNotNil(
	properties map[string]any,
	key string,
	value map[string]any,
) {
	if value != nil {
		properties[key] = value
	}
}
