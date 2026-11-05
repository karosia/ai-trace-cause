package graph

import (
	"context"
	"reflect"
	"time"
)

// NodeFilter narrows the nodes returned by Queryable.FindNodes. A
// zero-value field imposes no constraint on that dimension, so the
// zero-value NodeFilter matches every node.
type NodeFilter struct {
	// Type restricts results to nodes with this exact Type. Empty
	// matches any type.
	Type string

	// RecordedAfter restricts results to nodes recorded at or after
	// this time. Nil imposes no lower bound.
	RecordedAfter *time.Time

	// RecordedBefore restricts results to nodes recorded strictly
	// before this time. Nil imposes no upper bound.
	RecordedBefore *time.Time

	// Properties restricts results to nodes whose top-level
	// properties include every given key with an equal value
	// (reflect.DeepEqual, so dynamic types must match: int(5) does not
	// equal int64(5)). Nil or empty imposes no constraint.
	Properties map[string]any
}

// MatchesProperties reports whether properties satisfies the
// Properties constraint of filter. Store implementations can use it to
// apply that constraint consistently.
func (filter NodeFilter) MatchesProperties(
	properties map[string]any,
) bool {
	for key, want := range filter.Properties {
		got, ok := properties[key]
		if !ok || !reflect.DeepEqual(got, want) {
			return false
		}
	}

	return true
}

// Queryable is an optional capability a Store may implement to list
// nodes by criteria other than direct ID lookup or edge traversal, for
// example "every Decision recorded in the last hour." Stores that
// don't implement Queryable remain fully usable for everything else;
// Graph.FindNodes returns ErrStoreNotQueryable if the underlying Store
// doesn't implement it.
type Queryable interface {
	FindNodes(ctx context.Context, filter NodeFilter) ([]Node, error)
}

// FindNodes lists nodes matching filter by delegating to the
// underlying Store's Queryable implementation. It returns
// ErrStoreNotQueryable if the Store does not implement Queryable.
func (g *Graph) FindNodes(
	ctx context.Context,
	filter NodeFilter,
) ([]Node, error) {
	queryable, ok := g.store.(Queryable)
	if !ok {
		return nil, ErrStoreNotQueryable
	}

	return queryable.FindNodes(ctx, filter)
}
