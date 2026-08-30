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
	kindPersistence
	kindStoredAs
	kindRestrict
	kindValid
	kindInvalid
	kindClaim
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

	// domainType is the type a kindStoredAs assertion writes down.
	domainType reflect.Type

	// vectors carries the examples of a kindValid or kindInvalid assertion.
	vectors []string
}

// Restrict states the rule a value of this type must satisfy, in prose.
//
// # Why prose, and why that is not enough on its own
//
// A Go type carries what the type system can carry. That a field is a string
// is in the type; that it holds at most 64 characters, never a control
// character, and never one of the prefixes this platform reserves for itself
// is not, and no amount of reading the declaration will find it. It is a
// promise made to whoever is on the other end of the wire, and it is made in
// words.
//
// Words cannot be checked, which is exactly why a constraint on its own is
// refused. A rule stated only in prose is one a code generator drops without a
// sound, and both ends then agree about a shape while disagreeing about what
// may be in it. State it, and then give the examples that decide it — see
// [Valid] and [Invalid].
func Restrict(text string) Assertion {
	return Assertion{kind: kindRestrict, text: text}
}

// Valid gives examples a conforming implementation must accept.
//
// Written as ordinary Go string literals, so a control character is a control
// character and is visible as an escape rather than as an invisible byte.
func Valid(examples ...string) Assertion {
	return Assertion{kind: kindValid, vectors: examples}
}

// Invalid gives examples a conforming implementation must reject.
//
// The half that does the work. Accepting what should be accepted is what an
// implementation does by accident; refusing what must be refused is the part
// nobody gets right without being told, and the part a schema generator most
// reliably drops on the floor.
func Invalid(examples ...string) Assertion {
	return Assertion{kind: kindInvalid, vectors: examples}
}

// Claim marks a field as something the sender asserts and the receiver may not
// treat as true.
//
// The reason names what actually decides the matter, so the field can be read
// without going to look for it: "maßgeblich ist der öffentliche Schlüssel des
// vorgelegten Zertifikats".
//
// # Why this is worth a word of its own
//
// Because without one the two are indistinguishable. A field holding an
// identity that was verified and a field holding an identity the far end
// merely says it has look identical in the struct, in the JSON and in every
// generated schema. Somebody downstream reads the second as the first, and an
// assertion has quietly become an authorisation. That step is invisible in
// review precisely because nothing about the field is wrong.
func Claim(reason string) Assertion {
	return Assertion{kind: kindClaim, text: reason}
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
//
// Every event needs one. The set of states a thing can be in is the first
// question anybody asks about it, and left undeclared the answer exists only
// in the folds — where a sixth state added to something that had five passes
// without anybody deciding it. K15-EVENT-NO-TRANSITION asks for it, and
// K15-TRANSITION-UNKNOWN catches the reverse: this assertion accepts any type
// at all, so a transition naming a renamed or deleted event keeps compiling.
//
// It carries no from state, so this is a labelling of events rather than a
// machine. Whether that is enough will be decided against the process model
// rather than guessed at now.
func Transition[T any](to State) Assertion {
	return Assertion{kind: kindTransition, eventType: reflect.TypeFor[T](), state: to}
}

// External marks an event as arriving from outside — an anti-corruption layer
// or a federated system — so nothing in this codebase produces it.
//
// It is recorded and not yet read by any rule. The rule it exists for is the
// one that reports an event no process consumes, and that arrives with the
// process model; until then this states an intent rather than suspending a
// check. Said plainly because the previous wording named a reachability check
// that has never existed.
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
// tolerance mode: a finding is an error.
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

// Persistence marks a type as storage: an interface as a port, a struct as a
// shape that is written down.
//
// It exists because not every architecture says so by itself. A framework that
// gives out a Repository type has already stated it, and marking it again would
// be stating a fact twice — the one thing this language forbids. A hand written
// interface states nothing, and without a mark speclink cannot tell it from any
// other interface, so the questions that follow have nothing to attach to:
// whether the fields trace to a requirement, whether the shape was promised,
// whether the choice to store it this way was ever justified.
//
// It is therefore available only in the styles that need it, and a finding in
// the ones that do not. Which of those a project is in, its profile says.
func Persistence() Assertion {
	return Assertion{kind: kindPersistence}
}

// StoredAs marks the surrounding type as the form in which Domain is written
// down, and Domain as therefore free to change.
//
// A repository names the domain type it stores, and where that is all there is
// the domain type is what ends up on disk: every rename in it is a change to
// stored data. An adapter may instead keep a shape of its own and map between
// the two, which is what buys the freedom to restructure the domain without
// touching a byte. Nothing about the two structs says which arrangement is in
// force — they are both plain structs in different packages — so it is stated.
//
//	var _ = spec.For[RecordStore](
//		spec.StoredAs[sales.Record](),
//	)
//
// The effect is to move the promise: RecordStore is frozen and checked for
// compatible evolution, and sales.Record is not. Without it the promise stays
// on the domain type, which is the stricter reading and the safe default.
func StoredAs[Domain any]() Assertion {
	return Assertion{kind: kindStoredAs, domainType: reflect.TypeFor[Domain]()}
}
