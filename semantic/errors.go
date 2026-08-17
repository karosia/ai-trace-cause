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

	// ErrIDGeneratorUnavailable is returned when an entity or
	// relationship needs a generated ID but no IDGenerator is
	// configured.
	ErrIDGeneratorUnavailable = errors.New(
		"ID generator is not configured",
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
