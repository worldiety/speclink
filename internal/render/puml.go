// Package render turns the model into the inputs of a document.
//
// It writes and never runs. PlantUML and Typst are prerequisites of the
// environment, not dependencies of this program: nothing here imports them,
// shells out to them or pins a version of them, and every function below
// returns text. A project that has neither installed can still run every rule
// and every test; only the document is out of reach.
//
// That boundary is deliberate. A tool that renders is a tool that has to be
// installed with a Java runtime beside it, and the one thing this binary has
// promised throughout is that it is one binary.
package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// Process renders one course of business as a PlantUML state diagram.
//
// # Why a state diagram and not an activity diagram
//
// PlantUML's activity syntax is block structured: start, then a sequence, with
// fork and if nested inside one another. A process here is a graph, on purpose,
// because real courses come back and a nested form cannot express the jump
// backwards. Emitting activity syntax would mean recovering block structure
// from an arbitrary digraph — which is the problem the model was shaped to
// avoid, solved again in the renderer and this time without any rule to catch
// it going wrong.
//
// A state diagram is a graph. Nodes and edges map one to one, cycles come out
// as cycles, and the stereotypes carry the gateways: fork, join and choice are
// drawn as the bars and diamonds a reader expects.
func Process(p *ir.Process) string {
	var b strings.Builder

	fmt.Fprintf(&b, "@startuml %s\n", ident(p.ID))
	b.WriteString("hide empty description\n")
	if p.Title != "" {
		fmt.Fprintf(&b, "title %s\n", escape(p.Title))
	}
	b.WriteString("\n")

	byID := map[string]ir.ProcessNode{}
	for _, n := range p.Nodes {
		byID[n.ID] = n
	}

	// Nodes first, so that a reader of the source sees the vocabulary before
	// the wiring, and so PlantUML has the stereotype before the first edge.
	for _, n := range sortedNodes(p.Nodes) {
		if decl := declare(n); decl != "" {
			b.WriteString(decl + "\n")
		}
	}
	b.WriteString("\n")

	for _, e := range p.Edges {
		from, to := byID[e.From], byID[e.To]
		// A start and an end are not states. PlantUML spells both as [*], and
		// which one it means follows from the side of the arrow it is on.
		fromName, toName := ident(e.From), ident(e.To)
		if from.Kind == ir.NodeStart {
			fromName = "[*]"
		}

		label := e.When
		// The end's own label rides on the arrow. PlantUML draws a terminal
		// circle without a name, so the outcome has to be written where a
		// reader will see it.
		if to.Kind == ir.NodeEnd && to.Label != "" {
			if label != "" {
				label += " — "
			}
			label += to.Label
		}
		if from.Kind == ir.NodeStart && from.Label != "" && label == "" {
			label = from.Label
		}

		fmt.Fprintf(&b, "%s --> %s", fromName, toName)
		if label != "" {
			fmt.Fprintf(&b, " : %s", escape(label))
		}
		b.WriteString("\n")
	}

	b.WriteString("@enduml\n")
	return b.String()
}

// declare renders one node, or nothing for the endpoints PlantUML spells [*].
func declare(n ir.ProcessNode) string {
	switch n.Kind {
	case ir.NodeStart:
		// A start is spelled [*] on the left of an arrow and needs no
		// declaration of its own.
		return ""
	case ir.NodeEnd:
		// An end does need one. Spelling every outcome [*] would collapse them
		// into a single circle, and separate outcomes are the reason a process
		// records more than one — approved and withdrawn are not one ending
		// with two labels.
		return "state " + ident(n.ID) + " <<end>>"
	case ir.NodeFork:
		return "state " + ident(n.ID) + " <<fork>>"
	case ir.NodeJoin:
		return "state " + ident(n.ID) + " <<join>>"
	case ir.NodeChoice, ir.NodeMerge:
		// PlantUML has one diamond. A merge is a choice read backwards, and
		// inventing a second glyph would tell a reader there is a distinction
		// the drawing cannot show.
		return "state " + ident(n.ID) + " <<choice>>"
	case ir.NodeEmit:
		return "state " + quoted(short(n.Ref)) + " as " + ident(n.ID) + " <<event>>"
	case ir.NodeCatch:
		return "state " + quoted(short(n.Ref)) + " as " + ident(n.ID) + " <<awaits>>"
	}
	return "state " + quoted(short(n.Ref)) + " as " + ident(n.ID)
}

func sortedNodes(nodes []ir.ProcessNode) []ir.ProcessNode {
	out := append([]ir.ProcessNode(nil), nodes...)
	sort.Slice(out, func(i, j int) bool { return out[i].Pos.Less(out[j].Pos) })
	return out
}

// ident makes a name PlantUML accepts as a state identifier.
func ident(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "n"
	}
	// An identifier starting with a digit is read as a number.
	if s := b.String(); s[0] >= '0' && s[0] <= '9' {
		return "n" + s
	}
	return b.String()
}

// short is the last segment of a qualified name.
func short(name string) string {
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		return name[i+1:]
	}
	if name == "" {
		return "?"
	}
	return name
}

func quoted(s string) string { return `"` + escape(s) + `"` }

// escape keeps a label from ending the line or the quoted name it sits in.
func escape(s string) string {
	s = strings.ReplaceAll(s, `"`, `'`)
	s = strings.ReplaceAll(s, "\n", " ")
	return strings.TrimSpace(s)
}
