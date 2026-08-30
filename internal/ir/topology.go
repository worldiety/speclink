package ir

// ParticipantKind distinguishes who is on the far end.
type ParticipantKind int

const (
	// ParticipantActor is a person or role.
	ParticipantActor ParticipantKind = iota + 1
	// ParticipantForeign is a system somebody else answers for.
	ParticipantForeign
)

func (k ParticipantKind) String() string {
	switch k {
	case ParticipantActor:
		return "actor"
	case ParticipantForeign:
		return "foreign system"
	}
	return "unknown"
}

// Participant is something outside the code that the system deals with.
type Participant struct {
	Kind ParticipantKind
	ID   string
	Name string
	Role string
	Pos  Position
}

// Channel is one way across a boundary.
//
// Satisfies holds qualified Go identifiers rather than requirement IDs, as
// every other declaration here does, so that resolution happens once
// everything has been collected.
type Channel struct {
	From, To string

	Label    string
	Protocol string
	Data     string
	Auth     string
	Crypto   string

	Satisfies []string
	// Contract is the structure that crosses, where the declaration names a
	// type for it. Nil where none was stated, which is different from a
	// channel that carries nothing: one is unknown and the other is empty.
	Contract *WireShape

	Pos Position
}

// Topology is what a frontend can say about the world around the code.
//
// It carries the declarations and the facts they are held against in one value,
// because a rule needs both to say anything: a channel naming a package that
// does not exist and an adapter that no channel names are the same question
// asked from opposite ends, and answering only one of them is how a model
// becomes a picture.
type Topology struct {
	// Participants are the declared actors and foreign systems.
	Participants []Participant
	// Channels are the declared ways across.
	Channels []Channel

	// Packages are the repository relative directories the run measured, so
	// that a misspelled endpoint can be told from a package the scope left out.
	Packages map[string]bool
	// Adapters are the measured packages where this system touches something
	// outside. The architecture says which those are; nothing is guessed.
	Adapters []Adapter
}

// Adapter is one place the system reaches out.
//
// It carries both spellings of its name because they answer to different
// readers. Dir is what a channel endpoint says, because a channel is about a
// boundary a directory sits on. Pkg is what a waiver says, because a waiver is
// written in the language and keys on an import path.
type Adapter struct {
	Dir string
	Pkg string
	Pos Position
}

// Declared reports whether the project has a topology at all.
//
// Before the first declaration there is nothing to be incomplete against, and
// reporting every adapter would be demanding adoption rather than reporting a
// gap — the same ramp processes are on.
func (t Topology) Declared() bool {
	return len(t.Participants) > 0 || len(t.Channels) > 0
}

// Binding renders a channel as the satisfier it is, so that coverage counts it
// through the machinery every other satisfier goes through.
func (c Channel) Binding() Binding {
	as := make([]Assertion, 0, 1)
	if len(c.Satisfies) > 0 {
		as = append(as, Assertion{
			Kind:         AssertSatisfies,
			Pos:          c.Pos,
			Requirements: c.Satisfies,
		})
	}
	return Binding{
		Target:     Target{Kind: TargetChannel, Name: c.Name()},
		Assertions: as,
		Pos:        c.Pos,
	}
}

// Name renders the channel for a diagnostic and for the coverage table.
func (c Channel) Name() string {
	label := c.Label
	if label == "" {
		label = "channel"
	}
	return label + " (" + c.From + " → " + c.To + ")"
}
