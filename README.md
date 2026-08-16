# ai-trace-cause

**A Go-native foundation for tracing the causes behind AI agent decisions and actions.**

`ai-trace-cause` is an open-source Go library for building observable, traceable, and explainable AI agent systems.

Traditional observability tools are good at answering questions such as:

- What happened?
- Which service or tool was called?
- How long did it take?
- Where did an error occur?

AI agents introduce another important question:

> **Why did the agent do that?**

`ai-trace-cause` provides a semantic and causal graph that connects what an agent observed, what it accepted as evidence, what it decided, what it actually executed, and where the original information came from.

---

## Why ai-trace-cause?

An AI agent may execute a workflow such as:

```text
Read metrics
    ↓
Interpret the system state
    ↓
Make a decision
    ↓
Call a tool
```

Operational tracing may show:

```text
agent.run
├── metrics.fetch
├── llm.decide
└── tool.scale-service
```

But this does not necessarily explain why `tool.scale-service` was called.

`ai-trace-cause` models the semantic chain behind that execution:

```text
Source
  │
  │ PRODUCED
  ▼
Observation
  │
  │ SUPPORTS
  ▼
Fact
  │
  │ BASIS_OF
  ▼
Decision
  │
  │ CAUSED
  ▼
Action
```

This makes it possible to answer questions such as:

- Why did the agent call this tool?
- Which facts influenced this decision?
- Which observations supported those facts?
- Where did the original information come from?
- What did the agent know when the action occurred?
- Which OpenTelemetry trace or span produced this decision?
- Which action resulted from a particular fact or observation?

---

## Core Semantic Model

The current semantic model consists of five primary entity types.

### Source

Represents where information originated.

Examples:

```text
Prometheus
CloudWatch
REST API
Database
Uploaded document
User input
MCP tool
Another agent
```

### Observation

Represents something observed from the external world.

Example:

```text
CPU usage = 94%
```

### Fact

Represents information accepted as usable evidence.

Example:

```text
The service is under high CPU load.
```

### Decision

Represents a judgment or selected outcome.

Example:

```text
Scale the service.
```

### Action

Represents something actually executed.

Example:

```text
scale_service
target = payments-api
replicas = 5
```

Together:

```text
Prometheus
    │
    │ PRODUCED
    ▼
CPU usage = 94%
    │
    │ SUPPORTS
    ▼
CPU usage is high
    │
    │ BASIS_OF
    ▼
Scale the service
    │
    │ CAUSED
    ▼
scale_service(replicas=5)
```

---

## Installation

```bash
go get github.com/yourname/ai-trace-cause
```

Replace `github.com/yourname/ai-trace-cause` with the actual module path.

---

## Quick Start

Create a trace runtime:

```go
package main

import (
	"context"
	"fmt"
	"log"

	aitracecause "github.com/karosia/ai-trace-cause"
)

func main() {
	ctx := context.Background()

	trace, err := aitracecause.New()
	if err != nil {
		log.Fatal(err)
	}

	source := aitracecause.Source{
		ID:   "source-001",
		Kind: "Prometheus",
		URI:  "prometheus://production/cpu_usage",
	}

	observation := aitracecause.Observation{
		ID:    "observation-001",
		Name:  "cpu_usage",
		Value: 94,
		Metadata: map[string]any{
			"unit": "%",
		},
	}

	fact := aitracecause.Fact{
		ID:         "fact-001",
		Statement:  "CPU usage is high",
		Confidence: 0.98,
	}

	decision := aitracecause.Decision{
		ID:         "decision-001",
		Outcome:    "Scale the service",
		Rationale:  "CPU utilization is consistently above 90%",
		Confidence: 0.92,
	}

	action := aitracecause.Action{
		ID:     "action-001",
		Name:   "scale_service",
		Target: "payments-api",
		Parameters: map[string]any{
			"replicas": 5,
		},
	}

	if err := trace.RecordSource(ctx, source); err != nil {
		log.Fatal(err)
	}

	if err := trace.RecordObservation(ctx, observation); err != nil {
		log.Fatal(err)
	}

	if err := trace.RecordFact(ctx, fact); err != nil {
		log.Fatal(err)
	}

	if err := trace.RecordDecision(ctx, decision); err != nil {
		log.Fatal(err)
	}

	if err := trace.RecordAction(ctx, action); err != nil {
		log.Fatal(err)
	}

	if err := trace.Produced(
		ctx,
		"edge-produced",
		source.ID,
		observation.ID,
	); err != nil {
		log.Fatal(err)
	}

	if err := trace.Supports(
		ctx,
		"edge-supports",
		observation.ID,
		fact.ID,
	); err != nil {
		log.Fatal(err)
	}

	if err := trace.BasisOf(
		ctx,
		"edge-basis",
		fact.ID,
		decision.ID,
	); err != nil {
		log.Fatal(err)
	}

	if err := trace.Caused(
		ctx,
		"edge-caused",
		decision.ID,
		action.ID,
	); err != nil {
		log.Fatal(err)
	}

	causes, err := trace.TraceActionCause(
		ctx,
		action.ID,
		4,
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, visit := range causes {
		fmt.Printf(
			"%s %s depth=%d\n",
			visit.Node.Type,
			visit.Node.ID,
			visit.Depth,
		)
	}
}
```

