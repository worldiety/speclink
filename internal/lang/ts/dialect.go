package ts

import (
	"path"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// Dialect phrases the fixes in TypeScript.
type Dialect struct{}

var _ ir.Dialect = Dialect{}

func (Dialect) BindConstruct(construct string) string {
	return "bind(" + simple(construct) + ", satisfies(…))"
}

// BindField is phrased although nothing asks for it yet: this profile
// recognises no domain model, so no rule demands a binding for a field. The
// interface is the contract, and a dialect that answered by crashing would turn
// a rule that does not apply into a broken run.
func (Dialect) BindField(construct, field string) string {
	return "bind(" + simple(construct) + ", satisfies(…)) naming the field " + field
}

func (Dialect) BindFieldOptional(construct, field string) string {
	return "an optional marker on the field " + field + " of " + simple(construct)
}

func (Dialect) BindDecision(construct string) string {
	return "bind(" + simple(construct) + ", satisfies(RDecHexagonal))"
}

// AnnotationFile names the file beside the source that holds its bindings:
// Foo.ts and Foo.vue both get Foo.annotation.ts.
func (Dialect) AnnotationFile(sourceFile string) string {
	if sourceFile == "" {
		return "<module>.annotation.ts"
	}
	return annotationFileFor(sourceFile)
}

// RequirementFile is the ID with the requirement suffix. A hyphen is legal in
// a file name, and the ID is what a reader searches for.
func (Dialect) RequirementFile(id string) string { return id + ".req.ts" }

// Reference is the name the declaration was given, which is how it is written
// once it is imported.
func (Dialect) Reference(symbol string) string { return simple(symbol) }

func (Dialect) Verify(ref string) string { return "verified(" + simple(ref) + ")" }

func (Dialect) RecordEvidence() string {
	return "npx vitest run --reporter=json | speclink evidence"
}

func (Dialect) StoredName(wire string) string { return "the stored name '" + wire + "'" }

func (Dialect) Satisfy(ref string) string { return "satisfies(" + simple(ref) + ")" }

func (Dialect) Waive(rule string) string { return "waive('" + rule + "', …)" }

// Transition is phrased for completeness. This profile recognises no events,
// so nothing produces the line today.
func (Dialect) Transition(event, state string) string {
	return "transition(" + simple(event) + ", '" + state + "')"
}

func (Dialect) Term(name string) string {
	if name == "" {
		return name
	}
	return strings.ToLower(name[:1]) + name[1:] + "()"
}

func (Dialect) Status(name string) string { return "Status." + name }

// simple takes the declared name out of a qualified one. The module part is a
// path and may contain dots of its own — R-X.req — so it is the last dot that
// separates them.
func simple(qualified string) string {
	if i := strings.LastIndexByte(qualified, '.'); i >= 0 && !strings.Contains(qualified[i:], "/") {
		return qualified[i+1:]
	}
	return path.Base(qualified)
}
