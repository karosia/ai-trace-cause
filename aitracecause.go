package aitracecause

import (
	"context"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
)

type Trace struct {
	graph    *graph.Graph
	semantic *semantic.Service
}

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

func (t *Trace) RecordSource(
	ctx context.Context,
	source Source,
) (Source, error) {
	return t.semantic.RecordSource(
		ctx,
		source,
	)
}

func (t *Trace) RecordObservation(
	ctx context.Context,
	observation Observation,
) (Observation, error) {
	return t.semantic.RecordObservation(
		ctx,
		observation,
	)
}

func (t *Trace) RecordFact(
	ctx context.Context,
	fact Fact,
) (Fact, error) {
	return t.semantic.RecordFact(
		ctx,
		fact,
	)
}

func (t *Trace) RecordDecision(
	ctx context.Context,
	decision Decision,
) (Decision, error) {
	return t.semantic.RecordDecision(
		ctx,
		decision,
	)
}

func (t *Trace) RecordAction(
	ctx context.Context,
	action Action,
) (Action, error) {
	return t.semantic.RecordAction(
		ctx,
		action,
	)
}

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
