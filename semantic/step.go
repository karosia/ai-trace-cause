package semantic

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/karosia/ai-trace-cause/graph"
)

// Run groups everything an agent records for one task or
// conversation. It is a lightweight handle: creating one writes
// nothing. Every entity committed through one of its Steps carries the
// run's ID (Entity.RunID), so a whole run can be listed with FindNodes.
type Run struct {
	service *Service
	id      string
}

// Run returns a handle for the run identified by runID, typically the
// agent's task, session, or conversation ID. Calling Run again with
// the same ID continues the same run. An empty runID is reported by
// Step.Commit as ErrEmptyRunID.
func (s *Service) Run(runID string) *Run {
	return &Run{service: s, id: runID}
}

// ID returns the run's ID.
func (r *Run) ID() string {
	return r.id
}

// FindNodes is like Service.FindNodes restricted to this run's nodes.
func (r *Run) FindNodes(
	ctx context.Context,
	filter graph.NodeFilter,
) ([]graph.Node, error) {
	properties := maps.Clone(filter.Properties)
	if properties == nil {
		properties = make(map[string]any, 1)
	}

	properties[PropertyRunID] = r.id
	filter.Properties = properties

	return r.service.FindNodes(ctx, filter)
}

// Step starts recording one agent step, typically one model call and
// what it led to. stepID identifies the step, e.g. the model
// response's ID; if empty, one is generated.
//
// A Step buffers entities and relationships in memory and writes them
// all at once in Commit, so a step is recorded completely or (with an
// atomic store) not at all, and all its entities and relationships
// share one RecordedAt. Builder methods return the new entity's ID
// immediately, so it can be used as a cause of later entities in the
// same step. The first error is kept and returned by Commit, so
// building code needs no error checks between calls.
//
// A Step is safe for concurrent use. It cannot be reused after a
// successful Commit.
func (r *Run) Step(stepID string) *Step {
	step := &Step{
		service: r.service,
		runID:   r.id,
		known:   make(map[string]NodeType),
	}

	id, err := r.service.resolveID(stepID)
	if err != nil {
		step.err = fmt.Errorf("generating step ID: %w", err)
	}

	step.id = id

	return step
}

// Step is a buffered batch of entities and relationships; see
// Run.Step.
type Step struct {
	service *Service
	runID   string
	id      string

	mu        sync.Mutex
	nodes     []graph.Node
	known     map[string]NodeType
	links     []pendingEdge
	err       error
	committed bool
}

// StepResult describes a committed Step.
type StepResult struct {
	RunID      string    `json:"runId"`
	StepID     string    `json:"stepId"`
	RecordedAt time.Time `json:"recordedAt"`

	// NodeIDs lists the recorded entities in the order they were
	// added.
	NodeIDs []string `json:"nodeIds"`

	// EdgeIDs lists the relationships written by this step. A
	// relationship that already existed is not rewritten and not
	// listed.
	EdgeIDs []string `json:"edgeIds"`
}

// ID returns the step's ID.
func (st *Step) ID() string {
	return st.id
}

// Err returns the first error encountered while building the step, if
// any. Commit returns the same error.
func (st *Step) Err() error {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.err
}

// Source adds a Source and returns its ID.
func (st *Step) Source(source Source) string {
	node, err := source.toNode()

	return st.add(node, err, nil)
}

// Observation adds an Observation, linked from causeIDs (Sources or
// Actions, here or already recorded), and returns its ID.
func (st *Step) Observation(
	observation Observation,
	causeIDs ...string,
) string {
	node, err := observation.toNode()

	return st.add(node, err, causeIDs)
}

// Fact adds a Fact, linked from causeIDs (Observations or premise
// Facts), and returns its ID.
func (st *Step) Fact(fact Fact, causeIDs ...string) string {
	node, err := fact.toNode()

	return st.add(node, err, causeIDs)
}

// Decision adds a Decision, linked from causeIDs (Facts or a parent
// Decision), and returns its ID.
func (st *Step) Decision(
	decision Decision,
	causeIDs ...string,
) string {
	node, err := decision.toNode()

	return st.add(node, err, causeIDs)
}

// Action adds an Action, linked from causeIDs (Decisions), and returns
// its ID.
func (st *Step) Action(action Action, causeIDs ...string) string {
	node, err := action.toNode()

	return st.add(node, err, causeIDs)
}

