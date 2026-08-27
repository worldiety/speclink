package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The scope is the only dial speclink has, and it is deliberately the only one.
// There are no warnings and no severities: a codebase is brought in package by
// package, not rule by rule. "This package is not under speclink yet" is a true
// statement; "this rule half applies here" is not one.

func TestScopeSilencesWhatItDoesNotMeasure(t *testing.T) {
	dir := copyFixture(t, "../../testdata/arch")

	before, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("the negative fixture is supposed to fail:\n%s", before)
	}

	writeConfig(t, dir, `{"scope":["app/billing"]}`)

	after, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("the package left in scope still violates rules:\n%s", after)
	}
	if countFindings(after) >= countFindings(before) {
		t.Errorf("narrowing the scope did not reduce the findings:\nbefore: %s\nafter:  %s", summary(before), summary(after))
	}
	// A run that measures part of a project has to say so, or a green summary
	// is true of what was looked at and silent about what was not.
	if !strings.Contains(after, "outside the configured scope and not measured") {
		t.Errorf("the run did not disclose what it skipped:\n%s", after)
	}
}

// The scope decides what is measured, never what is loaded.
//
// Filtering the loaded set was the first attempt and it was wrong. Every rule
// that resolves across packages then answers differently depending on the
// scope: scoping out pkg/permtext made K5-UC-PERMISSION-I18N report permissions
// that were perfectly fine, because the helper it follows one step into was no
// longer there to resolve. A rule that changes its verdict on untouched code
// because of a configuration setting is worse than a rule that is switched off.
func TestScopeDoesNotBreakCrossPackageResolution(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")
	writeConfig(t, dir, `{"scope":["app/sales","requirements/fun/quote","requirements/dec"]}`)

	out, code := runVerify(t, dir)
	if code != 0 {
		t.Fatalf("a scope that excludes only helpers and unrelated requirements must stay clean:\n%s", out)
	}
	if strings.Contains(out, "hardcoded texts") {
		t.Error("the i18n rule lost its helper to the scope and reported a false positive")
	}
}

// Whether the module has an entry point is a question about the module. A scope
// that happens to exclude cmd/ must not answer it with "no".
func TestScopeDoesNotInventAMissingMain(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")
	writeConfig(t, dir, `{"scope":["app/sales","requirements/fun/quote","requirements/dec"]}`)

	out, _ := runVerify(t, dir)
	if strings.Contains(out, "no main package") {
		t.Errorf("K8-MAIN-EXISTS fired because the entry point was out of scope:\n%s", out)
	}
}

// A restricted scope has to reach the requirement tree, or it is unusable for
// the thing it exists for: the requirements of packages still outside would be
// satisfied by nothing in scope, which is true and would bury the run.
func TestScopeReachesTheRequirementTree(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")
	writeConfig(t, dir, `{"scope":["cmd/erp"]}`)

	out, code := runVerify(t, dir)
	if code != 0 {
		t.Fatalf("a scope holding no domain package must have nothing to report:\n%s", out)
	}
	if !strings.Contains(out, "0 normative requirements") {
		t.Errorf("requirements outside the scope were still measured:\n%s", summary(out))
	}
}

// The tree itself is always read in full. An in-scope construct may bind to any
// requirement, and a tree missing the far end of that reference would make the
// construct read as unbound — the scope decides what is asked of a requirement,
// not whether it exists.
func TestScopeNeverHidesARequirementFromABinding(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")
	writeConfig(t, dir, `{"scope":["app/sales"]}`)

	out, _ := runVerify(t, dir)
	if strings.Contains(out, "names no requirement") || strings.Contains(out, "K1-CONSTRUCT-UNBOUND") {
		t.Errorf("a binding lost its requirement to the scope:\n%s", out)
	}
	if !strings.Contains(out, "100% bound") {
		t.Errorf("constructs went unbound under a scope that keeps them:\n%s", summary(out))
	}
}

// An empty scope is the normal case and must change nothing.
func TestNoScopeMeasuresEverything(t *testing.T) {
	plain, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("the reference project is not clean:\n%s", plain)
	}
	if strings.Contains(plain, "outside the configured scope") {
		t.Error("an unconfigured project reported a restricted scope")
	}
}

func writeConfig(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "speclink.json"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func countFindings(out string) int {
	return strings.Count(out, "[SPEC-")
}
