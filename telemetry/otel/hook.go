package otel

import (
	"context"

	"github.com/karosia/ai-trace-cause/graph"

	oteltrace "go.opentelemetry.io/otel/trace"
)

type Hook struct{}

func New() *Hook {
	return &Hook{}
}

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
