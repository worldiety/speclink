package check

import (
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
)

// Rule IDs of the topology checks.
const (
	// RuleChannelEndpointUnknown fires when an end of a channel is neither a
	// declared participant nor a package of the module.
	RuleChannelEndpointUnknown = "K17-CHANNEL-ENDPOINT-UNKNOWN"
	// RuleChannelIncomplete fires when a channel leaves out what crosses it,
	// who may, or what protects it.
	RuleChannelIncomplete = "K17-CHANNEL-INCOMPLETE"
	// RuleChannelUnbound fires when a channel answers to no requirement.
	RuleChannelUnbound = "K17-CHANNEL-UNBOUND"
	// RuleParticipantUnused fires when nothing reaches a declared participant.
	RuleParticipantUnused = "K17-PARTICIPANT-UNUSED"
	// RuleParticipantDuplicate fires when two participants share an ID.
	RuleParticipantDuplicate = "K17-PARTICIPANT-DUPLICATE"
	// RuleAdapterNoChannel fires when the system reaches outside at a place no
	// channel describes.
	RuleAdapterNoChannel = "K17-ADAPTER-NO-CHANNEL"
)

// TopologyReport is what the run can say about the world around the code.
type TopologyReport struct {
	// Declared is whether the project has a topology at all.
	Declared bool
	// Channels is how many ways across were declared.
	Channels int
	// Adapters is how many places the system touches something outside, and
	// Described how many of them a channel accounts for.
	Adapters, Described int
}

// Topology checks the declared world against the code that lives in it.
//
// # Why this is declared rather than read
//
// Everywhere else a fact that is inferable must not be annotated. Here nothing
// is inferable. No Go module states that an end user exists, that the object
// store is somebody else's responsibility, or that the channel to it carries
// customer data under a short lived key. That is not missing from the code; it
// is knowledge the code cannot hold.
//
// # What keeps it from being a picture
//
// An adapter is where a system touches something outside, and the architecture
// says which packages those are — read from the layout, not guessed from
// imports. A framework import crosses no boundary, and a heuristic over import
// paths would call every one of them a channel.
//
// So both ends are enumerated. A channel naming a package that is not there is
// a mistake; an adapter no channel names is a way out of the system that
// nobody wrote down. Neither is findable from one side alone, and a model that
// only checked the first would be a drawing with a lint pass.
func Topology(tree *reqtree.Tree, t ir.Topology, bindings []ir.Binding, d ir.Dialect, out *diag.Set) TopologyReport {
	rep := TopologyReport{Declared: t.Declared(), Adapters: len(t.Adapters), Channels: len(t.Channels)}
	if !rep.Declared {
		// Nothing to be incomplete against. Reporting every adapter here would
		// demand adoption rather than report a gap, and no figure claims a
		// share of a boundary nobody described.
		return rep
	}

	parts := indexParticipants(t, out)
	touched := map[string]bool{}
	described := map[string]bool{}

	channels := append([]ir.Channel(nil), t.Channels...)
	sort.Slice(channels, func(i, j int) bool { return channels[i].Pos.Less(channels[j].Pos) })

	for _, c := range channels {
		checkChannelFields(c, d, out)
		checkChannelRequirements(tree, c, out)
		for _, end := range []string{c.From, c.To} {
			switch {
			case end == "":
				// Already reported as a missing field.
			case parts[end]:
				touched[end] = true
			case t.Packages[end]:
				described[end] = true
			case looksLikePackage(end):
				// A directory that the run did not measure. The scope decides
				// what is looked at, so this is unmeasured rather than wrong,
				// and saying otherwise would fail a project for its own
				// setting.
			default:
				out.Add(diag.Finding{
					Code: diag.Code(diag.PhaseSemantic, 90),
					Pos:  c.Pos,
					Rule: RuleChannelEndpointUnknown,
					What: quote(end) + " is neither a declared participant nor a package of this module.",
					Why:  "An end of a channel that resolves to nothing describes a boundary with only one side, and the far side is the half that decides who is responsible.",
					How:  "Name a declared actor or foreign system, or a package as its repository relative directory such as app/sales/adapter/fs.",
				})
			}
		}
	}

	checkUnusedParticipants(t, touched, out)
	rep.Described = describeAdapters(t, described, ir.CollectWaivers(bindings), out)
	return rep
}

func indexParticipants(t ir.Topology, out *diag.Set) map[string]bool {
	parts := map[string]bool{}
	seen := map[string]ir.Participant{}

	sorted := append([]ir.Participant(nil), t.Participants...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Pos.Less(sorted[j].Pos) })

	for _, p := range sorted {
		if first, dup := seen[p.ID]; dup {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 94),
				Pos:  p.Pos,
				Rule: RuleParticipantDuplicate,
				What: quote(p.ID) + " is declared twice.",
				Why:  "Channels name participants by ID. Two under one name make every channel touching it ambiguous about who is on the far end.",
				How:  "Rename one of them; the other is at " + first.Pos.String() + ".",
			})
			continue
		}
		seen[p.ID] = p
		parts[p.ID] = true
	}
	return parts
}

