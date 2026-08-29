package ir

// ConstructKind is an architectural role recognised in the host code.
//
// These roles are never annotated. They are inferred from framework usage,
// because the framework already states them unambiguously — and a fact that is
// inferable must not be annotated (exactly one source per fact). What cannot be
// inferred is which requirement a construct was written for; that is what the
// annotation carries.
//
// The set is supplied by the frontend, not fixed here, and the reason is that
// the vocabulary below is one framework's. Use case, command, event, aggregate
// and projection are the words of a domain driven, event sourced design. A
// project built on a different framework has other roles and the same question
// to answer about each of them, so what belongs here is the question, not the
// answer.
//
// The questions are the three attributes. Everything the rules do with a kind
// is ask one of them, which is worth stating plainly: for a long time the rules
// asked "is this an aggregate" when they meant "is this the domain model", and
// the two only happened to coincide.
type ConstructKind struct {
	name    string
	article string

	needsRequirement bool
	domainModel      bool
	storageDecision  bool
	movesLifecycle   bool
	performsWork     bool
}

// NewConstructKind declares a role a frontend can recognise.
//
// article is given rather than derived because the rule follows pronunciation,
// not spelling: "a use case", because "use" begins with a consonant sound
// despite starting with a vowel letter. A generic vowel test gets that wrong.
func NewConstructKind(name, article string, opts ...ConstructKindOption) ConstructKind {
	k := ConstructKind{name: name, article: article}
	for _, opt := range opts {
		opt(&k)
	}
	return k
}

// ConstructKindOption sets one of the questions a rule may ask about a role.
type ConstructKindOption func(*ConstructKind)

// NeedsRequirement marks a role whose instances must each name a requirement.
//
// It is for the roles that carry business meaning. The structural ones are
// covered through whatever uses them, and demanding a binding for each would
// only produce noise.
func NeedsRequirement() ConstructKindOption {
	return func(k *ConstructKind) { k.needsRequirement = true }
}

// IsDomainModel marks a role whose fields must each trace to a requirement.
//
// Types are created deliberately and reviewed; fields accrete afterwards, which
// is where the drift is. A field of the domain model states what the system
// believes about the thing it describes, and one that traces to nothing is
// either an unrecorded promise or dead weight.
func IsDomainModel() ConstructKindOption {
	return func(k *ConstructKind) { k.domainModel = true }
}

// EmbodiesStorageDecision marks a role that exists because somebody chose how
// data is kept, and must therefore point at that choice.
//
// Forward coverage does not ask this, because these roles carry no requirement
// of their own. The decision behind them is invisible in the type and is
// exactly what a later reader needs.
func EmbodiesStorageDecision() ConstructKindOption {
	return func(k *ConstructKind) { k.storageDecision = true }
}

// MovesLifecycle marks a role whose instances each move an aggregate into a
// state, and must therefore say which one.
//
// It is the question behind a lifecycle: an event that does not name its
// resulting state leaves the aggregate's states knowable only by reading every
// fold. The set of states a thing can be in is the first question anybody asks
// about it and the last one the code answers.
func MovesLifecycle() ConstructKindOption {
	return func(k *ConstructKind) { k.movesLifecycle = true }
}

// PerformsWork marks a role that is a step somebody takes, and can therefore
// appear in a process.
//
// It is not the same question as NeedsRequirement, which is why it is its own.
// A command carries business meaning and needs a requirement, but it is a
// message rather than an action; a projection is a read model that nobody
// performs. Conflating the two would put both into a process diagram, where
// neither is a step.
func PerformsWork() ConstructKindOption {
	return func(k *ConstructKind) { k.performsWork = true }
}

// WithArticle renders the kind as a noun phrase for diagnostics.
func (k ConstructKind) WithArticle() string {
	if k.article == "" {
		return "an unknown construct"
	}
	return k.article
}

func (k ConstructKind) String() string {
	if k.name == "" {
		return "unknown"
	}
	return k.name
}

// NeedsRequirement reports whether a construct of this kind has to be bound to
// a requirement.
func (k ConstructKind) NeedsRequirement() bool { return k.needsRequirement }

// IsDomainModel reports whether the fields of this construct must each trace to
// a requirement.
func (k ConstructKind) IsDomainModel() bool { return k.domainModel }

// EmbodiesStorageDecision reports whether this construct must point at the
// decision that put it there.
func (k ConstructKind) EmbodiesStorageDecision() bool { return k.storageDecision }

// MovesLifecycle reports whether this construct must name the state it leads to.
func (k ConstructKind) MovesLifecycle() bool { return k.movesLifecycle }

// PerformsWork reports whether this construct is a step somebody takes.
func (k ConstructKind) PerformsWork() bool { return k.performsWork }

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
	// Fields are the exported fields when the construct is a struct, empty
	// otherwise.
	//
	// This is the domain shape, which is not the same set as the persisted one
	// and is not covered by it. A stored entity carries the historical values,
	// which is what a review or a migration has to work from; the domain type
	// beside it carries the meaning. Neither answers for the other.
	Fields []SchemaField
	Pos    Position
}

// Target renders the construct as the binding target that would annotate it.
func (c Construct) Target() Target {
	return Target{Kind: TargetType, Package: c.Package, Name: c.Name}
}
