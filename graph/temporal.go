package graph

import (
	"context"
	"sort"
	"time"
)

func validateValidityInterval(
	validFrom *time.Time,
	validUntil *time.Time,
) error {
	if validFrom == nil || validUntil == nil {
		return nil
	}

	if !validUntil.After(*validFrom) {
		return ErrInvalidValidityInterval
	}

	return nil
}

func nodeVisibleAt(
	node Node,
	at time.Time,
) bool {
	if !node.RecordedAt.IsZero() &&
		node.RecordedAt.After(at) {
		return false
	}

	return validAt(
		node.ValidFrom,
		node.ValidUntil,
		at,
	)
}

func edgeVisibleAt(
	edge Edge,
	at time.Time,
) bool {
	if !edge.RecordedAt.IsZero() &&
		edge.RecordedAt.After(at) {
		return false
	}

	return validAt(
		edge.ValidFrom,
		edge.ValidUntil,
		at,
	)
}

func validAt(
	validFrom *time.Time,
	validUntil *time.Time,
	at time.Time,
) bool {
	if validFrom != nil &&
		at.Before(*validFrom) {
		return false
	}

	if validUntil != nil &&
		!at.Before(*validUntil) {
		return false
	}

	return true
}

// BFSAt is like BFS, but restricts the traversal to nodes and edges
// that were recorded and valid at the given time at, reconstructing
// the graph as it was known at that point in time. It returns
// ErrNodeNotVisibleAt if the start node itself was not visible at at,
// in addition to the errors BFS may return.
func (g *Graph) BFSAt(
	ctx context.Context,
	startID string,
	direction Direction,
	maxDepth int,
	at time.Time,
) ([]Visit, error) {
	if maxDepth < 0 {
		return nil, ErrInvalidMaxDepth
	}

	if err := validateDirection(direction); err != nil {
		return nil, err
	}

	startNode, err := g.store.GetNode(
		ctx,
		startID,
	)
	if err != nil {
		return nil, err
	}

	if !nodeVisibleAt(startNode, at) {
		return nil, ErrNodeNotVisibleAt
	}

	type queueItem struct {
		node         Node
		depth        int
		parentNodeID string
		viaEdgeID    string
	}

	queue := []queueItem{
		{
			node:  startNode,
			depth: 0,
		},
	}

	visited := map[string]struct{}{
		startID: {},
	}

	results := make([]Visit, 0)

	for len(queue) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		current := queue[0]
		queue = queue[1:]

		results = append(
			results,
			Visit{
				Node:         current.node,
				Depth:        current.depth,
				ParentNodeID: current.parentNodeID,
				ViaEdgeID:    current.viaEdgeID,
			},
		)

		if current.depth >= maxDepth {
			continue
		}

		edges, err := g.edgesForDirection(
			ctx,
			current.node.ID,
			direction,
		)
		if err != nil {
			return nil, err
		}

		sort.Slice(edges, func(i, j int) bool {
			return edges[i].ID < edges[j].ID
		})

		for _, edge := range edges {
			if !edgeVisibleAt(edge, at) {
				continue
			}

			neighborID := neighborIDForDirection(
				edge,
				direction,
			)

			if _, exists := visited[neighborID]; exists {
				continue
			}

			neighbor, err := g.store.GetNode(
				ctx,
				neighborID,
			)
			if err != nil {
				return nil, err
			}

			if !nodeVisibleAt(neighbor, at) {
				continue
			}

			visited[neighborID] = struct{}{}

			queue = append(
				queue,
				queueItem{
					node:         neighbor,
					depth:        current.depth + 1,
					parentNodeID: current.node.ID,
					viaEdgeID:    edge.ID,
				},
			)
		}
	}

	return results, nil
}
