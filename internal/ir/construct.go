package ir

// ConstructKind is an architectural role recognised in the host code.
//
// These roles are never annotated. They are inferred from framework usage,
// because the framework already states them unambiguously — and a fact that is
// inferable must not be annotated (P2: exactly one source per fact). What
// cannot be inferred is which requirement a construct was written for; that is
// what the annotation carries.
type ConstructKind int

const (
	// ConstructUseCase is a named func type taking an auth subject first.
	ConstructUseCase ConstructKind = iota + 1
	// ConstructCommand implements Decide and is the write side of an aggregate.
	ConstructCommand
	// ConstructEvent implements Evolve plus Discriminator and is a domain fact.
	ConstructEvent
	// ConstructAggregate is a consistency boundary with an identity.
	ConstructAggregate
	// ConstructPermission is declared by permission.Declare and bound to a use
	// case through its type parameter.
	ConstructPermission
	// ConstructQuery is a read side function taking an auth subject.
	ConstructQuery
	// ConstructProjection is an event folded read model, built by
	// evs.NewProjection and fed by evs.Project.
	ConstructProjection
	// ConstructRepository is a named type standing for data.Repository or
	// data.ReadRepository over an aggregate.
	ConstructRepository
)

// WithArticle renders the kind as a noun phrase for diagnostics.
//
// The article follows pronunciation, not spelling: "a use case", because "use"
// begins with a consonant sound despite starting with a vowel letter. A generic
// vowel test gets this wrong, so the phrase is stated per kind.
func (k ConstructKind) WithArticle() string {
	switch k {
	case ConstructUseCase:
		return "a use case"
	case ConstructCommand:
		return "a command"
	case ConstructEvent:
		return "an event"
	case ConstructAggregate:
		return "an aggregate"
	case ConstructPermission:
		return "a permission"
	case ConstructQuery:
		return "a query"
	case ConstructProjection:
		return "a projection"
	case ConstructRepository:
		return "a repository"
	}
	return "an unknown construct"
}

func (k ConstructKind) String() string {
	switch k {
	case ConstructUseCase:
		return "use case"
	case ConstructCommand:
		return "command"
	case ConstructEvent:
		return "event"
	case ConstructAggregate:
		return "aggregate"
	case ConstructPermission:
		return "permission"
	case ConstructQuery:
		return "query"
	case ConstructProjection:
		return "projection"
	case ConstructRepository:
		return "repository"
	}
	return "unknown"
}

// NeedsRequirement reports whether a construct of this kind has to be bound to
// a requirement (forward coverage, K1/K3).
//
// Use cases, queries, commands, events and projections carry business meaning
// and must trace back to something that was asked for.
//
// A query is included for the same reason a use case is. Reading is a promise
// too — that someone may see a thing — and every architecture rule already
// treats a query as a use case: it needs its own uc_ file, its constructor, its
// permission and its place in the bundle. Exempting it from coverage alone was
// an omission rather than a decision, and it made a recogniser defect
// expensive: while evs.SeqID went unrecognised, every writing use case was
// classified as a query and silently left the denominator.
//
// Permissions, aggregates and repositories are structural. They are covered
// through the use case that guards, writes or holds them, and demanding a
// binding for each would only produce noise.
func (k ConstructKind) NeedsRequirement() bool {
	switch k {
	case ConstructUseCase, ConstructQuery, ConstructCommand, ConstructEvent, ConstructProjection:
		return true
	}
	return false
}

// Persisted reports whether the shape of this construct outlives the code that
// declares it, and is therefore subject to the evolution rules.
//
// An event is the clear case: the struct is the wire format. It is written into
// a log that is replayed forever, so every field name and every field shape is
// a promise from the moment the first message is stored.
//
// An aggregate is not included here. In an event sourced context it is folded
// from the log and never stored; where it is stored, it is stored through a
// repository, and the repository is what says so. That distinction is made
// where the repository is recognised, not here.
func (k ConstructKind) Persisted() bool {
	return k == ConstructEvent
}

// Construct is an architectural fact recognised in the host language.
type Construct struct {
	Kind ConstructKind
	// Name is the fully qualified name, e.g. "example.com/m/sales.SubmitQuoteUC".
	Name string
	// Package is the import path the construct was found in.
	Package string
	// Evidence names the framework marker that revealed the construct. It goes
	// into diagnostics so a reader can see why speclink believes this.
	Evidence string
	Pos      Position
}

// Target renders the construct as the binding target that would annotate it.
func (c Construct) Target() Target {
	return Target{Kind: TargetType, Package: c.Package, Name: c.Name}
}
