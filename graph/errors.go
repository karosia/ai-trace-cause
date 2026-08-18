package graph

import "errors"

// Sentinel errors returned by the graph package and its Store
// implementations.
var (
	// ErrNilStore is returned by New when given a nil Store.
	ErrNilStore = errors.New("store cannot be nil")

	// ErrEmptyNodeID is returned when a Node has an empty ID.
	ErrEmptyNodeID = errors.New("node ID cannot be empty")
	// ErrEmptyNodeType is returned when a Node has an empty Type.
	ErrEmptyNodeType = errors.New("node type cannot be empty")
	// ErrNodeNotFound is returned when a referenced node does not
	// exist in the store.
	ErrNodeNotFound = errors.New("node not found")
	// ErrNodeAlreadyExists is returned when adding a node whose ID
	// already exists in the store.
	ErrNodeAlreadyExists = errors.New("node already exists")

	// ErrEmptyEdgeID is returned when an Edge has an empty ID.
	ErrEmptyEdgeID = errors.New("edge ID cannot be empty")
	// ErrEmptyEdgeType is returned when an Edge has an empty Type.
	ErrEmptyEdgeType = errors.New("edge type cannot be empty")
	// ErrEdgeNotFound is returned when a referenced edge does not
	// exist in the store.
	ErrEdgeNotFound = errors.New("edge not found")
	// ErrEdgeAlreadyExists is returned when adding an edge whose ID
	// already exists in the store.
	ErrEdgeAlreadyExists = errors.New("edge already exists")

	// ErrInvalidDirection is returned when a traversal is given a
	// Direction other than DirectionOutgoing or DirectionIncoming.
	ErrInvalidDirection = errors.New("invalid traversal direction")
	// ErrInvalidMaxDepth is returned when a traversal is given a
	// negative maxDepth.
	ErrInvalidMaxDepth = errors.New("max depth cannot be negative")

	// ErrInvalidValidityInterval is returned when a node or edge's
	// ValidUntil does not come after its ValidFrom.
	ErrInvalidValidityInterval = errors.New(
		"valid until must be after valid from",
	)

	// ErrNodeNotVisibleAt is returned by BFSAt when the starting node
	// was not yet recorded, or was not valid, at the requested time.
	ErrNodeNotVisibleAt = errors.New(
		"node is not visible at the requested time",
	)

	// ErrStoreNotQueryable is returned by Graph.FindNodes when the
	// underlying Store does not implement Queryable.
	ErrStoreNotQueryable = errors.New(
		"store does not support querying",
	)
)
