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
// The side effect surface of the whole language is therefore exactly the five
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
