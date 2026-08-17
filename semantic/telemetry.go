package semantic

import (
	"context"

	"github.com/karosia/ai-trace-cause/graph"
)

// TelemetryHook correlates a recording context with an active
// OpenTelemetry trace and span, so recorded entities and relationships
// can be tied back to the distributed trace that produced them.
type TelemetryHook interface {
	// CorrelationFromContext extracts a TelemetryRef from ctx. It
	// returns false if ctx carries no active trace/span to correlate
	// with.
	CorrelationFromContext(
		ctx context.Context,
	) (graph.TelemetryRef, bool)
}
