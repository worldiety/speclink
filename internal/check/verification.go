// Package check, this file: verification coverage.
package check

import (
	"sort"
	"strconv"
	"strings"

	"github.com/worldiety/speclink/internal/baseline"

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
// requirement. A waiver attaches to a construct of the host language and a
// requirement declaration is not one, so the narrow form has to be written
// where the implementation is: the same place somebody would look to decide whether a
// test is possible. K3-REQ-UNCOVERED settles for the undirected waiver here;
// this is the sharper answer, and it costs one extra lookup.
const RuleRequirementUnverified = "K14-REQ-UNVERIFIED"

// RuleVerificationStale fires when a test claims a requirement but no run has
// shown it doing so.
//
// It is the difference between what is written and what happened, and it is the
// only rule in speclink that can report a defect nothing in the working tree
// contains. The call is there, the test compiles, the source is beyond
// reproach — and the last recorded run either never reached the call, ended in
// a failure, or was made against a wording the requirement no longer has.
//
// Three quite different mistakes reduce to it, which is deliberate. A call
// behind a condition that never holds, a test that fails before reaching the
// end, and a requirement rewritten under a test that still passes are all the
// same thing from here: nobody has shown this, whatever the source suggests.
// The fix is the same too, which is why one rule is enough.
const RuleVerificationStale = "K14-VERIFICATION-STALE"

// Verification is the result of the verification direction.
type Verification struct {
	// Normative is the number of requirements that had to be demonstrated.
	Normative int
	// Verified is how many are claimed by at least one test.
	Verified int
	// ByTest maps a requirement ID to the tests naming it.
	ByTest map[string][]string
	// Shown is how many were not merely claimed but borne out by a recorded
	// run. It is filled in by [Demonstrated], which needs the baseline that
	// CoverVerification deliberately does not read.
	Shown int
}

// Ratio is the claimed share, 1 for an empty set.
func (v Verification) Ratio() float64 {
	if v.Normative == 0 {
		return 1
	}
	return float64(v.Verified) / float64(v.Normative)
}

// ShownRatio is the share that a recorded run actually bore out.
//
// It is reported next to Ratio rather than instead of it, because the gap
// between them is the interesting number: it is exactly the set of tests that
// exist, compile, claim something, and have not been seen to do it.
func (v Verification) ShownRatio() float64 {
	if v.Normative == 0 {
		return 1
	}
	return float64(v.Shown) / float64(v.Normative)
}

// CoverVerification checks that every normative requirement is named by a test.
//
// It reads claims, not evidence. A verification term proves that somebody
// wrote it down, and that is all this can see: the term may sit behind a
// condition that never holds, or in a test that fails before reaching it.
// Distinguishing a claim from a demonstration needs the line the call writes
// when it runs, which is recorded separately and checked separately.
//
// Reporting the claim is still worth doing on its own, because it is the half
// that can be forgotten. A test that was never written leaves nothing behind at
// all, and no amount of reading test output will produce it.
func CoverVerification(tree *reqtree.Tree, verifications []ir.Binding, cov Coverage, measured map[string]bool, waived ir.Waivers, d ir.Dialect, out *diag.Set) Verification {
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
		if measured != nil && !measured[r.ID] {
			continue
		}
		v.Normative++
		if len(v.ByTest[r.ID]) > 0 {
			v.Verified++
			continue
		}
		if waivedForRequirement(r.ID, cov, RuleRequirementUnverified, waived) {
			v.Verified++
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 120),
			Rule: RuleRequirementUnverified,
			Pos:  r.Pos,
			What: "no test demonstrates " + r.ID + ".",
			Why:  "Coverage says code was written for this requirement. It has never said the code does what the requirement asks, and nothing else does either: the implementation and the tests come from the same place, and a review that samples will not find the one that is missing.",
			How:  "Write a test and end it with " + d.Verify(callSiteName(r.GoIdent)) + ". If it cannot be demonstrated by a test, put " + d.Waive(RuleRequirementUnverified) + " with a reason on a construct that satisfies it: " + satisfierHint(r.ID, cov) + ".",
		})
	}
	return v
}

// waivedForRequirement reports whether the rule is waived, either globally or
// on one of the constructs written for this requirement.
func waivedForRequirement(id string, cov Coverage, rule string, waived ir.Waivers) bool {
	if waived.Has("", rule) {
		return true
	}
	for _, t := range cov.BySatisfier[id] {
		if waived.Has(t.String(), rule) {
			return true
		}
	}
	return false
}

