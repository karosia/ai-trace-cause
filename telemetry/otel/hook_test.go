package otel_test

import (
	"context"
	"testing"

	"github.com/karosia/ai-trace-cause/telemetry/otel"

	oteltrace "go.opentelemetry.io/otel/trace"
)

func TestCorrelationFromContext(t *testing.T) {
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

	hook := otel.New()

	ref, ok := hook.CorrelationFromContext(ctx)
	if !ok {
		t.Fatal(
			"CorrelationFromContext() ok = false, want true",
		)
	}

	if ref.TraceID != traceID.String() {
		t.Errorf(
			"TraceID = %q, want %q",
			ref.TraceID,
			traceID.String(),
		)
	}

	if ref.SpanID != spanID.String() {
		t.Errorf(
			"SpanID = %q, want %q",
			ref.SpanID,
			spanID.String(),
		)
	}
}

func TestCorrelationFromContextWithoutSpan(
	t *testing.T,
) {
	hook := otel.New()

	ref, ok := hook.CorrelationFromContext(
		context.Background(),
	)

	if ok {
		t.Fatal(
			"CorrelationFromContext() ok = true, want false",
		)
	}

	if ref.TraceID != "" {
		t.Errorf(
			"TraceID = %q, want empty",
			ref.TraceID,
		)
	}

	if ref.SpanID != "" {
		t.Errorf(
			"SpanID = %q, want empty",
			ref.SpanID,
		)
	}
}
