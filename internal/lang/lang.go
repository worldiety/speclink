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

// Checker is implemented by a frontend with rules of its own — a syntax
// whitelist, an architecture, a ban on constructs it cannot analyse.
//
// They stay on this side of the boundary because they are rules about a
// language and a framework, and another frontend would replace them wholesale
// rather than pick from them.
type Checker interface {
	Check(out *diag.Set)
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
	Checks        bool
}

// Of reports what a model can do.
func Of(m Model) Capabilities {
	_, constructs := m.(ConstructInferrer)
	_, schemas := m.(SchemaReader)
	_, verifications := m.(VerificationReader)
	_, checks := m.(Checker)
	return Capabilities{
		Constructs:    constructs,
		Schemas:       schemas,
		Verifications: verifications,
		Checks:        checks,
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
	return out
}
