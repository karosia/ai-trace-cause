// Package otel implements semantic.TelemetryHook using OpenTelemetry,
// correlating recorded entities and relationships with the active
// trace and span found in the recording context.
package otel

import (
	"context"

	"github.com/karosia/ai-trace-cause/graph"

	oteltrace "go.opentelemetry.io/otel/trace"
)

// Hook is a semantic.TelemetryHook backed by OpenTelemetry's span
// context propagation. It reads the active span from context but does
// not create, own, or end spans itself.
type Hook struct{}

// New creates a Hook.
func New() *Hook {
	return &Hook{}
}

// CorrelationFromContext extracts the trace ID and span ID from the
// OpenTelemetry span active in ctx. It returns false if ctx carries no
// valid span context.
func (h *Hook) CorrelationFromContext(
	ctx context.Context,
) (graph.TelemetryRef, bool) {
	spanContext := oteltrace.SpanContextFromContext(
		ctx,
	)

	if !spanContext.IsValid() {
		return graph.TelemetryRef{}, false
	}

	return graph.TelemetryRef{
		TraceID: spanContext.TraceID().String(),
		SpanID:  spanContext.SpanID().String(),
	}, true
}
