package check

import (
	"sort"
	"strconv"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
)

// Rule IDs of the process checks.
const (
	// RuleProcessNoStart fires when a process has no beginning.
	RuleProcessNoStart = "K16-PROCESS-NO-START"
	// RuleProcessNoEnd fires when a process has no way to finish.
	RuleProcessNoEnd = "K16-PROCESS-NO-END"
	// RuleProcessNoActivity fires when a process performs nothing.
	RuleProcessNoActivity = "K16-PROCESS-NO-ACTIVITY"
	// RuleProcessDuplicate fires when two processes claim one ID.
	RuleProcessDuplicate = "K16-PROCESS-DUPLICATE"
	// RuleProcessUnbound fires when a process names no requirement.
	RuleProcessUnbound = "K16-PROCESS-UNBOUND"
	// RuleNodeDuplicate fires when two nodes of one process share a name.
	RuleNodeDuplicate = "K16-NODE-DUPLICATE"
	// RuleEdgeDangling fires when an edge names a node that does not exist.
	RuleEdgeDangling = "K16-EDGE-DANGLING"
	// RuleNodeDegree fires when the wiring at a node cannot mean anything.
	RuleNodeDegree = "K16-NODE-DEGREE"
	// RuleChoiceUnconditional fires when an alternative states no condition.
	RuleChoiceUnconditional = "K16-CHOICE-UNCONDITIONAL"
	// RuleNodeUnreachable fires when no start reaches a node.
	RuleNodeUnreachable = "K16-NODE-UNREACHABLE"
	// RuleNodeTrapped fires when a node reaches no end.
	RuleNodeTrapped = "K16-NODE-TRAPPED"
	// RuleNodeRefUnknown fires when a node names a construct that is not what
	// that kind of node performs.
	RuleNodeRefUnknown = "K16-NODE-REF-UNKNOWN"
)

// ProcessReport is what the run can say about the declared processes.
type ProcessReport struct {
	// Declared is how many processes were read. Zero means the project has not
	// adopted them, which is why no figure about processes is printed then:
	// a share of nothing is not a hundred percent, it is no claim at all.
	Declared int
	// Sound is how many came through every graph rule without a finding.
	Sound int
}

// Processes checks the declared courses of business.
//
// # What the compiler no longer does
//
// Nodes are joined by name because real processes branch and come back, and a
// nested form cannot express the jump backwards. The cost is that the wiring
// is strings, and every mistake a type checker would have caught has to be
// caught here instead: a name that exists twice, an edge that points at
// nothing, a node nothing reaches, a node from which nothing finishes.
//
// # The tracer, and what it does not prove
//
// Two depth first searches. Forward from every start finds what is
// unreachable; backward from every end finds what is trapped. The second is
// the one that matters, and it is why cycles are allowed at all: a process may
// loop, but every point in the loop must still have a way out.
//
// It deliberately proves less than soundness. Whether every fork is matched by
// exactly one join on every path is reachability in a Petri net and is not
// cheaply decidable; claiming it would be worse than not checking it. The
// degree rules make the common shapes right, and the run says plainly that
// deadlock freedom was not established.
func Processes(tree *reqtree.Tree, procs []*ir.Process, constructs []ir.Construct, scope map[string]bool, d ir.Dialect, out *diag.Set) ProcessReport {
	rep := ProcessReport{Declared: len(procs)}
	if len(procs) == 0 {
		return rep
	}

	known := map[string]ir.Construct{}
	for _, c := range constructs {
		known[c.Name] = c
	}

	sorted := append([]*ir.Process(nil), procs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Pos.Less(sorted[j].Pos) })

	seenID := map[string]*ir.Process{}
	for _, p := range sorted {
		if first, dup := seenID[p.ID]; dup {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 81),
				Pos:  p.Pos,
				Rule: RuleProcessDuplicate,
				What: "process " + quote(p.ID) + " is declared twice.",
				Why:  "The ID is how a process is referred to from a document and from a diagnostic. Two declarations under one name make every such reference ambiguous.",
				How:  "Rename one of them; the other is at " + first.Pos.String() + ".",
			})
		} else {
			seenID[p.ID] = p
		}

		before := out.Len()
		checkProcess(tree, p, known, scope, d, out)
		if out.Len() == before {
			rep.Sound++
		}
	}
	return rep
}

func checkProcess(tree *reqtree.Tree, p *ir.Process, known map[string]ir.Construct, scope map[string]bool, d ir.Dialect, out *diag.Set) {
	nodes := indexNodes(p, out)
	checkRequirements(tree, p, d, out)
	checkEdges(p, nodes, out)
	checkDegrees(p, nodes, out)
	checkRefs(p, known, scope, out)
	checkEndpoints(p, out)
	trace(p, nodes, out)
}

