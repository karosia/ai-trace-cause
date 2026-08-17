# ai-trace-cause

**Causal tracing for AI agents in Go.**

`ai-trace-cause` is a Go library for recording and tracing **why an AI agent made a decision or performed an action**.

It complements operational observability systems such as OpenTelemetry by adding a semantic causal graph that connects:

```text
Source
  │ PRODUCED
  ▼
Observation
  │ SUPPORTS
  ▼
Fact
  │ BASIS_OF
  ▼
Decision
  │ CAUSED
  ▼
Action
```

This makes it possible to answer questions such as:

- Why did the agent perform this action?
- Which facts were used to make this decision?
- Which observations supported those facts?
- Where did the information originally come from?
- What did the agent know at the time?
- Which OpenTelemetry trace and span produced a decision or action?

---

## Features

- Semantic causal graph for AI agent execution
- Source, Observation, Fact, Decision, and Action entities
- Typed causal relationships
- Automatic UUIDv7 ID generation
- Caller-provided IDs when external IDs already exist
- Causal traversal from Action back to Source
- Breadth-first and depth-first graph traversal
- Cycle protection and maximum traversal depth
- Temporal context with `RecordedAt`, `ValidFrom`, and `ValidUntil`
- Historical causal reconstruction
- Optional OpenTelemetry Trace ID and Span ID correlation
- Pluggable storage through a `graph.Store` interface
- Concurrent-safe in-memory storage
- Provider-agnostic core with no dependency on a specific LLM vendor

---

## Installation

```bash
go get github.com/karosia/ai-trace-cause
```

---

