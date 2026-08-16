package graph

import "time"

type Node struct {
	ID         string
	Type       string
	Properties map[string]any

	RecordedAt time.Time

	ValidFrom  *time.Time
	ValidUntil *time.Time
}

type Edge struct {
	ID         string
	From       string
	To         string
	Type       string
	Properties map[string]any

	RecordedAt time.Time

	ValidFrom  *time.Time
	ValidUntil *time.Time
}
