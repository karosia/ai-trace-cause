package graph

import (
	"context"
	"sort"
	"time"
)

// Subgraph is the result of a traversal that keeps every edge between
// visited nodes, not just the one edge each node was first reached
// through.
//
// Visits alone describe a spanning tree: when a node is reachable via
// several edges (for example two Facts both supported by the same
// Observation), BFS records only the first. Anything that must account
// for every relationship, such as an audit explanation, should use
// Edges instead.
type Subgraph struct {
	// Visits holds the visited nodes in breadth-first order, exactly
	// as BFS (or BFSAt) returns them.
	Visits []Visit `json:"visits"`

	// Edges holds every edge, in the traversal direction, whose two
	// endpoints were both visited. Edges are ordered by the visit
	// order of the node they were collected from, then by the visit
	// order of the opposite endpoint, then by RecordedAt and edge ID.
	// Self-loops are included.
	Edges []Edge `json:"edges"`
}

// Subgraph performs a BFS from startID (see BFS for parameters and
// errors) and additionally collects every edge connecting two visited
// nodes.
func (g *Graph) Subgraph(
	ctx context.Context,
	startID string,
	direction Direction,
	maxDepth int,
) (Subgraph, error) {
	visits, err := g.BFS(ctx, startID, direction, maxDepth)
	if err != nil {
		return Subgraph{}, err
	}

	edges, err := g.collectEdges(
		ctx,
		visits,
		direction,
		func(Edge) bool { return true },
	)
	if err != nil {
		return Subgraph{}, err
	}

	return Subgraph{Visits: visits, Edges: edges}, nil
}

// SubgraphAt is like Subgraph but uses BFSAt, and only collects edges
// that were recorded and valid at at.
func (g *Graph) SubgraphAt(
	ctx context.Context,
	startID string,
	direction Direction,
	maxDepth int,
	at time.Time,
) (Subgraph, error) {
	visits, err := g.BFSAt(ctx, startID, direction, maxDepth, at)
	if err != nil {
		return Subgraph{}, err
	}

	edges, err := g.collectEdges(
		ctx,
		visits,
		direction,
		func(edge Edge) bool { return edgeVisibleAt(edge, at) },
	)
	if err != nil {
		return Subgraph{}, err
	}

	return Subgraph{Visits: visits, Edges: edges}, nil
}

// collectEdges returns every edge in direction whose endpoints are
// both among visits and which satisfies keep. Each edge is collected
// from exactly one endpoint (its To for incoming traversals, its From
// for outgoing ones), so no edge is returned twice.
func (g *Graph) collectEdges(
	ctx context.Context,
	visits []Visit,
	direction Direction,
	keep func(Edge) bool,
) ([]Edge, error) {
	order := make(map[string]int, len(visits))
	for i, visit := range visits {
		order[visit.Node.ID] = i
	}

	edges := make([]Edge, 0, len(visits))

	for _, visit := range visits {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		candidates, err := g.edgesForDirection(
			ctx,
			visit.Node.ID,
			direction,
		)
		if err != nil {
			return nil, err
		}

		kept := candidates[:0]

		for _, edge := range candidates {
			neighborID := neighborIDForDirection(edge, direction)
			if _, visited := order[neighborID]; !visited {
				continue
			}

			if !keep(edge) {
				continue
			}

			kept = append(kept, edge)
		}

		sort.Slice(kept, func(i, j int) bool {
			oi := order[neighborIDForDirection(kept[i], direction)]
			oj := order[neighborIDForDirection(kept[j], direction)]
			if oi != oj {
				return oi < oj
			}

			if !kept[i].RecordedAt.Equal(kept[j].RecordedAt) {
				return kept[i].RecordedAt.Before(kept[j].RecordedAt)
			}

			return kept[i].ID < kept[j].ID
		})

		edges = append(edges, kept...)
	}

	return edges, nil
}
