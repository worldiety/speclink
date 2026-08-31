package spec

// The world outside the code, and the ways the system reaches it.
//
// # Why any of this is declared
//
// Everything else speclink knows it reads. A use case, a repository, an event —
// the recognisers find them, and a fact that is inferable must not be
// annotated. None of that works here. There is no declaration anywhere in a Go
// module that says an end user exists, that the object store is somebody else's
// responsibility, or that the channel to it carries customer data under a
// short lived key. That is not an omission in the code; it is knowledge the
// code cannot hold.
//
// So it is stated, and then held against what the code does show. An adapter is
// where a system touches something outside — the architecture says so itself,
// and speclink can see every one of them. A channel that names no adapter and
// an adapter that no channel names are both reported, which is the only reason
// this is a model rather than a picture.

// Actor is a person or role outside the system boundary.
type Actor struct {
	// ID is what a channel names, and is stable.
	ID string
	// Name is what a reader sees.
	Name string
	// Role says what this person does with the system, in one sentence.
	Role string
	// Satisfies names the requirements this participant answers to.
	//
	// A person outside the boundary is rarely there for no reason: somebody
	// decided this role exists and that the system deals with it. Where that
	// decision is written down, naming it here is what keeps the drawing and
	// the requirement tree from drifting into two different accounts of who
	// uses this system.
	Satisfies []Requirement
	// Topics are the themes this participant belongs to. Optional, as
	// everywhere else: a carelessly assigned theme is worse than none.
	Topics []Topic
}

// Foreign is a system outside the boundary that this one talks to.
//
// The distinction from an actor is not decoration. What a person does is the
// system's problem to authorise; what a foreign system does is somebody else's
// to answer for, and where that line falls decides how far any assurance about
// this system reaches.
type Foreign struct {
	// ID is what a channel names, and is stable.
	ID string
	// Name is what a reader sees.
	Name string
	// Role says what it is to this system, and whose responsibility it is.
	Role string
	// Satisfies names the requirements this participant answers to.
	//
	// A person outside the boundary is rarely there for no reason: somebody
	// decided this role exists and that the system deals with it. Where that
	// decision is written down, naming it here is what keeps the drawing and
	// the requirement tree from drifting into two different accounts of who
	// uses this system.
	Satisfies []Requirement
	// Topics are the themes this participant belongs to. Optional, as
	// everywhere else: a carelessly assigned theme is worse than none.
	Topics []Topic
}

// Channel is one way across the boundary, or one way between two parts.
//
// # Why the four descriptive fields are required
//
// Protocol, Data, Auth and Crypto are not adornment. They are the answer to
// "what crosses here, who may, and what protects it" — the question every
// review of an interface starts with, and the one that is normally answered by
// reading the code of both ends and hoping. A channel that leaves one of them
// empty is refused, because the empty one is always the interesting one.
type Channel struct {
	// From and To name an actor, a foreign system, or a package of this module
	// written as its repository relative directory: app/sales/adapter/fs.
	//
	// A package rather than a type, because a channel is not a thing in the
	// code — it is the boundary a directory sits on. The rules check that the
	// name resolves to one of the three, which is the check a string endpoint
	// costs and buys.
	From, To string

	// Label names the channel in a diagram.
	Label string
	// Protocol is how it is carried: HTTPS, SSH, wss, systemd.
	Protocol string
	// Data says what crosses, in the terms a reader of a data protection
	// register would use.
	Data string
	// Auth says how the far end is established, or plainly that it is not.
	Auth string
	// Crypto says what protects it in transit, or plainly that nothing does.
	Crypto string

	// Satisfies names the requirements this channel answers to. A way across
	// the boundary that answers to nothing is a way somebody opened without
	// being asked.
	Satisfies []Requirement

	// Topics are the themes this channel belongs to. Optional, as everywhere
	// else.
	Topics []Topic

	// Envelope is the Go type wrapping every message on this channel, given as
	// a zero value. Stated once here rather than repeated in each message,
	// because it is one fact about the channel and fourteen copies of it would
	// disagree the first time somebody added a field.
	Envelope any

	// Messages are the kinds of thing that cross, for a channel that carries a
	// protocol rather than a single payload.
	//
	// Declared separately and named here, so that a process step can send one
	// and a message can name the one that answers it. Contract is the short
	// form for a boundary carrying exactly one shape; stating both is refused,
	// because then two fields describe what crosses and a reader has to work
	// out which is authoritative.
	Messages []Message

	// Contract is the Go type whose shape crosses here, given as a zero value:
	//
	//	Contract: fs.QuoteFile{}
	//
	// # Why this exists
	//
	// speclink freezes the surface this system offers. Every route it serves
	// carries a recorded shape, and a field removed from a response is a
	// finding. It knew nothing at all about the surfaces this system depends
	// on, which is the same edge seen from the other side and the more
	// dangerous one: a promise this system makes is broken by somebody who can
	// be told, and a promise it relies on is broken by somebody who cannot.
	//
	// The four descriptive fields say what crosses in the words of a data
	// protection register. That is the right register for a reviewer and no use
	// at all to a compiler. This is the same fact in a form that can be
	// compared: the structure is recorded in speclink.lock, and a change to it
	// is reported the next run.
	//
	// # Why a type and not a schema file
	//
	// The type is already in the code — it is what the adapter marshals. A
	// separate schema would be the same fact written twice, and the copy would
	// be the one that rots. Where the far end is not modelled in this module
	// at all, leave it empty: an unstated contract is honest, and a fabricated
	// one is not.
	Contract any
}