// Link adds a relationship with an explicit relation type, for causes
// whose relation the schema cannot infer, or for linking two entities
// that were both recorded earlier.
func (st *Step) Link(
	fromID string,
	relation RelationType,
	toID string,
) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if !st.usableLocked() {
		return
	}

	st.links = append(st.links, pendingEdge{
		from:     fromID,
		to:       toID,
		relation: relation,
	})
}

func (st *Step) add(
	node graph.Node,
	buildErr error,
	causeIDs []string,
) string {
	st.mu.Lock()
	defer st.mu.Unlock()

	if !st.usableLocked() {
		return ""
	}

	if buildErr != nil {
		st.err = buildErr

		return ""
	}

	id, err := st.service.resolveID(node.ID)
	if err != nil {
		st.err = err

		return ""
	}

	if _, duplicate := st.known[id]; duplicate {
		st.err = fmt.Errorf(
			"%w: %s added twice to step %s",
			graph.ErrNodeAlreadyExists,
			id,
			st.id,
		)

		return ""
	}

	node.ID = id
	node.Properties[PropertyRunID] = st.runID
	node.Properties[PropertyStepID] = st.id

	st.nodes = append(st.nodes, node)
	st.known[id] = NodeType(node.Type)

	for _, causeID := range causeIDs {
		st.links = append(st.links, pendingEdge{
			from: causeID,
			to:   id,
		})
	}

	return id
}

func (st *Step) usableLocked() bool {
	if st.committed {
		st.err = ErrStepCommitted
	}

	return st.err == nil
}

// Commit validates the step and writes all of its entities and
// relationships in one batch: atomically if the Store implements
// graph.AtomicBatchStore (the built-in memory store does). Telemetry
// is taken from ctx.
//
// It returns the first error from building the step, ErrEmptyRunID,
// ErrStepCommitted, or a schema or store error. Nothing is written if
// validation fails. After a failed Commit caused by the store, the
// step may be committed again.
func (st *Step) Commit(ctx context.Context) (StepResult, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	if st.committed {
		return StepResult{}, ErrStepCommitted
	}

	if st.err != nil {
		return StepResult{}, st.err
	}

	if st.runID == "" {
		return StepResult{}, ErrEmptyRunID
	}

	edges, err := st.service.resolveEdges(
		ctx,
		st.links,
		maps.Clone(st.known),
	)
	if err != nil {
		return StepResult{}, err
	}

	edges, err = st.withoutExistingEdges(ctx, edges)
	if err != nil {
		return StepResult{}, err
	}

	recordedAt := st.service.now().UTC()

	nodes := make([]graph.Node, len(st.nodes))
	nodeIDs := make([]string, len(st.nodes))

	for i, node := range st.nodes {
		node.RecordedAt = recordedAt
		st.service.applyTelemetry(ctx, &node)

		nodes[i] = node
		nodeIDs[i] = node.ID
	}

	edgeIDs := make([]string, len(edges))

	for i := range edges {
		edges[i].RecordedAt = recordedAt
		st.service.applyEdgeTelemetry(ctx, &edges[i])
		edgeIDs[i] = edges[i].ID
	}

	if err := st.service.graph.AddBatch(ctx, nodes, edges); err != nil {
		return StepResult{}, err
	}

	st.committed = true

	return StepResult{
		RunID:      st.runID,
		StepID:     st.id,
		RecordedAt: recordedAt,
		NodeIDs:    nodeIDs,
		EdgeIDs:    edgeIDs,
	}, nil
}

// withoutExistingEdges drops relationships between two previously
// recorded nodes that already exist, so re-linking them is a no-op as
// it is with Service.Link. Edges touching a node of this step are new
// by construction.
func (st *Step) withoutExistingEdges(
	ctx context.Context,
	edges []graph.Edge,
) ([]graph.Edge, error) {
	kept := edges[:0]

	for _, edge := range edges {
		_, fromInStep := st.known[edge.From]
		_, toInStep := st.known[edge.To]

		if !fromInStep && !toInStep {
			_, err := st.service.graph.GetEdge(ctx, edge.ID)
			if err == nil {
				continue
			}

			if !errors.Is(err, graph.ErrEdgeNotFound) {
				return nil, err
			}
		}

		kept = append(kept, edge)
	}

	return kept, nil
}
