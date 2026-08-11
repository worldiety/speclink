// Package data is a stub of the nago repository API.
package data

// IDType constrains identifier types.
type IDType interface {
	~int | ~int64 | ~int32 | ~string
}

// Aggregate is an entity that owns others and defines a transaction boundary.
type Aggregate[Ident comparable] interface {
	Identity() Ident
}

// Repository stores aggregates.
type Repository[E Aggregate[ID], ID IDType] interface {
	FindByID(id ID) (E, error)
	Save(E) error
	DeleteByID(id ID) error
	Name() string
}