// indexNodes maps every node by ID and reports the ones declared twice.
func indexNodes(p *ir.Process, out *diag.Set) map[string]ir.ProcessNode {
	nodes := map[string]ir.ProcessNode{}
	for _, n := range p.Nodes {
		if _, dup := nodes[n.ID]; dup {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 73),
				Pos:  n.Pos,
				Rule: RuleNodeDuplicate,
				What: "process " + p.ID + " has two nodes called " + quote(n.ID) + ".",
				Why:  "Edges join nodes by name. Two nodes under one name make every edge touching it point at whichever the reader happens to pick.",
				How:  "Give each node a name of its own.",
			})
			continue
		}
		nodes[n.ID] = n
	}
	return nodes
}

func checkRequirements(tree *reqtree.Tree, p *ir.Process, d ir.Dialect, out *diag.Set) {
	if len(p.Satisfies) == 0 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 80),
			Pos:  p.Pos,
			Rule: RuleProcessUnbound,
			What: "process " + p.ID + " names no requirement.",
			Why:  "A process is the promise that the separate actions add up to something somebody asked for. Which promise that is cannot be read off the graph.",
			How:  "Add `Satisfies` naming the requirement this course of business answers to.",
		})
		return
	}
	for _, ref := range p.Satisfies {
		if tree.ByGoIdent(ref) != nil {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 80),
			Pos:  p.Pos,
			Rule: RuleProcessUnbound,
			What: "process " + p.ID + " names " + shortName(ref) + ", which is not a requirement.",
			Why:  "A reference that resolves to nothing looks like traceability and provides none.",
			How:  "Name a declared requirement, or add the one this process answers to.",
		})
	}
}

func checkEdges(p *ir.Process, nodes map[string]ir.ProcessNode, out *diag.Set) {
	for _, e := range p.Edges {
		for _, end := range []string{e.From, e.To} {
			if _, ok := nodes[end]; ok {
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 74),
				Pos:  e.Pos,
				Rule: RuleEdgeDangling,
				What: "process " + p.ID + " has an edge to " + quote(end) + ", which is not a node of it.",
				Why:  "Nodes are joined by name, so a misspelling is not a compile error. It is a step that silently never runs.",
				How:  "Correct the name, or declare the node.",
			})
		}
		if e.When == "" {
			if from, ok := nodes[e.From]; ok && from.Kind == ir.NodeChoice {
				out.Add(diag.Finding{
					Code: diag.Code(diag.PhaseSemantic, 76),
					Pos:  e.Pos,
					Rule: RuleChoiceUnconditional,
					What: "the branch from " + quote(e.From) + " to " + quote(e.To) + " states no condition.",
					Why:  "A choice takes exactly one branch. Which one is the decision the process exists to record, and an unlabelled alternative leaves it unwritten.",
					How:  "Set `When` to the condition under which this branch is taken.",
				})
			}
		}
	}
}

