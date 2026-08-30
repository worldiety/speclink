package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// writeChapters replaces the chapter declarations of a copied fixture.
func writeChapters(t *testing.T, dir, body string) {
	t.Helper()
	src := "package requirements\n\nimport \"github.com/worldiety/speclink/spec\"\n\n" + body
	path := filepath.Join(dir, "requirements", "chapters.spec.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestProseIsSetWhereTheOutlineSaysItGoes is the reason chapters exist.
//
// Every other chapter of the document is derived: it states what the model
// states, and each sentence is written in exactly one place. Why the system is
// cut this way follows from no model, and a specification without it describes
// a module to somebody who already knows what the module is for.
func TestProseIsSetWhereTheOutlineSaysItGoes(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	out, code := runSpeclink(t, "generate", dir)
	if code != 0 {
		t.Fatalf("generate failed with %d:\n%s", code, out)
	}

	title := "Warum es diese Anwendung gibt"
	if !strings.Contains(out, title) {
		t.Fatalf("the declared chapter is missing from the document:\n%s", out)
	}
	// The prose stands before the figures, because that is the place it
	// declares. A chapter that landed anywhere the assembly happened to put it
	// would make the outline decorative.
	// Against the heading, not the first mention of the words: the chapter on
	// how to read the document names the later ones in a table.
	if strings.Index(out, "## "+title) > strings.Index(out, "## Where it stands") {
		t.Error("the chapter was set after the summary, but it declares spec.Beginning")
	}
	// Its own structure survives the crossing: a numbered list says the order
	// carries meaning, and rendering it as bullets would change the claim.
	if !strings.Contains(out, "1. Die Nummernvergabe ist lückenlos") {
		t.Errorf("the numbered list of the prose was not kept:\n%s", out)
	}
	// The anchor comes from the declared ID and never from the words of the
	// heading, so a cross reference survives somebody improving the title.
	if !strings.Contains(out, `<a id="req-warum"></a>`) {
		t.Errorf("the chapter carries no anchor from its ID:\n%s", out)
	}
}

// TestAChapterNamingNoProseIsReported is what the declaration buys.
//
// This is the whole argument for an outline that is declared rather than
// configured. A chapter whose file was moved is caught while the specification
// is checked, at the line that names it. Left until the document is written it
// would be found by whoever next reads the document, and a chapter that is
// missing looks exactly like a chapter nobody ever wrote.
func TestAChapterNamingNoProseIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	writeChapters(t, dir, `var Warum = spec.Chapter{
	ID:  "warum",
	Doc: "doc/verschwunden.md",
	At:  spec.Beginning,
}
`)

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a chapter naming a file that is not there must be reported:\n%s", summary(out))
	}
	if !strings.Contains(out, "SPEC-V6-181") {
		t.Errorf("expected the missing prose to be reported, got:\n%s", summary(out))
	}
	// Against the declaration, not against the document, because that is where
	// somebody can act on it.
	if !strings.Contains(out, "chapters.spec.go") {
		t.Errorf("the finding does not point at the declaration:\n%s", summary(out))
	}
}

// TestAChapterThatCannotBePlacedIsReported covers the three ways an outline
// entry produces nothing.
func TestAChapterThatCannotBePlacedIsReported(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		body string
		code string
	}{
		{
			// Two chapters under one name make both ambiguous, and a
			// reference to either lands wherever the assembly put it.
			name: "duplicate",
			body: `var A = spec.Chapter{ID: "warum", Doc: "doc/warum.md", At: spec.Beginning}

var B = spec.Chapter{ID: "warum", Doc: "doc/warum.md", At: spec.Appendix}
`,
			code: "SPEC-V6-180",
		},
		{
			// The heading of the file is the title of the chapter. Prose that
			// opens without one has no name in the table of contents.
			name: "untitled",
			body: `var A = spec.Chapter{ID: "kopflos", Doc: "doc/kopflos.md", At: spec.Appendix}
`,
			code: "SPEC-V6-182",
		},
		{
			// One rule for every missing field, because the consequence is
			// the same and a reader has to fix it once.
			name: "incomplete",
			body: `var A = spec.Chapter{ID: "luecke"}
`,
			code: "SPEC-V5-061",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir := copyFixture(t, "../../testdata/bare")
			if tc.name == "untitled" {
				headless := filepath.Join(dir, "doc", "kopflos.md")
				if err := os.WriteFile(headless, []byte("Dieser Text hat keine Überschrift.\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			writeChapters(t, dir, tc.body)

			out, code := runVerify(t, dir)
			if code == 0 {
				t.Fatalf("expected %s, but the run was clean:\n%s", tc.code, summary(out))
			}
			if !strings.Contains(out, tc.code) {
				t.Errorf("expected %s, got:\n%s", tc.code, summary(out))
			}
		})
	}
}

// TestADecisionMustSayWhatItCosts is the half nobody writes unprompted.
//
// A justification is pleasant to write and gets written at length. Admitting
// what the ruling makes worse is not, and in a single field it quietly
// disappears — leaving a record that reads as advocacy for the choice rather
// than an account of it.
func TestADecisionMustSayWhatItCosts(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")

	path := filepath.Join(dir, "requirements", "dec", "R-DEC-NUMBERING.spec.go")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cut := regexpConsequences.ReplaceAll(src, nil)
	if len(cut) == len(src) {
		t.Fatal("the fixture no longer declares Consequences")
	}
	if err := os.WriteFile(path, cut, 0o644); err != nil {
		t.Fatal(err)
	}

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a decision without its cost must be reported:\n%s", summary(out))
	}
	if !strings.Contains(out, "SPEC-V5-006") {
		t.Errorf("expected the missing consequences to be reported, got:\n%s", summary(out))
	}
}

var regexpConsequences = regexp.MustCompile(`(?m)^\tConsequences:.*\n`)
