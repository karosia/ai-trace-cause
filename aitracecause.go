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
// With no options, New uses a concurrent-safe in-memory store, a
// UUIDv7 ID generator, DefaultSchema, and DefaultMaxVisits. Use WithStore, WithMemoryStore, WithClock,
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

	g, err := graph.New(
		cfg.store,
		graph.WithMaxVisits(cfg.maxVisits),
	)
	if err != nil {
		return nil, err
	}

	semanticOptions := make(
		[]semantic.Option,
		0,
		4,
	)

	if cfg.schema != nil {
		semanticOptions = append(
			semanticOptions,
			semantic.WithSchema(*cfg.schema),
		)
	}

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
// source, such as a metric value or a tool response. Pass
// WithCauses(sourceIDs...) to link the producing Source(s) in the same
// call.
func (t *Trace) RecordObservation(
	ctx context.Context,
	observation Observation,
	options ...RecordOption,
) (Observation, error) {
	return t.semantic.RecordObservation(
		ctx,
		observation,
		options...,
	)
}

// RecordFact records information accepted as evidence for a decision.
// Pass WithCauses(observationIDs...) to link the supporting
// Observation(s) in the same call.
func (t *Trace) RecordFact(
	ctx context.Context,
	fact Fact,
	options ...RecordOption,
) (Fact, error) {
	return t.semantic.RecordFact(
		ctx,
		fact,
		options...,
	)
}

// RecordDecision records a selected outcome or judgment made by the
// agent. Rationale is intended for concise, explicit justification and
// is not intended to store private model chain-of-thought. Pass
// WithCauses(factIDs...) to link the Fact(s) it was based on in the
// same call.
func (t *Trace) RecordDecision(
	ctx context.Context,
	decision Decision,
	options ...RecordOption,
) (Decision, error) {
	return t.semantic.RecordDecision(
		ctx,
		decision,
		options...,
	)
}

// RecordAction records something the agent actually executed or
// attempted to execute. Pass WithCauses(decisionIDs...) to link the
// Decision(s) that caused it in the same call; this is what lets
// TraceActionCauseAtActionTime see them.
func (t *Trace) RecordAction(
	ctx context.Context,
	action Action,
	options ...RecordOption,
) (Action, error) {
	return t.semantic.RecordAction(
		ctx,
		action,
		options...,
	)
}

// Produced records a Source -> Observation PRODUCED relationship,
// indicating that the source produced the observation. Recording the
// same relationship again is a no-op, so retries are safe; the same
// holds for Supports, BasisOf, and Caused.
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

// ResultedIn records an Action -> Observation RESULTED_IN
// relationship, such as a tool call and the result it returned. This
// is what lets a trace follow an agent across tool calls.
func (t *Trace) ResultedIn(
	ctx context.Context,
	actionID string,
	observationID string,
) error {
	return t.semantic.ResultedIn(
		ctx,
		actionID,
		observationID,
	)
}

// Infers records a Fact -> Fact INFERS relationship from a premise to
// a conclusion drawn from it.
func (t *Trace) Infers(
	ctx context.Context,
	premiseFactID string,
	conclusionFactID string,
) error {
	return t.semantic.Infers(
		ctx,
		premiseFactID,
		conclusionFactID,
	)
}

// Refines records a Decision -> Decision REFINES relationship from a
// broader decision, such as a plan, to a narrower one made within it.
func (t *Trace) Refines(
	ctx context.Context,
	parentDecisionID string,
	childDecisionID string,
) error {
	return t.semantic.Refines(
		ctx,
		parentDecisionID,
		childDecisionID,
	)
}

// Link records a relationship of any type the configured Schema
// allows, from fromID (the cause) to toID (the effect). It returns a
// *semantic.RelationNotAllowedError otherwise. Recording the same
// relationship again is a no-op.
func (t *Trace) Link(
	ctx context.Context,
	fromID string,
	relation RelationType,
	toID string,
) error {
	return t.semantic.Link(
		ctx,
		fromID,
		relation,
		toID,
	)
}

