package semantic

type NodeType string

const (
	NodeTypeObservation NodeType = "Observation"
	NodeTypeFact        NodeType = "Fact"
)

type RelationType string

const (
	RelationSupports RelationType = "SUPPORTS"
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
