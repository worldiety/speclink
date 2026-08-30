package ir

// Kind classifies what a requirement is. Mirrors spec.Kind.
type Kind int

const (
	Functional Kind = iota + 1
	NonFunctional
	Constraint
	Decision
)

func (k Kind) String() string {
	switch k {
	case Functional:
		return "functional"
	case NonFunctional:
		return "non-functional"
	case Constraint:
		return "constraint"
	case Decision:
		return "decision"
	}
	return "unknown"
}

// Dir is the first level directory of the requirement tree that a Kind implies.
//
// Functional requirements are the majority and carry no natural cross cutting
// home, so they are grouped by domain below fun/. The other three kinds are
// cross cutting by definition and are grouped by kind.
func (k Kind) Dir() string {
	switch k {
	case Functional:
		return "fun"
	case NonFunctional:
		return "nfr"
	case Constraint:
		// Not "con": CON is a reserved device name on Windows, and Go refuses
		// it as an import path element. A directory nobody can compile is not
		// a convention, it is a trap.
		return "cst"
	case Decision:
		return "dec"
	}
	return ""
}

// Discipline records who owns a requirement. Mirrors spec.Discipline.
type Discipline int

const (
	Business Discipline = iota + 1
	Technical
	Mixed
)

func (d Discipline) String() string {
	switch d {
	case Business:
		return "business"
	case Technical:
		return "technical"
	case Mixed:
		return "mixed"
	}
	return "unknown"
}

// Status drives the backward direction of the coverage check. Mirrors
// spec.Status.
type Status int

const (
	Normative Status = iota + 1
	Abstract
	Planned
	OutOfScope
	Informative
	Superseded
)

func (s Status) String() string {
	switch s {
	case Normative:
		return "normative"
	case Abstract:
		return "abstract"
	case Planned:
		return "planned"
	case OutOfScope:
		return "out-of-scope"
	case Informative:
		return "informative"
	case Superseded:
		return "superseded"
	}
	return "unknown"
}

// MustBeCovered reports whether a requirement of this status has to be
// referenced by at least one construct. Only Normative must; every other value
// is an explicit, justified exemption.
func (s Status) MustBeCovered() bool { return s == Normative }

// Role classifies accompanying material. Mirrors spec.Role.
type Role int

const (
	Mockup Role = iota + 1
	Scribble
	Diagram
	AcceptanceCriteria
	Protocol
	Document
)

func (r Role) String() string {
	switch r {
	case Mockup:
		return "mockup"
	case Scribble:
		return "scribble"
	case Diagram:
		return "diagram"
	case AcceptanceCriteria:
		return "acceptance-criteria"
	case Protocol:
		return "protocol"
	case Document:
		return "document"
	}
	return "unknown"
}

// Source names where a requirement originates.
type Source struct {
	Doc    string
	Anchor string
	Extern string
	Note   string
	Pos    Position
}

// Attachment is accompanying material of a requirement.
type Attachment struct {
	Path string
	Role Role
	Note string
}

// Requirement is a resolved requirement node.
//
// DerivedFrom and Supersedes hold requirement IDs rather than pointers: the
// graph is assembled after all declarations have been collected, so forward
// references are legal and order is irrelevant.
type Requirement struct {
	ID           string
	GoIdent      string // qualified Go identifier, e.g. "…/requirements/fun/quote.RQuoteSubmit"
	Kind         Kind
	Discipline   Discipline
	Status       Status
	Title        string
	Text         string
	Detail       string
	Rationale    string
	Consequences string
	DerivedFrom  []string
	Supersedes   []string
	Sources      []Source
	Attachments  []Attachment
	// Topics holds qualified Go identifiers until they are resolved, then
	// topic IDs — the same two pass shape DerivedFrom uses, for the same
	// reason: order of declaration must not matter.
	Topics []string
	Pos    Position
}

// Topic is a theme requirements are grouped under, and a chapter.
type Topic struct {
	ID          string
	GoIdent     string
	Title       string
	Description string
	Pos         Position
}

// Model is the fully collected, not yet checked specification model.
type Model struct {
	// Requirements is keyed by requirement ID.
	Requirements map[string]*Requirement
	// Bindings are all assertions found in annotation files.
	Bindings []Binding
}

// NewModel returns an empty model.
func NewModel() *Model {
	return &Model{Requirements: map[string]*Requirement{}}
}
