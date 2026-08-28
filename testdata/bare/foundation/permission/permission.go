// Package permission declares what a use case is allowed to do.
//
// A permission is a compile time thing: it exists because somebody wrote a use
// case, and the set of them is fixed once the binary is linked. Declaring them
// as package level variables is what makes that true, and it is why an admin
// screen or a listing can be generated from the registry rather than maintained
// beside it.
package permission

import (
	"fmt"
	"regexp"
	"slices"
	"sync"
)

// ID names one capability, e.g. "sales.quote.submit".
type ID string

var valid = regexp.MustCompile(`^[a-z][a-z0-9_]*(\.[a-z0-9_]+)*$`)

// Valid reports whether the ID follows the dotted lower case convention.
func (id ID) Valid() bool { return valid.MatchString(string(id)) }

// Auditable is the question a use case asks.
//
// It is declared here rather than in auth so that the two packages do not have
// to import each other.
type Auditable interface {
	// Audit checks the permission and records that it was asked.
	Audit(ID) error
	// HasPermission is the unaudited question, for a presentation deciding what
	// to show. A use case uses Audit.
	HasPermission(ID) bool
}

// Permission is a declared capability with the texts a human sees.
type Permission struct {
	ID          ID
	Name        string
	Description string
	// UseCase is the type this permission guards, for the listing.
	UseCase string
}

var (
	mu       sync.Mutex
	declared = map[ID]Permission{}
)

// Declare registers a permission and binds it to the use case it guards.
//
// The type parameter is the binding: one use case, at most one permission, and
// the compiler checks that the use case exists. A duplicate or an invalid ID
// panics, because this runs at init — the one moment where failing loudly costs
// nothing and failing quietly costs an unreachable check.
func Declare[UseCase any](id ID, name, description string) ID {
	if !id.Valid() {
		panic(fmt.Sprintf("permission %q is not a valid id", id))
	}

	mu.Lock()
	defer mu.Unlock()
	if _, dup := declared[id]; dup {
		panic(fmt.Sprintf("permission %q is declared twice", id))
	}
	declared[id] = Permission{ID: id, Name: name, Description: description, UseCase: nameOf[UseCase]()}
	return id
}

// All returns every declared permission, ordered by ID.
func All() []Permission {
	mu.Lock()
	defer mu.Unlock()

	out := make([]Permission, 0, len(declared))
	for _, p := range declared {
		out = append(out, p)
	}
	slices.SortFunc(out, func(a, b Permission) int {
		return int(a.ID[0]) - int(b.ID[0])
	})
	return out
}
