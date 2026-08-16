package graph

import (
	"context"
	"sort"
)

type Graph struct {
	store Store
}

func New(store Store) (*Graph, error) {
	if store == nil {
		return nil, ErrNilStore
	}

	return &Graph{
		store: store,
	}, nil
}

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

func (g *Graph) GetNode(
	ctx context.Context,
	id string,
) (Node, error) {
	return g.store.GetNode(ctx, id)
}

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

func (g *Graph) GetEdge(
	ctx context.Context,
	id string,
) (Edge, error) {
	return g.store.GetEdge(ctx, id)
}

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