Expected traversal:

```text
Action action-001 depth=0
Decision decision-001 depth=1
Fact fact-001 depth=2
Observation observation-001 depth=3
Source source-001 depth=4
```

---

## Public SDK

The root package provides a simplified API for applications.

Instead of manually creating:

```go
store := memory.New()

g, err := graph.New(store)

service, err := semantic.NewService(g)
```

applications can simply use:

```go
trace, err := aitracecause.New()
```

The root SDK acts as a facade over:

```text
Storage
Graph
Traversal
Semantic model
Temporal context
Telemetry correlation
```

The main public APIs are:

```go
trace.RecordSource(...)
trace.RecordObservation(...)
trace.RecordFact(...)
trace.RecordDecision(...)
trace.RecordAction(...)
```

Relationships:

```go
trace.Produced(...)
trace.Supports(...)
trace.BasisOf(...)
trace.Caused(...)
```

Tracing:

```go
trace.TraceActionCause(...)
trace.TraceActionCauseAt(...)
```

Configuration:

```go
aitracecause.WithStore(...)
aitracecause.WithMemoryStore()
aitracecause.WithClock(...)
aitracecause.WithTelemetryHook(...)
```

---

## Architecture

```text
                     External Application

                              │
                              ▼

                       aitracecause.New()

                              │
                              ▼

                          Trace SDK
                              │
               ┌──────────────┴──────────────┐
               │                             │
               ▼                             ▼

        Semantic Service                 Graph Engine
               │                             │
     ┌─────────┼──────────┐        ┌─────────┴─────────┐
     │         │          │        │                   │
     ▼         ▼          ▼        ▼                   ▼

 Provenance  Causality  Temporal  BFS / DFS           Store

                                                     │
                                                     ▼

                                              Store Interface
                                                     │
                                                     ▼

                                               Memory Store


Optional Telemetry:

OpenTelemetry
      │
      ▼
telemetry/otel
      │
      ▼
TelemetryHook
      │
      ▼
Semantic Service
```

---

## Project Structure

```text
ai-trace-cause/
├── aitracecause.go
├── aitracecause_test.go
├── options.go
├── types.go
│
├── graph/
│   ├── errors.go
│   ├── graph.go
│   ├── graph_test.go
│   ├── store.go
│   ├── temporal.go
│   ├── temporal_test.go
│   ├── traversal.go
│   ├── traversal_test.go
│   └── types.go
│
├── semantic/
│   ├── errors.go
│   ├── service.go
│   ├── service_test.go
│   ├── story_test.go
│   ├── telemetry.go
│   ├── telemetry_test.go
│   └── types.go
│
├── storage/
│   └── memory/
│       ├── store.go
│       └── store_test.go
│
├── telemetry/
│   └── otel/
│       ├── hook.go
│       └── hook_test.go
│
└── go.mod
```

The structure may evolve as additional storage engines and agent integrations are introduced.

---

## Graph Core

The graph package remains generic and independent of AI-specific concepts.

```go
type Node struct {
	ID         string
	Type       string
	Properties map[string]any

	RecordedAt time.Time

	ValidFrom  *time.Time
	ValidUntil *time.Time

	Telemetry *TelemetryRef
}
```

