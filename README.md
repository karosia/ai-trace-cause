# ai-trace-cause

**A Go-native foundation for tracing the causes behind AI agent decisions and actions.**

`ai-trace-cause` is an open-source Go library for building observable and explainable AI agent systems.

The project is designed to capture and connect the semantic relationships between what an AI agent **observed**, what it considered a **fact**, what it **decided**, what **action** it took, and where that information originally came from.

Instead of only answering:

> What happened during an agent execution?

`ai-trace-cause` aims to help answer:

> Why did the agent do that?

---

## Motivation

Traditional observability tools are good at describing how a system executed.

For example, distributed tracing can show:

```text
Agent Run
├── Retrieve Context
├── LLM Call
├── Tool Call
└── Database Write
```

This is useful for understanding:

* latency
* errors
* service dependencies
* tool calls
* execution flow

However, it does not necessarily explain the semantic reasoning behind an AI agent's behavior.

For example:

```text
Why did the agent call this tool?

Which observation influenced this decision?

Which fact was used as evidence?

Where did that fact come from?

What decision caused this action?
```

`ai-trace-cause` is intended to represent those relationships as a directed graph.

---

## Core Idea

An agent execution can be represented as a causal graph.

```text
Source
  │
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

For example:

```text
Prometheus
    │
    ▼
CPU Usage = 94%
    │
    ▼
High CPU Load
    │
    ▼
Scale Service
    │
    ▼
Increase Replicas to 5
```

Each item can be represented as a **Node**, while the relationship between them is represented as an **Edge**.

---

## Project Goals

`ai-trace-cause` is being designed as infrastructure for AI agent observability and explainability.

The long-term goals include:

* graph-based agent execution context
* observation and fact tracking
* decision lineage
* action causality
* provenance tracking
* causal graph traversal
* temporal context
* OpenTelemetry integration
* pluggable storage backends
* AI provider integrations

The core library is intended to remain independent of specific AI providers.

The library itself does not need to directly call OpenAI, Anthropic, Gemini, or another LLM provider.

Instead, an application can use its preferred AI SDK and record relevant execution information through `ai-trace-cause`.

```text
OpenAI
Anthropic
Gemini
Custom Agent
     │
     ▼
Application
     │
     ▼
ai-trace-cause
     │
     ▼
Causal / Decision Graph
```

---

## Current Status

The project is currently in an early development stage.

The following foundations are implemented:

* directed graph representation
* nodes and edges
* node lookup
* edge lookup
* incoming relationship indexes
* outgoing relationship indexes
* neighbor lookup
* storage abstraction
* in-memory storage backend
* concurrent-safe in-memory operations
* context-aware storage API

Higher-level AI concepts such as `Observation`, `Fact`, `Decision`, `Action`, and `Provenance` will be built on top of this graph foundation.

---

## Architecture

The graph layer does not directly depend on a specific database.

Instead, it depends on a small `Store` interface.

```text
                     Graph
                       │
                       ▼
                Store Interface
                       ▲
             ┌─────────┼─────────┐
             │         │         │
             ▼         ▼         ▼
          Memory    PostgreSQL   Neo4j
```

Currently, only the in-memory implementation exists.

This separation allows graph and causal reasoning logic to remain independent from the storage backend.

---

## Project Structure

```text
ai-trace-cause/
├── graph/
│   ├── types.go
│   ├── errors.go
│   ├── store.go
│   ├── graph.go
│   └── graph_test.go
│
├── storage/
│   └── memory/
│       ├── store.go
│       └── store_test.go
│
└── go.mod
```

### `graph`

Contains the graph domain and storage contract.

The graph package defines concepts such as:

```go
type Node struct {
    ID         string
    Type       string
    Properties map[string]any
}

type Edge struct {
    ID         string
    From       string
    To         string
    Type       string
    Properties map[string]any
}
```

It also defines the storage interface used by the graph.

```go
type Store interface {
    PutNode(ctx context.Context, node Node) error
    GetNode(ctx context.Context, id string) (Node, error)

    PutEdge(ctx context.Context, edge Edge) error
    GetEdge(ctx context.Context, id string) (Edge, error)

    OutgoingEdges(ctx context.Context, nodeID string) ([]Edge, error)
    IncomingEdges(ctx context.Context, nodeID string) ([]Edge, error)
}
```

### `storage/memory`

Provides the current in-memory implementation of `graph.Store`.

Internally it maintains:

```text
nodes

