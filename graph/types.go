package graph

type Node struct {
	ID         string
	Type       string
	Properties map[string]any
}

type Edge struct {
	ID         string
	From       string
	To         string
	Type       string
	Properties map[string]any
}