```go
type Edge struct {
	ID         string
	From       string
	To         string
	Type       string
	Properties map[string]any

	RecordedAt time.Time

	ValidFrom  *time.Time
	ValidUntil *time.Time

	Telemetry *TelemetryRef
}
```

The graph layer does not need to understand concepts such as:

```text
Observation
Fact
Decision
Action
```

Those rules belong to the semantic layer.

---

## Storage Abstraction

Persistence is defined through a storage interface.

```go
type Store interface {
	PutNode(ctx context.Context, node Node) error
	GetNode(ctx context.Context, id string) (Node, error)

	PutEdge(ctx context.Context, edge Edge) error
	GetEdge(ctx context.Context, id string) (Edge, error)

	OutgoingEdges(
		ctx context.Context,
		nodeID string,
	) ([]Edge, error)

	IncomingEdges(
		ctx context.Context,
		nodeID string,
	) ([]Edge, error)
}
```

This keeps the graph independent from a specific database.

Current implementation:

```text
Graph
  │
  ▼
Store Interface
  ▲
  │
Memory Store
```

Future storage implementations may include:

```text
PostgreSQL
Neo4j
Embedded persistent stores
Distributed graph stores
```

---

## In-Memory Store

The built-in in-memory store maintains:

```text
nodes

NodeID
  ↓
Node
```

```text
edges

EdgeID
  ↓
Edge
```

and relationship indexes:

```text
outgoing

NodeID
  ↓
Edge IDs
```

```text
incoming

NodeID
  ↓
Edge IDs
```

The indexes allow traversal without scanning every edge.

The memory store uses:

```go
sync.RWMutex
```

for concurrent access.

---

## Graph Traversal

The graph engine currently supports:

- Breadth-First Search
- Depth-First Search
- Incoming traversal
- Outgoing traversal
- Maximum depth
- Cycle protection
- Parent tracking
- Edge tracking
- Deterministic ordering

Example:

```go
visits, err := g.BFS(
	ctx,
	"action-001",
	graph.DirectionIncoming,
	4,
)
```

Given:

```text
Source
  ↓
Observation
  ↓
Fact
  ↓
Decision
  ↓
Action
```

incoming traversal returns:

```text
Action        depth 0
Decision      depth 1
Fact          depth 2
Observation   depth 3
Source        depth 4
```

Each traversal result contains additional information:

```go
type Visit struct {
	Node Node

	Depth int

	ParentNodeID string
	ViaEdgeID    string
}
```

This preserves information about how each node was reached.

---

## Semantic Relationships

The semantic layer currently defines four causal relationships.

### PRODUCED

```text
Source
  │
  │ PRODUCED
  ▼
Observation
```

### SUPPORTS

```text
Observation
  │
  │ SUPPORTS
  ▼
Fact
```

### BASIS_OF

```text
Fact
  │
  │ BASIS_OF
  ▼
Decision
```

### CAUSED

```text
Decision
  │
  │ CAUSED
  ▼
Action
```

The full causal flow is:

```text
Source
  │
  ▼
Observation
  │
  ▼
Fact
  │
  ▼
Decision
  │
  ▼
Action
```

The semantic layer validates these relationships before storing them.

For example:

```text
Observation ──SUPPORTS──> Fact
```

is valid.

But:

```text
Fact ──SUPPORTS──> Action
```

is rejected by the semantic layer.

---

## Trace Why an Action Happened

The primary causal tracing API is:

```go
visits, err := trace.TraceActionCause(
	ctx,
	actionID,
	4,
)
```

The traversal starts from the action and follows incoming causal relationships.

```text
Action
  ↑
Decision
  ↑
Fact
  ↑
Observation
  ↑
Source
```

This allows applications to reconstruct the evidence chain behind an action.

---

## Temporal Context

AI decisions often need to be explained using the information available **at the time the decision was made**.

`ai-trace-cause` therefore distinguishes between:

```text
RecordedAt
```

and:

```text
ValidFrom
ValidUntil
```

### RecordedAt

Represents when the information became known to `ai-trace-cause`.

Example:

```text
09:00
Subscription actually expired.

09:05
The agent discovered that the subscription expired.
```

The fact may have:

```text
ValidFrom  = 09:00
RecordedAt = 09:05
```

At `09:03`, the fact was true in the external world, but the agent did not know it yet.

