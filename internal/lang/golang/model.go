package golang

import (
	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang"
)

// Model is the Go frontend's view of a project.
//
// It implements every capability, which is why the interface was written after
// a second frontend existed rather than before: against this one alone, the
// distinction between what a frontend must do and what it happens to do could
// not have been seen.
type Model struct {
	// All is every loaded package, kept whole for resolution. A rule that
	// follows a helper one step has to see the helper even when the helper is
	// out of scope, or it changes its verdict on untouched code.
	All []*Package
	// Measured is the subset the scope says is reported on.
	Measured []*Package
	// TestPackages are the test variants of the measured packages.
	TestPackages []*Package

	Layout config.Config
	Root   string
	// Style is the convention half of the architecture: how a use case file is
	// named, what its constructor is called, which terms it admits.
	Style Style
	// Framework names the packages the recognisers match on. For nago they are
	// fixed; for an architecture whose foundation is its own code they are
	// derived from the module.
	Framework Framework
	// Layered says whether this architecture separates presentation and
	// adapters into their own packages, which is what the layering rules
	// check. nago has neither directory, so asking would be asking about
	// something that cannot exist.
	Layered bool
}

var (
	_ lang.Model               = (*Model)(nil)
	_ lang.ConstructInferrer   = (*Model)(nil)
	_ lang.SchemaReader        = (*Model)(nil)
	_ lang.VerificationReader  = (*Model)(nil)
	_ lang.SyntaxChecker       = (*Model)(nil)
	_ lang.ArchitectureChecker = (*Model)(nil)
	_ lang.DomainScoper        = (*Model)(nil)
)

func (m *Model) Name() string        { return "go" }
func (m *Model) Dialect() ir.Dialect { return Dialect{} }

func (m *Model) Requirements(out *diag.Set) []*ir.Requirement {
	var reqs []*ir.Requirement
	for _, p := range m.Measured {
		reqs = append(reqs, p.ReadRequirements(out)...)
	}
	return reqs
}

func (m *Model) Bindings(out *diag.Set) []ir.Binding {
	var bindings []ir.Binding
	for _, p := range m.Measured {
		bindings = append(bindings, p.ReadBindings(m.Style, out)...)
	}
	return bindings
}

func (m *Model) Constructs(out *diag.Set) []ir.Construct {
	var constructs []ir.Construct

	if m.Framework.Name == "bare" {
		// The marks are collected across the whole project before anything is
		// inferred: an annotation file names its neighbour, and the neighbour
		// may sit in a package the recogniser reaches first.
		marked := PersistenceMarks(m.Bindings(&diag.Set{}))
		for _, p := range m.Measured {
			constructs = append(constructs, p.InferBare(m.Framework, marked)...)
		}
		return constructs
	}

	for _, p := range m.Measured {
		constructs = append(constructs, p.Infer()...)
	}
	return constructs
}

// Schemas reads the shape of every persisted type.
//
// The persisted set is collected from everything loaded rather than from the
// measured packages, because a repository is usually built in the wiring
// package, far from the type it stores. A scope excluding the wiring would
// otherwise leave the stored shapes unrecognised rather than unmeasured.
func (m *Model) Schemas(out *diag.Set) []ir.SchemaType {
	models := map[string]bool{}
	for _, p := range m.All {
		for name := range p.PersistedModels() {
			models[name] = true
		}
	}
	var schema []ir.SchemaType
	for _, p := range m.Measured {
		schema = append(schema, p.ReadSchema(models)...)
	}
	return schema
}

func (m *Model) Scope() map[string]bool {
	scope := map[string]bool{}
	for _, p := range m.Measured {
		scope[p.PkgPath()] = true
	}
	return scope
}

func (m *Model) Verifications(out *diag.Set) []ir.Binding {
	var verifications []ir.Binding
	for _, p := range m.TestPackages {
		verifications = append(verifications, p.ReadVerifications(out)...)
	}
	return verifications
}

// CheckSyntax enforces the closed subset the carrier form is written in.
//
// It runs wherever the model is read, including where only the requirement tree
// was loaded, because a requirement file is written in the same subset as an
// annotation file and the whitelist is what keeps it readable without
// evaluating it.
func (m *Model) CheckSyntax(out *diag.Set) {
	for _, p := range m.Measured {
		p.CheckWhitelist(out)
		p.CheckOrphans(out)
	}
}

// CheckArchitecture enforces the invariants of the framework.
//
// It takes every loaded package, not the measured ones. A rule that follows a
// helper one step has to see the helper even when the helper is out of scope,
// or it changes its verdict on untouched code — which is how scoping out
// pkg/permtext once made the permission i18n rule report permissions that were
// perfectly fine.
func (m *Model) CheckArchitecture(out *diag.Set) {
	for _, p := range m.Measured {
		if m.Framework.Name == "nago" {
			// The ban is on that framework's own factories. An architecture
			// without them has nothing to forbid, and running the check would
			// be asking a question with no possible answer.
			p.CheckGenericCRUD(out)
		}
		p.CheckVerifiedOutsideTests(out)
	}
	CheckArchitecture(m.All, m.Layout, m.Root, m.Style, ir.CollectWaivers(m.Bindings(&diag.Set{})), out)
	if m.Layered {
		CheckLayering(m.All, m.Layout, m.Root, out)
	}
}

// DomainPackages implements lang.DomainScoper.
func (m *Model) DomainPackages() map[string]bool {
	return DomainPackages(m.All, m.Layout, m.Root)
}
