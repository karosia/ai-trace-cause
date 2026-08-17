// Package aitracecause provides causal tracing for AI agents.
//
// It records why an AI agent made a decision or performed an action by
// connecting a chain of semantic entities — Source, Observation, Fact,
// Decision, and Action — through typed causal relationships. This
// complements operational observability systems such as OpenTelemetry,
// which answer how an agent executed, by answering why it acted.
//
// The zero-value entry point is Trace, created with New. See the
// semantic and graph subpackages for the underlying domain model and
// graph engine, respectively.
package aitracecause

import (
	"context"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
)

// Trace is the public entry point for recording and querying an AI
// agent's causal graph. It wraps a graph.Graph for storage and a
// semantic.Service for validating and recording the semantic model.
//
// A Trace is safe for concurrent use as long as the underlying
// graph.Store is safe for concurrent use.
type Trace struct {
	graph    *graph.Graph
	semantic *semantic.Service
}

// New creates a Trace configured with the given Options.
//
// With no options, New uses a concurrent-safe in-memory store and a
// UUIDv7 ID generator. Use WithStore, WithMemoryStore, WithClock,
// WithTelemetryHook, and WithIDGenerator to customize behavior.
func New(
	options ...Option,
) (*Trace, error) {
	cfg := defaultConfig()

	for _, option := range options {
		if option != nil {
			option(cfg)
		}
	}

	g, err := graph.New(cfg.store)
	if err != nil {
		return nil, err
	}

	semanticOptions := make(
		[]semantic.Option,
		0,
		3,
	)

	if cfg.clock != nil {
		semanticOptions = append(
			semanticOptions,
			semantic.WithClock(cfg.clock),
		)
	}

	if cfg.telemetry != nil {
		semanticOptions = append(
			semanticOptions,
			semantic.WithTelemetryHook(
				cfg.telemetry,
			),
		)
	}

	if cfg.idGenerator != nil {
		semanticOptions = append(
			semanticOptions,
			semantic.WithIDGenerator(
				cfg.idGenerator,
			),
		)
	}

	service, err := semantic.NewService(
		g,
		semanticOptions...,
	)
	if err != nil {
		return nil, err
	}

	return &Trace{
		graph:    g,
		semantic: service,
	}, nil
}

// RecordSource records where a piece of information originated, such
// as an API response, a document, a database, or another agent.
func (t *Trace) RecordSource(
	ctx context.Context,
	source Source,
) (Source, error) {
	return t.semantic.RecordSource(
		ctx,
		source,
	)
}

// RecordObservation records something observed from an external
// source, such as a metric value or a tool response.
func (t *Trace) RecordObservation(
	ctx context.Context,
	observation Observation,
) (Observation, error) {
	return t.semantic.RecordObservation(
		ctx,
		observation,
	)
}

// RecordFact records information accepted as evidence for a decision.
func (t *Trace) RecordFact(
	ctx context.Context,
	fact Fact,
) (Fact, error) {
	return t.semantic.RecordFact(
		ctx,
		fact,
	)
}

// RecordDecision records a selected outcome or judgment made by the
// agent. Rationale is intended for concise, explicit justification and
// is not intended to store private model chain-of-thought.
func (t *Trace) RecordDecision(
	ctx context.Context,
	decision Decision,
) (Decision, error) {
	return t.semantic.RecordDecision(
		ctx,
		decision,
	)
}

// RecordAction records something the agent actually executed or
// attempted to execute.
func (t *Trace) RecordAction(
	ctx context.Context,
	action Action,
) (Action, error) {
	return t.semantic.RecordAction(
		ctx,
		action,
	)
}

// Produced records a Source -> Observation PRODUCED relationship,
// indicating that the source produced the observation.
func (t *Trace) Produced(
	ctx context.Context,
	sourceID string,
	observationID string,
) error {
	return t.semantic.Produced(
		ctx,
		sourceID,
		observationID,
	)
}

// Supports records an Observation -> Fact SUPPORTS relationship,
// indicating that the observation supports the fact.
func (t *Trace) Supports(
	ctx context.Context,
	observationID string,
	factID string,
) error {
	return t.semantic.Supports(
		ctx,
		observationID,
		factID,
	)
}

// BasisOf records a Fact -> Decision BASIS_OF relationship, indicating
// that the fact was a basis for the decision.
func (t *Trace) BasisOf(
	ctx context.Context,
	factID string,
	decisionID string,
) error {
	return t.semantic.BasisOf(
		ctx,
		factID,
		decisionID,
	)
}

// Caused records a Decision -> Action CAUSED relationship, indicating
// that the decision caused the action.
func (t *Trace) Caused(
	ctx context.Context,
	decisionID string,
	actionID string,
) error {
	return t.semantic.Caused(
		ctx,
		decisionID,
		actionID,
	)
}

// TraceActionCause walks backward from an action through its causal
// chain (Decision, Fact, Observation, Source), up to maxDepth hops,
// and returns the visited nodes in breadth-first order.
func (t *Trace) TraceActionCause(
	ctx context.Context,
	actionID string,
	maxDepth int,
) ([]Visit, error) {
	return t.semantic.TraceActionCause(
		ctx,
		actionID,
		maxDepth,
	)
}

// TraceActionCauseAt is like TraceActionCause but reconstructs the
// causal chain as it existed at the given point in time, including
// only entities and relationships that were recorded and valid at at.
func (t *Trace) TraceActionCauseAt(
	ctx context.Context,
	actionID string,
	at time.Time,
	maxDepth int,
) ([]Visit, error) {
	return t.semantic.TraceActionCauseAt(
		ctx,
		actionID,
		at,
		maxDepth,
	)
}
