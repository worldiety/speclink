package jvm

import (
	"path/filepath"
	"testing"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

func fixture(t *testing.T) (string, []*Class) {
	t.Helper()

	root, err := filepath.Abs(filepath.Join("..", "..", "..", "testdata", "java"))
	if err != nil {
		t.Fatal(err)
	}
	classes, errs := Load(root, nil)
	for _, e := range errs {
		t.Errorf("load: %v", e)
	}
	if len(classes) == 0 {
		t.Fatal("no classes found — run testdata/java/build.sh")
	}
	return root, classes
}

func read(t *testing.T) (*Reader, *diag.Set) {
	t.Helper()
	root, classes := fixture(t)
	return NewReader(root, classes, nil, ""), &diag.Set{}
}

func TestReadsTheRequirementTree(t *testing.T) {
	r, out := read(t)
	reqs := r.ReadRequirements(out)

	if out.Len() != 0 {
		t.Fatalf("unexpected findings: %d", out.Len())
	}
	if len(reqs) != 3 {
		t.Fatalf("read %d requirements, want 3", len(reqs))
	}

	var submit *ir.Requirement
	for _, req := range reqs {
		if req.ID == "R-QUOTE-SUBMIT" {
			submit = req
		}
	}
	if submit == nil {
		t.Fatal("R-QUOTE-SUBMIT is missing")
	}
	if submit.Kind != ir.Functional || submit.Status != ir.Normative {
		t.Errorf("kind and status read as %v / %v", submit.Kind, submit.Status)
	}
	if submit.Title == "" || submit.Text == "" {
		t.Errorf("title or text lost: %+v", submit)
	}
}

// The derivation edge is a class literal, so the Java compiler resolves it. If
// this ever became a string, speclink would be verifying a reference nobody
// else checked — which is the property the whole carrier form was chosen for.
func TestDerivationIsAResolvedReference(t *testing.T) {
	r, out := read(t)
	reqs := r.ReadRequirements(out)

	for _, req := range reqs {
		if req.ID != "R-QUOTE-SUBMIT" {
			continue
		}
		if len(req.DerivedFrom) != 1 {
			t.Fatalf("derivedFrom is %v", req.DerivedFrom)
		}
		// It carries the class, which is the identity a reference resolves to
		// here — the same role the qualified Go identifier plays elsewhere.
		if req.DerivedFrom[0] != "com.example.requirements.dec.RDecNumbering" {
			t.Errorf("derivedFrom resolved to %q", req.DerivedFrom[0])
		}
		return
	}
	t.Fatal("R-QUOTE-SUBMIT is missing")
}

func TestReadsSourcesIncludingExternal(t *testing.T) {
	r, out := read(t)

	byID := map[string]*ir.Requirement{}
	for _, req := range r.ReadRequirements(out) {
		byID[req.ID] = req
	}

	if got := byID["R-QUOTE-SUBMIT"].Sources; len(got) != 1 || got[0].Anchor != "8-abgabe" {
		t.Errorf("document source read as %+v", got)
	}
	if got := byID["R-DEC-NUMBERING"].Sources; len(got) != 1 || got[0].Extern != "GoBD Rz. 36" {
		t.Errorf("external source read as %+v", got)
	}
}

// An annotation attaches to whatever declaration it precedes, so a binding
// targets a type or a method depending on where it was written — and both have
// to arrive with the target the rules expect, or a method binding silently
// marks its whole class as accounted for.
func TestReadsBindings(t *testing.T) {
	r, out := read(t)
	bindings := r.ReadBindings(out)

	types, methods := 0, 0
	for _, b := range bindings {
		switch b.Target.Kind {
		case ir.TargetType:
			types++
		case ir.TargetFunc:
			methods++
		default:
			t.Errorf("%s bound as %v", b.Target.Name, b.Target.Kind)
		}
		if len(b.Assertions) != 1 || len(b.Assertions[0].Requirements) != 1 {
			t.Errorf("%s carries %+v", b.Target.Name, b.Assertions)
		}
	}
	if types != 2 || methods != 3 {
		t.Errorf("read %d type and %d method bindings, want 2 and 3", types, methods)
	}
}

// The class file has no line for a class and none at all for a field, so the
// line comes from a text search over the source. What it must never do is
// answer confidently and wrongly: a reader sent to the wrong line concludes the
// tool is confused, where a reader given only the file opens it and looks.
func TestPositionsPointAtTheSource(t *testing.T) {
	r, out := read(t)

	for _, req := range r.ReadRequirements(out) {
		if req.ID != "R-QUOTE-SUBMIT" {
			continue
		}
		if got := filepath.ToSlash(req.Pos.File); got != "src/com/example/requirements/fun/quote/RQuoteSubmit.java" {
			t.Errorf("position points at %q, not the source", got)
		}
		if req.Pos.Line == 0 {
			t.Error("no line was recovered for a class that is plainly declared in its file")
		}
		return
	}
	t.Fatal("R-QUOTE-SUBMIT is missing")
}

func TestFindDeclarationIgnoresComments(t *testing.T) {
	lines := []string{
		"package com.example;",
		"",
		"/**",
		" * Talks about notes at length before anything is declared.",
		" */",
		"public final class Customer {",
		"    // notes in a line comment",
		"    private List<String> notes;",
		"}",
	}
	if got := findDeclaration(lines, "notes"); got != 8 {
		t.Errorf("found the declaration at line %d, want 8", got)
	}
}

// Without a whole word test, looking for "id" matches "quoteId", "valid" and
// "identity" — and one of those is on almost every line of a data class.
func TestFindDeclarationMatchesWholeWords(t *testing.T) {
	lines := []string{
		"    private String quoteId;",
		"    private boolean valid;",
		"    private String id;",
	}
	if got := findDeclaration(lines, "id"); got != 3 {
		t.Errorf("found %q at line %d, want 3", "id", got)
	}
}

func TestUnknownNameYieldsFileWithoutLine(t *testing.T) {
	root, classes := fixture(t)
	p := newPositions(root, nil)

	for _, c := range classes {
		if c.Name != "com.example.sales.SubmitQuote" {
			continue
		}
		file, line := p.Of(c, "somethingNobodyWrote")
		if line != 0 {
			t.Errorf("a name that is not there produced line %d", line)
		}
		if file == "" {
			t.Error("the file was lost along with the line")
		}
		return
	}
	t.Fatal("the fixture class is missing")
}

func TestClassNameOfRequirementID(t *testing.T) {
	for id, want := range map[string]string{
		"R-QUOTE-SUBMIT":  "RQuoteSubmit",
		"R-DEC-NUMBERING": "RDecNumbering",
		"R-NFR-AUDIT":     "RNfrAudit",
	} {
		if got := ClassNameOf(id); got != want {
			t.Errorf("%s became %q, want %q", id, got, want)
		}
	}
}

func TestDialectSpeaksJava(t *testing.T) {
	var d ir.Dialect = Dialect{}

	if got := d.RequirementFile("R-QUOTE-SUBMIT"); got != "RQuoteSubmit.java" {
		t.Errorf("requirement file is %q", got)
	}
	if got := d.Satisfy("com.example.requirements.RQuoteSubmit"); got != "@Satisfies(RQuoteSubmit.class)" {
		t.Errorf("satisfy reads %q", got)
	}
	// There is no sidecar in Java and there cannot be: an annotation attaches
	// to the declaration it precedes, and a public type is alone in its file.
	if got := d.AnnotationFile("src/com/example/sales/SubmitQuote.java"); got != "SubmitQuote.java" {
		t.Errorf("annotation file reads %q", got)
	}
}

// A status is an enum constant, and Java spells those in screaming snake case.
// Upper casing OutOfScope gives OUTOFSCOPE, which does not exist and cannot be
// pasted — the small kind of wrong that makes a reader distrust the rest.
func TestStatusNamesAreValidJava(t *testing.T) {
	var d ir.Dialect = Dialect{}

	for name, want := range map[string]string{
		"Planned":     "Status.PLANNED",
		"OutOfScope":  "Status.OUT_OF_SCOPE",
		"Informative": "Status.INFORMATIVE",
	} {
		if got := d.Status(name); got != want {
			t.Errorf("%s became %q, want %q", name, got, want)
		}
	}
}
