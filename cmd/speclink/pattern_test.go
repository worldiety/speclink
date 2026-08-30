package main

import (
	"strings"
	"testing"
)

// A package pattern on the command line is a scope, and these are the tests
// that say so.
//
// Their absence is the whole story of the defect they now guard. The same
// mistake had already been made once for the configured scope, diagnosed, and
// fixed — "the scope decides what is measured, never what is loaded" is written
// out in scope_test.go. Nobody wrote the counterpart for patterns, so the
// second narrowing path kept filtering the load for a long time afterwards, and
// every rule that resolves across packages quietly changed its verdict with the
// arguments the operator typed.
//
// Each test below is one of those in scope_test.go, asked of a pattern instead
// of a configuration key. Where the two mechanisms answer differently, one of
// them is wrong.

// TestPatternNarrowsWhatIsMeasuredNotWhatIsLoaded is the headline.
//
// The fixture is clean and stays clean at every width. Before, this produced
// nine to eleven findings against untouched code: a withdrawn address that was
// still mounted, aggregates resting on no decision because the decisions live
// in requirements/, document sections that became no requirement because the
// tree had not been read, and a module with no entry point because cmd/ had not
// been loaded.
func TestPatternNarrowsWhatIsMeasuredNotWhatIsLoaded(t *testing.T) {
	t.Parallel()
	for _, pattern := range []string{
		"./...",
		"./app/...",
		"./app/sales/...",
		"./app/sales/rest/...",
		"./foundation/...",
		"./cmd/...",
	} {
		out, code := runSpeclink(t, "verify", "../../testdata/bare", pattern)
		if code != 0 {
			t.Errorf("the clean fixture failed at width %q:\n%s", pattern, out)
		}
	}
}

// TestPatternDoesNotInventAMissingMain is the case whose fix was already
// written down in the rule itself.
//
// CheckMainPackages counts entry points before consulting the scope, with the
// comment that whether the module has one is a question about the module. That
// reasoning is exactly right and was defeated by a loader that had already
// thrown cmd/ away before the rule ran.
func TestPatternDoesNotInventAMissingMain(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "verify", "../../testdata/bare", "./app/sales/...")
	if strings.Contains(out, "no main package") {
		t.Errorf("K8-MAIN-EXISTS fired because the entry point was not loaded:\n%s", out)
	}
}

// TestPatternNeverHidesTheRequirementTree guards the most expensive symptom.
//
// The tree lives in requirements/, so a pattern naming a code package used to
// leave it empty — and an empty tree is not an error, it is a hundred percent
// of nothing. The run announced full coverage having read no requirement at
// all, which is the one way this tool could mislead without stating a
// falsehood.
func TestPatternNeverHidesTheRequirementTree(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "verify", "../../testdata/bare", "./app/sales/...")
	if strings.Contains(out, "100% covered") {
		t.Errorf("a run that measured no requirement claimed full coverage:\n%s", summary(out))
	}
	if !strings.Contains(out, "no normative requirement in this scope") {
		t.Errorf("the figure must name what it left out:\n%s", summary(out))
	}
	if strings.Contains(out, "became no requirement") {
		t.Errorf("K12-SOURCE-UNCOVERED fired because the tree was not loaded:\n%s", out)
	}
}

// TestPatternDoesNotBreakCrossPackageResolution is the failure that first
// taught this lesson, asked of the other mechanism.
//
// A rule that follows a helper one step into another package answers
// differently when that package is not there. A verdict on untouched code that
// depends on the command line is worse than a rule switched off, because it
// looks like a finding.
func TestPatternDoesNotBreakCrossPackageResolution(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "verify", "../../testdata/example", "./app/sales/...")
	if code != 0 {
		t.Fatalf("a pattern that excludes only helpers must stay clean:\n%s", out)
	}
	if strings.Contains(out, "hardcoded texts") {
		t.Error("the i18n rule lost its helper to the pattern and reported a false positive")
	}
}

