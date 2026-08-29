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
}

// NodeKind discriminates what a node is. It is opaque to callers.
type NodeKind int

const (
	nodeStart NodeKind = iota + 1
	nodeEnd
	nodeActivity
	nodeEmit
	nodeCatch
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
}

// Start is where a process begins. A process may have more than one.
func Start(id, label string) Node { return Node{kind: nodeStart, id: id, label: label} }

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