// checkDegrees enforces the wiring rules that make a graph readable.
//
// Only the outgoing side is constrained for ordinary nodes. Several edges
// arriving is how a loop comes back and how alternatives rejoin, and demanding
// an explicit merge for those would be ceremony. Several edges leaving is a
// different matter: it is a fan out that nothing named, and a reader cannot
// tell whether all of them run or one of them does.
func checkDegrees(p *ir.Process, nodes map[string]ir.ProcessNode, out *diag.Set) {
	for _, n := range sortedNodes(p) {
		in, outgoing := len(p.In(n.ID)), len(p.Out(n.ID))

		var what, why, how string
		switch {
		case n.Kind == ir.NodeStart && in > 0:
			what, why, how = "the start "+quote(n.ID)+" has "+count(in, "incoming edge")+".",
				"A start is where the process begins. Something leading into it means it is also a step, and then it is not the beginning.",
				"Remove the edge, or make this an ordinary node and add a start before it."
		case n.Kind == ir.NodeStart && outgoing != 1:
			what, why, how = "the start "+quote(n.ID)+" has "+count(outgoing, "outgoing edge")+".",
				"A start begins one course. Two would be a fan out that nothing named.",
				"Give it exactly one outgoing edge, into a gateway if the process really does split here."
		case n.Kind == ir.NodeEnd && outgoing > 0:
			what, why, how = "the end "+quote(n.ID)+" has "+count(outgoing, "outgoing edge")+".",
				"An end is where the process finishes. Something leading out of it means it does not.",
				"Remove the edge, or make this an ordinary node."
		// An end with nothing arriving is not reported here. It is
		// unreachable by definition, and the tracer says so in the words that
		// name the actual problem. Two findings for one fact teach a reader to
		// skim.
		case n.Kind.Splits() && outgoing < 2:
			what, why, how = quote(n.ID)+" is "+article(n.Kind)+" with "+count(outgoing, "outgoing edge")+".",
				"A gateway exists to divide the course. One branch divides nothing and only adds a box to the picture.",
				"Give it at least two branches, or remove it and join the neighbours directly."
		case n.Kind.Merges() && in < 2:
			what, why, how = quote(n.ID)+" is "+article(n.Kind)+" with "+count(in, "incoming edge")+".",
				"A gateway exists to bring branches back together. One branch joins nothing.",
				"Give it at least two incoming edges, or remove it and join the neighbours directly."
		case n.Kind.Merges() && outgoing != 1:
			what, why, how = quote(n.ID)+" is "+article(n.Kind)+" with "+count(outgoing, "outgoing edge")+".",
				"Bringing branches together yields one course. Several would split again without saying so.",
				"Give it exactly one outgoing edge."
		case !n.Kind.Gateway() && n.Kind != ir.NodeEnd && outgoing != 1:
			what, why, how = quote(n.ID)+" has "+count(outgoing, "outgoing edge")+".",
				"A step leads on to one thing. Where several follow, a reader cannot tell whether all of them run or one of them does, and that is precisely what a gateway is for.",
				"Give it exactly one outgoing edge, into a fork if all follow or a choice if one does."
		default:
			continue
		}

		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 75),
			Pos:  n.Pos,
			Rule: RuleNodeDegree,
			What: what,
			Why:  why,
			How:  how,
		})
	}
}

// checkRefs holds the graph against the code it claims to describe.
//
// A reference into a package the scope left out is skipped rather than
// reported. The recognisers never looked there, so "not a recognised construct"
// would be a statement about the setting rather than about the process — the
// mistake of reading silence as an answer, one step removed.
func checkRefs(p *ir.Process, known map[string]ir.Construct, scope map[string]bool, out *diag.Set) {
	for _, n := range sortedNodes(p) {
		if !n.Kind.References() || n.Ref == "" {
			continue
		}
		if scope != nil && n.RefPackage != "" && !scope[n.RefPackage] {
			continue
		}
		c, found := known[n.Ref]
		switch {
		case !found:
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 79),
				Pos:  n.Pos,
				Rule: RuleNodeRefUnknown,
				What: shortName(n.Ref) + " is named as " + article(n.Kind) + " but is not a recognised construct.",
				Why:  "A step that names something the recognisers never found describes work nothing performs.",
				How:  "Name the use case or event this step is, or remove the step.",
			})
		case n.Kind == ir.NodeActivity && c.Kind.MovesLifecycle():
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 79),
				Pos:  n.Pos,
				Rule: RuleNodeRefUnknown,
				What: shortName(n.Ref) + " is " + c.Kind.WithArticle() + " and cannot be an activity.",
				Why:  "An event is a fact that was recorded, not work that is done. Reading one as a step confuses what happened with what happens.",
				How:  "Use a node that raises or awaits the event instead.",
			})
		case n.Kind == ir.NodeActivity && !c.Kind.NeedsRequirement():
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 79),
				Pos:  n.Pos,
				Rule: RuleNodeRefUnknown,
				What: shortName(n.Ref) + " is " + c.Kind.WithArticle() + " and performs nothing.",
				Why:  "An activity is a step somebody takes. A structural role carries no action of its own and is reached through whatever uses it.",
				How:  "Name the use case that performs this step.",
			})
		case (n.Kind == ir.NodeEmit || n.Kind == ir.NodeCatch) && !c.Kind.MovesLifecycle():
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 79),
				Pos:  n.Pos,
				Rule: RuleNodeRefUnknown,
				What: shortName(n.Ref) + " is " + c.Kind.WithArticle() + " and is not an event.",
				Why:  "Raising and awaiting are about facts that outlive the code that wrote them. Anything else has no moment at which it occurs.",
				How:  "Name an event, or make this an activity.",
			})
		}
	}
}

