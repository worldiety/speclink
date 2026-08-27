package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
)

// RulePersistenceUnjustified fires when an aggregate or a repository traces to
// no architectural decision.
const RulePersistenceUnjustified = "K1-PERSISTENCE-UNJUSTIFIED"

// JustifyPersistence asks the one question the coverage check leaves out: not
// which requirement this construct serves, but why it is shaped this way.
//
// Aggregates and repositories are exempt from needing a requirement of their
// own, and rightly so — they are reached through the use case that writes or
// holds them, and demanding a binding for each would only produce noise. But
// that exemption quietly skips something else. An aggregate and a repository
// are not two implementations of the same thing. One keeps every change as a
// fact and can answer what was true last March; the other overwrites and can
// only answer what is true now. Choosing between them decides which questions
// the system will be able to answer for the rest of its life, and it is close
// to irreversible once there is data.
//
// The choice is therefore required to trace to a Kind: Decision, whose
// Rationale is already mandatory. Nothing here judges the choice — only that
// somebody made it on purpose and wrote down what the alternative would have
// cost.
//
// It is asked per construct rather than per package because the two mix within
// one context as a matter of course: the write side of a domain can be event
// sourced while its reference data sits in a repository, and that is not a
// defect but the normal shape.
func JustifyPersistence(tree *reqtree.Tree, constructs []ir.Construct, bindings []ir.Binding, domain map[string]bool, d ir.Dialect, out *diag.Set) {
	decided := decisionTargets(tree, bindings)
	waived := ir.CollectWaivers(bindings)

	sorted := append([]ir.Construct(nil), constructs...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Pos.File != sorted[j].Pos.File {
			return sorted[i].Pos.File < sorted[j].Pos.File
		}
		return sorted[i].Pos.Line < sorted[j].Pos.Line
	})

	for _, c := range sorted {
		if !c.Kind.EmbodiesStorageDecision() {
			continue
		}
		// Infrastructure is out of scope. A helper type that carries an
		// identity is not a persistence choice anybody made about the domain.
		if !domain[c.Package] {
			continue
		}
		if justified(c, decided) || waived.Has(c.Name, RulePersistenceUnjustified) {
			continue
		}
		short := shortName(c.Name)
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 21),
			Pos:  c.Pos,
			Rule: RulePersistenceUnjustified,
			What: short + " is " + c.Kind.WithArticle() + " and rests on no recorded decision.",
			Why:  "Keeping every change as a fact and keeping only the current state answer different questions, and the choice is close to irreversible once there is data. A system that cannot say why it stores things the way it does cannot weigh the cost of that later.",
			How:  "Bind a requirement of kind decision, e.g. `" + d.BindDecision(c.Name) + "` — on the type, on its constructor, or on the package when the whole context made one choice.",
		})
	}
}

// decisionTargets collects every target that satisfies a decision.
func decisionTargets(tree *reqtree.Tree, bindings []ir.Binding) map[ir.Target]bool {
	out := map[ir.Target]bool{}
	for _, b := range bindings {
		for _, a := range b.Assertions {
			if a.Kind != ir.AssertSatisfies {
				continue
			}
			for _, id := range a.Requirements {
				if r := requirementOf(tree, id); r != nil && r.Kind == ir.Decision {
					out[b.Target] = true
				}
			}
		}
	}
	return out
}

// justified reports whether a decision reaches this construct.
//
// Three forms count. The type itself is the obvious one. The constructor is
// accepted because a repository is usually reached through New… and that is
// where a project naturally writes down why it chose one — a rule that only
// took the type would report the very projects that had thought about it. The
// package covers the ordinary case where a whole context made one choice.
func justified(c ir.Construct, decided map[ir.Target]bool) bool {
	if decided[ir.Target{Kind: ir.TargetType, Package: c.Package, Name: c.Name}] {
		return true
	}
	if decided[ir.Target{Kind: ir.TargetPackage, Package: c.Package}] {
		return true
	}
	ctor := c.Package + ".New" + shortName(c.Name)
	return decided[ir.Target{Kind: ir.TargetFunc, Package: c.Package, Name: ctor}]
}
