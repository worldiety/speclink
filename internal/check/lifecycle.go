package check

import (
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// Rule IDs of the lifecycle checks.
const (
	// RuleEventNoTransition fires when an event does not say which state it
	// leaves the aggregate in.
	RuleEventNoTransition = "K15-EVENT-NO-TRANSITION"
	// RuleTransitionUnknown fires when a transition names something that is not
	// a recognised event.
	RuleTransitionUnknown = "K15-TRANSITION-UNKNOWN"
)

// Lifecycle checks that the states an aggregate can be in are written down.
//
// An event is a fact that outlives the code that wrote it, and the set of
// states it can leave behind is the first question anybody asks about an
// aggregate — can this quote still be withdrawn, is that invoice already
// settled. Today the answer exists only in the folds: to learn the states you
// read every Evolve and reconstruct the machine in your head, and nothing tells
// you when somebody adds a sixth state to a thing that had five.
//
// The declaration is not inferable, which is why it is asked for rather than
// read. A fold assigns fields; which coarse state those fields amount to is a
// judgement about the business, and two events writing the same field may well
// mean different things. That is exactly the kind of fact this tool collects.
//
// # Both directions
//
// An event without a transition is one half. The other is a transition naming
// something that is not an event — a type that was renamed, an event that was
// deleted, a fold that moved. That reference compiles, because the assertion
// takes any type at all, and would otherwise rot silently.
func Lifecycle(constructs []ir.Construct, bindings []ir.Binding, d ir.Dialect, out *diag.Set) {
	events := map[string]ir.Construct{}
	for _, c := range constructs {
		if c.Kind.MovesLifecycle() {
			events[c.Name] = c
		}
	}
	// Nothing to say about an architecture that has no such role. This is not
	// the unmeasured case the capability lines exist for: go_bare has no events
	// because it is not event sourced, which its profile summary already says,
	// and no figure claims a share of lifecycles that were checked.
	if len(events) == 0 {
		return
	}

	waived := ir.CollectWaivers(bindings)
	declared := map[string]bool{}

	for _, b := range bindings {
		for _, a := range b.Assertions {
			if a.Kind != ir.AssertTransition || a.EventType == "" {
				continue
			}
			declared[a.EventType] = true
			if _, isEvent := events[a.EventType]; isEvent {
				continue
			}
			// The waiver sits on the binding that carries the transition,
			// because that is the line somebody has to look at.
			if waived.Has(b.Target.Name, RuleTransitionUnknown) {
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 61),
				Pos:  a.Pos,
				Rule: RuleTransitionUnknown,
				What: shortName(a.EventType) + " is named as a transition but is not an event.",
				Why:  "A transition says which state a fact leaves behind. Naming something that folds nothing describes a lifecycle nobody can reach, and the reference compiles because the assertion accepts any type at all.",
				How:  "Name the event this state follows from, or remove the transition if the event is gone.",
			})
		}
	}

	missing := make([]ir.Construct, 0, len(events))
	for name, c := range events {
		if declared[name] || waived.Has(name, RuleEventNoTransition) {
			continue
		}
		missing = append(missing, c)
	}
	sort.Slice(missing, func(i, j int) bool {
		if missing[i].Pos.File != missing[j].Pos.File {
			return missing[i].Pos.File < missing[j].Pos.File
		}
		return missing[i].Pos.Line < missing[j].Pos.Line
	})

	for _, c := range missing {
		short := shortName(c.Name)
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 60),
			Pos:  c.Pos,
			Rule: RuleEventNoTransition,
			What: short + " does not say which state it leaves the aggregate in.",
			Why:  "The states a thing can be in is the first question asked about it. Left undeclared, the answer exists only in the folds, and a sixth state added to something that had five passes without anybody deciding it.",
			How:  "Add `" + d.Transition(c.Name, suggestState(short)) + "` to the binding, naming the coarse state that follows.",
		})
	}
}

// suggestState turns an event name into a plausible state, so the How line
// carries a shape rather than an ellipsis.
//
// QuoteWithdrawn becomes "withdrawn": the last word of the event is the past
// participle, and the past participle is what a coarse state is named after.
// It is a suggestion in a diagnostic, so being wrong costs a word of editing —
// which is cheaper than a placeholder that has to be decoded first.
func suggestState(event string) string {
	words := splitWords(event)
	last := words[len(words)-1]
	if last == "" {
		return "…"
	}
	return strings.ToLower(last)
}

// splitWords breaks a CamelCase identifier into its words.
func splitWords(name string) []string {
	var (
		words []string
		start int
	)
	runes := []rune(name)
	for i, r := range runes {
		if i == 0 || !isUpperRune(r) {
			continue
		}
		// A capital continues a run of capitals rather than starting a word,
		// so QuoteIDAssigned splits into Quote, ID, Assigned.
		if isUpperRune(runes[i-1]) && (i+1 >= len(runes) || isUpperRune(runes[i+1])) {
			continue
		}
		words = append(words, string(runes[start:i]))
		start = i
	}
	return append(words, string(runes[start:]))
}

func isUpperRune(r rune) bool { return r >= 'A' && r <= 'Z' }
