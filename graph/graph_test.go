package graph

import (
	"errors"
	"testing"
)

func TestAddAndGetNode(t *testing.T) {
	g := New()

	node := Node{
		ID:   "fact-001",
		Type: "Fact",
		Properties: map[string]any{
			"statement": "CPU usage is above 90%",
		},
	}

	if err := g.AddNode(node); err != nil {
		t.Fatalf("AddNode() error = %v", err)
	}

	got, err := g.GetNode("fact-001")
	if err != nil {
		t.Fatalf("GetNode() error = %v", err)
	}

	if got.ID != node.ID {
		t.Errorf(
			"GetNode() ID = %q, want %q",
			got.ID,
			node.ID,
		)
	}

	if got.Type != node.Type {
		t.Errorf(
			"GetNode() Type = %q, want %q",
			got.Type,
			node.Type,
		)
	}
}

func TestAddNodeRejectsDuplicate(t *testing.T) {
	g := New()

	node := Node{
		ID:   "fact-001",
		Type: "Fact",
	}

	if err := g.AddNode(node); err != nil {
		t.Fatalf("first AddNode() error = %v", err)
	}

	err := g.AddNode(node)

	if !errors.Is(err, ErrNodeAlreadyExists) {
		t.Fatalf(
			"second AddNode() error = %v, want ErrNodeAlreadyExists",
			err,
		)
	}
}

func TestAddAndGetEdge(t *testing.T) {
	g := New()

	if err := g.AddNode(Node{
		ID:   "fact-001",
		Type: "Fact",
	}); err != nil {
		t.Fatalf("AddNode(fact-001) error = %v", err)
	}

	if err := g.AddNode(Node{
		ID:   "decision-001",
		Type: "Decision",
	}); err != nil {
		t.Fatalf("AddNode(decision-001) error = %v", err)
	}

	edge := Edge{
		ID:   "edge-001",
		From: "fact-001",
		To:   "decision-001",
		Type: "SUPPORTS",
	}

	if err := g.AddEdge(edge); err != nil {
		t.Fatalf("AddEdge() error = %v", err)
	}

	got, err := g.GetEdge("edge-001")
	if err != nil {
		t.Fatalf("GetEdge() error = %v", err)
	}

	if got.From != "fact-001" {
		t.Errorf("From = %q, want %q", got.From, "fact-001")
	}

	if got.To != "decision-001" {
		t.Errorf("To = %q, want %q", got.To, "decision-001")
	}
}

func TestAddEdgeRejectsUnknownNode(t *testing.T) {
	g := New()

	if err := g.AddNode(Node{
		ID:   "fact-001",
		Type: "Fact",
	}); err != nil {
		t.Fatalf("AddNode() error = %v", err)
	}

	err := g.AddEdge(Edge{
		ID:   "edge-001",
		From: "fact-001",
		To:   "decision-001",
		Type: "SUPPORTS",
	})

	if !errors.Is(err, ErrNodeNotFound) {
		t.Fatalf(
			"AddEdge() error = %v, want ErrNodeNotFound",
			err,
		)
	}
}

func TestOutgoingNeighbors(t *testing.T) {
	g := New()

	nodes := []Node{
		{
			ID:   "observation-001",
			Type: "Observation",
		},
		{
			ID:   "fact-001",
			Type: "Fact",
		},
		{
			ID:   "fact-002",
			Type: "Fact",
		},
	}

	for _, node := range nodes {
		if err := g.AddNode(node); err != nil {
			t.Fatalf(
				"AddNode(%q) error = %v",
				node.ID,
				err,
			)
		}
	}

	edges := []Edge{
		{
			ID:   "edge-001",
			From: "observation-001",
			To:   "fact-001",
			Type: "SUPPORTS",
		},
		{
			ID:   "edge-002",
			From: "observation-001",
			To:   "fact-002",
			Type: "SUPPORTS",
		},
	}

	for _, edge := range edges {
		if err := g.AddEdge(edge); err != nil {
			t.Fatalf(
				"AddEdge(%q) error = %v",
				edge.ID,
				err,
			)
		}
	}

	neighbors, err := g.OutgoingNeighbors("observation-001")
	if err != nil {
		t.Fatalf(
			"OutgoingNeighbors() error = %v",
			err,
		)
	}

	if len(neighbors) != 2 {
		t.Fatalf(
			"len(neighbors) = %d, want 2",
			len(neighbors),
		)
	}

	if neighbors[0].ID != "fact-001" {
		t.Errorf(
			"neighbors[0].ID = %q, want fact-001",
			neighbors[0].ID,
		)
	}

	if neighbors[1].ID != "fact-002" {
		t.Errorf(
			"neighbors[1].ID = %q, want fact-002",
			neighbors[1].ID,
		)
	}
}
