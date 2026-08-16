package semantic

import "time"

type NodeType string

const (
	NodeTypeSource      NodeType = "Source"
	NodeTypeObservation NodeType = "Observation"
	NodeTypeFact        NodeType = "Fact"
	NodeTypeDecision    NodeType = "Decision"
	NodeTypeAction      NodeType = "Action"
)

type RelationType string

const (
	RelationProduced RelationType = "PRODUCED"
	RelationSupports RelationType = "SUPPORTS"
	RelationBasisOf  RelationType = "BASIS_OF"
	RelationCaused   RelationType = "CAUSED"
)

type Source struct {
	ID string

	Kind string
	URI  string

	Metadata map[string]any

	Validity Validity
}

type Observation struct {
	ID string

	Name  string
	Value any

	Metadata map[string]any

	Validity Validity
}

type Fact struct {
	ID string

	Statement string

	Confidence float64

	Metadata map[string]any

	Validity Validity
}

type Decision struct {
	ID string

	Outcome    string
	Rationale  string
	Confidence float64

	Metadata map[string]any

	Validity Validity
}

type Action struct {
	ID string

	Name   string
	Target string

	Parameters map[string]any
	Metadata   map[string]any

	Validity Validity
}

type Validity struct {
	ValidFrom  *time.Time
	ValidUntil *time.Time
}
