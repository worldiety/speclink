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
}
