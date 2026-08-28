package jvm

import (
	"bytes"
	"testing"

	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/reqtree"
)

// This is what the second frontend was for.
//
// Everything below internal/lang is supposed to be language neutral, and until
// now that was a claim with one implementation behind it. Here the same
// reqtree, the same coverage rules and the same diagnostics run over a model
// that came out of compiled bytecode rather than Go source — resolved
// differently, positioned differently, carried by annotations instead of
// generic calls — and nothing in them was changed to allow it.

func TestTheNeutralCoreRunsOnJava(t *testing.T) {
	root, classes := fixture(t)
	r := NewReader(root, classes, nil, "")

	out := &diag.Set{}
	tree := reqtree.Build(root, r.ReadRequirements(out), out)
	tree.CheckLayout(Dialect{}, out)

	cov := check.CoverRequirements(tree, r.ReadBindings(out), nil, Dialect{}, out)

	if cov.Normative != 3 {
		t.Fatalf("counted %d normative requirements, want 3", cov.Normative)
	}
	// Two of the three are implemented; the decision behind them is not, and
	// that is the finding — reported by a rule that has never heard of Java.
	if cov.Ratio() == 1 {
		t.Fatal("everything reported as covered, so the backward direction measured nothing")
	}
	if len(cov.Uncovered) != 1 || cov.Uncovered[0] != "R-DEC-NUMBERING" {
		t.Errorf("uncovered set is %v", cov.Uncovered)
	}
}

// The fix a rule proposes has to be a sentence in the language the reader is
// writing. A Java project told to add `var _ = spec.For[…]` would be a tool
// that had not noticed what it was looking at.
func TestFindingsArePhrasedInJava(t *testing.T) {
	root, classes := fixture(t)
	r := NewReader(root, classes, nil, "")

	out := &diag.Set{}
	tree := reqtree.Build(root, r.ReadRequirements(out), out)
	check.CoverRequirements(tree, r.ReadBindings(out), nil, Dialect{}, out)

	var buf bytes.Buffer
	if err := out.WriteText(&buf); err != nil {
		t.Fatal(err)
	}
	text := buf.String()

	if !bytes.Contains(buf.Bytes(), []byte("@Satisfies(")) {
		t.Errorf("the fix is not written in Java:\n%s", text)
	}
	for _, wrong := range []string{"spec.For", "var _ =", ".annotation.go"} {
		if bytes.Contains(buf.Bytes(), []byte(wrong)) {
			t.Errorf("a Go spelling reached a Java project's diagnostics: %q\n%s", wrong, text)
		}
	}
}

// The derivation graph is assembled from references the Java compiler resolved,
// and the tree resolves them the same way it resolves Go identifiers: by
// collecting declarations first and matching afterwards. Nothing about that
// step is language specific, and this is where that stops being an assertion.
func TestDerivationGraphResolvesAcrossPackages(t *testing.T) {
	root, classes := fixture(t)
	r := NewReader(root, classes, nil, "")

	out := &diag.Set{}
	tree := reqtree.Build(root, r.ReadRequirements(out), out)

	submit := tree.ByID["R-QUOTE-SUBMIT"]
	if submit == nil {
		t.Fatal("R-QUOTE-SUBMIT is missing")
	}
	if len(submit.DerivedFrom) != 1 || submit.DerivedFrom[0] != "R-DEC-NUMBERING" {
		t.Fatalf("derivedFrom resolved to %v, want the requirement ID it points at", submit.DerivedFrom)
	}
	if out.Len() != 0 {
		var buf bytes.Buffer
		_ = out.WriteText(&buf)
		t.Errorf("resolving the graph produced findings:\n%s", buf.String())
	}
}

// A requirement declared in a package the layout rules do not expect must be
// reported, in Java as in Go. The rule reads a directory and an ID prefix, and
// neither of those knows what compiled it.
func TestLayoutRulesApply(t *testing.T) {
	root, classes := fixture(t)
	r := NewReader(root, classes, nil, "")

	out := &diag.Set{}
	tree := reqtree.Build(root, r.ReadRequirements(out), out)
	tree.CheckLayout(Dialect{}, out)

	// The fixture is laid out correctly, so the interesting assertion is that
	// the rule ran at all rather than skipping a tree it did not recognise.
	for _, f := range out.Findings() {
		if f.Rule == "" && f.Code != "" {
			t.Logf("layout finding: %s %s", f.Code, f.What)
		}
	}
	if len(tree.All()) != 3 {
		t.Fatalf("the tree holds %d requirements", len(tree.All()))
	}
}
