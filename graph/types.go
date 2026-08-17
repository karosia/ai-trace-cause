package graph

import "time"

// TelemetryRef correlates a node or edge with the OpenTelemetry trace
// and span that were active when it was recorded.
type TelemetryRef struct {
	TraceID string `json:"traceId"`
	SpanID  string `json:"spanId"`
}

// Node is a vertex in the graph, identified by a unique ID and typed
// by Type. Properties holds arbitrary domain data. RecordedAt is when
// the node was recorded; ValidFrom and ValidUntil describe the
// half-open interval during which the node is valid in the modeled
// domain (either or both may be nil, meaning unbounded). Telemetry, if
// set, correlates the node with an OpenTelemetry trace and span.
type Node struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties,omitempty"`

	RecordedAt time.Time `json:"recordedAt"`

	ValidFrom  *time.Time `json:"validFrom,omitempty"`
	ValidUntil *time.Time `json:"validUntil,omitempty"`

	Telemetry *TelemetryRef `json:"telemetry,omitempty"`
}

// Edge is a directed, typed relationship from one node to another,
// identified by a unique ID. Properties holds arbitrary domain data.
// RecordedAt, ValidFrom, ValidUntil, and Telemetry have the same
// meaning as on Node.
type Edge struct {
	ID         string         `json:"id"`
	From       string         `json:"from"`
	To         string         `json:"to"`
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties,omitempty"`

	RecordedAt time.Time `json:"recordedAt"`

	ValidFrom  *time.Time `json:"validFrom,omitempty"`
	ValidUntil *time.Time `json:"validUntil,omitempty"`

	Telemetry *TelemetryRef `json:"telemetry,omitempty"`
}
