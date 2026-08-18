package graph

import "context"

// NodeBatchStore is an optional capability a Store may implement to
// persist multiple nodes in a single call, e.g. inside one database
// transaction instead of one round trip per node. Graph.AddNodes uses
// it when the underlying Store implements it, and falls back to
// sequential AddNode calls otherwise.
type NodeBatchStore interface {
	PutNodes(ctx context.Context, nodes []Node) error
}

// EdgeBatchStore is the edge counterpart to NodeBatchStore. Graph.AddEdges
// uses it when the underlying Store implements it, and falls back to
// sequential AddEdge calls otherwise.
type EdgeBatchStore interface {
	PutEdges(ctx context.Context, edges []Edge) error
}

// AddNodes validates and adds nodes to the graph. If the underlying
// Store implements NodeBatchStore, all nodes are persisted through a
// single PutNodes call; otherwise AddNodes falls back to calling
// AddNode for each node in order and stops at the first error. Either
// way, every node is validated up front, so a validation failure never
// leaves a partially-applied batch behind — whether a batch that fails
// inside the store itself is all-or-nothing depends on the Store
// implementation.
func (g *Graph) AddNodes(
	ctx context.Context,
	nodes []Node,
) error {
	for _, node := range nodes {
		if node.ID == "" {
			return ErrEmptyNodeID
		}

		if node.Type == "" {
			return ErrEmptyNodeType
		}

		if err := validateValidityInterval(
			node.ValidFrom,
			node.ValidUntil,
		); err != nil {
			return err
		}
	}

	if batch, ok := g.store.(NodeBatchStore); ok {
		return batch.PutNodes(ctx, nodes)
	}

	for _, node := range nodes {
		if err := g.store.PutNode(ctx, node); err != nil {
			return err
		}
	}

	return nil
}

// AddEdges validates and adds edges to the graph, using EdgeBatchStore
// when the underlying Store implements it and falling back to
// sequential AddEdge calls otherwise. See AddNodes for the atomicity
// caveat.
func (g *Graph) AddEdges(
	ctx context.Context,
	edges []Edge,
) error {
	for _, edge := range edges {
		if edge.ID == "" {
			return ErrEmptyEdgeID
		}

		if edge.Type == "" {
			return ErrEmptyEdgeType
		}

		if err := validateValidityInterval(
			edge.ValidFrom,
			edge.ValidUntil,
		); err != nil {
			return err
		}
	}

	if batch, ok := g.store.(EdgeBatchStore); ok {
		return batch.PutEdges(ctx, edges)
	}

	for _, edge := range edges {
		if err := g.store.PutEdge(ctx, edge); err != nil {
			return err
		}
	}

	return nil
}
