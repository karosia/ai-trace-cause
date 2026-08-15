package semantic_test

import (
	"testing"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

func newTestService(
	t *testing.T,
) (*semantic.Service, *graph.Graph) {
	t.Helper()

	store := memory.New()

	g, err := graph.New(store)
	if err != nil {
		t.Fatalf(
			"graph.New() error = %v",
			err,
		)
	}

	service, err := semantic.NewService(g)
	if err != nil {
		t.Fatalf(
			"semantic.NewService() error = %v",
			err,
		)
	}

	return service, g
}
