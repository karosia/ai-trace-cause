package aitracecause

import (
	"time"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

type config struct {
	store graph.Store

	clock func() time.Time

	telemetry semantic.TelemetryHook
}

type Option func(*config)

func defaultConfig() *config {
	return &config{
		store: memory.New(),
	}
}

func WithStore(
	store graph.Store,
) Option {
	return func(cfg *config) {
		if store != nil {
			cfg.store = store
		}
	}
}

func WithMemoryStore() Option {
	return func(cfg *config) {
		cfg.store = memory.New()
	}
}

func WithClock(
	clock func() time.Time,
) Option {
	return func(cfg *config) {
		if clock != nil {
			cfg.clock = clock
		}
	}
}

func WithTelemetryHook(
	hook TelemetryHook,
) Option {
	return func(cfg *config) {
		cfg.telemetry = hook
	}
}