// checkChannelFields refuses a channel that leaves out what crosses it, who may
// and what protects it.
//
// These are the answer to the question every review of an interface begins
// with, and the one normally answered by reading the code of both ends and
// hoping. The empty field is always the interesting one.
func checkChannelFields(c ir.Channel, _ ir.Dialect, out *diag.Set) {
	var missing []string
	for _, f := range []struct{ name, value string }{
		{"From", c.From}, {"To", c.To}, {"Label", c.Label},
		{"Protocol", c.Protocol}, {"Data", c.Data},
		{"Auth", c.Auth}, {"Crypto", c.Crypto},
	} {
		if strings.TrimSpace(f.value) == "" {
			missing = append(missing, f.name)
		}
	}
	if len(missing) == 0 {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 91),
		Pos:  c.Pos,
		Rule: RuleChannelIncomplete,
		What: "channel " + c.Name() + " leaves out " + strings.Join(missing, ", ") + ".",
		Why:  "What crosses a boundary, who is allowed across and what protects it in transit are the three things any review of an interface asks first. A blank one reads as nothing to say, and it never is.",
		How:  "State each of them, plainly where the answer is that there is none — \"entfällt, lokal\" is an answer and an empty field is not.",
	})
}

func checkChannelRequirements(tree *reqtree.Tree, c ir.Channel, out *diag.Set) {
	if len(c.Satisfies) == 0 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 92),
			Pos:  c.Pos,
			Rule: RuleChannelUnbound,
			What: "channel " + c.Name() + " answers to no requirement.",
			Why:  "A way across the boundary that answers to nothing is a way somebody opened without being asked. Every one of them widens what has to be defended.",
			How:  "Add `Satisfies` naming the requirement this channel exists for.",
		})
		return
	}
	for _, ref := range c.Satisfies {
		if tree.ByGoIdent(ref) == nil {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 92),
				Pos:  c.Pos,
				Rule: RuleChannelUnbound,
				What: "channel " + c.Name() + " names " + shortName(ref) + ", which is not a requirement.",
				Why:  "A reference that resolves to nothing looks like traceability and provides none.",
				How:  "Name a declared requirement.",
			})
		}
	}
}

func checkUnusedParticipants(t ir.Topology, touched map[string]bool, out *diag.Set) {
	sorted := append([]ir.Participant(nil), t.Participants...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Pos.Less(sorted[j].Pos) })

	for _, p := range sorted {
		if touched[p.ID] {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 93),
			Pos:  p.Pos,
			Rule: RuleParticipantUnused,
			What: p.Name + " is declared as " + participantArticle(p.Kind) + " that nothing reaches.",
			Why:  "A participant exists in this model to be at the end of a channel. One that is at the end of none is either a channel somebody forgot or a leftover, and both read as part of the picture.",
			How:  "Declare the channel that reaches it, or remove it.",
		})
	}
}

// describeAdapters is the backward direction: every place the system touches
// something outside must be a place some channel describes.
func describeAdapters(t ir.Topology, described map[string]bool, waived ir.Waivers, out *diag.Set) int {
	n := 0
	for _, a := range t.Adapters {
		if described[a.Dir] {
			n++
			continue
		}
		// An adapter that crosses no boundary is the one case this rule cannot
		// see: an in memory implementation of a port is structurally an
		// adapter and reaches nothing. The waiver is the right way out, and it
		// carries the reason into the generated document.
		// Counted as described, because that is what a waiver does everywhere
		// else here: it satisfies the figure and reappears in the generated
		// document as an accepted gap with its reason beside it. A waiver that
		// silently held the number down would be the one kind nobody reviews.
		if waived.Has(a.Pkg, RuleAdapterNoChannel) {
			n++
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 95),
			Pos:  a.Pos,
			Rule: RuleAdapterNoChannel,
			What: a.Dir + " reaches outside the system and no channel describes it.",
			Why:  "An adapter is where this system touches something it does not control. One that no channel names is a way out that never appeared in any interface list — which is exactly the kind that is missed when somebody asks what leaves the building.",
			How:  "Declare a channel from " + a.Dir + " to the actor or foreign system behind it, or waive this with the reason it crosses no boundary.",
		})
	}
	return n
}

// looksLikePackage distinguishes a directory from an identifier, so that an
// endpoint into an unmeasured package is not reported as a misspelling.
func looksLikePackage(end string) bool { return strings.Contains(end, "/") }

func participantArticle(k ir.ParticipantKind) string {
	if k == ir.ParticipantActor {
		return "an " + k.String()
	}
	return "a " + k.String()
}
