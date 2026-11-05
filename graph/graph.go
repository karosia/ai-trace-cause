// Package graph implements a generic, storage-agnostic directed graph
// engine: nodes, edges, traversal (BFS/DFS), and temporal visibility.
//
// The graph layer intentionally has no knowledge of AI-specific
// semantics; see the semantic package for the domain model built on
// top of it.
package graph

import (
	"context"
	"fmt"
	"sort"

	"github.com/karosia/ai-trace-cause/internal/propcopy"
)

// Graph is a directed graph backed by a pluggable Store. It validates
// nodes and edges before delegating persistence to the store.
type Graph struct {
	store Store

	maxVisits int
}

// Option configures a Graph created with New.
type Option func(*Graph)

// WithMaxVisits caps how many nodes a single traversal (BFS, BFSAt,
// DFS, Subgraph, SubgraphAt) may visit. A traversal that would visit
// more returns an error wrapping ErrMaxVisitsExceeded instead of a
// silently truncated result. Zero or a negative value, the default,
// means no limit.
//
// Depth limits alone do not bound a traversal: a single Source feeding
// thousands of Observations, or an agent loop chaining Actions and
// Observations, can reach very many nodes within a few hops.
func WithMaxVisits(maxVisits int) Option {
	return func(g *Graph) {
		g.maxVisits = maxVisits
	}
}

// New creates a Graph backed by store. It returns ErrNilStore if store
// is nil.
func New(store Store, options ...Option) (*Graph, error) {
	if store == nil {
		return nil, ErrNilStore
	}

	g := &Graph{
		store: store,
	}

	for _, option := range options {
		if option != nil {
			option(g)
		}
	}

	return g, nil
}

// checkVisitLimit returns ErrMaxVisitsExceeded once visited exceeds
// the configured limit.
func (g *Graph) checkVisitLimit(visited int, startID string) error {
	if g.maxVisits > 0 && visited > g.maxVisits {
		return fmt.Errorf(
			"%w: more than %d nodes reachable from %q; lower maxDepth or raise the limit",
			ErrMaxVisitsExceeded,
			g.maxVisits,
			startID,
		)
	}

	return nil
}

// AddNode validates and adds node to the graph. It returns
// ErrEmptyNodeID or ErrEmptyNodeType if the corresponding field is
// empty, ErrInvalidValidityInterval if node's validity interval is
// malformed, or ErrUnsupportedPropertyValue if a property value is not
// JSON-representable.
//
// Properties are deep-copied before they reach the Store, so mutating
// the caller's maps or slices afterwards never alters the recorded
// node.
func (g *Graph) AddNode(
	ctx context.Context,
	node Node,
) error {
	node, err := prepareNode(node)
	if err != nil {
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
// empty, ErrInvalidValidityInterval if edge's validity interval is
// malformed, or ErrUnsupportedPropertyValue if a property value is not
// JSON-representable. Properties are deep-copied as in AddNode.
func (g *Graph) AddEdge(
	ctx context.Context,
	edge Edge,
) error {
	edge, err := prepareEdge(edge)
	if err != nil {
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

// prepareNode validates node and returns a copy whose Properties are
// deep-copied and independent of the caller's.
func prepareNode(node Node) (Node, error) {
	if node.ID == "" {
		return Node{}, ErrEmptyNodeID
	}

	if node.Type == "" {
		return Node{}, ErrEmptyNodeType
	}

	if err := validateValidityInterval(
		node.ValidFrom,
		node.ValidUntil,
	); err != nil {
		return Node{}, err
	}

	properties, err := propcopy.CloneMap(node.Properties)
	if err != nil {
		return Node{}, fmt.Errorf(
			"%w: node %q: %w",
			ErrUnsupportedPropertyValue,
			node.ID,
			err,
		)
	}

	node.Properties = properties

	return node, nil
}

// prepareEdge is the Edge counterpart to prepareNode.
func prepareEdge(edge Edge) (Edge, error) {
	if edge.ID == "" {
		return Edge{}, ErrEmptyEdgeID
	}

	if edge.Type == "" {
		return Edge{}, ErrEmptyEdgeType
	}

	if err := validateValidityInterval(
		edge.ValidFrom,
		edge.ValidUntil,
	); err != nil {
		return Edge{}, err
	}

	properties, err := propcopy.CloneMap(edge.Properties)
	if err != nil {
		return Edge{}, fmt.Errorf(
			"%w: edge %q: %w",
			ErrUnsupportedPropertyValue,
			edge.ID,
			err,
		)
	}

	edge.Properties = properties

	return edge, nil
}
