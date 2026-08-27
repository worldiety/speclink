// Package check, this file: verification coverage.
package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
)

// RuleRequirementUnverified fires on a normative requirement that no test
// claims to demonstrate.
//
// It is the question coverage never asked. K3 says a requirement is satisfied
// by some construct, which means code was written for it; it says nothing about
// whether that code does what the requirement asks. In a loop where the same
// model writes the implementation and the tests, and people review by sampling,
// that gap is where the tool stops being worth anything: every figure reads a
// hundred percent and the only thing demonstrated is that somebody typed.
//
// The rule is waivable, and genuinely needs to be. Some requirements are not
// testable by a unit test at all — a structural decision about how data is
// stored is discharged by the type existing at all — and demanding one anyway
// would be answered by writing a test that asserts nothing, which is worse than
// a waiver with a reason.
//
// The waiver goes on a construct satisfying the requirement, not on the
// requirement. spec.Waive attaches to a Go construct and a requirement
// declaration is not one, so the narrow form has to be written where the
// implementation is: the same place somebody would look to decide whether a
// test is possible. K3-REQ-UNCOVERED settles for the undirected waiver here;
// this is the sharper answer, and it costs one extra lookup.
const RuleRequirementUnverified = "K14-REQ-UNVERIFIED"

// Verification is the result of the verification direction.
type Verification struct {
	// Normative is the number of requirements that had to be demonstrated.
	Normative int
	// Verified is how many are claimed by at least one test.
	Verified int
	// ByTest maps a requirement ID to the tests naming it.
	ByTest map[string][]string
}

// Ratio is the demonstrated share, 1 for an empty set.
func (v Verification) Ratio() float64 {
	if v.Normative == 0 {
		return 1
	}
	return float64(v.Verified) / float64(v.Normative)
}

// CoverVerification checks that every normative requirement is named by a test.
//
// It reads claims, not evidence. A spec.Verified call proves that somebody
// wrote it down, and that is all this can see: the call may sit behind a
// condition that never holds, or in a test that fails before reaching it.
// Distinguishing a claim from a demonstration needs the line the call writes
// when it runs, which is recorded separately and checked separately.
//
// Reporting the claim is still worth doing on its own, because it is the half
// that can be forgotten. A test that was never written leaves nothing behind at
// all, and no amount of reading test output will produce it.
func CoverVerification(tree *reqtree.Tree, verifications []ir.Binding, cov Coverage, waived ir.Waivers, out *diag.Set) Verification {
	v := Verification{ByTest: map[string][]string{}}

	for _, b := range verifications {
		for _, a := range b.Assertions {
			if a.Kind != ir.AssertVerified {
				continue
			}
			for _, ref := range a.Requirements {
				r := requirementOf(tree, ref)
				if r == nil {
					// Unresolvable references are a Go compile error long
					// before speclink runs.
					continue
				}
				v.ByTest[r.ID] = append(v.ByTest[r.ID], b.Target.Name)
			}
		}
	}
	for id := range v.ByTest {
		sort.Strings(v.ByTest[id])
		v.ByTest[id] = dedupe(v.ByTest[id])
	}

	for _, r := range tree.All() {
		if !r.Status.MustBeCovered() {
			continue
		}
		v.Normative++
		if len(v.ByTest[r.ID]) > 0 {
			v.Verified++
			continue
		}
		if waivedForRequirement(r.ID, cov, waived) {
			v.Verified++
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 120),
			Rule: RuleRequirementUnverified,
			Pos:  r.Pos,
			What: "no test demonstrates " + r.ID + ".",
			Why:  "Coverage says code was written for this requirement. It has never said the code does what the requirement asks, and nothing else does either: the implementation and the tests come from the same place, and a review that samples will not find the one that is missing.",
			How:  "Write a test and end it with spec.Verified(t, " + callSiteName(r.GoIdent) + "). If it cannot be demonstrated by a test, put spec.Waive(" + quoted(RuleRequirementUnverified) + ", …) with a reason on a construct that satisfies it: " + satisfierHint(r.ID, cov) + ".",
		})
	}
	return v
}

// waivedForRequirement reports whether the rule is waived, either globally or
// on one of the constructs written for this requirement.
func waivedForRequirement(id string, cov Coverage, waived ir.Waivers) bool {
	if waived.Has("", RuleRequirementUnverified) {
		return true
	}
	for _, t := range cov.BySatisfier[id] {
		if waived.Has(t.String(), RuleRequirementUnverified) {
			return true
		}
	}
	return false
}

// callSiteName renders a requirement's Go identifier the way it is written at a
// call site, so the How line can be pasted.
//
// It differs from shortName, which drops the package entirely: a requirement is
// always referenced through its package, because the tree lives elsewhere than
// the code binding to it.
func callSiteName(goIdent string) string {
	pkg, name := goIdent, ""
	for i := len(goIdent) - 1; i >= 0; i-- {
		if goIdent[i] == '.' {
			pkg, name = goIdent[:i], goIdent[i+1:]
			break
		}
	}
	if name == "" {
		return goIdent
	}
	for i := len(pkg) - 1; i >= 0; i-- {
		if pkg[i] == '/' {
			return pkg[i+1:] + "." + name
		}
	}
	return pkg + "." + name
}

func dedupe(in []string) []string {
	out := in[:0]
	var last string
	for i, s := range in {
		if i == 0 || s != last {
			out = append(out, s)
		}
		last = s
	}
	return out
}

// quoted renders a rule ID as it is written in a spec.Waive call.
func quoted(s string) string { return `"` + s + `"` }

// satisfierHint names a construct the waiver can go on, so the fix is a place
// rather than a search. Without one there is nothing satisfying the requirement
// and K3-REQ-UNCOVERED has already said so.
func satisfierHint(id string, cov Coverage) string {
	targets := cov.BySatisfier[id]
	if len(targets) == 0 {
		return "there is none yet, so write the implementation first"
	}
	names := make([]string, 0, len(targets))
	for _, t := range targets {
		names = append(names, t.String())
	}
	sort.Strings(names)
	return names[0]
}