// Demonstrated checks the claims against the record of what actually ran.
//
// CoverVerification reads the source and reports the test nobody wrote.
// This reads the baseline and reports the test nobody ran, which no reading of
// the source can find: the call is there and the code is beyond reproach.
//
// A requirement with no claim at all is not reported twice. K14-REQ-UNVERIFIED
// has already said the only thing worth saying about it, and a second finding
// would only make the first one look like half of a bigger problem.
func Demonstrated(tree *reqtree.Tree, v Verification, cov Coverage, measured map[string]bool, base *baseline.File, waived ir.Waivers, out *diag.Set) int {
	// Counted down from what was claimed, not from the whole. Counting down
	// from the whole reported a hundred percent demonstrated next to zero
	// percent verified, because nothing claimed and therefore nothing was
	// stale — a figure that was arithmetically defensible and plainly false.
	stale := 0
	for _, r := range tree.All() {
		if !r.Status.MustBeCovered() {
			continue
		}
		if measured != nil && !measured[r.ID] {
			continue
		}
		claimants := v.ByTest[r.ID]
		if len(claimants) == 0 {
			continue // K14-REQ-UNVERIFIED, or waived there
		}
		if len(base.VerifiedBy(r.ID, baseline.HashText(r.Text, r.Title))) > 0 {
			continue
		}
		if waivedForRequirement(r.ID, cov, RuleVerificationStale, waived) {
			continue
		}
		stale++
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 121),
			Rule: RuleVerificationStale,
			Pos:  r.Pos,
			What: staleWhat(claimants) + " " + r.ID + ", but no run has shown it.",
			Why:  "A verification term writes its line when control reaches it, and only a passing test has its line recorded. So one of three things is true and the source cannot tell them apart: the call was never reached, the test failed before the end, or the requirement was rewritten after the last run and the evidence was against the old wording.",
			How:  "Run the tests and hand the result over: go test -json ./... | speclink evidence. If they do not pass, that is the finding.",
		})
	}
	return v.Verified - stale
}

func staleWhat(claimants []string) string {
	if len(claimants) == 1 {
		return shortName(claimants[0]) + " claims"
	}
	return strconv.Itoa(len(claimants)) + " tests claim"
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

// RecordVerifications folds a set of demonstrations into the baseline.
//
// demonstrated maps a requirement ID to the tests that passed while claiming
// it. Each record is bound to the wording the requirement has now, which is the
// wording those tests just ran against — that binding is what makes a later
// rewrite void its own evidence without a second mechanism.
//
// Records for a requirement are replaced rather than merged. The input is one
// complete test run, so a test that no longer demonstrates a requirement has to
// disappear from the record; merging would keep it alive forever.
func RecordVerifications(base *baseline.File, tree *reqtree.Tree, demonstrated map[string][]string) (changed []string, unknown []string) {
	for _, id := range sortedKeys(demonstrated) {
		r := tree.ByID[id]
		if r == nil {
			unknown = append(unknown, id)
			continue
		}
		text := baseline.HashText(r.Text, r.Title)

		tests := append([]string(nil), demonstrated[id]...)
		sort.Strings(tests)

		records := make([]baseline.Verification, 0, len(tests))
		for _, test := range tests {
			records = append(records, baseline.Verification{Test: test, Text: text})
		}
		if !sameVerifications(base.Verifications[id], records) {
			changed = append(changed, "verified  "+id+" ("+strings.Join(tests, ", ")+")")
		}
		base.Verifications[id] = records
	}

	// A requirement nothing demonstrated in this run keeps no record. The run
	// is the whole truth about what passed, and leaving an old entry would let
	// a deleted test go on vouching for something.
	//
	// A removal is a change like any other. Reporting only additions was a real
	// bug: with nothing new to record the caller returned early and never wrote
	// the file, so the evidence for a test that had just stopped passing stayed
	// in the baseline and the run went green.
	for _, id := range sortedKeys(base.Verifications) {
		if _, ran := demonstrated[id]; ran {
			continue
		}
		delete(base.Verifications, id)
		changed = append(changed, "withdrawn "+id+": no passing test demonstrated it in this run")
	}
	sort.Strings(changed)
	return changed, unknown
}

func sameVerifications(a, b []baseline.Verification) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