// TestPatternDisclosesWhatItSkipped is the obligation that comes with silence.
//
// A run that measures part of a project has to say so. Narrowing correctly and
// then reporting a clean summary would be true of what was looked at and silent
// about what was not.
func TestPatternDisclosesWhatItSkipped(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "verify", "../../testdata/bare", "./app/sales/...")
	if !strings.Contains(out, "outside the configured scope and not measured") {
		t.Errorf("a narrowed run did not disclose what it skipped:\n%s", out)
	}
}

// TestWholeModulePatternIsNotARestriction pins the normal case.
func TestWholeModulePatternIsNotARestriction(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "verify", "../../testdata/bare", "./...")
	if code != 0 {
		t.Fatalf("the fixture is not clean:\n%s", out)
	}
	if strings.Contains(out, "outside the configured scope") {
		t.Error("naming the whole module reported a restricted scope")
	}
}

// TestSeveralPatternsCompose is how a narrowed run reaches the tree.
//
// A pattern is a scope entry, so naming the code and the requirements together
// is the same thing as listing both in the configuration — which is what the
// configured scope has always required for exactly this reason.
func TestSeveralPatternsCompose(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "verify", "../../testdata/bare", "./app/...", "./requirements/...")
	if code != 0 {
		t.Fatalf("naming the code and the tree together must verify:\n%s", out)
	}
	if !strings.Contains(out, "normative requirements") {
		t.Errorf("the tree was named and still not measured:\n%s", summary(out))
	}
}

// TestUnsupportedPatternIsRefusedRatherThanApproximated is the rule this whole
// change is about, applied to its own input.
//
// An import path or a meta pattern could be mapped onto directories with a
// couple of assumptions. A pattern mapped slightly wrong does not fail — it
// measures something other than what was asked for and reports the result as
// though the question had been answered.
func TestUnsupportedPatternIsRefusedRatherThanApproximated(t *testing.T) {
	t.Parallel()
	for _, pattern := range []string{"example.com/bare/app/sales", "all", "./app/*/rest"} {
		out, code := runSpeclink(t, "verify", "../../testdata/bare", pattern)
		if code == 0 {
			t.Errorf("pattern %q was accepted and silently approximated:\n%s", pattern, out)
		}
		if !strings.Contains(out, "package pattern") {
			t.Errorf("pattern %q failed without saying why:\n%s", pattern, out)
		}
	}
}

// TestRequirementsMayStillNarrowTheLoad is the exception, and it is an
// exception because of what that command does not claim.
//
// Reading the tree has to work while the implementation around it is in pieces,
// so it loads only what it was given. That is sound for exactly one reason: it
// never asks a question about the module. Every other command does, which is
// why every other command loads all of it.
func TestRequirementsMayStillNarrowTheLoad(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "app/sales/model.go", "type Quote struct {", "type Quote struct {\n\tBroken int = \"\"")

	out, code := runSpeclink(t, "requirements", dir, "./requirements/...")
	if code != 0 {
		t.Fatalf("the tree must be readable while the code is broken:\n%s", out)
	}
	if _, verifyCode := runVerify(t, dir); verifyCode == 0 {
		t.Error("the fixture was supposed to be left uncompilable")
	}
}

func TestScopeFromPatterns(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		in   []string
		want []string
	}{
		{nil, nil},
		{[]string{"./..."}, nil},
		{[]string{"./app/sales/..."}, []string{"app/sales/**"}},
		{[]string{"./app/sales"}, []string{"app/sales"}},
		{[]string{"./app/...", "./requirements/..."}, []string{"app/**", "requirements/**"}},
		// The whole module and a part of it is the whole module. Intersecting
		// them would answer a question nobody put.
		{[]string{"./app/...", "./..."}, nil},
	} {
		got, err := scopeFromPatterns(tc.in)
		if err != nil {
			t.Errorf("scopeFromPatterns(%q): %v", tc.in, err)
			continue
		}
		if strings.Join(got, ",") != strings.Join(tc.want, ",") {
			t.Errorf("scopeFromPatterns(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
