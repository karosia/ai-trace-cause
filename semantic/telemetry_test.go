package semantic_test

import (
	"context"
	"testing"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
	"github.com/karosia/ai-trace-cause/storage/memory"
	oteltelemetry "github.com/karosia/ai-trace-cause/telemetry/otel"

	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestDecisionCapturesOTelContext(
	t *testing.T,
) {
	store := memory.New()

	g, err := graph.New(store)
	if err != nil {
		t.Fatal(err)
	}

	service, err := semantic.NewService(
		g,
		semantic.WithTelemetryHook(
			oteltelemetry.New(),
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	traceID, err := oteltrace.TraceIDFromHex(
		"4bf92f3577b34da6a3ce929d0e0e4736",
	)
	if err != nil {
		t.Fatal(err)
	}

	spanID, err := oteltrace.SpanIDFromHex(
		"00f067aa0ba902b7",
	)
	if err != nil {
		t.Fatal(err)
	}

	spanContext := oteltrace.NewSpanContext(
		oteltrace.SpanContextConfig{
			TraceID: traceID,
			SpanID:  spanID,
		},
	)

	ctx := oteltrace.ContextWithSpanContext(
		context.Background(),
		spanContext,
	)

	decision := semantic.Decision{
		ID:         "decision-001",
		Outcome:    "Scale service",
		Rationale:  "CPU usage is high",
		Confidence: 0.92,
	}

	recorded, err := service.RecordDecision(
		ctx,
		decision,
	)
	if err != nil {
		t.Fatal(err)
	}

	node, err := g.GetNode(
		ctx,
		recorded.ID,
	)
	if err != nil {
		t.Fatal(err)
	}

	if node.Telemetry == nil {
		t.Fatal(
			"node.Telemetry = nil",
		)
	}

	if node.Telemetry.TraceID != traceID.String() {
		t.Errorf(
			"TraceID = %q, want %q",
			node.Telemetry.TraceID,
			traceID.String(),
		)
	}

	if node.Telemetry.SpanID != spanID.String() {
		t.Errorf(
			"SpanID = %q, want %q",
			node.Telemetry.SpanID,
			spanID.String(),
		)
	}
}