func checkEndpoints(p *ir.Process, out *diag.Set) {
	var starts, ends, activities int
	for _, n := range p.Nodes {
		switch {
		case n.Kind == ir.NodeStart:
			starts++
		case n.Kind == ir.NodeEnd:
			ends++
		case n.Kind == ir.NodeActivity:
			activities++
		}
	}

	if starts == 0 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 70),
			Pos:  p.Pos,
			Rule: RuleProcessNoStart,
			What: "process " + p.ID + " has no start.",
			Why:  "Without one there is no point from which the course can be read, and nothing can be said about which steps are reachable.",
			How:  "Add a start and join it to the first step.",
		})
	}
	if ends == 0 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 71),
			Pos:  p.Pos,
			Rule: RuleProcessNoEnd,
			What: "process " + p.ID + " has no end.",
			Why:  "A course of business that cannot finish is either incompletely written down or a defect worth knowing about. Both are worth stopping for.",
			How:  "Add an end for each outcome, naming what each one means.",
		})
	}
	if activities == 0 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 72),
			Pos:  p.Pos,
			Rule: RuleProcessNoActivity,
			What: "process " + p.ID + " performs nothing.",
			Why:  "Gateways and endpoints route and finish; they do no work. A process built only from them describes a shape rather than a course of business.",
			How:  "Add an activity naming the use case that performs the work.",
		})
	}
}

// trace walks the graph from both ends.
func trace(p *ir.Process, nodes map[string]ir.ProcessNode, out *diag.Set) {
	forward := map[string][]string{}
	backward := map[string][]string{}
	for _, e := range p.Edges {
		if _, ok := nodes[e.From]; !ok {
			continue
		}
		if _, ok := nodes[e.To]; !ok {
			continue
		}
		forward[e.From] = append(forward[e.From], e.To)
		backward[e.To] = append(backward[e.To], e.From)
	}

	reached := walk(nodes, forward, ir.NodeStart)
	finishes := walk(nodes, backward, ir.NodeEnd)

	// A process that cannot finish at all is one mistake, not one per node.
	// Listing every step of a graph that loops forever buries the fact in its
	// own consequences.
	if trapped := countTrapped(nodes, reached, finishes); trapped > 0 && trapped == len(reached) {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 78),
			Pos:  p.Pos,
			Rule: RuleNodeTrapped,
			What: "process " + p.ID + " can never finish: no end is reachable from anywhere in it.",
			Why:  "Every point must have a way out. This is what makes a loop legitimate and a loop without an exit a defect — and the difference is invisible in a drawing.",
			How:  "Add the edge that leaves the loop towards an end.",
		})
		return
	}

	for _, n := range sortedNodes(p) {
		if _, ok := nodes[n.ID]; !ok {
			continue
		}
		if !reached[n.ID] {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 77),
				Pos:  n.Pos,
				Rule: RuleNodeUnreachable,
				What: quote(n.ID) + " cannot be reached from any start.",
				Why:  "A step nothing leads to never runs. It reads as part of the process and is not, which is worse than its absence.",
				How:  "Join it to the step it follows, or remove it.",
			})
			continue
		}
		if !finishes[n.ID] {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 78),
				Pos:  n.Pos,
				Rule: RuleNodeTrapped,
				What: "from " + quote(n.ID) + " the process can never finish.",
				Why:  "Every point must have a way out. This is what makes a loop legitimate and a loop without an exit a defect — and the difference is invisible in a drawing.",
				How:  "Add the edge that leaves this part of the graph towards an end.",
			})
		}
	}
}

// countTrapped counts the reachable nodes from which no end can be reached.
func countTrapped(nodes map[string]ir.ProcessNode, reached, finishes map[string]bool) int {
	n := 0
	for id := range nodes {
		if reached[id] && !finishes[id] {
			n++
		}
	}
	return n
}

// walk collects everything reachable from the nodes of one kind.
func walk(nodes map[string]ir.ProcessNode, adj map[string][]string, from ir.NodeKind) map[string]bool {
	seen := map[string]bool{}

	var visit func(string)
	visit = func(id string) {
		if seen[id] {
			return
		}
		seen[id] = true
		for _, next := range adj[id] {
			visit(next)
		}
	}

	ids := make([]string, 0, len(nodes))
	for id, n := range nodes {
		if n.Kind == from {
			ids = append(ids, id)
		}
	}
	sort.Strings(ids)
	for _, id := range ids {
		visit(id)
	}
	return seen
}

// sortedNodes returns the nodes in source order, so findings come out in the
// order somebody reads the file.
func sortedNodes(p *ir.Process) []ir.ProcessNode {
	out := append([]ir.ProcessNode(nil), p.Nodes...)
	sort.Slice(out, func(i, j int) bool { return out[i].Pos.Less(out[j].Pos) })
	return out
}

func article(k ir.NodeKind) string {
	switch k {
	case ir.NodeActivity, ir.NodeEmit, ir.NodeCatch:
		return "an " + k.String()
	}
	return "a " + k.String()
}

func count(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return strconv.Itoa(n) + " " + noun + "s"
}
