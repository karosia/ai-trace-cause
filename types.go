package aitracecause

import (
	"github.com/karosia/ai-trace-cause/graph"
	"github.com/karosia/ai-trace-cause/semantic"
)

// Source represents where a piece of information originated, such as
// an API response, user input, a document, or a metrics system.
type Source = semantic.Source

// Observation represents something observed from an external Source.
type Observation = semantic.Observation

// Fact represents information accepted as evidence for a Decision.
type Fact = semantic.Fact

// Decision represents a selected outcome or judgment made by the
// agent.
type Decision = semantic.Decision

// Action represents something the agent actually executed or
// attempted to execute.
type Action = semantic.Action

// Validity describes the half-open interval [ValidFrom, ValidUntil)
// during which an entity or relationship is considered valid in the
// modeled domain.
type Validity = semantic.Validity

// Retraction records that a node should no longer be relied upon,
// without modifying or deleting it. See Trace.Retract.
type Retraction = semantic.Retraction

// CausalGraph is every node reached by a causal trace, every causal
// relationship between them, and any retractions for those nodes.
type CausalGraph = semantic.CausalGraph

// RecordOption configures a single Record* call. See WithCauses.
type RecordOption = semantic.RecordOption

// WithCauses links the entity being recorded to its direct causes in
// the same call: an Observation to Source(s), a Fact to
// Observation(s), a Decision to Fact(s), or an Action to Decision(s).
// Causes are validated before anything is written, and the
// relationships share the entity's RecordedAt.
func WithCauses(causeIDs ...string) RecordOption {
	return semantic.WithCauses(causeIDs...)
}

// WithCausesVia is like WithCauses but names the relation explicitly,
// for cause types the schema links by more than one relation.
func WithCausesVia(
	relation RelationType,
	causeIDs ...string,
) RecordOption {
	return semantic.WithCausesVia(relation, causeIDs...)
}

// NodeType identifies an entity type in the causal model.
type NodeType = semantic.NodeType

// RelationType identifies a relationship type in the causal model.
type RelationType = semantic.RelationType

// Entity types.
const (
	NodeTypeSource      = semantic.NodeTypeSource
	NodeTypeObservation = semantic.NodeTypeObservation
	NodeTypeFact        = semantic.NodeTypeFact
	NodeTypeDecision    = semantic.NodeTypeDecision
	NodeTypeAction      = semantic.NodeTypeAction
)

// Relationship types in DefaultSchema. Each points from cause to
// effect.
const (
	RelationProduced   = semantic.RelationProduced
	RelationResultedIn = semantic.RelationResultedIn
	RelationSupports   = semantic.RelationSupports
	RelationInfers     = semantic.RelationInfers
	RelationBasisOf    = semantic.RelationBasisOf
	RelationRefines    = semantic.RelationRefines
	RelationCaused     = semantic.RelationCaused
)

// Schema is the set of relationships a Trace accepts. See
// DefaultSchema, NewSchema, and Schema.With.
type Schema = semantic.Schema

// RelationRule allows one relationship type between two entity types.
type RelationRule = semantic.RelationRule

// DefaultSchema returns the built-in causal model, including the agent
// loop (Action RESULTED_IN Observation), inference between Facts, and
// refinement between Decisions.
func DefaultSchema() Schema {
	return semantic.DefaultSchema()
}

// NewSchema returns a Schema allowing exactly rules.
func NewSchema(rules ...RelationRule) (Schema, error) {
	return semantic.NewSchema(rules...)
}

// Run groups everything an agent records for one task or
// conversation. See Trace.Run.
type Run = semantic.Run

// Step buffers one agent step's entities and relationships and writes
// them together on Commit. See Run.Step.
type Step = semantic.Step

// StepResult describes a committed Step.
type StepResult = semantic.StepResult

// ExplainGraph renders a CausalGraph as one plain-English sentence per
// causal relationship, annotating retracted nodes. Use it with
// Trace.TraceActionCauseAtActionTime to explain an action as it was
// understood when it happened.
func ExplainGraph(causal CausalGraph) string {
	return semantic.ExplainGraph(causal)
}

// Visit is a single node visited while tracing a causal chain, along
// with the depth at which it was found and the edge it was reached
// through.
type Visit = graph.Visit

// Node is a single entity in the underlying causal graph, as returned
// by FindNodes and embedded in Visit.
type Node = graph.Node

// NodeFilter narrows the nodes returned by FindNodes. A zero-value
// field imposes no constraint on that dimension, so the zero-value
// NodeFilter matches every node.
type NodeFilter = graph.NodeFilter

// TelemetryHook correlates a context with an active OpenTelemetry
// trace and span, so recorded entities and relationships can be tied
// back to the distributed trace that produced them.
type TelemetryHook = semantic.TelemetryHook

// IDGenerator generates unique IDs for entities and relationships
// that are recorded without a caller-provided ID.
type IDGenerator = semantic.IDGenerator
