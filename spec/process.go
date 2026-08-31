package spec

import "reflect"

// Process is a business process written down as a graph.
//
// # Why a graph and not a list of steps
//
// Real processes branch, run things in parallel and come back to an earlier
// point. A quote that is sent back for rework returns to the drafting step; a
// submission fans out into invoicing and notification and waits for both. A
// list of steps cannot say any of that, and a nested form — sequence inside
// parallel inside choice — can say everything except the one thing that occurs
// constantly, which is the jump backwards.
//
// The price is that the compiler no longer checks the wiring: nodes are joined
// by name, and a name is a string. That is exactly the trade speclink makes
// everywhere and immediately pays for. What the compiler cannot check, a rule
// checks: K16 refuses a dangling edge, a duplicate name, an unreachable node
// and a node from which no end can be reached.
//
// # What a process is for
//
// It answers the question the building block view cannot: not what there is,
// but what happens, in which order, and where it can end. A use case is a
// promise about one action; a process is the promise that the actions add up.
//
// It satisfies requirements the way a construct does, so a requirement about
// the course of business is covered by the process rather than by whichever
// use case happens to be named in it.
type Process struct {
	// ID is stable and appears in diagnostics and in the generated document.
	ID string
	// Title is the one line a reader sees above the diagram.
	Title string
	// Purpose says what the process is for, in prose. The graph says how.
	Purpose string
	// Satisfies names the requirements this course of business answers to.
	Satisfies []Requirement
	// Nodes are the steps, gateways and endpoints, in no particular order.
	Nodes []Node
	// Edges join them. Order is irrelevant; the graph is what counts.
	Edges []Edge

	// Drawn says how the course is pictured. The default is a graph of what
	// happens; AsSequence pictures the same graph as who says what to whom.
	//
	// # Why one model and two drawings
	//
	// Because a sequence diagram and an activity graph are the same
	// information whenever control passes with the message, which is what a
	// synchronous request and response is. The participants are not declared:
	// each activity names a use case, the use case lives in a package, and the
	// package is the part of the system performing it. Declaring the lane as
	// well would be that fact written twice.
	//
	// The equality is not universal. Two branches of a fork have no order
	// between them, and a sequence drawing puts them side by side in a frame
	// rather than pretending there is one.
	Drawn View
}

// View is how a course of business is pictured.
type View int

const (
	// AsFlow is the default: a graph of what happens in what order.
	AsFlow View = iota
	// AsSequence pictures the same graph as an exchange between the parts of
	// the system, with time running downwards.
	AsSequence
)

// NodeKind discriminates what a node is. It is opaque to callers.
type NodeKind int

const (
	nodeStart NodeKind = iota + 1
	nodeEnd
	nodeActivity
	nodeEmit
	nodeCatch
	nodeSend
	nodeFork
	nodeJoin
	nodeChoice
	nodeMerge
)

// Node is one point in a process.
//
// The set is the smallest one that still describes a real process: a beginning,
// something happening, a fact being recorded, a decision, a fan out, and one or
// more ways to finish. Deliberately absent are inclusive and event based
// gateways, boundary events, timers, compensation and subprocesses. Each is
// expressible the moment a real case asks for it, and each one added before
// then would be another vocabulary nobody reads — which is the state
// spec.Transition was in until K15.
type Node struct {
	kind  NodeKind
	id    string
	label string
	ref   reflect.Type
	note  string
	actor string
}

// Note attaches an aside to a node.
//
// For what a reader has to know at that step and cannot read off the graph:
// that a lock is held elsewhere, that a comparison is against the previous run,
// that the field is a code and not a checksum. Without somewhere to put it, it
// ends up in the label — where it makes the box unreadable — or nowhere.
func (n Node) Note(text string) Node { n.note = text; return n }

// Start is where a process begins. A process may have more than one.
func Start(id, label string) Node { return Node{kind: nodeStart, id: id, label: label} }

// StartedBy is a beginning somebody outside the system brings about.
//
// The actor is named because in a drawing of who talks to whom, a process that
// simply begins has nobody on the left of it. It is the same fact a channel
// states about a boundary, said at the moment it is crossed rather than about
// the boundary in general.
func StartedBy(actor Actor, id, label string) Node {
	return Node{kind: nodeStart, id: id, label: label, actor: actor.ID}
}

// End is where a process finishes.
//
// Several are normal and are the point: "accepted" and "rejected" are different
// outcomes, and a process that models them as one endpoint has thrown away the
// distinction anybody actually cares about.
func End(id, label string) Node { return Node{kind: nodeEnd, id: id, label: label} }

// Do is an activity, and names the use case that performs it.
//
// The type parameter is the link into the code: the compiler checks that the
// use case exists, and a rule checks that it really is one. A process step that
// named its activity in prose would be a caption, and a caption cannot go stale
// in a way anybody notices.
func Do[T any](id string) Node {
	return Node{kind: nodeActivity, id: id, ref: reflect.TypeFor[T]()}
}

// Emit records that a domain event is raised at this point.
func Emit[T any](id string) Node {
	return Node{kind: nodeEmit, id: id, ref: reflect.TypeFor[T]()}
}

// On waits for a domain event before continuing.
func On[T any](id string) Node {
	return Node{kind: nodeCatch, id: id, ref: reflect.TypeFor[T]()}
}

// Send is a message crossing a boundary of this system.
//
// The type parameter is the payload of a declared message, so the compiler
// checks the type exists and a rule checks that some channel actually carries
// it.
//
// # Why this is not an activity
//
// Because control does not pass to it. An activity is work this module
// performs and answers for; a message goes to another program, which may be
// slow, absent, or a different version of itself. Drawing the two the same way
// is what makes a picture of a distributed system read like a call stack, and
// reading it that way is how the failure modes get forgotten.
func Send[T any](id string) Node {
	return Node{kind: nodeSend, id: id, ref: reflect.TypeFor[T]()}
}

// Fork splits into branches that all run.
func Fork(id string) Node { return Node{kind: nodeFork, id: id} }

// Join waits for every branch of a Fork.
func Join(id string) Node { return Node{kind: nodeJoin, id: id} }

// Choice takes exactly one branch. Every edge leaving it states its condition.
func Choice(id string) Node { return Node{kind: nodeChoice, id: id} }

// Merge brings alternative branches back together without waiting.
func Merge(id string) Node { return Node{kind: nodeMerge, id: id} }

// Edge joins two nodes.
type Edge struct {
	// From and To name nodes of the same process.
	From, To string
	// When is the condition, and is required on every edge leaving a Choice.
	// An unlabelled alternative is a decision nobody wrote down.
	When string
}
