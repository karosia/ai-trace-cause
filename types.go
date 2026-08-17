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

// Visit is a single node visited while tracing a causal chain, along
// with the depth at which it was found and the edge it was reached
// through.
type Visit = graph.Visit

// TelemetryHook correlates a context with an active OpenTelemetry
// trace and span, so recorded entities and relationships can be tied
// back to the distributed trace that produced them.
type TelemetryHook = semantic.TelemetryHook

// IDGenerator generates unique IDs for entities and relationships
// that are recorded without a caller-provided ID.
type IDGenerator = semantic.IDGenerator