## Quick Start

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

    source, err := trace.RecordSource(
        ctx,
        aitracecause.Source{
            Kind: "Prometheus",
            URI:  "prometheus://production/cpu_usage",
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    observation, err := trace.RecordObservation(
        ctx,
        aitracecause.Observation{
            Name:  "cpu_usage",
            Value: 94,
            Metadata: map[string]any{
                "unit": "%",
            },
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    fact, err := trace.RecordFact(
        ctx,
        aitracecause.Fact{
            Statement:  "CPU usage is high",
            Confidence: 0.98,
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    decision, err := trace.RecordDecision(
        ctx,
        aitracecause.Decision{
            Outcome:    "Scale the service",
            Rationale:  "CPU utilization is above the expected threshold",
            Confidence: 0.92,
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    action, err := trace.RecordAction(
        ctx,
        aitracecause.Action{
            Name:   "scale_service",
            Target: "payments-api",
            Parameters: map[string]any{
                "replicas": 5,
            },
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    if err := trace.Produced(ctx, source.ID, observation.ID); err != nil {
        log.Fatal(err)
    }

    if err := trace.Supports(ctx, observation.ID, fact.ID); err != nil {
        log.Fatal(err)
    }

    if err := trace.BasisOf(ctx, fact.ID, decision.ID); err != nil {
        log.Fatal(err)
    }

    if err := trace.Caused(ctx, decision.ID, action.ID); err != nil {
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
            "type=%s id=%s depth=%d\n",
            visit.Node.Type,
            visit.Node.ID,
            visit.Depth,
        )
    }
}
```

A trace from the action walks backward through the causal chain:

```text
Action        depth=0
Decision      depth=1
Fact          depth=2
Observation   depth=3
Source        depth=4
```

---

## Semantic Model

### Source

Represents where information originated.

Typical sources include:

- API responses
- user input
- documents
- databases
- metrics systems
- tool responses
- other agents

```go
source, err := trace.RecordSource(
    ctx,
    aitracecause.Source{
        Kind: "Prometheus",
        URI:  "prometheus://production/cpu_usage",
    },
)
```

### Observation

Represents something observed from an external source.

```go
observation, err := trace.RecordObservation(
    ctx,
    aitracecause.Observation{
        Name:  "cpu_usage",
        Value: 94,
    },
)
```

### Fact

Represents information accepted as evidence for a decision.

```go
fact, err := trace.RecordFact(
    ctx,
    aitracecause.Fact{
        Statement:  "CPU usage is high",
        Confidence: 0.98,
    },
)
```

### Decision

Represents a selected outcome or judgment.

```go
decision, err := trace.RecordDecision(
    ctx,
    aitracecause.Decision{
        Outcome:    "Scale the service",
        Rationale:  "CPU utilization is above the expected threshold",
        Confidence: 0.92,
    },
)
```

`Rationale` is intended for concise, explicit decision justification. It is not intended to store private model chain-of-thought.

### Action

Represents something the agent actually executed or attempted to execute.

```go
action, err := trace.RecordAction(
    ctx,
    aitracecause.Action{
        Name:   "scale_service",
        Target: "payments-api",
        Parameters: map[string]any{
            "replicas": 5,
        },
    },
)
```

---

## Causal Relationships

The semantic API validates relationship types before they are stored.

### Source → Observation

```go
err := trace.Produced(
    ctx,
    source.ID,
    observation.ID,
)
```

```text
Source ──PRODUCED──> Observation
```

### Observation → Fact

```go
err := trace.Supports(
    ctx,
    observation.ID,
    fact.ID,
)
```

```text
Observation ──SUPPORTS──> Fact
```

### Fact → Decision

```go
err := trace.BasisOf(
    ctx,
    fact.ID,
    decision.ID,
)
```

```text
Fact ──BASIS_OF──> Decision
```

### Decision → Action

```go
err := trace.Caused(
    ctx,
    decision.ID,
    action.ID,
)
```

```text
Decision ──CAUSED──> Action
```

Relationship IDs are generated automatically.

---

## Automatic IDs

Entity and relationship IDs use UUIDv7 by default.

```go
observation, err := trace.RecordObservation(
    ctx,
    aitracecause.Observation{
        Name:  "cpu_usage",
        Value: 94,
    },
)
```

The returned entity contains the generated ID:

```go
fmt.Println(observation.ID)
```

If an application already has a stable external ID, it can provide one explicitly:

```go
observation, err := trace.RecordObservation(
    ctx,
    aitracecause.Observation{
        ID:    externalEventID,
        Name:  "payment_status",
        Value: "failed",
    },
)
```

Provided IDs are preserved.

A custom ID generator can also be supplied:

```go
trace, err := aitracecause.New(
    aitracecause.WithIDGenerator(myGenerator),
)
```

The generator implements:

```go
type IDGenerator interface {
    NewID() (string, error)
}
```

---

## Tracing Causes

Use `TraceActionCause` to walk backward from an action through its causal dependencies.

```go
visits, err := trace.TraceActionCause(
    ctx,
    action.ID,
    4,
)
```

Each visit contains:

```go
type Visit struct {
    Node Node

    Depth int

    ParentNodeID string
    ViaEdgeID    string
}
```

This preserves both the discovered entity and the edge through which it was reached.

```text
Action
  ↑ CAUSED
Decision
  ↑ BASIS_OF
Fact
  ↑ SUPPORTS
Observation
  ↑ PRODUCED
Source
```

---

## Temporal Context

Causal explanations often need to reflect **what the agent knew at the time**, rather than what is known now.

Entities and relationships can carry temporal information:

```text
RecordedAt
ValidFrom
ValidUntil
```

`RecordedAt` represents when the information was recorded by `ai-trace-cause`.

`ValidFrom` and `ValidUntil` represent when the information is valid in the modeled domain.

Validity uses a half-open interval:

```text
[ValidFrom, ValidUntil)
```

For example:

```go
validFrom := time.Date(
    2026,
    time.August,
    17,
    10,
    0,
    0,
    0,
    time.UTC,
)

fact, err := trace.RecordFact(
    ctx,
    aitracecause.Fact{
        Statement:  "The subscription is expired",
        Confidence: 1.0,
        Validity: aitracecause.Validity{
            ValidFrom: &validFrom,
        },
    },
)
```

### Historical Cause Tracing

Use `TraceActionCauseAt` to reconstruct the causal graph at a specific point in time:

```go
visits, err := trace.TraceActionCauseAt(
    ctx,
    action.ID,
    at,
    4,
)
```

Only entities and relationships that were known and valid at `at` are included.

---

## OpenTelemetry Integration

`ai-trace-cause` complements OpenTelemetry rather than replacing it.

OpenTelemetry primarily answers:

```text
How did the agent execute?
```

`ai-trace-cause` focuses on:

```text
Why did the agent make this decision or perform this action?
```

The two can be correlated using the active OpenTelemetry `TraceID` and `SpanID`.

### Setup

Add OpenTelemetry:

```bash
go get go.opentelemetry.io/otel
```

Configure the adapter:

```go
import (
    aitracecause "github.com/karosia/ai-trace-cause"
    oteltelemetry "github.com/karosia/ai-trace-cause/telemetry/otel"
)

trace, err := aitracecause.New(
    aitracecause.WithTelemetryHook(
        oteltelemetry.New(),
    ),
)
```

If `ctx` contains an active OpenTelemetry span, semantic entities and relationships recorded with that context are automatically correlated with its Trace ID and Span ID.

```go
ctx, span := tracer.Start(
    ctx,
    "agent.decide",
)
defer span.End()

decision, err := trace.RecordDecision(
    ctx,
    aitracecause.Decision{
        Outcome:    "Scale the service",
        Rationale:  "CPU usage is high",
        Confidence: 0.92,
    },
)
```

`ai-trace-cause` reads the active span context but does not create, own, or end application spans.

---

## Storage

The public SDK uses the in-memory store by default:

```go
trace, err := aitracecause.New()
```

Equivalent explicit configuration:

```go
trace, err := aitracecause.New(
    aitracecause.WithMemoryStore(),
)
```

For persistent or application-specific storage, provide an implementation of `graph.Store`:

```go
trace, err := aitracecause.New(
    aitracecause.WithStore(customStore),
)
```

The store interface is:

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

The built-in memory store is concurrent-safe and maintains incoming and outgoing relationship indexes for traversal.

---

## Custom Clock

A custom clock can be injected for deterministic tests or controlled environments:

```go
trace, err := aitracecause.New(
    aitracecause.WithClock(
        func() time.Time {
            return fixedTime
        },
    ),
)
```

---

## Low-Level Graph API

Most applications should use the root `aitracecause` package.

The lower-level `graph` package is available for applications that need direct graph operations such as:

- custom traversal
- direct node or edge access
- graph-specific storage implementations
- infrastructure integrations

The graph layer is intentionally generic and does not depend on AI-specific semantic types.

---

## Provider Agnostic

`ai-trace-cause` does not require or directly call a specific model provider.

It can be used with:

```text
OpenAI
Anthropic
Gemini
local models
custom agent runtimes
MCP-based applications
non-LLM decision systems
```

Applications decide what should be represented as Sources, Observations, Facts, Decisions, and Actions.

Provider-specific instrumentation can be added separately without coupling the core graph and semantic model to a particular vendor.

---

## Concurrency

The built-in memory store uses `sync.RWMutex` and supports concurrent access.

Run the race detector when developing or modifying storage and traversal code:

```bash
go test -race ./...
```

---

## Testing

Run the full test suite:

```bash
go test ./...
```

Run with race detection:

```bash
go test -race ./...
```

Package-specific tests:

```bash
go test ./graph/...
go test ./semantic/...
go test ./storage/...
go test ./telemetry/...
go test ./idgen/...
```

---

## Design Goals

`ai-trace-cause` is designed around a few constraints:

- Keep the causal graph separate from operational tracing.
- Keep the graph engine generic.
- Keep semantic relationships explicit and type-checked.
- Keep storage replaceable.
- Keep telemetry optional.
- Keep AI providers optional.
- Preserve historical context.
- Allow semantic entities to correlate with distributed traces.
- Keep the public API small enough to embed in existing agent runtimes.

---

## License

Add the license used by this repository here.
