package golang

// The framework contract.
//
// speclink recognises constructs by matching against the framework: a type
// whose first parameter is auth.Subject is a use case, a type with Evolve and
// Discriminator is an event, a call to permission.Declare is a permission. All
// of that matching is done against strings, and that is deliberate — one binary
// serves projects on different framework versions, so speclink must neither
// link the framework nor pin it. Its own go.mod does not require it and must
// not.
//
// The price is that every one of these strings is an unchecked claim, with a
// failure mode worse than being wrong: a stale claim does not fail, it silently
// disables the rule that depends on it. That is not hypothetical. The generic
// CRUD user interface moved from presentation/ui/ent to application/ent/ui, the
// constant kept pointing at the old location, and K4-NO-GENERIC-CRUD went on
// passing because it never matched anything again.
//
// The claims are therefore declared here, once, as data, and checked against a
// real framework at test time. Declaring them is what makes the check possible
// at all: a hand maintained list beside the constants is the same failure one
// level up, and it had already happened — the previous test never mentioned
// nagoDataJSON, so the JSON repository paths were unchecked from the day they
// were added.
//
// Each entry says which rule dies when it stops resolving, so a version bump
// that moves something reports what it broke instead of what it could not find.
//
// This is also the shape a second language frontend needs. Recognisers there
// resolve only within the project and cannot follow a type into a library, so
// the framework boundary has to be a declared set of named symbols no matter
// what the language is. Making it declared here is the part of that work that
// pays for itself immediately.

const (
	nagoAuth       = "go.wdy.de/nago/auth"
	nagoUser       = "go.wdy.de/nago/application/user"
	nagoPermission = "go.wdy.de/nago/application/permission"
	nagoEvs        = "go.wdy.de/nago/application/evs"
	nagoHapi       = "go.wdy.de/nago/application/hapi"
	nagoData       = "go.wdy.de/nago/pkg/data"
	nagoDataJSON   = "go.wdy.de/nago/pkg/data/json"
	nagoNdb        = "go.wdy.de/nago/pkg/ndb"
	nagoEnt        = "go.wdy.de/nago/application/ent"
	nagoEntCfg     = "go.wdy.de/nago/application/ent/cfg"
	nagoUIEnt      = "go.wdy.de/nago/application/ent/ui"
)

// frameworkSymbol is one thing a recogniser matches on.
//
// Name empty means the package itself is what is matched, which is how the
// generic CRUD user interface is banned: nothing in it is named, the import is
// the finding.
type frameworkSymbol struct {
	Pkg  string
	Name string
	// Breaks names the rule that stops working when this no longer resolves.
	Breaks string
}

// frameworkContract is every framework name speclink depends on.
//
// Adding a recogniser without adding its symbols here is the mistake this file
// exists to prevent, and the only defence against it is that the list is short
// and lives next to nothing else.
var frameworkContract = []frameworkSymbol{
	// A use case is a named func type whose first parameter is a subject.
	// auth.Subject is an alias of user.Subject, so the resolved type reports
	// the user package and matching only auth would recognise no use case at
	// all — silently, in every project.
	{nagoAuth, "Subject", "every rule that needs a use case: K1, K5, K6"},
	{nagoUser, "Subject", "the same, through the alias the type checker resolves to"},
	{nagoPermission, "Auditable", "recognising a use case that takes a permission subject"},

	// A query is a use case that returns neither an error nor a commit
	// sequence. evs.SeqID is an alias of ndb.Seq; an unmatched sequence makes
	// every writing use case look like a query.
	{nagoEvs, "SeqID", "telling a command from a query"},
	{nagoNdb, "Seq", "the same, through the alias the type checker resolves to"},

	// Event sourcing.
	{nagoEvs, "Evt", "recognising an event"},
	{nagoEvs, "Cmd", "recognising a command"},
	{nagoEvs, "Discriminator", "reading the serialisation tag, and with it every K9 rule"},
	{nagoEvs, "NewProjection", "recognising a projection"},
	{nagoEvs, "NewSingleton", "recognising a single instance projection"},

	// Storage.
	{nagoData, "Repository", "recognising a repository"},
	{nagoData, "Aggregate", "recognising an aggregate by its Identity method"},
	{nagoDataJSON, "NewJSONRepository", "finding the persistence model, and with it K1-PERSISTENCE-UNJUSTIFIED and the K9 rules over stored shapes"},
	{nagoDataJSON, "NewSloppyJSONRepository", "the same, where domain and persistence model are one type"},

	// Permissions. The recogniser matches any Declare prefix, so these are the
	// variants the reference project actually uses; a rename of the family
	// would show up on the first of them.
	{nagoPermission, "Declare", "K5-UC-PERMISSION"},
	{nagoPermission, "DeclareCreate", "K5-UC-PERMISSION for the create variant"},
	{nagoPermission, "DeclareFindByID", "K5-UC-PERMISSION for the read variant"},
	{nagoPermission, "DeclareFindAll", "K5-UC-PERMISSION for the list variant"},

	// The generic CRUD ban. These produce specification facts at run time,
	// which a static analysis can only see by reimplementing framework
	// internals.
	{nagoEntCfg, "Enable", "K4-NO-GENERIC-CRUD"},
	{nagoEnt, "DeclarePermissions", "K4-NO-GENERIC-CRUD"},
	{nagoEnt, "NewUseCases", "K4-NO-GENERIC-CRUD"},
	{nagoUIEnt, "", "K4-NO-GENERIC-CRUD, which bans the import itself"},
}
