package graph

import "errors"

var (
	ErrEmptyNodeID       = errors.New("node ID cannot be empty")
	ErrEmptyNodeType     = errors.New("node type cannot be empty")
	ErrNodeNotFound      = errors.New("node not found")
	ErrNodeAlreadyExists = errors.New("node already exists")

	ErrEmptyEdgeID       = errors.New("edge ID cannot be empty")
	ErrEmptyEdgeType     = errors.New("edge type cannot be empty")
	ErrEdgeNotFound      = errors.New("edge not found")
	ErrEdgeAlreadyExists = errors.New("edge already exists")
)