### Validity

Validity uses a half-open interval:

```text
[ValidFrom, ValidUntil)
```

which means:

```text
ValidFrom <= t < ValidUntil
```

---

## Temporal Traversal

The graph can reconstruct causal context at a particular time.

Graph-level API:

```go
visits, err := g.BFSAt(
	ctx,
	"decision-001",
	graph.DirectionIncoming,
	3,
	at,
)
```

SDK-level API:

```go
visits, err := trace.TraceActionCauseAt(
	ctx,
	action.ID,
	at,
	4,
)
```

This allows applications to answer:

> Why did the agent perform this action based on what it knew at that time?

---

## OpenTelemetry Integration

`ai-trace-cause` is not a replacement for OpenTelemetry.

The two systems answer different questions.

```text
OpenTelemetry                 ai-trace-cause
────────────────────          ────────────────────

HOW did it execute?           WHY did it happen?

Which service?                Which observation?
Which span?                   Which fact?
How long?                     Which decision?
Which error?                  Which source?
Which tool call?              Why this action?
```

The two systems can be correlated through:

```text
TraceID
SpanID
```

---

## Using OpenTelemetry

Add the OpenTelemetry dependency:

```bash
go get go.opentelemetry.io/otel
```

Create the adapter:

```go
hook := oteltelemetry.New()
```

Then configure `ai-trace-cause`:

```go
trace, err := aitracecause.New(
	aitracecause.WithTelemetryHook(
		hook,
	),
)
```

If the supplied `context.Context` contains an active OpenTelemetry `SpanContext`, the corresponding semantic node or edge automatically records:

```text
TraceID
SpanID
```

---

## OpenTelemetry Example

Assume an application already has a span:

```go
ctx, span := tracer.Start(
	ctx,
	"agent.run",
)
defer span.End()
```

Use the same context when recording a decision:

```go
err := trace.RecordDecision(
	ctx,
	aitracecause.Decision{
		ID:         "decision-001",
		Outcome:    "Scale the service",
		Rationale:  "CPU utilization is too high",
		Confidence: 0.92,
	},
)
```

Internally, the Decision may become correlated with:

```text
Decision
├── TraceID
└── SpanID
```

The application does not need to pass those IDs manually.

`ai-trace-cause` reads the active trace context but does not own or terminate application spans.

---

## HOW and WHY Together

An AI agent execution may look like:

```text
OpenTelemetry

Trace abc123

agent.run
├── metrics.fetch
├── llm.decide
└── tool.scale-service
```

The corresponding semantic trace may look like:

```text
ai-trace-cause

Observation
TraceID = abc123
SpanID  = metrics.fetch
      │
      ▼
Fact
TraceID = abc123
SpanID  = llm.decide
      │
      ▼
Decision
TraceID = abc123
SpanID  = llm.decide
      │
      ▼
Action
TraceID = abc123
SpanID  = tool.scale-service
```

Together:

```text
Operational Trace
       +
Causal Trace
       =
Explainable Agent Execution
```

---

## Provider Agnostic by Design

The core library does not call an LLM.

It does not require:

```text
OpenAI
Anthropic
Gemini
Ollama
LangChain
MCP
```

The intended architecture is:

```text
AI Application
      │
      │ OpenAI / Anthropic / Gemini / Custom Agent
      ▼

ai-trace-cause
      │
      ▼

Semantic / Causal Graph
```

The application decides which events represent:

```text
Source
Observation
Fact
Decision
Action
```

and records those relationships through the SDK.

Provider-specific integrations can be added independently without coupling the core runtime to a particular AI provider.

---

## Why Not Store Everything in OpenTelemetry Spans?

Operational execution structure and semantic causality are not always the same graph.

An OpenTelemetry trace may look like:

```text
agent.run
├── retrieve
├── model.generate
└── tool.call
```

But semantic causality might look like:

```text
Source A ───> Observation A ───┐
                               │
                               ▼
                              Fact A ───┐
                                        │
Source B ───> Observation B ───> Fact B ─┼──> Decision
                                        │
                              Fact C ───┘
                                        │
                                        ▼
                                      Action
```

These structures serve different purposes.

`ai-trace-cause` preserves the semantic graph independently while allowing it to be correlated with operational telemetry.

---

## Configuration

