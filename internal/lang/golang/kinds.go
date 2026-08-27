package golang

import "github.com/worldiety/speclink/internal/ir"

// The architectural roles this frontend recognises.
//
// They are the vocabulary of a domain driven, event sourced design as the nago
// framework spells it, which is why they are declared here and not in the
// language neutral model. Another framework has other roles and the same three
// questions to answer about each of them.
//
// Each answer is a decision worth reading rather than a flag worth copying.
var (
	// A named func type whose first parameter is an auth subject. It carries
	// business meaning and must say which requirement it was written for.
	ConstructUseCase = ir.NewConstructKind("use case", "a use case", ir.NeedsRequirement())

	// A use case that returns neither an error nor a commit sequence.
	//
	// It needs a requirement for the same reason a use case does. Reading is a
	// promise too — that someone may see a thing — and every architecture rule
	// already treats a query as a use case: its own file, its constructor, its
	// permission, its place in the bundle. Exempting it from coverage alone was
	// an omission rather than a decision, and it made a recogniser defect
	// expensive: while evs.SeqID went unrecognised, every writing use case was
	// classified as a query and silently left the denominator.
	ConstructQuery = ir.NewConstructKind("query", "a query", ir.NeedsRequirement())

	// Implements Decide, and is the write side of an aggregate.
	ConstructCommand = ir.NewConstructKind("command", "a command", ir.NeedsRequirement())

	// Implements Evolve plus Discriminator, and is a domain fact that outlives
	// the code that wrote it.
	ConstructEvent = ir.NewConstructKind("event", "an event", ir.NeedsRequirement())

	// An event folded read model, built by evs.NewProjection or
	// evs.NewSingleton.
	ConstructProjection = ir.NewConstructKind("projection", "a projection", ir.NeedsRequirement())

	// A consistency boundary with an identity.
	//
	// It carries no requirement of its own — it is covered through whatever
	// writes it — but its fields do, and its existence answers a question about
	// how the data is kept.
	ConstructAggregate = ir.NewConstructKind("aggregate", "an aggregate",
		ir.IsDomainModel(), ir.EmbodiesStorageDecision())

	// A named type standing for data.Repository over an aggregate. Structural,
	// and the plainest statement there is that something is stored as state.
	ConstructRepository = ir.NewConstructKind("repository", "a repository",
		ir.EmbodiesStorageDecision())

	// Declared by permission.Declare and bound to a use case through its type
	// parameter. Structural: it is covered through the use case it guards.
	ConstructPermission = ir.NewConstructKind("permission", "a permission")
)
