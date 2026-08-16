package semantic

import (
	"context"

	"github.com/karosia/ai-trace-cause/graph"
)

type TelemetryHook interface {
	CorrelationFromContext(
		ctx context.Context,
	) (graph.TelemetryRef, bool)
}
