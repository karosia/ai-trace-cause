package graph

import (
	"context"
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
