// Package lang is the boundary between the rules and the language they read.
//
// It was written after the second frontend rather than before it, on purpose.
// An interface designed against one implementation comes out shaped like that
// implementation, and the whole reason for a boundary here was to avoid
// exactly that. What is below is a description of where two frontends actually
// differ, not a guess at where a third might.
//
// # What they had in common
//
// Reading. Both turn their own model into []*ir.Requirement and []ir.Binding,
// reporting what they could not read through a diag.Set. That is the whole of
// the overlap and it is the whole of [Model].
//
// # What they did not
//
// Loading. The Go frontend takes package patterns and hands back packages, with
// type errors as a separate concept because a Go build either succeeds or there
// is nothing to say. The JVM frontend takes directories of compiled classes and
// hands back classes, with per-file parse errors, because a class file is
// readable or it is not and its neighbours are unaffected. Neither signature
// generalises without lying about one of them, so loading stays outside the
// interface and each command asks its frontend directly.
//
// Capability. The Go frontend infers architectural roles, reads persisted
// shapes, finds verification claims and enforces an architecture. The JVM
// frontend does none of that yet, and a third one might never. Requiring them
// would mean stub methods returning nil, which reads as "measured, found
// nothing" — the exact confusion between an empty answer and no answer that
// this tool spends most of its rules preventing.
//
// So they are asked for rather than required, and a run reports which
// directions it was able to measure. A summary saying a hundred percent of a
// question that was never put is worse than one that admits the gap.
package lang

