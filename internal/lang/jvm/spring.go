package jvm

import "github.com/worldiety/speclink/internal/ir"

// The Spring and Jakarta names the recognisers match on.
//
// The same coupling as the Go frontend's, for the same reason: speclink must
// not depend on the framework it recognises, because one binary serves projects
// that are not all on the same version of it. So these are strings, and a
// string is an unchecked claim whose failure is silence — a renamed annotation
// does not break a rule, it stops the rule matching anything.
//
// Reading bytecode makes the claim narrower than it was in Go. There, matching
// meant reasoning about method sets, type aliases and signatures, and two real
// bugs came out of it: auth.Subject is an alias for user.Subject, and evs.SeqID
// for ndb.Seq, each of which silently disabled a recogniser. A class file has
// no aliases and no imports. An annotation is present under its fully qualified
// name or it is not.
//
// What the Go frontend needed inference for, Spring simply declares. That is
// the substance of the comparison: an architecture written in annotations is
// visible to a reader that only sees declarations, where an architecture
// written in signatures is not.
const (
	springStereotype = "org.springframework.stereotype"
	springWeb        = "org.springframework.web.bind.annotation"
	springData       = "org.springframework.data.repository"
	jakartaPersist   = "jakarta.persistence"
	javaxPersist     = "javax.persistence"
)

// frameworkSymbol is one name a recogniser matches on.
type frameworkSymbol struct {
	Name string
	// Breaks names the role that stops being recognised when this no longer
	// matches, so a framework upgrade reports what it took with it rather than
	// what it could not find.
	Breaks string
}

// frameworkContract is every Spring or Jakarta name this frontend depends on.
//
// Declared as data so one list serves both the recognisers and the guard that
// checks them. A hand maintained copy beside the constants is the same failure
// one level up, and in the Go frontend it had already happened.
var frameworkContract = []frameworkSymbol{
	{springWeb + ".RestController", "endpoints, and with them forward coverage of the web layer"},
	{springStereotype + ".Controller", "endpoints declared the older way"},
	{springStereotype + ".Service", "application services"},
	{springStereotype + ".Repository", "repositories declared by stereotype"},

	{springWeb + ".RequestMapping", "recognising which methods of a controller are endpoints"},
	{springWeb + ".GetMapping", "the read variant"},
	{springWeb + ".PostMapping", "the write variant"},
	{springWeb + ".PutMapping", "the replace variant"},
	{springWeb + ".DeleteMapping", "the delete variant"},
	{springWeb + ".PatchMapping", "the partial update variant"},

	{jakartaPersist + ".Entity", "entities, and with them field coverage and the storage decision"},
	{javaxPersist + ".Entity", "the same, for projects still on the javax namespace"},

	{springData + ".Repository", "repositories declared by extending Spring Data"},
	{springData + ".CrudRepository", "the same, through the commonest subinterface"},
}

// mappingAnnotations are the ones that make a controller method an endpoint.
var mappingAnnotations = []string{
	springWeb + ".RequestMapping",
	springWeb + ".GetMapping",
	springWeb + ".PostMapping",
	springWeb + ".PutMapping",
	springWeb + ".DeleteMapping",
	springWeb + ".PatchMapping",
}

// The architectural roles this frontend recognises.
//
// Spring's vocabulary, not nago's, which is the point of declaring roles in a
// frontend rather than in the neutral model. The three questions underneath are
// the same ones, and answering them is what a frontend is for.
var (
	// A method of a controller that answers a request. It is the boundary of
	// the system and the plainest thing a requirement can be written about.
	ConstructEndpoint = ir.NewConstructKind("endpoint", "an endpoint", ir.NeedsRequirement())

	// A public method of an application service: one operation somebody asked
	// for, which is what a use case is.
	ConstructService = ir.NewConstructKind("service operation", "a service operation", ir.NeedsRequirement())

	// A persistent entity. It carries no requirement of its own — it is covered
	// through whatever writes it — but its fields do, and its existence answers
	// a question about how the data is kept.
	ConstructEntity = ir.NewConstructKind("entity", "an entity",
		ir.IsDomainModel(), ir.EmbodiesStorageDecision())

	// A repository, by stereotype or by extending Spring Data. Structural, and
	// the plainest statement there is that something is stored as state.
	ConstructRepository = ir.NewConstructKind("repository", "a repository",
		ir.EmbodiesStorageDecision())
)
