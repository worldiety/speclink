package ir

// Architecture is the shape a project has agreed to keep, described rather
// than checked.
//
// The rules that enforce this have always existed, and until now the only way
// to learn them was to break one. Every sentence in a finding is written at
// the moment of violation and names the offending file, so a project with no
// findings — the normal case, and the one somebody hands to a reviewer — said
// nothing at all about the shape it holds itself to.
//
// This describes what is enforced, and only what is enforced. A rule that this
// profile does not run must not appear here: a document that lists a
// prohibition nothing checks is worse than one that lists none, because a
// reader will believe it.
type Architecture struct {
	// Style names the convention, as the profile does.
	Style string
	// Layers are the places code lives, in the order a reader should meet
	// them. Empty when the profile classifies nothing.
	Layers []Layer
	// Rules are the prohibitions and obligations actually in force.
	Rules []Rule
}

// Declared reports whether anything is known about the architecture.
func (a Architecture) Declared() bool { return len(a.Layers) > 0 || len(a.Rules) > 0 }

// Layer is one place code is expected to live.
type Layer struct {
	// Name is what the convention calls it.
	Name string
	// Where is the path it occupies, relative to the module root, using the
	// project's own configured roots rather than the profile's defaults.
	Where string
	// Purpose is why it is separate from the others.
	Purpose string
}

// Rule is one thing the architecture forbids or requires.
type Rule struct {
	// ID is the rule identifier a finding would carry, so a reader who has
	// seen the failure can find the sentence and the other way round.
	ID string
	// Statement is the obligation in one sentence, in the present tense.
	Statement string
	// Why is the reason it is worth having. A rule whose reason cannot be
	// written is a rule that should not be enforced.
	Why string
}
