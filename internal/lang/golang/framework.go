package golang

import (
	"errors"
	"go/types"
)

// Framework names the packages whose types and calls state an architecture.
//
// The recognisers match by import path, on purpose: one binary serves projects
// on different versions of whatever they build on, so speclink must not link
// it. What that costs is that a path is a claim nothing checks at run time, and
// the failure is silence — a moved package does not break a rule, it stops the
// rule matching anything.
//
// Making the paths a value rather than constants is what lets a second
// architecture exist at all. nago's live at go.wdy.de/nago/…, fixed for every
// project. A project with no framework has the same shapes at paths derived
// from its own module, which is not knowable until the module is read.
type Framework struct {
	// Name identifies it in diagnostics.
	Name string

	// Subjects are the types whose presence as a first parameter makes a named
	// func type a use case.
	//
	// More than one, because an alias resolves to the type it names: nago's
	// auth.Subject is user.Subject, so matching only the first would recognise
	// no use case at all, silently, in every project.
	Subjects []Symbol

	// Data holds Repository and ReadRepository, over which a named type marks
	// a repository. Empty when the architecture has no such type, in which
	// case a repository is whatever spec.Persistence says it is.
	Data string

	// Permission holds the Declare family that binds a permission to the use
	// case it guards.
	Permission string
}

// Symbol is one named thing in one package.
type Symbol struct{ Pkg, Name string }

// Nago is the framework the go_nago_ddd1 profile recognises.
var Nago = Framework{
	Name: "nago",
	Subjects: []Symbol{
		{nagoAuth, "Subject"},
		{nagoUser, "Subject"},
		{nagoPermission, "Auditable"},
	},
	Data:       nagoData,
	Permission: nagoPermission,
}

// BareFoundation returns the framework of a project that has none.
//
// The paths come from the module rather than from a constant, because there is
// nothing external to point at: the types live in the project's own foundation
// packages, which the profile's template writes and the project then owns.
//
// It is derived rather than matched by suffix. A suffix would also accept a
// dependency that happened to end in /foundation/auth, and being approximately
// right about which package declares a subject is the kind of wrong that shows
// up as a rule quietly matching nothing.
func BareFoundation(module, root string) Framework {
	if root == "" {
		root = "foundation"
	}
	base := module + "/" + root
	return Framework{
		Name:       "bare",
		Subjects:   []Symbol{{base + "/auth", "Subject"}},
		Data:       base + "/data",
		Permission: base + "/permission",
	}
}

// isSubject reports whether a type is one of the framework's subject types.
func (f Framework) isSubject(t types.Type) bool {
	// Since Go 1.23 an alias is its own type node rather than the aliased named
	// type, so a plain assertion on *types.Named silently fails on one.
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj.Pkg() == nil {
		return false
	}
	for _, s := range f.Subjects {
		if obj.Pkg().Path() == s.Pkg && obj.Name() == s.Name {
			return true
		}
	}
	return false
}

// ModulePath returns the module the loaded packages belong to.
//
// An architecture whose foundation is its own code has no fixed import path to
// match on, so the paths are derived from here. Failing loudly matters: a
// recogniser built on the wrong module matches nothing, and matching nothing is
// indistinguishable from a project that has nothing.
func ModulePath(pkgs []*Package) (string, error) {
	for _, p := range pkgs {
		if p.pkg.Module != nil && p.pkg.Module.Path != "" {
			return p.pkg.Module.Path, nil
		}
	}
	return "", errors.New("no module found; this architecture derives the paths it recognises from go.mod, so speclink has to be run inside a module")
}
