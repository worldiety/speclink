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
	_ lang.ProcessReader       = (*Model)(nil)
	_ lang.TopicReader         = (*Model)(nil)
	_ lang.TopologyReader      = (*Model)(nil)
	_ lang.EndpointReader      = (*Model)(nil)
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

// Topics reads the themes declared in the requirement tree.
func (m *Model) Topics() []*ir.Topic {
	var topics []*ir.Topic
	for _, p := range m.Measured {
		topics = append(topics, p.ReadTopics()...)
	}
	return topics
}

func (m *Model) Bindings(out *diag.Set) []ir.Binding {
	var bindings []ir.Binding
	for _, p := range m.Measured {
		bindings = append(bindings, p.ReadBindings(m.Style, out)...)
	}
	return bindings
}

// Processes reads the declared courses of business.
//
// Measured packages only, like every other reader here. A process lives above
// the contexts it names, so a scope that lists only contexts will not include
// it — but that is what a scope is for, and inventing an exception would make
// the one setting whose meaning is "what is measured" mean something else here.
func (m *Model) Processes(out *diag.Set) []*ir.Process {
	var procs []*ir.Process
	for _, p := range m.Measured {
		procs = append(procs, p.ReadProcesses(out)...)
	}
	return procs
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
		m.foldConstructors(constructs)
		return constructs
	}

	for _, p := range m.Measured {
		constructs = append(constructs, p.Infer()...)
	}
	m.foldConstructors(constructs)
	return constructs
}

// foldConstructors folds a use case's constructor into its fingerprint.
//
// Without this the record of who read what would be close to worthless for the
// one role it matters most for. A use case is declared as a named func type —
// a signature and nothing else — while everything it actually does lives in the
// constructor beside it. A reviewer reads the logic; a fingerprint over the
// signature alone would survive any rewrite of it and go on claiming the review
// still holds.
//
// The constructor is found by the style's own spelling, which is the same
// source the architecture rules use to demand one. Where a style names none, or
// a construct has none, the fingerprint stays what it was.
func (m *Model) foldConstructors(constructs []ir.Construct) {
	if m.Style.Constructor == nil {
		return
	}
	byPkg := map[string]*Package{}
	for _, p := range m.Measured {
		byPkg[p.PkgPath()] = p
	}

	for i := range constructs {
		c := &constructs[i]
		if c.Fingerprint == "" || !c.Kind.PerformsWork() {
			continue
		}
		p, ok := byPkg[c.Package]
		if !ok {
			continue
		}
		start, end, found := p.FuncExtent(m.Style.Constructor(last(c.Name)))
		if !found {
			continue
		}
		p.FoldInto(c, start, end)
	}
}

// Schemas reads the shape of every persisted type.
//
// The persisted set is collected from everything loaded rather than from the
// measured packages, because a repository is usually built in the wiring
// package, far from the type it stores. A scope excluding the wiring would
// otherwise leave the stored shapes unrecognised rather than unmeasured.
func (m *Model) Schemas(out *diag.Set) []ir.SchemaType {
	models := m.persisted()
	var schema []ir.SchemaType
	for _, p := range m.Measured {
		schema = append(schema, p.ReadSchema(models)...)
	}
	return schema
}

// persisted names the types whose shape is promised.
//
// How they are found is the framework's business, and the two answers are not
// variations of one another. nago states it in a constructor: NewJSONRepository
// names a persistence model distinct from the domain model, NewSloppyJSON names
// none and ties the domain type to the wire. A project with no framework has no
// such call, so the structural statement is the element type of its repository
// ports, and an adapter that keeps a shape of its own says so with StoredAs.
//
// The set is collected from everything loaded rather than from the measured
// packages, because a repository is usually built in the wiring package, far
// from the type it stores. A scope excluding the wiring would otherwise leave
// the stored shapes unrecognised rather than unmeasured.
func (m *Model) persisted() map[string]bool {
	models := map[string]bool{}

	if m.Framework.Name != "bare" {
		for _, p := range m.All {
			for name := range p.PersistedModels() {
				models[name] = true
			}
		}
		return models
	}

	for _, p := range m.All {
		for _, name := range p.RepositoryElements(m.Framework) {
			models[name] = true
		}
	}

	// A declared mapping moves the promise onto the stored shape and releases
	// the domain type. Without one the domain type stays promised, which is
	// the stricter reading: whatever the adapter does with it, nothing has
	// said the two may drift apart.
	for domain, store := range StoredForms(m.Bindings(&diag.Set{})) {
		if models[domain] {
			delete(models, domain)
			models[store] = true
		}
	}
	return models
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
