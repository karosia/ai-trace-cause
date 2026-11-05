package semantic

import (
	"time"

	"github.com/karosia/ai-trace-cause/graph"
)

// NodeType identifies which stage of the causal chain a graph.Node
// represents.
type NodeType string

const (
	NodeTypeSource      NodeType = "Source"
	NodeTypeObservation NodeType = "Observation"
	NodeTypeFact        NodeType = "Fact"
	NodeTypeDecision    NodeType = "Decision"
	NodeTypeAction      NodeType = "Action"
)

// RelationType identifies which causal relationship a graph.Edge
// represents.
type RelationType string

const (
	// RelationProduced connects a Source to an Observation it
	// produced.
	RelationProduced RelationType = "PRODUCED"
	// RelationSupports connects an Observation to a Fact it
	// supports.
	RelationSupports RelationType = "SUPPORTS"
	// RelationBasisOf connects a Fact to a Decision it was a basis
	// for.
	RelationBasisOf RelationType = "BASIS_OF"
	// RelationCaused connects a Decision to the Action it caused.
	RelationCaused RelationType = "CAUSED"

	// RelationRetracted is a self-loop annotation edge recording that
	// a node was retracted. It is not a causal relationship and never
	// appears in CausalGraph.Edges; see Service.Retract.
	RelationRetracted RelationType = "RETRACTED"
)

// Source represents where a piece of information originated, such as
// an API response, user input, a document, or a metrics system. Kind
// must be non-empty.
type Source struct {
	ID string `json:"id,omitempty"`

	Kind string `json:"kind"`
	URI  string `json:"uri,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`

	Validity Validity `json:"validity,omitzero"`

	// RecordedAt is set by the Service when the entity is recorded;
	// any value supplied by the caller is ignored.
	RecordedAt time.Time `json:"recordedAt,omitzero"`
}

// Observation represents something observed from an external Source.
// Name must be non-empty.
type Observation struct {
	ID string `json:"id,omitempty"`

	Name  string `json:"name"`
	Value any    `json:"value,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`

	Validity Validity `json:"validity,omitzero"`

	// RecordedAt is set by the Service when the entity is recorded;
	// any value supplied by the caller is ignored.
	RecordedAt time.Time `json:"recordedAt,omitzero"`
}

// Fact represents information accepted as evidence for a Decision.
// Statement must be non-empty and Confidence must be within [0, 1].
type Fact struct {
	ID string `json:"id,omitempty"`

	Statement string `json:"statement"`

	Confidence float64 `json:"confidence"`

	Metadata map[string]any `json:"metadata,omitempty"`

	Validity Validity `json:"validity,omitzero"`

	// RecordedAt is set by the Service when the entity is recorded;
	// any value supplied by the caller is ignored.
	RecordedAt time.Time `json:"recordedAt,omitzero"`
}

// Decision represents a selected outcome or judgment made by the
// agent. Outcome must be non-empty and Confidence must be within
// [0, 1]. Rationale is intended for concise, explicit justification
// and is not intended to store private model chain-of-thought.
type Decision struct {
	ID string `json:"id,omitempty"`

	Outcome    string  `json:"outcome"`
	Rationale  string  `json:"rationale,omitempty"`
	Confidence float64 `json:"confidence"`

	Metadata map[string]any `json:"metadata,omitempty"`

	Validity Validity `json:"validity,omitzero"`

	// RecordedAt is set by the Service when the entity is recorded;
	// any value supplied by the caller is ignored.
	RecordedAt time.Time `json:"recordedAt,omitzero"`
}

// Action represents something the agent actually executed or
// attempted to execute. Name must be non-empty.
type Action struct {
	ID string `json:"id,omitempty"`

	Name   string `json:"name"`
	Target string `json:"target,omitempty"`

	Parameters map[string]any `json:"parameters,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`

	Validity Validity `json:"validity,omitzero"`

	// RecordedAt is set by the Service when the entity is recorded;
	// any value supplied by the caller is ignored.
	RecordedAt time.Time `json:"recordedAt,omitzero"`
}

// Validity describes the half-open interval [ValidFrom, ValidUntil)
// during which an entity or relationship is valid in the modeled
// domain, as distinct from when it was recorded. Either or both
// fields may be nil, meaning unbounded in that direction.
type Validity struct {
	ValidFrom  *time.Time `json:"validFrom,omitempty"`
	ValidUntil *time.Time `json:"validUntil,omitempty"`
}

// Retraction records that a previously recorded node (typically a Fact
// or Observation) should no longer be relied upon. Recorded entities
// are never modified or deleted; a retraction is appended alongside
// them so that historical traces can still show what was believed at
// the time.
type Retraction struct {
	// NodeID identifies the retracted node. It is required.
	NodeID string `json:"nodeId"`

	// Reason explains why the node was retracted. It is required.
	Reason string `json:"reason"`

	// EffectiveAt is when the retraction takes effect in the modeled
	// domain. It may be earlier than RecordedAt for a retroactive
	// correction ("this reading was wrong since 09:00"). Zero means
	// the time of recording.
	EffectiveAt time.Time `json:"effectiveAt"`

	// RecordedAt is set by the Service when the retraction is
	// recorded; any value supplied by the caller is ignored.
	RecordedAt time.Time `json:"recordedAt,omitzero"`
}

// CausalGraph is the causal neighborhood of a node: every node
// reached by a trace, every causal relationship between them, and any
// retractions known for those nodes.
type CausalGraph struct {
	// Visits holds the reached nodes in breadth-first order. Each
	// Visit records only the first edge its node was reached through;
	// use Edges for the complete set of relationships.
	Visits []graph.Visit `json:"visits"`

	// Edges holds every causal relationship between nodes in Visits.
	// Edge.From is always the cause and Edge.To the effect.
	Edges []graph.Edge `json:"edges"`

	// Retractions maps a node ID to its retraction, for nodes in
	// Visits that were retracted as of the trace's point in time.
	Retractions map[string]Retraction `json:"retractions,omitempty"`
}
