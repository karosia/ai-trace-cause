package graph

import (
	"context"
	"fmt"
	"sort"
)

// Direction selects which edges a traversal follows: outgoing edges
// (Node.ID == Edge.From) or incoming edges (Node.ID == Edge.To).
type Direction uint8

const (
	// DirectionOutgoing follows edges away from each visited node.
	DirectionOutgoing Direction = iota + 1
	// DirectionIncoming follows edges toward each visited node.
	DirectionIncoming
)

// Visit is a single node reached during a traversal, along with the
// depth (number of hops from the start node) at which it was found
// and the parent node and edge it was reached through. The start node
// itself is visited at Depth 0 with an empty ParentNodeID and
// ViaEdgeID.
type Visit struct {
	Node Node

	Depth int

	ParentNodeID string
	ViaEdgeID    string
}

// BFS performs a breadth-first traversal starting at startID,
// following edges in direction up to maxDepth hops, and returns every
// visited node in the order it was discovered. It returns
// ErrInvalidMaxDepth if maxDepth is negative, ErrInvalidDirection if
// direction is not recognized, or ErrNodeNotFound if startID does not
// exist. Each node is visited at most once (cycle-safe).
func (g *Graph) BFS(
	ctx context.Context,
	startID string,
	direction Direction,
	maxDepth int,
) ([]Visit, error) {
	if maxDepth < 0 {
		return nil, ErrInvalidMaxDepth
	}

	if err := validateDirection(direction); err != nil {
		return nil, err
	}

	startNode, err := g.store.GetNode(ctx, startID)
	if err != nil {
		return nil, err
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

		results = append(results, Visit{
			Node:         current.node,
			Depth:        current.depth,
			ParentNodeID: current.parentNodeID,
			ViaEdgeID:    current.viaEdgeID,
		})

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

			visited[neighborID] = struct{}{}

			queue = append(queue, queueItem{
				node:         neighbor,
				depth:        current.depth + 1,
				parentNodeID: current.node.ID,
				viaEdgeID:    edge.ID,
			})
		}
	}

	return results, nil
}

// DFS performs a depth-first traversal starting at startID, following
// edges in direction up to maxDepth hops, and returns every visited
// node in the order it was discovered. It returns the same errors as
// BFS for invalid input. Each node is visited at most once
// (cycle-safe).
func (g *Graph) DFS(
	ctx context.Context,
	startID string,
	direction Direction,
	maxDepth int,
) ([]Visit, error) {
	if maxDepth < 0 {
		return nil, ErrInvalidMaxDepth
	}

	if err := validateDirection(direction); err != nil {
		return nil, err
	}

	startNode, err := g.store.GetNode(ctx, startID)
	if err != nil {
		return nil, err
	}

	type stackItem struct {
		node         Node
		depth        int
		parentNodeID string
		viaEdgeID    string
	}

	stack := []stackItem{
		{
			node:  startNode,
			depth: 0,
		},
	}

	visited := make(map[string]struct{})

	results := make([]Visit, 0)

	for len(stack) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		lastIndex := len(stack) - 1

		current := stack[lastIndex]
		stack = stack[:lastIndex]

		if _, exists := visited[current.node.ID]; exists {
			continue
		}

		visited[current.node.ID] = struct{}{}

		results = append(results, Visit{
			Node:         current.node,
			Depth:        current.depth,
			ParentNodeID: current.parentNodeID,
			ViaEdgeID:    current.viaEdgeID,
		})

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

		for i := len(edges) - 1; i >= 0; i-- {
			edge := edges[i]

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

			stack = append(stack, stackItem{
				node:         neighbor,
				depth:        current.depth + 1,
				parentNodeID: current.node.ID,
				viaEdgeID:    edge.ID,
			})
		}
	}

	return results, nil
}

func (g *Graph) edgesForDirection(
	ctx context.Context,
	nodeID string,
	direction Direction,
) ([]Edge, error) {
	switch direction {

	case DirectionOutgoing:
		return g.store.OutgoingEdges(
			ctx,
			nodeID,
		)

	case DirectionIncoming:
		return g.store.IncomingEdges(
			ctx,
			nodeID,
		)

	default:
		return nil, ErrInvalidDirection
	}
}

func neighborIDForDirection(
	edge Edge,
	direction Direction,
) string {
	if direction == DirectionOutgoing {
		return edge.To
	}

	return edge.From
}

func validateDirection(
	direction Direction,
) error {
	switch direction {

	case DirectionOutgoing,
		DirectionIncoming:
		return nil

	default:
		return fmt.Errorf(
			"%w: %d",
			ErrInvalidDirection,
			direction,
		)
	}
}
