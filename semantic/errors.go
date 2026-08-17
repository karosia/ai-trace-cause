package semantic

import (
	"errors"
	"fmt"
)

var (
	ErrNilGraph = errors.New("graph cannot be nil")

	ErrEmptyObservationName = errors.New(
		"observation name cannot be empty",
	)

	ErrEmptyFactStatement = errors.New(
		"fact statement cannot be empty",
	)

	ErrEmptyDecisionOutcome = errors.New(
		"decision outcome cannot be empty",
	)
	ErrEmptyActionName = errors.New(
		"action name cannot be empty",
	)

	ErrInvalidConfidence = errors.New(
		"confidence must be between 0 and 1",
	)

	ErrUnexpectedNodeType = errors.New(
		"unexpected node type",
	)

	ErrEmptySourceKind = errors.New(
		"source kind cannot be empty",
	)

	ErrIDGeneratorUnavailable = errors.New(
		"ID generator is not configured",
	)
)

type UnexpectedNodeTypeError struct {
	NodeID   string
	Expected NodeType
	Actual   string
}

func (e *UnexpectedNodeTypeError) Error() string {
	return fmt.Sprintf(
		"%v: node %q has type %q, expected %q",
		ErrUnexpectedNodeType,
		e.NodeID,
		e.Actual,
		e.Expected,
	)
}

func (e *UnexpectedNodeTypeError) Unwrap() error {
	return ErrUnexpectedNodeType
}
