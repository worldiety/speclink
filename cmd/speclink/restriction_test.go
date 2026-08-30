package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// appendAnnotation adds a binding to the annotation file of the file store.
func appendAnnotation(t *testing.T, dir, body string) {
	t.Helper()
	path := filepath.Join(dir, "app", "sales", "adapter", "fs", "quotes.annotation.go")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(src, []byte("\n"+body+"\n")...), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestARuleAboutValuesMustBeDecidable is the argument for vectors.
//
// A restriction is written in prose because the type system cannot carry it.
// Prose is exactly what a schema generator drops without a sound, and both ends
// then agree about the shape of a message while disagreeing about what may be
// in it. The examples are the only part that survives translation into another
// language and another team.
func TestARuleAboutValuesMustBeDecidable(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		drop string
		want string
	}{
		{
			// Accepting what should be accepted is what an implementation
			// does by accident. Refusing what must be refused is the part
			// nobody gets right unprompted.
			name: "without the cases it must refuse",
			drop: "spec.Invalid(",
			want: "no example it must reject",
		},
		{
			name: "without the cases it must accept",
			drop: "spec.Valid(",
			want: "no example it must accept",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := copyFixture(t, "../../testdata/bare")
			dropLine(t, dir, "app/sales/model.annotation.go", tc.drop)

			out, code := runVerify(t, dir)
			if code == 0 {
				t.Fatalf("a rule nothing decides must be reported:\n%s", summary(out))
			}
			if !strings.Contains(out, "SPEC-V6-191") || !strings.Contains(out, tc.want) {
				t.Errorf("expected %q, got:\n%s", tc.want, summary(out))
			}
		})
	}
}

// dropLine removes the first line of a file containing the marker.
func dropLine(t *testing.T, dir, rel, marker string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(src), "\n")
	for i, l := range lines {
		if strings.Contains(l, marker) {
			out := append(append([]string{}, lines[:i]...), lines[i+1:]...)
			if err := os.WriteFile(path, []byte(strings.Join(out, "\n")), 0o644); err != nil {
				t.Fatal(err)
			}
			return
		}
	}
	t.Fatalf("%s no longer contains %q", rel, marker)
}

// TestExamplesWithoutARuleAreReported covers the other direction.
//
// An example decides a rule. Without one written down a reader can see that a
// value was refused and not why, and the next case nobody thought of has
// nothing to be judged against.
func TestExamplesWithoutARuleAreReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	dropLine(t, dir, "app/sales/model.annotation.go", "spec.Restrict(")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("examples deciding nothing must be reported:\n%s", summary(out))
	}
	if !strings.Contains(out, "SPEC-V6-190") {
		t.Errorf("expected the unstated rule to be reported, got:\n%s", summary(out))
	}
}

// TestTheRuleAndItsCasesReachTheDocument is what the whole thing is for.
//
// The restriction exists so that somebody on the far end can program against
// it. Recording it and not printing it would leave the tool knowing the one
// thing the reader needed.
func TestTheRuleAndItsCasesReachTheDocument(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	out, code := runSpeclink(t, "generate", dir)
	if code != 0 {
		t.Fatalf("generate failed with %d:\n%s", code, out)
	}
	if !strings.Contains(out, "Values with a rule") {
		t.Fatalf("the chapter of restricted values is missing:\n%s", out)
	}
	// Quoted in Go notation, because an example ending in an invisible byte is
	// one nobody can copy correctly.
	if !strings.Contains(out, `"Q-1\n"`) {
		t.Errorf("the escaped example is not printed as written:\n%s", out)
	}
}

// TestAClaimIsCarriedToWhoeverReadsTheField is why the marker exists.
//
// A field holding a verified identity and a field holding one the far end
// merely says it has are the same string in the struct, in the payload and in
// every generated schema. Nothing about the second is wrong, which is why a
// review does not catch it, and why somebody downstream reads it as the first
// — at which point an assertion has become an authorisation.
func TestAClaimIsCarriedToWhoeverReadsTheField(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	// Written here rather than into the fixture, because the file store is
	// this system's own and the ruling on it says the stored state is the
	// authoritative one. A demonstration that contradicts the fixture would
	// teach the wrong thing about when to reach for this.
	appendAnnotation(t, dir, `var _ = spec.ForField[QuoteStore]("Note",
	spec.Claim("Wer Rechte auf das Verzeichnis hat, kann ihn geschrieben haben."),
)`)

	out, code := runSpeclink(t, "generate", dir)
	if code != 0 {
		t.Fatalf("generate failed with %d:\n%s", code, out)
	}
	if !strings.Contains(out, "may not take them as true") {
		t.Fatalf("the claim is nowhere in the document:\n%s", out)
	}
	if !strings.Contains(out, "Wer Rechte auf das Verzeichnis hat") {
		t.Errorf("what decides the matter instead is missing:\n%s", out)
	}
}

// TestAClaimWithoutItsReasonIsReported covers the half that makes it usable.
func TestAClaimWithoutItsReasonIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	appendAnnotation(t, dir, `var _ = spec.ForField[QuoteStore]("Note", spec.Claim(""))`)

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a claim naming no authority must be reported:\n%s", summary(out))
	}
	if !strings.Contains(out, "SPEC-V3-009") {
		t.Errorf("expected the missing reason to be reported, got:\n%s", summary(out))
	}
}

// TestAnUnstatedWireShapeSaysSo covers the difference between two silences.
//
// On a builder that states its types, an empty response means the route
// promises no shape. On the standard library's router it means nobody asked.
// Printing nothing in the second case lets a reader take it for the first.
func TestAnUnstatedWireShapeSaysSo(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	out, code := runSpeclink(t, "generate", dir)
	if code != 0 {
		t.Fatalf("generate failed with %d:\n%s", code, out)
	}
	if !strings.Contains(out, "What crosses each address") {
		t.Fatalf("the section is missing entirely:\n%s", out)
	}
	if !strings.Contains(out, "Nothing here is known") {
		t.Errorf("the reason nothing is known is not given:\n%s", out)
	}
}
