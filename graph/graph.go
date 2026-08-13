package graph

import (
	"fmt"
	"sort"
	"sync"
)

type Graph struct {
	mu sync.RWMutex

	nodes map[string]Node
	edges map[string]Edge

	outgoing map[string]map[string]struct{}
	incoming map[string]map[string]struct{}
}

func New() *Graph {
	return &Graph{
		nodes:    make(map[string]Node),
		edges:    make(map[string]Edge),
		outgoing: make(map[string]map[string]struct{}),
		incoming: make(map[string]map[string]struct{}),
	}
}

func (g *Graph) AddNode(node Node) error {
	if node.ID == "" {
		return ErrEmptyNodeID
	}

	if node.Type == "" {
		return ErrEmptyNodeType
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.nodes[node.ID]; exists {
		return fmt.Errorf("%w: %s", ErrNodeAlreadyExists, node.ID)
	}

	g.nodes[node.ID] = cloneNode(node)

	return nil
}

func (g *Graph) GetNode(id string) (Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.nodes[id]
	if !exists {
		return Node{}, fmt.Errorf("%w: %s", ErrNodeNotFound, id)
	}

	return cloneNode(node), nil
}

func (g *Graph) AddEdge(edge Edge) error {
	if edge.ID == "" {
		return ErrEmptyEdgeID
	}

	if edge.Type == "" {
		return ErrEmptyEdgeType
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if _, exists := g.edges[edge.ID]; exists {
		return fmt.Errorf("%w: %s", ErrEdgeAlreadyExists, edge.ID)
	}

	if _, exists := g.nodes[edge.From]; !exists {
		return fmt.Errorf("%w: %s", ErrNodeNotFound, edge.From)
	}

	if _, exists := g.nodes[edge.To]; !exists {
		return fmt.Errorf("%w: %s", ErrNodeNotFound, edge.To)
	}

	g.edges[edge.ID] = cloneEdge(edge)

	if g.outgoing[edge.From] == nil {
		g.outgoing[edge.From] = make(map[string]struct{})
	}

	if g.incoming[edge.To] == nil {
		g.incoming[edge.To] = make(map[string]struct{})
	}

	g.outgoing[edge.From][edge.ID] = struct{}{}
	g.incoming[edge.To][edge.ID] = struct{}{}

	return nil
}

func (g *Graph) GetEdge(id string) (Edge, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edge, exists := g.edges[id]
	if !exists {
		return Edge{}, fmt.Errorf("%w: %s", ErrEdgeNotFound, id)
	}

	return cloneEdge(edge), nil
}

func (g *Graph) OutgoingNeighbors(nodeID string) ([]Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.nodes[nodeID]; !exists {
		return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, nodeID)
	}

	edgeIDs := g.outgoing[nodeID]

	neighbors := make([]Node, 0, len(edgeIDs))

	for edgeID := range edgeIDs {
		edge := g.edges[edgeID]
		node := g.nodes[edge.To]

		neighbors = append(neighbors, cloneNode(node))
	}

	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].ID < neighbors[j].ID
	})

	return neighbors, nil
}

func (g *Graph) IncomingNeighbors(nodeID string) ([]Node, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if _, exists := g.nodes[nodeID]; !exists {
		return nil, fmt.Errorf("%w: %s", ErrNodeNotFound, nodeID)
	}

	edgeIDs := g.incoming[nodeID]

	neighbors := make([]Node, 0, len(edgeIDs))

	for edgeID := range edgeIDs {
		edge := g.edges[edgeID]
		node := g.nodes[edge.From]

		neighbors = append(neighbors, cloneNode(node))
	}

	sort.Slice(neighbors, func(i, j int) bool {
		return neighbors[i].ID < neighbors[j].ID
	})

	return neighbors, nil
}

func cloneNode(node Node) Node {
	return Node{
		ID:         node.ID,
		Type:       node.Type,
		Properties: cloneProperties(node.Properties),
	}
}

func cloneEdge(edge Edge) Edge {
	return Edge{
		ID:         edge.ID,
		From:       edge.From,
		To:         edge.To,
		Type:       edge.Type,
		Properties: cloneProperties(edge.Properties),
	}
}

func cloneProperties(properties map[string]any) map[string]any {
	if properties == nil {
		return nil
	}

	cloned := make(map[string]any, len(properties))

	for key, value := range properties {
		cloned[key] = value
	}

	return cloned
}
