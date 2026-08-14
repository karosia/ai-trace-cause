package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/karosia/ai-trace-cause/graph"
)

type Store struct {
	mu sync.RWMutex

	nodes map[string]graph.Node
	edges map[string]graph.Edge

	outgoing map[string]map[string]struct{}
	incoming map[string]map[string]struct{}
}

func New() *Store {
	return &Store{
		nodes:    make(map[string]graph.Node),
		edges:    make(map[string]graph.Edge),
		outgoing: make(map[string]map[string]struct{}),
		incoming: make(map[string]map[string]struct{}),
	}
}

func (s *Store) PutNode(
	ctx context.Context,
	node graph.Node,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if node.ID == "" {
		return graph.ErrEmptyNodeID
	}

	if node.Type == "" {
		return graph.ErrEmptyNodeType
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.nodes[node.ID]; exists {
		return fmt.Errorf(
			"%w: %s",
			graph.ErrNodeAlreadyExists,
			node.ID,
		)
	}

	s.nodes[node.ID] = cloneNode(node)

	return nil
}

func (s *Store) GetNode(
	ctx context.Context,
	id string,
) (graph.Node, error) {
	if err := ctx.Err(); err != nil {
		return graph.Node{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	node, exists := s.nodes[id]
	if !exists {
		return graph.Node{}, fmt.Errorf(
			"%w: %s",
			graph.ErrNodeNotFound,
			id,
		)
	}

	return cloneNode(node), nil
}

func (s *Store) PutEdge(
	ctx context.Context,
	edge graph.Edge,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if edge.ID == "" {
		return graph.ErrEmptyEdgeID
	}

	if edge.Type == "" {
		return graph.ErrEmptyEdgeType
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.edges[edge.ID]; exists {
		return fmt.Errorf(
			"%w: %s",
			graph.ErrEdgeAlreadyExists,
			edge.ID,
		)
	}

	if _, exists := s.nodes[edge.From]; !exists {
		return fmt.Errorf(
			"%w: %s",
			graph.ErrNodeNotFound,
			edge.From,
		)
	}

	if _, exists := s.nodes[edge.To]; !exists {
		return fmt.Errorf(
			"%w: %s",
			graph.ErrNodeNotFound,
			edge.To,
		)
	}

	s.edges[edge.ID] = cloneEdge(edge)

	if s.outgoing[edge.From] == nil {
		s.outgoing[edge.From] = make(map[string]struct{})
	}

	if s.incoming[edge.To] == nil {
		s.incoming[edge.To] = make(map[string]struct{})
	}

	s.outgoing[edge.From][edge.ID] = struct{}{}
	s.incoming[edge.To][edge.ID] = struct{}{}

	return nil
}

func (s *Store) GetEdge(
	ctx context.Context,
	id string,
) (graph.Edge, error) {
	if err := ctx.Err(); err != nil {
		return graph.Edge{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	edge, exists := s.edges[id]
	if !exists {
		return graph.Edge{}, fmt.Errorf(
			"%w: %s",
			graph.ErrEdgeNotFound,
			id,
		)
	}

	return cloneEdge(edge), nil
}

func (s *Store) OutgoingEdges(
	ctx context.Context,
	nodeID string,
) ([]graph.Edge, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.nodes[nodeID]; !exists {
		return nil, fmt.Errorf(
			"%w: %s",
			graph.ErrNodeNotFound,
			nodeID,
		)
	}

	edgeIDs := s.outgoing[nodeID]

	edges := make([]graph.Edge, 0, len(edgeIDs))

	for edgeID := range edgeIDs {
		edges = append(
			edges,
			cloneEdge(s.edges[edgeID]),
		)
	}

	sort.Slice(edges, func(i, j int) bool {
		return edges[i].ID < edges[j].ID
	})

	return edges, nil
}

func (s *Store) IncomingEdges(
	ctx context.Context,
	nodeID string,
) ([]graph.Edge, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.nodes[nodeID]; !exists {
		return nil, fmt.Errorf(
			"%w: %s",
			graph.ErrNodeNotFound,
			nodeID,
		)
	}

	edgeIDs := s.incoming[nodeID]

	edges := make([]graph.Edge, 0, len(edgeIDs))

	for edgeID := range edgeIDs {
		edges = append(
			edges,
			cloneEdge(s.edges[edgeID]),
		)
	}

	sort.Slice(edges, func(i, j int) bool {
		return edges[i].ID < edges[j].ID
	})

	return edges, nil
}

func cloneNode(node graph.Node) graph.Node {
	return graph.Node{
		ID:         node.ID,
		Type:       node.Type,
		Properties: cloneProperties(node.Properties),
	}
}

func cloneEdge(edge graph.Edge) graph.Edge {
	return graph.Edge{
		ID:         edge.ID,
		From:       edge.From,
		To:         edge.To,
		Type:       edge.Type,
		Properties: cloneProperties(edge.Properties),
	}
}

func cloneProperties(
	properties map[string]any,
) map[string]any {
	if properties == nil {
		return nil
	}

	cloned := make(
		map[string]any,
		len(properties),
	)

	for key, value := range properties {
		cloned[key] = value
	}

	return cloned
}
