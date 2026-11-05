package semantic

import (
	"errors"
	"fmt"
)

// RelationRule allows relationships of type Relation from nodes of
// type From to nodes of type To. Relationships always point from cause
// to effect.
type RelationRule struct {
	From     NodeType     `json:"from"`
	Relation RelationType `json:"relation"`
	To       NodeType     `json:"to"`
}

// Schema is the set of relationships a Service accepts. Keeping it as
// data, rather than one method per relationship, lets applications
// model their own agent loops by extending DefaultSchema with
// Schema.With.
//
// A Schema is immutable; With returns a new one. The zero value
// allows nothing; use DefaultSchema or NewSchema.
type Schema struct {
	rules  []RelationRule
	byPair map[[2]NodeType][]RelationType
}

// Errors returned when building a Schema.
var (
	// ErrInvalidRelationRule is returned by NewSchema and Schema.With
	// for a rule with an empty field, a reserved relation, or a
	// duplicate.
	ErrInvalidRelationRule = errors.New("invalid relation rule")
)

// DefaultSchema returns the built-in causal model:
//
//	Source      --PRODUCED-->    Observation
//	Action      --RESULTED_IN--> Observation  (a tool call's result)
//	Observation --SUPPORTS-->    Fact
//	Fact        --INFERS-->      Fact         (premise to conclusion)
//	Fact        --BASIS_OF-->    Decision
//	Decision    --REFINES-->     Decision     (plan to sub-decision)
//	Decision    --CAUSED-->      Action
//
// RESULTED_IN closes the agent loop: what an action observed can feed
// the next decision, so traces follow an agent across tool calls.
func DefaultSchema() Schema {
	schema, err := NewSchema(
		RelationRule{From: NodeTypeSource, Relation: RelationProduced, To: NodeTypeObservation},
		RelationRule{From: NodeTypeAction, Relation: RelationResultedIn, To: NodeTypeObservation},
		RelationRule{From: NodeTypeObservation, Relation: RelationSupports, To: NodeTypeFact},
		RelationRule{From: NodeTypeFact, Relation: RelationInfers, To: NodeTypeFact},
		RelationRule{From: NodeTypeFact, Relation: RelationBasisOf, To: NodeTypeDecision},
		RelationRule{From: NodeTypeDecision, Relation: RelationRefines, To: NodeTypeDecision},
		RelationRule{From: NodeTypeDecision, Relation: RelationCaused, To: NodeTypeAction},
	)
	if err != nil {
		panic(err) // the built-in rules are valid by construction
	}

	return schema
}

// NewSchema returns a Schema allowing exactly rules. It returns an
// error wrapping ErrInvalidRelationRule if a rule has an empty field,
// uses the reserved RelationRetracted, or is listed twice.
func NewSchema(rules ...RelationRule) (Schema, error) {
	return Schema{}.With(rules...)
}

// With returns a new Schema that also allows rules. The receiver is
// not modified.
func (s Schema) With(rules ...RelationRule) (Schema, error) {
	extended := Schema{
		rules:  make([]RelationRule, 0, len(s.rules)+len(rules)),
		byPair: make(map[[2]NodeType][]RelationType, len(s.byPair)+len(rules)),
	}

	for _, rule := range append(append([]RelationRule(nil), s.rules...), rules...) {
		if rule.From == "" || rule.Relation == "" || rule.To == "" {
			return Schema{}, fmt.Errorf(
				"%w: empty field in %+v",
				ErrInvalidRelationRule,
				rule,
			)
		}

		if rule.Relation == RelationRetracted {
			return Schema{}, fmt.Errorf(
				"%w: %s is reserved for retractions",
				ErrInvalidRelationRule,
				RelationRetracted,
			)
		}

		if extended.Allows(rule.From, rule.Relation, rule.To) {
			return Schema{}, fmt.Errorf(
				"%w: duplicate rule %s --%s--> %s",
				ErrInvalidRelationRule,
				rule.From,
				rule.Relation,
				rule.To,
			)
		}

		pair := [2]NodeType{rule.From, rule.To}
		extended.rules = append(extended.rules, rule)
		extended.byPair[pair] = append(extended.byPair[pair], rule.Relation)
	}

	return extended, nil
}

// Rules returns a copy of the schema's rules, in the order they were
// added.
func (s Schema) Rules() []RelationRule {
	return append([]RelationRule(nil), s.rules...)
}

// Allows reports whether the schema permits relation from a node of
// type from to a node of type to.
func (s Schema) Allows(
	from NodeType,
	relation RelationType,
	to NodeType,
) bool {
	for _, allowed := range s.byPair[[2]NodeType{from, to}] {
		if allowed == relation {
			return true
		}
	}

	return false
}

// relationBetween returns the single relation the schema allows from
// a node of type from to a node of type to. It fails if there is none,
// or more than one (then the caller must name the relation).
func (s Schema) relationBetween(
	from NodeType,
	to NodeType,
) (RelationType, error) {
	relations := s.byPair[[2]NodeType{from, to}]

	switch len(relations) {
	case 1:
		return relations[0], nil
	case 0:
		return "", &RelationNotAllowedError{From: from, To: to}
	default:
		return "", fmt.Errorf(
			"%w: %s -> %s allows %v; name the relation with WithCausesVia or Link",
			ErrAmbiguousRelation,
			from,
			to,
			relations,
		)
	}
}

// relationVerbs phrases each built-in relation for ExplainGraph as
// "<effect> <verb> <cause>".
var relationVerbs = map[RelationType]string{
	RelationProduced:   "was produced by",
	RelationResultedIn: "resulted from",
	RelationSupports:   "was supported by",
	RelationInfers:     "was inferred from",
	RelationBasisOf:    "was based on",
	RelationRefines:    "refines",
	RelationCaused:     "was caused by",
}