Node ID
   ↓
Node
```

```text
edges

Edge ID
   ↓
Edge
```

and indexes for efficient relationship lookup:

```text
outgoing

Node ID
   ↓
Outgoing Edge IDs
```

```text
incoming

Node ID
   ↓
Incoming Edge IDs
```

---

## Basic Usage

Create an in-memory store:

```go
store := memory.New()
```

Create a graph using that store:

```go
g, err := graph.New(store)
if err != nil {
    panic(err)
}
```

Create a context:

```go
ctx := context.Background()
```

Add an observation:

```go
observation := graph.Node{
    ID:   "observation-001",
    Type: "Observation",
    Properties: map[string]any{
        "metric": "cpu_usage",
        "value":  94,
    },
}

if err := g.AddNode(ctx, observation); err != nil {
    panic(err)
}
```

Add a fact:

```go
fact := graph.Node{
    ID:   "fact-001",
    Type: "Fact",
    Properties: map[string]any{
        "statement": "CPU usage is high",
    },
}

if err := g.AddNode(ctx, fact); err != nil {
    panic(err)
}
```

Connect them:

```go
edge := graph.Edge{
    ID:   "edge-001",
    From: observation.ID,
    To:   fact.ID,
    Type: "SUPPORTS",
}

if err := g.AddEdge(ctx, edge); err != nil {
    panic(err)
}
```

The resulting graph is:

```text
Observation
CPU Usage = 94%
      │
      │ SUPPORTS
      ▼
Fact
CPU Usage Is High
```

Outgoing neighbors can then be retrieved with:

```go
neighbors, err := g.OutgoingNeighbors(
    ctx,
    observation.ID,
)
```

---

## Why a Storage Interface?

The graph should not need to know whether its data is stored in:

* Go memory
* PostgreSQL
* Neo4j
* SQLite
* another graph database

Instead, the graph only depends on the `Store` abstraction.

```text
Graph
  │
  ▼
Store
```

A storage implementation is responsible for satisfying that interface.

For example:

```text
MemoryStore
    │
    └── implements graph.Store
```

This makes it possible to add new persistence implementations without changing higher-level graph logic.

---

## Concurrency

The in-memory store uses `sync.RWMutex` to support concurrent access.

This is important because AI agent systems may have many executions writing observations, decisions, and actions concurrently.

Conceptually:

```text
Agent A ──┐
Agent B ──┼──> ai-trace-cause
Agent C ──┘
```

Read operations use a read lock, while write operations use an exclusive lock.

The project can be tested for race conditions with:

```bash
go test -race ./...
```

---

## Development

Run all tests:

```bash
go test ./...
```

Run tests with the Go race detector:

```bash
go test -race ./...
```

---

## Roadmap

The current development plan is:

```text
Step 1  Graph Core
   ↓
Step 2  Storage Abstraction        ← current
   ↓
Step 3  Graph Traversal
   ↓
Step 4  Observation / Fact
   ↓
Step 5  Decision Graph
   ↓
Step 6  Action & Causal Trace
   ↓
Step 7  Provenance
   ↓
Step 8  Temporal Context
   ↓
Step 9  OpenTelemetry Integration
   ↓
Step 10 Agent SDK
```

### Upcoming: Graph Traversal

The next major feature will introduce graph traversal such as BFS and DFS.

This will eventually allow causal chains such as:

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

to be traced programmatically.

---

## Design Principle

`ai-trace-cause` is not intended to become another LLM framework.

The goal is to remain provider-agnostic infrastructure that can be added to existing AI applications.

For example:

```text
Your AI Application
        │
        ├── OpenAI
        ├── Anthropic
        ├── Gemini
        ├── MCP
        └── Custom Tools
                │
                ▼
         ai-trace-cause
                │
                ▼
          Causal Graph
```

The AI application performs the work.

`ai-trace-cause` records the relationships needed to explain **why** that work happened.

---

## Vision

Operational observability answers:

> How did the agent execute?

`ai-trace-cause` aims to answer:

> Why did the agent make that decision?

The long-term goal is to complement traditional observability systems with semantic and causal traces for AI agents.
