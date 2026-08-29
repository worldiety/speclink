package ir

// NodeKind is what a process node is, independent of how a language spells it.
type NodeKind int

const (
	// NodeStart begins a process.
	NodeStart NodeKind = iota + 1
	// NodeEnd finishes one.
	NodeEnd
	// NodeActivity performs a use case.
	NodeActivity
	// NodeEmit raises a domain event.
	NodeEmit
	// NodeCatch waits for one.
	NodeCatch
	// NodeFork splits into branches that all run.
	NodeFork
	// NodeJoin waits for every branch of a fork.
	NodeJoin
	// NodeChoice takes exactly one branch.
	NodeChoice
	// NodeMerge brings alternatives back together without waiting.
	NodeMerge
)

func (k NodeKind) String() string {
	switch k {
	case NodeStart:
		return "start"
	case NodeEnd:
		return "end"
	case NodeActivity:
		return "activity"
	case NodeEmit:
		return "event raised"
	case NodeCatch:
		return "event awaited"
	case NodeFork:
		return "fork"
	case NodeJoin:
		return "join"
	case NodeChoice:
		return "choice"
	case NodeMerge:
		return "merge"
	}
	return "unknown"
}

// Splits reports whether this kind may have more than one outgoing edge.
func (k NodeKind) Splits() bool { return k == NodeFork || k == NodeChoice }

// Merges reports whether this kind may have more than one incoming edge.
func (k NodeKind) Merges() bool { return k == NodeJoin || k == NodeMerge }

// Gateway reports whether the node routes rather than does anything.
//
// A gateway carries no meaning of its own, which is why the rules treat it
// differently in both directions: it needs no requirement, and a process built
// only from gateways does nothing at all.
func (k NodeKind) Gateway() bool { return k.Splits() || k.Merges() }

// References reports whether this kind names a construct in the code.
func (k NodeKind) References() bool {
	return k == NodeActivity || k == NodeEmit || k == NodeCatch
}

// ProcessNode is one point in a process graph.
type ProcessNode struct {
	Kind NodeKind
	// ID is unique within the process and is what edges name.
	ID string
	// Label is the human readable text, set on start and end nodes.
	Label string
	// Ref is the fully qualified construct this node names, empty for
	// gateways and endpoints.
	Ref string
	// RefPackage is the package Ref lives in, so that a rule can tell a
	// misspelled reference from one into a package the scope left out.
	RefPackage string
	Pos        Position
}

// ProcessEdge joins two nodes of one process.
type ProcessEdge struct {
	From, To string
	// When is the condition on a branch leaving a choice.
	When string
	Pos  Position
}

// Process is a course of business written down as a graph.
//
// Satisfies holds qualified Go identifiers rather than requirement IDs, for the
// same reason DerivedFrom does: the declarations are collected in one pass and
// resolved in a second, which is what makes forward references legal and the
// input order irrelevant.
type Process struct {
	ID      string
	Title   string
	Purpose string

	Satisfies []string

	Nodes []ProcessNode
	Edges []ProcessEdge

	// GoIdent is the declaration this came from, for diagnostics.
	GoIdent string
	Pos     Position
}

// Node returns the node of this ID, and whether it exists.
func (p *Process) Node(id string) (ProcessNode, bool) {
	for _, n := range p.Nodes {
		if n.ID == id {
			return n, true
		}
	}
	return ProcessNode{}, false
}

// Out returns the edges leaving a node.
func (p *Process) Out(id string) []ProcessEdge {
	var out []ProcessEdge
	for _, e := range p.Edges {
		if e.From == id {
			out = append(out, e)
		}
	}
	return out
}

// In returns the edges arriving at a node.
func (p *Process) In(id string) []ProcessEdge {
	var in []ProcessEdge
	for _, e := range p.Edges {
		if e.To == id {
			in = append(in, e)
		}
	}
	return in
}
