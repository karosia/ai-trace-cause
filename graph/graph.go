// Package graph implements a generic, storage-agnostic directed graph
// engine: nodes, edges, traversal (BFS/DFS), and temporal visibility.
//
// The graph layer intentionally has no knowledge of AI-specific
// semantics; see the semantic package for the domain model built on
// top of it.
package graph

import (
	"context"
	"sort"
)

// Graph is a directed graph backed by a pluggable Store. It validates
// nodes and edges before delegating persistence to the store.
type Graph struct {
	store Store
}

// New creates a Graph backed by store. It returns ErrNilStore if store
// is nil.
func New(store Store) (*Graph, error) {
	if store == nil {
		return nil, ErrNilStore
	}

	return &Graph{
		store: store,
	}, nil
}

// AddNode validates and adds node to the graph. It returns
// ErrEmptyNodeID or ErrEmptyNodeType if the corresponding field is
// empty, or ErrInvalidValidityInterval if node's validity interval is
// malformed.
func (g *Graph) AddNode(
	ctx context.Context,
	node Node,
) error {
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

	return g.store.PutNode(ctx, node)
}

// GetNode returns the node with the given id, or ErrNodeNotFound if no
// such node exists.
func (g *Graph) GetNode(
	ctx context.Context,
	id string,
) (Node, error) {
	return g.store.GetNode(ctx, id)
}

// AddEdge validates and adds edge to the graph. It returns
// ErrEmptyEdgeID or ErrEmptyEdgeType if the corresponding field is
// empty, or ErrInvalidValidityInterval if edge's validity interval is
// malformed.
func (g *Graph) AddEdge(
	ctx context.Context,
	edge Edge,
) error {
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

	return g.store.PutEdge(ctx, edge)
}

// GetEdge returns the edge with the given id, or ErrEdgeNotFound if no
// such edge exists.
func (g *Graph) GetEdge(
	ctx context.Context,
	id string,
) (Edge, error) {
	return g.store.GetEdge(ctx, id)
}

// OutgoingNeighbors returns the nodes reachable from nodeID by a
// single outgoing edge, sorted by node ID.
func (g *Graph) OutgoingNeighbors(
	ctx context.Context,
	nodeID string,
) ([]Node, error) {
	edges, err := g.store.OutgoingEdges(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	neighbors := make([]Node, 0, len(edges))

	for _, edge := range edges {
		node, err := g.store.GetNode(ctx, edge.To)
		if err != nil {
			return nil, err
		}

		neighbors = append(neighbors, node)
	}

	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].ID < neighbors[j].ID
	})

	return neighbors, nil
}

// IncomingNeighbors returns the nodes that reach nodeID by a single
// outgoing edge, sorted by node ID.
func (g *Graph) IncomingNeighbors(
	ctx context.Context,
	nodeID string,
) ([]Node, error) {
	edges, err := g.store.IncomingEdges(ctx, nodeID)
	if err != nil {
		return nil, err
	}

	neighbors := make([]Node, 0, len(edges))

	for _, edge := range edges {
		node, err := g.store.GetNode(ctx, edge.From)
		if err != nil {
			return nil, err
		}

		neighbors = append(neighbors, node)
	}

	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].ID < neighbors[j].ID
	})

	return neighbors, nil
}
