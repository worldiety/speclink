package check

import "github.com/worldiety/speclink/internal/ir"

// plainDialect phrases fixes in no language at all.
//
// Its existence is the point of the interface. These rules are supposed to
// decide without knowing what language they are looking at, and until now the
// only way to test them was to accept Go syntax in their output. If a rule
// stops working against this, it has learned something about Go that belongs in
// a frontend.
type plainDialect struct{}

var _ ir.Dialect = plainDialect{}

func (plainDialect) BindConstruct(construct string) string {
	return "a binding naming a requirement for " + construct
}

func (plainDialect) BindField(construct, field string) string {
	return "a binding for field " + field + " of " + construct
}

func (plainDialect) BindFieldOptional(construct, field string) string {
	return "a term marking field " + field + " of " + construct + " optional"
}

func (plainDialect) BindDecision(construct string) string {
	return "a binding naming a decision for " + construct
}

func (plainDialect) AnnotationFile(sourceFile string) string {
	return "the annotation file beside " + sourceFile
}

func (plainDialect) RequirementFile(id string) string { return "the file of " + id }

func (plainDialect) Verify(ref string) string  { return "a verification of " + ref }
func (plainDialect) Satisfy(ref string) string { return "a reference to " + ref }
func (plainDialect) Waive(rule string) string  { return "a waiver of " + rule }
func (plainDialect) Term(name string) string   { return "the " + name + " term" }
func (plainDialect) Status(name string) string { return name }
