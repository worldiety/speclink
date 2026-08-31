package render

import (
	"fmt"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// Sequence renders one course of business as a PlantUML sequence diagram.
//
// # Where the participants come from
//
// Nothing declares them. Each activity names a use case, the use case lives in
// a package, and the package is the part of the system performing it. Stating
// the lane as well would be that fact written twice, and the copy would be the
// one that goes stale when a use case moves.
//
// # What the drawing does not pretend
//
// Time runs downwards, so every arrow is ordered — and two branches of a fork
// have no order between them. They are drawn inside a par frame rather than
// stacked, because stacking them would invent a sequence the model does not
// have. A choice is drawn the same way, as alt.
func Sequence(p *ir.Process, participants map[string]ir.Participant) string {
	var b strings.Builder

	fmt.Fprintf(&b, "@startuml %s\n", ident(p.ID))
	if p.Title != "" {
		fmt.Fprintf(&b, "title %s\n", escape(p.Title))
	}
	b.WriteString("skinparam shadowing false\nskinparam sequenceMessageAlign left\n\n")

	lanes := lanesOf(p, participants)
	for _, l := range lanes.order {
		fmt.Fprintf(&b, "%s %s as %s\n", lanes.kind[l], quoted(lanes.label[l]), ident(l))
	}
	b.WriteString("\n")

	byID := map[string]ir.ProcessNode{}
	for _, n := range p.Nodes {
		byID[n.ID] = n
	}

	// The graph is walked from every start, so a course with two beginnings
	// draws both. Visited nodes are not drawn twice: a cycle in a sequence has
	// no rendering, and drawing it until the walk gives up would produce a
	// picture that is longer than the process.
	seen := map[string]bool{}
	for _, n := range sortedNodes(p.Nodes) {
		if n.Kind == ir.NodeStart {
			walk(&b, p, byID, lanes, n, lanes.of[n.ID], seen, 0)
		}
	}

	b.WriteString("@enduml\n")
	return b.String()
}

// laneSet is the participants of one drawing, in the order they first speak.
type laneSet struct {
	order []string
	label map[string]string
	kind  map[string]string
	// of maps a node ID onto the lane that performs it.
	of map[string]string
}

func lanesOf(p *ir.Process, participants map[string]ir.Participant) laneSet {
	l := laneSet{label: map[string]string{}, kind: map[string]string{}, of: map[string]string{}}
	add := func(id, label, kind string) {
		if _, ok := l.label[id]; !ok {
			l.order = append(l.order, id)
			l.label[id] = label
			l.kind[id] = kind
		}
	}

	for _, n := range sortedNodes(p.Nodes) {
		switch {
		case n.Actor != "":
			part, known := participants[n.Actor]
			label, id := shortIdent(n.Actor), shortIdent(n.Actor)
			if known {
				label, id = or(part.Name, part.ID), part.ID
			}
			// An actor is drawn as one, so a reader can tell at a glance which
			// side of the picture is a person and which is a program.
			add(id, label, "actor")
			l.of[n.ID] = id

		case n.Kind.PerformedHere() && n.RefPackage != "":
			id := laneID(n.RefPackage)
			add(id, id, "participant")
			l.of[n.ID] = id

		case n.Kind == ir.NodeSend:
			// The far end of a send is whoever the channel reaches. Without
			// the channel in hand the payload names it, which is still a
			// truthful label and never an invented participant.
			id := shortIdent(n.Ref)
			add(id, id, "participant")
			l.of[n.ID] = id
		}
	}
	return l
}

// laneID is the part of the system a package belongs to.
//
// The bounded context rather than the whole path: app/sales/adapter/fs and
// app/sales are the same participant in a picture of who talks to whom, and
// drawing them apart would put the internals of one context on the same
// footing as the boundary between two.
func laneID(pkg string) string {
	parts := strings.Split(pkg, "/")
	for i, part := range parts {
		if part == "app" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	if len(parts) == 0 {
		return pkg
	}
	return parts[len(parts)-1]
}

// walk draws the graph from one node onwards.
//
// caller is the participant control is currently with, because that is who an
// arrow comes from. A step is drawn as a message to whoever performs it: the
// performer is a property of the step, and the sender is wherever the previous
// step left off. Taking the performer as the sender draws every arrow one
// participant too early and labels it with the wrong step.
//
// depth guards against a cycle rather than a rule: the graph rules already
// report an unreachable or trapped node, and a renderer that hangs on one is
// worse than a picture that stops.
func walk(b *strings.Builder, p *ir.Process, byID map[string]ir.ProcessNode, lanes laneSet, n ir.ProcessNode, caller string, seen map[string]bool, depth int) {
	if seen[n.ID] || depth > 200 {
		return
	}
	seen[n.ID] = true

	out := p.Out(n.ID)

	switch n.Kind {
	case ir.NodeChoice, ir.NodeFork:
		frame := "alt"
		if n.Kind == ir.NodeFork {
			// Two branches of a fork have no order between them, which is
			// what par says and what stacking them would deny.
			frame = "par"
		}
		for i, e := range out {
			keyword := frame
			if i > 0 {
				keyword = "else"
			}
			fmt.Fprintf(b, "%s %s\n", keyword, escape(e.When))
			if next, ok := byID[e.To]; ok {
				walk(b, p, byID, lanes, next, caller, seen, depth+1)
			}
		}
		if len(out) > 0 {
			b.WriteString("end\n")
		}
		return

	case ir.NodeStart, ir.NodeEnd:
		if n.Label != "" {
			fmt.Fprintf(b, "note over %s: %s\n", ident(firstLane(lanes, caller)), escape(n.Label))
		}

	default:
		if to := lanes.of[n.ID]; to != "" {
			from := caller
			if from == "" {
				// Nothing has spoken yet, so the step is drawn as one the
				// performer takes on its own rather than invented a sender.
				from = to
			}
			fmt.Fprintf(b, "%s -> %s: %s\n", ident(from), ident(to), escape(labelOf(n)))
			caller = to
		}
		if n.Note != "" {
			fmt.Fprintf(b, "note right: %s\n", escape(n.Note))
		}
	}

	for _, e := range out {
		if next, ok := byID[e.To]; ok {
			walk(b, p, byID, lanes, next, caller, seen, depth+1)
		}
	}
}

// firstLane is somewhere to hang a note that belongs to no participant.
func firstLane(l laneSet, preferred string) string {
	if preferred != "" {
		return preferred
	}
	if len(l.order) > 0 {
		return l.order[0]
	}
	return "system"
}

// labelOf is what a step is called in a drawing.
func labelOf(n ir.ProcessNode) string {
	if n.Ref != "" {
		return shortIdent(n.Ref)
	}
	return or(n.Label, n.ID)
}

func shortIdent(qualified string) string {
	if i := strings.LastIndexByte(qualified, '.'); i >= 0 {
		return qualified[i+1:]
	}
	return qualified
}

// or is the first non empty of two.
func or(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
