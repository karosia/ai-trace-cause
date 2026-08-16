package semantic

type NodeType string

const (
	NodeTypeObservation NodeType = "Observation"
	NodeTypeFact        NodeType = "Fact"
	NodeTypeDecision    NodeType = "Decision"
)

type RelationType string

const (
	RelationSupports RelationType = "SUPPORTS"
	RelationBasisOf  RelationType = "BASIS_OF"
)

type Observation struct {
	ID string

	Name  string
	Value any

	Metadata map[string]any
}

type Fact struct {
	ID string

	Statement string

	Confidence float64

	Metadata map[string]any
}

type Decision struct {
	ID string

	Outcome    string
	Rationale  string
	Confidence float64

	Metadata map[string]any
}
