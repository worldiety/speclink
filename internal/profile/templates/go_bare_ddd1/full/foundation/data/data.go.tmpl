// Package data is the little of a persistence framework the rules need to see,
// and nothing more.
//
// It exists so that a repository is recognisable. Without a type to stand for
// one, nothing tells a stored aggregate from any other struct, and the
// questions that follow — does this field trace to a requirement, has this
// shape been promised — have nothing to attach to.
//
// A project that would rather write its own interfaces can: marking them with
// spec.Persistence says the same thing by hand.
package data

import (
	"context"
	"iter"
)

// Aggregate is a consistency boundary with an identity.
type Aggregate[ID comparable] interface {
	Identity() ID
}

// ReadRepository is the half a query needs.
type ReadRepository[E Aggregate[ID], ID comparable] interface {
	FindByID(ctx context.Context, id ID) (E, bool, error)
	All(ctx context.Context) iter.Seq2[E, error]
}

// Repository stores aggregates as their current state.
//
// A named type over it is what marks a repository:
//
//	type CustomerRepository data.Repository[Customer, CustomerID]
type Repository[E Aggregate[ID], ID comparable] interface {
	ReadRepository[E, ID]

	Save(ctx context.Context, e E) error
	DeleteByID(ctx context.Context, id ID) error
}
