package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// Context renders the system boundary: who is outside, and every way across it.
//
// The whole module is one box. That is the point of the view — a reader asking
// where the system ends does not want to know that it has a sales context, and
// a diagram that shows everything shows nothing. Channels between two packages
// are left out here for the same reason: they do not cross the boundary this
// picture is about.
func Context(t ir.Topology, system string) string {
	var b strings.Builder

	b.WriteString("@startuml context\n")
	b.WriteString("hide empty description\n")
	b.WriteString("left to right direction\n\n")

	fmt.Fprintf(&b, "rectangle %s as system\n", quoted(system))

	touched := map[string]bool{}
	for _, c := range t.Channels {
		for _, end := range []string{c.From, c.To} {
			touched[end] = true
		}
	}

	for _, p := range sortedParticipants(t.Participants) {
		if !touched[p.ID] {
			continue
		}
		switch p.Kind {
		case ir.ParticipantActor:
			fmt.Fprintf(&b, "actor %s as %s\n", quoted(p.Name), ident(p.ID))
		default:
			fmt.Fprintf(&b, "rectangle %s as %s <<extern>>\n", quoted(p.Name), ident(p.ID))
		}
	}
	b.WriteString("\n")

	// Two packages either side of a channel means it does not cross the
	// boundary, so it belongs in the building block view and not here.
	seen := map[string]bool{}
	for _, c := range sortedChannels(t.Channels) {
		from, fromOut := endpoint(t, c.From)
		to, toOut := endpoint(t, c.To)
		if !fromOut && !toOut {
			continue
		}
		line := fmt.Sprintf("%s --> %s : %s", from, to, escape(c.Label+labelDetail(c)))
		if seen[line] {
			continue
		}
		seen[line] = true
		b.WriteString(line + "\n")
	}

	b.WriteString("@enduml\n")
	return b.String()
}

// Blocks renders what is inside the boundary and how the parts reach out.
//
// Only the packages some channel names appear. A view containing every package
// of a module is a dependency graph, which answers a different question and
// answers it worse than the compiler already does.
func Blocks(t ir.Topology, system string) string {
	var b strings.Builder

	b.WriteString("@startuml blocks\n")
	b.WriteString("hide empty description\n\n")

	inside := map[string]bool{}
	for _, c := range t.Channels {
		for _, end := range []string{c.From, c.To} {
			if t.Packages[end] {
				inside[end] = true
			}
		}
	}

	fmt.Fprintf(&b, "rectangle %s {\n", quoted(system))
	for _, dir := range sortedKeys(inside) {
		fmt.Fprintf(&b, "  rectangle %s as %s\n", quoted(dir), ident(dir))
	}
	b.WriteString("}\n\n")

	for _, p := range sortedParticipants(t.Participants) {
		switch p.Kind {
		case ir.ParticipantActor:
			fmt.Fprintf(&b, "actor %s as %s\n", quoted(p.Name), ident(p.ID))
		default:
			fmt.Fprintf(&b, "rectangle %s as %s <<extern>>\n", quoted(p.Name), ident(p.ID))
		}
	}
	b.WriteString("\n")

	for _, c := range sortedChannels(t.Channels) {
		fmt.Fprintf(&b, "%s --> %s : %s\n", ident(c.From), ident(c.To), escape(c.Label+labelDetail(c)))
	}

	b.WriteString("@enduml\n")
	return b.String()
}

// endpoint resolves one end for the context view, collapsing anything inside
// the module into the single system box.
func endpoint(t ir.Topology, end string) (name string, outside bool) {
	for _, p := range t.Participants {
		if p.ID == end {
			return ident(end), true
		}
	}
	return "system", false
}

// labelDetail puts the protocol under the name, which is where a reader looks
// for it and where it does not lengthen the arrow.
func labelDetail(c ir.Channel) string {
	if c.Protocol == "" {
		return ""
	}
	return "\\n<size:10>" + c.Protocol + "</size>"
}

func sortedParticipants(ps []ir.Participant) []ir.Participant {
	out := append([]ir.Participant(nil), ps...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func sortedChannels(cs []ir.Channel) []ir.Channel {
	out := append([]ir.Channel(nil), cs...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		return out[i].To < out[j].To
	})
	return out
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
