package graph

import "context"

// Store is the persistence interface a Graph delegates to. It must
// maintain enough indexing to answer OutgoingEdges and IncomingEdges
// without a full scan, and implementations intended for concurrent use
// must guard their own state.
//
// PutNode and PutEdge must reject duplicate IDs with an error wrapping
// ErrNodeAlreadyExists or ErrEdgeAlreadyExists, checked atomically with
// the write (e.g. a unique constraint): the semantic layer relies on
// this to make relationship recording idempotent under concurrent
// retries. GetNode, GetEdge, OutgoingEdges, and IncomingEdges should
// return ErrNodeNotFound when referencing a node that does not exist.
type Store interface {
	PutNode(ctx context.Context, node Node) error
	GetNode(ctx context.Context, id string) (Node, error)

	PutEdge(ctx context.Context, edge Edge) error
	GetEdge(ctx context.Context, id string) (Edge, error)

	OutgoingEdges(ctx context.Context, nodeID string) ([]Edge, error)
	IncomingEdges(ctx context.Context, nodeID string) ([]Edge, error)
}
