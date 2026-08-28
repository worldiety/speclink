package golang

import (
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// Dialect phrases the fixes in Go.
//
// It is the sentence half of every rule: internal/check decides that a
// construct names no requirement, this decides that saying so means
// `var _ = spec.For[SubmitQuote](spec.Satisfies(…))`. Before this existed the
// two were the same string in the same file, which read as language neutral
// code with Go in the quotes.
//
// A second frontend needs exactly this and nothing else phrased differently.
// The reasons a rule gives — a promise cannot be withdrawn, a section that
// became no requirement is invisible to every other check — do not change with
// the language.
type Dialect struct{}

var _ ir.Dialect = Dialect{}

func (Dialect) BindConstruct(construct string) string {
	return "var _ = spec.For[" + last(construct) + "](spec.Satisfies(…))"
}

func (Dialect) BindField(construct, field string) string {
	return "var _ = spec.ForField[" + last(construct) + "](\"" + field + "\", spec.Satisfies(…))"
}

func (Dialect) BindFieldOptional(construct, field string) string {
	return "var _ = spec.ForField[" + last(construct) + "](\"" + field + "\", spec.Optional())"
}

func (Dialect) BindDecision(construct string) string {
	return "var _ = spec.For[" + last(construct) + "](spec.Satisfies(dec.RDecEventSourcing))"
}

// AnnotationFile is the sidecar convention: commands.go -> commands.annotation.go.
//
// The fallback matters more than it looks. A construct with no file behind it
// is not a Go source file at all — a synthesised position, a package level
// finding — and naming a file that cannot exist would send the reader looking
// for it.
func (Dialect) AnnotationFile(sourceFile string) string {
	file := sourceFile
	if i := strings.LastIndexByte(file, '/'); i >= 0 {
		file = file[i+1:]
	}
	if base, ok := strings.CutSuffix(file, ".go"); ok && base != "" {
		return base + AnnotationSuffix
	}
	return "the annotation file next to it"
}

func (Dialect) RequirementFile(id string) string { return id + RequirementSuffix }

func (Dialect) Verify(ref string) string  { return "spec.Verified(t, " + ref + ")" }
func (Dialect) Satisfy(ref string) string { return "spec.Satisfies(" + ref + ")" }
func (Dialect) Waive(rule string) string  { return `spec.Waive("` + rule + `", …)` }

func (Dialect) Transition(event, state string) string {
	return "spec.Transition[" + last(event) + `]("` + state + `")`
}

func (Dialect) Term(name string) string   { return "spec." + name + "()" }
func (Dialect) Status(name string) string { return "spec." + name }

// last returns the unqualified name, which is how a type is written at a call
// site inside its own package — where an annotation file always is.
func last(qualified string) string {
	if i := strings.LastIndexByte(qualified, '.'); i >= 0 {
		return qualified[i+1:]
	}
	return qualified
}
