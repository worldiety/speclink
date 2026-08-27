package ir

// Dialect is how a rule says what to do, in the language the reader is writing.
//
// The rules in internal/check are language neutral in what they decide and were
// never language neutral in how they say it. That is not an oversight to be
// tidied away: the How line of a finding is the interface. It is meant to be
// the literal fix, and a literal fix is a sentence in a particular language.
// "Add a binding" is not actionable; `var _ = spec.For[SubmitQuote](…)` is.
//
// So the rule decides and the frontend phrases. The split is exactly where the
// existing one already runs — internal/check owns the reasoning, the frontend
// owns the syntax — and this interface is the part of it that was left implicit
// and therefore hardcoded in Go.
//
// Everything here returns a fragment to be embedded in a sentence, not a whole
// sentence. The Why of a finding is the rule's own and stays where it is: the
// reason a promise cannot be withdrawn is the same in every language.
type Dialect interface {
	// BindConstruct spells a binding that names a requirement for a construct.
	BindConstruct(construct string) string
	// BindField spells a binding for one field of a construct.
	BindField(construct, field string) string
	// BindFieldOptional spells a binding marking a field as optional.
	BindFieldOptional(construct, field string) string
	// BindDecision spells a binding naming a decision requirement, with an
	// example reference so the shape of one is visible.
	BindDecision(construct string) string

	// AnnotationFile names the file a binding for this construct belongs in.
	AnnotationFile(sourceFile string) string
	// RequirementFile names the file a requirement of this ID belongs in.
	RequirementFile(id string) string

	// Verify spells the statement a test ends with to demonstrate a
	// requirement. ref is the requirement as it is written at a call site.
	Verify(ref string) string
	// Satisfy spells the assertion binding a construct to a requirement.
	Satisfy(ref string) string
	// Waive spells the escape hatch for one rule.
	Waive(rule string) string

	// Term spells one of the declaration terms by name, so a rule can refer to
	// spec.Draft or spec.Optional without knowing how a language writes a call.
	Term(name string) string
	// Status spells one of the requirement statuses by name.
	Status(name string) string
}
