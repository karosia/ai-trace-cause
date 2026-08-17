package semantic

import "time"

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
}

// Observation represents something observed from an external Source.
// Name must be non-empty.
type Observation struct {
	ID string `json:"id,omitempty"`

	Name  string `json:"name"`
	Value any    `json:"value,omitempty"`

	Metadata map[string]any `json:"metadata,omitempty"`

	Validity Validity `json:"validity,omitzero"`
}

// Fact represents information accepted as evidence for a Decision.
// Statement must be non-empty and Confidence must be within [0, 1].
type Fact struct {
	ID string `json:"id,omitempty"`

	Statement string `json:"statement"`

	Confidence float64 `json:"confidence"`

	Metadata map[string]any `json:"metadata,omitempty"`

	Validity Validity `json:"validity,omitzero"`
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
}

// Validity describes the half-open interval [ValidFrom, ValidUntil)
// during which an entity or relationship is valid in the modeled
// domain, as distinct from when it was recorded. Either or both
// fields may be nil, meaning unbounded in that direction.
type Validity struct {
	ValidFrom  *time.Time `json:"validFrom,omitempty"`
	ValidUntil *time.Time `json:"validUntil,omitempty"`
}