import (
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// Model is what every frontend can produce.
type Model interface {
	// Requirements returns the declared requirements.
	Requirements(out *diag.Set) []*ir.Requirement
	// Bindings returns the assertions tying constructs to requirements.
	Bindings(out *diag.Set) []ir.Binding
	// Dialect phrases the fixes in the language this model came from.
	Dialect() ir.Dialect
	// Name identifies the frontend in diagnostics and in the summary.
	Name() string
}

// ConstructInferrer is implemented by a frontend that recognises architectural
// roles on its own.
//
// This is the capability that makes forward coverage possible, and the only one
// whose absence changes what a run means rather than what it checks. Without it
// there is no set of constructs to hold accountable, so "every construct names
// a requirement" is not a weaker claim — it is not a claim.
type ConstructInferrer interface {
	Constructs(out *diag.Set) []ir.Construct
}

// ProcessReader is implemented by a frontend that can find the process
// declarations of a project.
//
// Separate from ConstructInferrer because the two answer different questions
// and a frontend may well be able to do one and not the other: reading a
// declaration needs a parser, recognising a use case needs to know a framework.
type ProcessReader interface {
	Processes(out *diag.Set) []*ir.Process
}

// TopicReader is implemented by a frontend that can find the theme
// declarations of the requirement tree.
//
// A capability rather than part of Model, because a theme is a convenience of
// the document and a frontend that cannot read one is not thereby unable to
// check anything. Nothing depends on it but the chapters.
type TopicReader interface {
	Topics() []*ir.Topic
}

// TopologyReader is implemented by a frontend that can say what surrounds the
// code and where it reaches out.
type TopologyReader interface {
	Topology(out *diag.Set) ir.Topology
}

// EndpointReader is implemented by a frontend that can say what addresses the
// system answers on.
//
// Separate from TopologyReader although both describe the edge, because they
// are known in different ways. A channel is declared: no module states that an
// object store is somebody else's responsibility. An endpoint is recognised:
// the code that mounts it already says everything there is to say. A frontend
// may well be able to do one and not the other.
type EndpointReader interface {
	Endpoints() []ir.Endpoint
}

// SchemaReader is implemented by a frontend that can say which types are
// persisted and what shape they have on the wire.
type SchemaReader interface {
	Schemas(out *diag.Set) []ir.SchemaType
	// Scope names the packages the run looked at, which is what lets the
	// evolution rules tell a deleted type from an unexamined one.
	Scope() map[string]bool
}

// VerificationReader is implemented by a frontend that can find the claims
// tests make about requirements.
type VerificationReader interface {
	Verifications(out *diag.Set) []ir.Binding
}

// SyntaxChecker is implemented by a frontend that constrains how its carrier
// form may be written — a closed subset, a file naming convention.
//
// It is separate from ArchitectureChecker because the two have different
// reaches, which only became visible when they were briefly the same method:
// the tree command started reporting that a module had no entry point, having
// been asked to look at the requirement tree alone. A requirement file is
// written in the same subset as an annotation file, so its syntax is checked
// wherever it is read; an architecture is a statement about a whole project and
// can only be judged when the whole project was loaded.
type SyntaxChecker interface {
	CheckSyntax(out *diag.Set)
}

// ArchitectureChecker is implemented by a frontend that enforces invariants of
// its framework.
//
// These stay on that side of the boundary because another frontend would
// replace them wholesale rather than pick from them. They are also not
// decoration: the recognisers work because the architecture holds, and the ban
// on framework factories that produce specification facts at run time exists so
// that a static analysis can see them at all.
type ArchitectureChecker interface {
	CheckArchitecture(out *diag.Set)
}

// DomainScoper is implemented by a frontend that can say which packages hold
// the domain, as opposed to the plumbing around it.
//
// It is what confines the field level rules to where they mean something. A
// field of a domain model states what the system believes; a field of a request
// object states what a caller sent, and demanding a requirement for the second
// would teach people to switch the rule off.
type DomainScoper interface {
	DomainPackages() map[string]bool
}

// Capabilities describes what a model turned out to be able to do.
//
// It is reported rather than assumed, because the honest failure of a partial
// frontend is silence. A direction that was never measured must not read as a
// direction that came out clean.
type Capabilities struct {
	Constructs    bool
	Schemas       bool
	Verifications bool
	Syntax        bool
	Architecture  bool
	Endpoints     bool
	Processes     bool
	Topics        bool
	Topology      bool
}

// Of reports what a model can do.
func Of(m Model) Capabilities {
	_, constructs := m.(ConstructInferrer)
	_, schemas := m.(SchemaReader)
	_, verifications := m.(VerificationReader)
	_, syntax := m.(SyntaxChecker)
	_, architecture := m.(ArchitectureChecker)
	_, endpoints := m.(EndpointReader)
	_, processes := m.(ProcessReader)
	_, topics := m.(TopicReader)
	_, topology := m.(TopologyReader)
	return Capabilities{
		Constructs:    constructs,
		Schemas:       schemas,
		Verifications: verifications,
		Syntax:        syntax,
		Architecture:  architecture,
		Endpoints:     endpoints,
		Processes:     processes,
		Topics:        topics,
		Topology:      topology,
	}
}

// Missing lists the directions this model cannot measure, for a run to say so.
func (c Capabilities) Missing() []string {
	var out []string
	if !c.Constructs {
		out = append(out, "forward coverage, because this frontend infers no constructs")
	}
	if !c.Schemas {
		out = append(out, "schema evolution, because this frontend reads no persisted shapes")
	}
	if !c.Verifications {
		out = append(out, "verification, because this frontend finds no test claims")
	}
	if !c.Topology {
		out = append(out, "the boundary, because this frontend reads no topology declarations")
	}
	if !c.Processes {
		out = append(out, "courses of business, because this frontend reads no process declarations")
	}
	if !c.Endpoints {
		// The most dangerous of these to leave unsaid. A framework whose
		// routes this cannot read does not expose fewer addresses; it exposes
		// the same ones, unlisted. Every other figure in the summary would be
		// about a system whose whole outside edge went unmentioned.
		out = append(out, "the exposed surface, because this frontend recognises no routes")
	}
	return out
}

// OnlyTree narrows a model to the questions a part of a project can answer.
//
// A model built from a subset of the packages knows what those packages say and
// nothing about the module. That distinction was carried in a comment for a
// long time, and the comment did not hold: three separate defects came from a
// module-wide question being put to a partial view — an entry point reported as
// missing because cmd/ had not been read, a use case reported as absent because
// its package had not been, and a requirement tree reported as fully covered
// having never been loaded.
//
// Making it a type is what stops the fourth. A partial model does not implement
// the capabilities that answer for a whole module, so the question cannot be
// asked rather than being asked and answered wrongly — and because every caller
// already handles a missing capability by saying the direction was not
// measured, the honest report comes out with no further work.
//
// The wrapper is deliberately not a struct embedding the model. Embedding would
// promote whatever the underlying type happens to implement, which is exactly
// the leak this exists to close: the whole point is that some methods are not
// reachable.
func OnlyTree(m Model) Model { return treeOnly{m} }

type treeOnly struct{ m Model }

func (t treeOnly) Requirements(out *diag.Set) []*ir.Requirement { return t.m.Requirements(out) }
func (t treeOnly) Bindings(out *diag.Set) []ir.Binding          { return t.m.Bindings(out) }
func (t treeOnly) Dialect() ir.Dialect                          { return t.m.Dialect() }
func (t treeOnly) Name() string                                 { return t.m.Name() }

// CheckSyntax is forwarded because it is the one check that is genuinely local.
//
// A requirement file is written in a closed subset of the language, and whether
// this file obeys it is answered by this file. Nothing about the rest of the
// module changes the verdict, which is what separates it from the architecture
// rules that were also once reachable from here.
func (t treeOnly) CheckSyntax(out *diag.Set) {
	if c, ok := t.m.(SyntaxChecker); ok {
		c.CheckSyntax(out)
	}
}

var _ SyntaxChecker = treeOnly{}
