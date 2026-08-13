package graph

import "context"

type Store interface {
	PutNode(ctx context.Context, node Node) error
	GetNode(ctx context.Context, id string) (Node, error)

	PutEdge(ctx context.Context, edge Edge) error
	GetEdge(ctx context.Context, id string) (Edge, error)

	OutgoingEdges(ctx context.Context, nodeID string) ([]Edge, error)
	IncomingEdges(ctx context.Context, nodeID string) ([]Edge, error)
}
