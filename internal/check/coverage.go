// Package check holds the specification rules K1 to K4.
//
// Every rule carries a stable ID so it can be waived individually with a
// mandatory justification. Rules are Go code rather than data: a rule language
// would be a second compiler inside the compiler, and keeping the tool small
// matters more than making rules configurable.
//
// There are no severities. A finding is an error and the run fails.
package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
)

// Rule IDs. They appear in diagnostics and in spec.Waive calls, so they are
// part of the public surface and must stay stable.
const (
	// RuleRequirementUncovered fires when a normative requirement is
	// referenced by no construct at all.
	RuleRequirementUncovered = "K3-REQ-UNCOVERED"
	// RuleAbstractCovered fires when a pure derivation node is covered
	// directly.
	RuleAbstractCovered = "K3-ABSTRACT-COVERED"
	// RuleSupersededCovered fires when a construct still satisfies a
	// requirement that has been replaced.
	RuleSupersededCovered = "K3-SUPERSEDED-COVERED"
)

// Coverage is the result of the two directional coverage analysis.
type Coverage struct {
	// BySatisfier maps a requirement ID to the constructs referencing it.
	BySatisfier map[string][]ir.Target
	// Uncovered lists normative requirements no construct references.
	Uncovered []string
	// Normative is the number of requirements that had to be covered.
	Normative int
}

// Ratio returns the covered share of normative requirements in [0,1].
func (c Coverage) Ratio() float64 {
	if c.Normative == 0 {
		return 1
	}
	return float64(c.Normative-len(c.Uncovered)) / float64(c.Normative)
}

// CoverRequirements performs the coverage analysis in both directions.
//
// The forward direction — is every annotated construct tied to a requirement —
// is the usual one. The backward direction is the actual gain: it turns "I hope
// everything is implemented" into a number and makes hand maintained status
// tables unnecessary. A forgotten requirement is invisible to any one
// directional measurement, because it appears nowhere and therefore fails no
// test.
func CoverRequirements(tree *reqtree.Tree, bindings []ir.Binding, out *diag.Set) Coverage {
	cov := Coverage{BySatisfier: map[string][]ir.Target{}}
	waived := ir.CollectWaivers(bindings)

	// Forward: collect what each construct claims to satisfy.
	for _, b := range bindings {
		for _, a := range b.Assertions {
			if a.Kind != ir.AssertSatisfies {
				continue
			}
			for _, id := range a.Requirements {
				r := requirementOf(tree, id)
				if r == nil {
					continue // unresolvable references are a Go compile error
				}
				cov.BySatisfier[r.ID] = append(cov.BySatisfier[r.ID], b.Target)
				checkSatisfiable(r, b, a, waived, out)
			}
		}
	}

	// Backward: every normative requirement needs at least one construct.
	for _, id := range sortedIDs(tree) {
		r := tree.ByID[id]
		if !r.Status.MustBeCovered() {
			continue
		}
		cov.Normative++
		if len(cov.BySatisfier[id]) > 0 {
			continue
		}
		cov.Uncovered = append(cov.Uncovered, id)

		if waived.Has("", RuleRequirementUncovered) {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 1),
			Pos:  r.Pos,
			Rule: RuleRequirementUncovered,
			What: "normative requirement " + r.ID + " is satisfied by no construct.",
			Why:  "Backward coverage is what makes a forgotten requirement visible at all. A requirement nobody references appears nowhere and therefore breaks no test.",
			How:  "Bind an implementing construct with spec.Satisfies(" + shortIdent(r) + "), or mark the requirement as spec.Planned, spec.OutOfScope or spec.Informative if it is deliberately not implemented.",
		})
	}
	return cov
}

// checkSatisfiable rejects references to requirements that must not be covered.
func checkSatisfiable(r *ir.Requirement, b ir.Binding, a ir.Assertion, waived ir.Waivers, out *diag.Set) {
	switch r.Status {
	case ir.Abstract:
		if waived.Has(b.Target.String(), RuleAbstractCovered) {
			return
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 2),
			Pos:  a.Pos,
			Rule: RuleAbstractCovered,
			What: r.ID + " is abstract and cannot be satisfied directly.",
			Why:  "An abstract requirement is a pure derivation node. Satisfying it directly hides which concrete requirement the construct actually implements, and inflates the coverage figure.",
			How:  "Reference the concrete requirement derived from " + r.ID + " instead.",
		})
	case ir.Superseded:
		if waived.Has(b.Target.String(), RuleSupersededCovered) {
			return
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 3),
			Pos:  a.Pos,
			Rule: RuleSupersededCovered,
			What: r.ID + " has been superseded and should no longer be satisfied.",
			Why:  "A superseded requirement was replaced deliberately; code still pointing at it drifts away from what is actually required.",
			How:  "Point at the successor that lists " + r.ID + " in its Supersedes field.",
		})
	}
}

func requirementOf(tree *reqtree.Tree, goIdentOrID string) *ir.Requirement {
	if r, ok := tree.ByID[goIdentOrID]; ok {
		return r
	}
	return tree.ByGoIdent(goIdentOrID)
}

func sortedIDs(tree *reqtree.Tree) []string {
	ids := make([]string, 0, len(tree.ByID))
	for id := range tree.ByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// shortIdent renders the Go identifier a caller would write, e.g.
// "quote.RQuoteSubmit" rather than the fully qualified path.
func shortIdent(r *ir.Requirement) string {
	ident := r.GoIdent
	slash := -1
	dot := -1
	for i := len(ident) - 1; i >= 0; i-- {
		if ident[i] == '.' && dot < 0 {
			dot = i
		}
		if ident[i] == '/' {
			slash = i
			break
		}
	}
	if dot < 0 {
		return ident
	}
	return ident[slash+1:]
}