// Run returns a handle for recording one agent task or conversation
// step by step. See Run.Step.
func (t *Trace) Run(runID string) *Run {
	return t.semantic.Run(runID)
}

// TraceEffectsGraph walks forward from any node through everything it
// went on to influence, answering for example "this fact was
// retracted: which decisions and actions relied on it?".
func (t *Trace) TraceEffectsGraph(
	ctx context.Context,
	nodeID string,
	maxDepth int,
) (CausalGraph, error) {
	return t.semantic.TraceEffectsGraph(
		ctx,
		nodeID,
		maxDepth,
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

// TraceSourceEffects walks forward from the Source identified by
// sourceID through its downstream effects (Observation, Fact,
// Decision, Action), up to maxDepth hops, and returns the visited
// nodes in breadth-first order. It is the forward-tracing counterpart
// to TraceActionCause: instead of "why did this action happen?", it
// answers "what did this source go on to influence?".
func (t *Trace) TraceSourceEffects(
	ctx context.Context,
	sourceID string,
	maxDepth int,
) ([]Visit, error) {
	return t.semantic.TraceSourceEffects(
		ctx,
		sourceID,
		maxDepth,
	)
}

// Explain walks backward from the Action identified by actionID and
// renders every causal relationship it finds as a human-readable
// sentence, suitable for logging, debugging, or audit output. Retracted
// nodes are annotated. See ExplainGraph to explain a graph obtained at
// a specific point in time.
func (t *Trace) Explain(
	ctx context.Context,
	actionID string,
	maxDepth int,
) (string, error) {
	return t.semantic.Explain(
		ctx,
		actionID,
		maxDepth,
	)
}

// TraceActionCauseGraph walks backward from an action like
// TraceActionCause, but returns every causal relationship between the
// reached nodes, not only the first edge each node was reached
// through, plus any retractions recorded for those nodes.
func (t *Trace) TraceActionCauseGraph(
	ctx context.Context,
	actionID string,
	maxDepth int,
) (CausalGraph, error) {
	return t.semantic.TraceActionCauseGraph(
		ctx,
		actionID,
		maxDepth,
	)
}

// TraceActionCauseGraphAt is like TraceActionCauseGraph but
// reconstructs the graph as it was known at at, including only
// retractions that were recorded and in effect by then.
func (t *Trace) TraceActionCauseGraphAt(
	ctx context.Context,
	actionID string,
	at time.Time,
	maxDepth int,
) (CausalGraph, error) {
	return t.semantic.TraceActionCauseGraphAt(
		ctx,
		actionID,
		at,
		maxDepth,
	)
}

// TraceActionCauseAtActionTime answers "what did the agent know when
// it acted?" by reconstructing the causal graph as of the action's own
// RecordedAt. Link the action's decisions with WithCauses when
// recording it; a later Caused call is knowledge acquired after the
// action and is, correctly, not visible.
func (t *Trace) TraceActionCauseAtActionTime(
	ctx context.Context,
	actionID string,
	maxDepth int,
) (CausalGraph, error) {
	return t.semantic.TraceActionCauseAtActionTime(
		ctx,
		actionID,
		maxDepth,
	)
}

// Retract records that a node, typically a Fact or Observation, should
// no longer be relied upon. Nothing recorded is modified: traces as of
// an earlier time still show what was believed then, and current
// traces and explanations mark the node as retracted. Retrying an
// identical retraction is safe; a conflicting one returns an error
// wrapping semantic.ErrAlreadyRetracted.
func (t *Trace) Retract(
	ctx context.Context,
	retraction Retraction,
) (Retraction, error) {
	return t.semantic.Retract(
		ctx,
		retraction,
	)
}

// FindNodes lists nodes matching filter, for queries other than
// direct ID lookup or causal traversal, such as "every Decision
// recorded in the last hour." It returns an error wrapping
// graph.ErrStoreNotQueryable if the configured store does not support
// querying; the default in-memory store does.
func (t *Trace) FindNodes(
	ctx context.Context,
	filter NodeFilter,
) ([]Node, error) {
	return t.semantic.FindNodes(
		ctx,
		filter,
	)
}
