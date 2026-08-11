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
)

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
	}
	return "unknown"
}

// NeedsRequirement reports whether a construct of this kind has to be bound to
// a requirement (forward coverage, K1/K3).
//
// Use cases, commands and events carry business meaning and must trace back to
// something that was asked for. Permissions and aggregates are structural: they
// are covered transitively through the use case that guards or writes them, and
// demanding a binding for each would only produce noise.
func (k ConstructKind) NeedsRequirement() bool {
	switch k {
	case ConstructUseCase, ConstructCommand, ConstructEvent:
		return true
	}
	return false
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