### Default Memory Store

The simplest setup is:

```go
trace, err := aitracecause.New()
```

The default configuration uses the built-in memory store.

Equivalent explicit configuration:

```go
trace, err := aitracecause.New(
	aitracecause.WithMemoryStore(),
)
```

### Custom Store

Custom storage implementations can satisfy the `graph.Store` interface.

```go
trace, err := aitracecause.New(
	aitracecause.WithStore(customStore),
)
```

This allows the public SDK to remain unchanged when adding future database backends.

### Custom Clock

A clock can be injected for deterministic testing.

```go
trace, err := aitracecause.New(
	aitracecause.WithClock(
		func() time.Time {
			return fixedTime
		},
	),
)
```

### Telemetry Hook

Optional telemetry correlation can be configured through:

```go
trace, err := aitracecause.New(
	aitracecause.WithTelemetryHook(hook),
)
```

The core SDK remains independent from a specific telemetry provider.

---

## Concurrency

The current in-memory backend uses:

```go
sync.RWMutex
```

to protect graph state.

This allows multiple agent executions to concurrently record and inspect graph data.

Run Go's race detector during development:

```bash
go test -race ./...
```

---

## Testing

Run all tests:

```bash
go test ./...
```

Run all tests with race detection:

```bash
go test -race ./...
```

Package-specific tests:

```bash
go test ./graph/...
go test ./semantic/...
go test ./storage/...
go test ./telemetry/...
```

---

## Future Direction

### Persistent Storage

Possible storage implementations:

```text
PostgreSQL
Neo4j
Embedded databases
Distributed graph stores
```

### AI Provider Integrations

Possible adapters:

```text
OpenAI
Anthropic
Gemini
Local models
Custom agent runtimes
```

The core will remain provider independent.

### Automatic Tool Tracing

Future integrations may capture:

```text
Tool request
Tool response
Tool failure
Retry
```

and connect them automatically to semantic actions.

### Automatic LLM Tracing

Provider integrations may capture:

```text
Model request
Model response
Usage metadata
Latency
Model name
```

while allowing applications to explicitly define semantic decisions and evidence.

### Advanced Causal Queries

Possible APIs:

```go
trace.ExplainAction(...)
trace.TraceDecision(...)
trace.FindSources(...)
trace.FindInfluences(...)
trace.FindConsequences(...)
```

### Distributed Causal Tracing

OpenTelemetry context can eventually help correlate causal graph entities across:

```text
Agent Service
    │
    │ HTTP / gRPC
    ▼
Tool Service
    │
    ▼
Another Agent
```

### Visualization

Future visualization could represent graphs such as:

```text
Source A ──> Observation A ──> Fact A ──┐
                                        │
Source B ──> Observation B ──> Fact B ──┼──> Decision ──> Action
                                        │
Source C ──> Observation C ──> Fact C ──┘
```

---

## Design Principles

### Keep the graph generic

The graph package should not know what an AI agent, fact, decision, or action is.

### Keep semantic rules explicit

Relationships such as:

```text
Observation → Fact
Fact → Decision
Decision → Action
```

belong to the semantic layer.

### Keep the public API small

Applications should primarily interact with:

```go
aitracecause.New()
```

and the resulting `Trace` SDK.

Internal graph and storage concepts should remain optional for advanced users.

### Keep providers optional

The core must not require a specific AI provider SDK.

### Keep storage replaceable

The graph engine must not depend on a particular persistence implementation.

### Keep telemetry optional

Applications should be able to use `ai-trace-cause` with or without OpenTelemetry.

### Preserve causality

The system should make it possible to trace:

```text
Action
  ↑
Decision
  ↑
Fact
  ↑
Observation
  ↑
Source
```

### Preserve historical context

An explanation should reflect what the agent knew when the decision was made, not only what is known now.

### Preserve operational correlation

Semantic entities should be correlatable with the execution traces that created them.

---

## Vision

Operational observability answers:

> **How did the agent execute?**

`ai-trace-cause` aims to answer:

> **Why did the agent make that decision and perform that action?**

Together with OpenTelemetry, the goal is to provide a foundation for AI agent systems that are:

```text
Observable
Traceable
Auditable
Explainable
```

without coupling the core runtime to a specific AI provider, storage engine, or observability backend.
