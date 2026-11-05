package semantic

import (
	"errors"
	"fmt"
)

// Sentinel errors returned by Service.
var (
	// ErrNilGraph is returned by NewService when given a nil graph.
	ErrNilGraph = errors.New("graph cannot be nil")

	// ErrEmptyObservationName is returned when recording an
	// Observation with an empty Name.
	ErrEmptyObservationName = errors.New(
		"observation name cannot be empty",
	)

	// ErrEmptyFactStatement is returned when recording a Fact with
	// an empty Statement.
	ErrEmptyFactStatement = errors.New(
		"fact statement cannot be empty",
	)

	// ErrEmptyDecisionOutcome is returned when recording a Decision
	// with an empty Outcome.
	ErrEmptyDecisionOutcome = errors.New(
		"decision outcome cannot be empty",
	)
	// ErrEmptyActionName is returned when recording an Action with
	// an empty Name.
	ErrEmptyActionName = errors.New(
		"action name cannot be empty",
	)

	// ErrInvalidConfidence is returned when a Fact or Decision's
	// Confidence is outside [0, 1].
	ErrInvalidConfidence = errors.New(
		"confidence must be between 0 and 1",
	)

	// ErrUnexpectedNodeType is the sentinel wrapped by
	// UnexpectedNodeTypeError; use errors.Is to detect it.
	ErrUnexpectedNodeType = errors.New(
		"unexpected node type",
	)

	// ErrEmptySourceKind is returned when recording a Source with an
	// empty Kind.
	ErrEmptySourceKind = errors.New(
		"source kind cannot be empty",
	)

	// ErrIDGeneratorUnavailable is returned when an entity needs a
	// generated ID but no IDGenerator is configured.
	ErrIDGeneratorUnavailable = errors.New(
		"ID generator is not configured",
	)

	// ErrEmptyRetractionNodeID is returned when recording a
	// Retraction with an empty NodeID.
	ErrEmptyRetractionNodeID = errors.New(
		"retraction node ID cannot be empty",
	)

	// ErrEmptyRetractionReason is returned when recording a
	// Retraction with an empty Reason.
	ErrEmptyRetractionReason = errors.New(
		"retraction reason cannot be empty",
	)

	// ErrRelationNotAllowed is the sentinel wrapped by
	// RelationNotAllowedError; use errors.Is to detect it.
	ErrRelationNotAllowed = errors.New(
		"relation not allowed by schema",
	)

	// ErrAmbiguousRelation is returned when a cause is linked without
	// naming the relation and the schema allows more than one
	// relation between the two node types.
	ErrAmbiguousRelation = errors.New(
		"ambiguous relation",
	)

	// ErrEmptyRunID is returned when committing a Step of a Run that
	// was created with an empty ID.
	ErrEmptyRunID = errors.New(
		"run ID cannot be empty",
	)

	// ErrStepCommitted is returned when adding to, or committing, a
	// Step that was already committed.
	ErrStepCommitted = errors.New(
		"step already committed",
	)

	// ErrAlreadyRetracted is returned when retracting a node that
	// already has a retraction with a different Reason or
	// EffectiveAt. Retrying an identical retraction is not an error.
	ErrAlreadyRetracted = errors.New(
		"node is already retracted",
	)
)

// UnexpectedNodeTypeError is returned when an operation expects a node
// of type Expected but finds one of type Actual. It wraps
// ErrUnexpectedNodeType.
type UnexpectedNodeTypeError struct {
	NodeID   string
	Expected NodeType
	Actual   string
}

// Error implements the error interface.
func (e *UnexpectedNodeTypeError) Error() string {
	return fmt.Sprintf(
		"%v: node %q has type %q, expected %q",
		ErrUnexpectedNodeType,
		e.NodeID,
		e.Actual,
		e.Expected,
	)
}

// Unwrap returns ErrUnexpectedNodeType, so errors.Is(err,
// ErrUnexpectedNodeType) works on an UnexpectedNodeTypeError.
func (e *UnexpectedNodeTypeError) Unwrap() error {
	return ErrUnexpectedNodeType
}

// RelationNotAllowedError is returned when the Service's Schema does
// not allow a relationship between two node types. Relation is empty
// when no relation was named and none could be inferred. It wraps
// ErrRelationNotAllowed.
type RelationNotAllowedError struct {
	From     NodeType
	Relation RelationType
	To       NodeType
}

// Error implements the error interface.
func (e *RelationNotAllowedError) Error() string {
	if e.Relation == "" {
		return fmt.Sprintf(
			"%v: no relation from %s to %s",
			ErrRelationNotAllowed,
			e.From,
			e.To,
		)
	}

	return fmt.Sprintf(
		"%v: %s --%s--> %s",
		ErrRelationNotAllowed,
		e.From,
		e.Relation,
		e.To,
	)
}

// Unwrap returns ErrRelationNotAllowed.
func (e *RelationNotAllowedError) Unwrap() error {
	return ErrRelationNotAllowed
}
