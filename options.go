package aitracecause

import (
	"time"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/idgen/uuidv7"
	"github.com/karosia/ai-trace-cause/semantic"
	"github.com/karosia/ai-trace-cause/storage/memory"
)

type config struct {
	store graph.Store

	clock func() time.Time

	telemetry semantic.TelemetryHook

	idGenerator semantic.IDGenerator

	schema *semantic.Schema

	maxVisits int
}

// DefaultMaxVisits is the default cap on how many nodes a single trace
// may visit. See WithMaxVisits.
const DefaultMaxVisits = 10_000

// Option configures a Trace created with New.
type Option func(*config)

func defaultConfig() *config {
	return &config{
		store:       memory.New(),
		idGenerator: uuidv7.New(),
		maxVisits:   DefaultMaxVisits,
	}
}

// WithStore configures the Trace to persist its graph through store
// instead of the default in-memory store. A nil store is ignored.
func WithStore(
	store graph.Store,
) Option {
	return func(cfg *config) {
		if store != nil {
			cfg.store = store
		}
	}
}

// WithMemoryStore configures the Trace to use a fresh, concurrent-safe
// in-memory store. This is the default and is only needed to override
// an earlier WithStore option.
func WithMemoryStore() Option {
	return func(cfg *config) {
		cfg.store = memory.New()
	}
}

// WithClock configures the Trace to use clock instead of time.Now for
// timestamping recorded entities and relationships. This is primarily
// useful for deterministic tests. A nil clock is ignored.
func WithClock(
	clock func() time.Time,
) Option {
	return func(cfg *config) {
		if clock != nil {
			cfg.clock = clock
		}
	}
}

// WithTelemetryHook configures the Trace to correlate recorded
// entities and relationships with the active OpenTelemetry trace and
// span found in the recording context, using hook.
func WithTelemetryHook(
	hook TelemetryHook,
) Option {
	return func(cfg *config) {
		cfg.telemetry = hook
	}
}

// WithIDGenerator configures the Trace to use generator for producing
// IDs when a caller does not supply one. The default generator
// produces UUIDv7 values. A nil generator is ignored.
func WithIDGenerator(
	generator IDGenerator,
) Option {
	return func(cfg *config) {
		if generator != nil {
			cfg.idGenerator = generator
		}
	}
}

// WithSchema configures which relationships the Trace accepts. The
// default is DefaultSchema(); extend it with Schema.With to add
// application-specific relationships.
func WithSchema(
	schema Schema,
) Option {
	return func(cfg *config) {
		cfg.schema = &schema
	}
}

// WithMaxVisits caps how many nodes a single trace may visit
// (DefaultMaxVisits unless configured). A trace that would visit more
// fails with an error wrapping graph.ErrMaxVisitsExceeded rather than
// returning a silently truncated result. Zero or a negative value
// removes the limit.
func WithMaxVisits(
	maxVisits int,
) Option {
	return func(cfg *config) {
		cfg.maxVisits = maxVisits
	}
}
