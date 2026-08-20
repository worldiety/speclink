package spec

import "reflect"

// assertionKind discriminates the payload of an [Assertion].
type assertionKind int

const (
	kindSatisfies assertionKind = iota + 1
	kindTransition
	kindExternal
	kindHelp
	kindTerm
	kindRationale
	kindWaive
	kindDraft
	kindOptional
)

// Assertion states one fact about the construct named by the surrounding
// binding. It is opaque to callers.
//
// Assertions are pure: they compute a payload and return it, without side
// effects. This is not a convention but forced by Go — arguments are evaluated
// before the enclosing call, so an assertion cannot know its binding target and
// therefore cannot register itself:
//
//	var _ = spec.For[SubmitQuoteUC](
//		spec.Satisfies(quote.RQuoteSubmit),          // runs first
//		spec.Transition[QuoteSubmitted]("submitted"), // then
//	)                                                    // For runs last
//
// The side effect surface of the whole language is therefore exactly the four
// binding functions in binding.go.
type Assertion struct {
	kind      assertionKind
	reqs      []Requirement
	eventType reflect.Type
	state     State
	text      string
	term      Glossary
	rule      RuleID
}

// Satisfies binds the surrounding construct to one or more requirements.
//
// The central statement of the language and the only one that is impossible to
// infer from code: no amount of static analysis can tell which requirement a
// use case was written for.
func Satisfies(reqs ...Requirement) Assertion {
	return Assertion{kind: kindSatisfies, reqs: reqs}
}

// Transition states which coarse lifecycle state the aggregate assumes after
// event T. The state deliberately is not an event field; it follows from the
// event type.
func Transition[T any](to State) Assertion {
	return Assertion{kind: kindTransition, eventType: reflect.TypeFor[T](), state: to}
}

// External marks an event as arriving from outside (anti-corruption layer or
// federation) and exempts it from the journey reachability check.
func External() Assertion {
	return Assertion{kind: kindExternal}
}

// Help carries the end user instruction for documentation, help system and
// assistant. Use a raw string literal for multiple lines.
func Help(text string) Assertion {
	return Assertion{kind: kindHelp, text: text}
}

// Term anchors a glossary entry at the construct that defines it.
func Term(g Glossary) Assertion {
	return Assertion{kind: kindTerm, term: g}
}

// Rationale justifies a decision at the construct implementing it.
func Rationale(text string) Assertion {
	return Assertion{kind: kindRationale, text: text}
}

// Waive suspends one rule for this construct. The reason is mandatory and
// appears in the gap report, so the exemption leaves a trace.
//
// This is the only escape hatch of the tool. There are no severities and no
// tolerance mode: a finding is an error (see docs/annotations.md §1.8).
func Waive(rule RuleID, reason string) Assertion {
	return Assertion{kind: kindWaive, rule: rule, text: reason}
}

// Draft marks a persisted shape as not yet promised: it may still change in
// any way, and it may be deleted outright.
//
// What the term claims is precise, and it is not a statement about deployment.
// It says: we are willing to delete every stored message of this type. The
// framework can do exactly that — ndb.DeleteType removes them all — which is
// what makes the state real rather than aspirational.
//
// The term is therefore removed at the moment nobody would purge any more. That
// moment has nothing to do with going live: a development database can hold
// data somebody minds losing, and from then on the shape is promised whatever
// the source says. Conversely a deployed system whose data may be thrown away
// at will is still a draft. As long as you would call DeleteType without
// hesitating, the term is honest.
//
// Everything persisted is frozen by default. That inversion is deliberate. The
// number of drafts in a system shrinks over its lifetime, so marking the
// exception costs less than marking the rule, and forgetting to mark fails
// safe: an unmarked newcomer is frozen at once, which shows up as an error on
// the first attempt to change it rather than as silent data loss later.
//
// Committing to a shape is therefore not an act of writing something, but of
// deleting this term and recording what remains.
//
// It cascades downwards, and may be attached at the level that is actually
// true:
//
//	ForPackage  every persisted type in the package is a draft
//	For[T]      every field of T is a draft
//	ForField    only meaningful once the type itself is frozen
//
// Repeating it at a level that is already covered states nothing new and is
// reported as redundant.
func Draft() Assertion {
	return Assertion{kind: kindDraft}
}

// Optional marks a field that may be absent from stored data.
//
// It is required of every field added to a type whose shape has already been
// promised: messages written before the field existed do not carry it, and no
// amount of care in the writer changes that. Declaring it says the reader has
// to cope, which it has to anyway.
//
// It cannot be taken back. Once a field is recorded as optional, the data that
// lacks it exists, and a later claim that the field is always present would be
// false about messages nobody can rewrite.
func Optional() Assertion {
	return Assertion{kind: kindOptional}
}
