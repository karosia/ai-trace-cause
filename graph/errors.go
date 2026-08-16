package graph

import "errors"

var (
	ErrNilStore = errors.New("store cannot be nil")

	ErrEmptyNodeID       = errors.New("node ID cannot be empty")
	ErrEmptyNodeType     = errors.New("node type cannot be empty")
	ErrNodeNotFound      = errors.New("node not found")
	ErrNodeAlreadyExists = errors.New("node already exists")

	ErrEmptyEdgeID       = errors.New("edge ID cannot be empty")
	ErrEmptyEdgeType     = errors.New("edge type cannot be empty")
	ErrEdgeNotFound      = errors.New("edge not found")
	ErrEdgeAlreadyExists = errors.New("edge already exists")

	ErrInvalidDirection = errors.New("invalid traversal direction")
	ErrInvalidMaxDepth  = errors.New("max depth cannot be negative")

	ErrInvalidValidityInterval = errors.New(
		"valid until must be after valid from",
	)

	ErrNodeNotVisibleAt = errors.New(
		"node is not visible at the requested time",
	)
)
