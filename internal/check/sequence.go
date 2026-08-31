package check

import (
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

const (
	// RuleActivityOutsideModule fires when a step of a process names work in a
	// package this module does not contain.
	RuleActivityOutsideModule = "K16-ACTIVITY-OUTSIDE-MODULE"
	// RuleSendUncarried fires when a message sent by a process crosses no
	// declared channel.
	RuleSendUncarried = "K16-SEND-UNCARRIED"
	// RuleSequenceOneLane fires when a course drawn as an exchange has nobody
	// to exchange with.
	RuleSequenceOneLane = "K16-SEQUENCE-ONE-LANE"
)

// Sequences holds a course of business to what a drawing of it can honestly
// say.
//
// # Why control flow may not leave the module
//
// An activity is work this module performs and answers for. What goes to
// another program is a message: it may be slow, absent, or a different version
// of itself, and none of that is true of a function this module calls. Drawing
// the two the same way is what makes a picture of a distributed system read
// like a call stack, and reading it that way is how the failure modes get
// forgotten. So a step naming work outside this module is refused, and the way
// to say it is a send.
func Sequences(t ir.Topology, processes []*ir.Process, out *diag.Set) {
	carried := map[string]bool{}
	for _, c := range t.Channels {
		for _, m := range c.Messages {
			if m.PayloadType != "" {
				carried[m.PayloadType] = true
			}
		}
		if c.Contract != nil && c.Contract.Type != "" {
			carried[c.Contract.Type] = true
		}
	}

	sorted := append([]*ir.Process(nil), processes...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Pos.Less(sorted[j].Pos) })

	for _, p := range sorted {
		for _, n := range p.Nodes {
			checkNodeReach(t, carried, p, n, out)
		}
		if p.Drawn == ir.AsSequence {
			checkLanes(p, out)
		}
	}
}

func checkNodeReach(t ir.Topology, carried map[string]bool, p *ir.Process, n ir.ProcessNode, out *diag.Set) {
	switch {
	case n.Kind.PerformedHere() && n.RefPackage != "" && !withinModule(t, n.RefPackage):
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 220),
			Pos:  n.Pos,
			Rule: RuleActivityOutsideModule,
			What: p.ID + " performs " + shortName(n.Ref) + ", which is not in this module.",
			Why:  "An activity is work this module performs and answers for. What runs in another program may be slow, absent or a different version of itself, and none of that is true of a call this module makes. Drawing the two the same way makes a picture of a distributed system read like a call stack.",
			How:  "Send a message across a declared channel instead, with spec.Send.",
		})

	case n.Kind == ir.NodeSend && n.Ref != "" && !carried[n.Ref]:
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 221),
			Pos:  n.Pos,
			Rule: RuleSendUncarried,
			What: p.ID + " sends " + shortName(n.Ref) + ", which crosses no declared channel.",
			Why:  "A message that travels on no channel is a step in a drawing and nothing in the model: there is no boundary it crosses, nobody at the far end, and no record of what it carries.",
			How:  "List the message on the channel it crosses, or name a payload that one already carries.",
		})
	}
}

// withinModule reports whether a package was measured by this run.
//
// The measured set and not a prefix test on the import path, because a package
// that was left out of the analysed patterns is a fact about the run rather
// than about the process, and reporting it as work in another program would be
// a false statement made confidently.
func withinModule(t ir.Topology, pkg string) bool {
	if len(t.Packages) == 0 {
		return true // nothing was measured, so nothing can be said
	}
	for dir := range t.Packages {
		if strings.HasSuffix(pkg, "/"+dir) || pkg == dir {
			return true
		}
	}
	return false
}

// checkLanes refuses a drawing of an exchange that has nobody to exchange with.
//
// A sequence is a picture of who says what to whom. One participant makes it a
// list of steps drawn down the side of a single line, which is the flow view
// with extra ceremony and less room.
func checkLanes(p *ir.Process, out *diag.Set) {
	lanes := map[string]bool{}
	for _, n := range p.Nodes {
		switch {
		case n.Kind == ir.NodeSend:
			lanes["\x00send"] = true
		case n.Actor != "":
			lanes["\x00actor"] = true
		case n.Kind.PerformedHere() && n.RefPackage != "":
			lanes[n.RefPackage] = true
		}
	}
	if len(lanes) >= 2 {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 222),
		Pos:  p.Pos,
		Rule: RuleSequenceOneLane,
		What: p.ID + " is drawn as an exchange and has only one participant.",
		Why:  "A sequence pictures who says what to whom. With one participant it is a list of steps down the side of a single line, which is the ordinary view with more ceremony and less room for the labels.",
		How:  "Draw it as a flow, or name the other side: work in another package, a message across a channel, or an actor who begins it.",
	})
}
