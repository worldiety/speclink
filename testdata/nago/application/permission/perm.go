// Package permission is a stub of the nago permission API, reduced to the
// surface the speclink recognizers match on.
package permission

// ID is unique in the entire permission world.
type ID string

// Auditable checks whether an identity carries a use case permission.
type Auditable interface {
	Audit(permission ID) error
	HasPermission(permission ID) bool
}

// Declare connects a permission with the use case type it guards.
//
// The type parameter must be a named func type; nago panics otherwise. That is
// what makes the binding permission <-> use case statically recoverable.
func Declare[UseCase any](id ID, name string, description string) ID { return id }
