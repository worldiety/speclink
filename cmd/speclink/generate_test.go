package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The document is the reason for the tool. As long as speclink exists beside a
// hand written specification, a generator and a knowledge graph, it makes the
// situation worse: one more thing to keep in step. Nothing can be removed until
// what those produce comes out of here.

func TestGeneratedDocumentCarriesTheWholeChain(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "generate", "../../testdata/example", "./...")
	if code != 0 {
		t.Fatalf("generate failed with %d:\n%s", code, out)
	}

	// Every link of the chain has to be visible on one page, or the reader is
	// back to holding the graph in their head.
	for _, want := range []string{
		"R-QUOTE-SUBMIT", // the requirement
		"sequential, duplicate free quote number MUST be drawn",    // its words
		`requirements/\_sources/sales/quoteflow.md#8-abgabe`,       // where they came from
		`requirements/\_sources/sales/quotescreen.png#abgabeknopf`, // and from which mockup
		"sales.SubmitQuote",                  // what implements them
		"TestSubmitQuoteDrawsAGaplessNumber", // what demonstrated them
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the document does not mention %q", want)
		}
	}
}

// A waiver without its reason would be the worst of both: the gap reads as an
// oversight, and the sentence somebody was made to write disappears. The reason
// is mandatory precisely so it can be read here.
func TestGapsCarryTheReasonTheyWereAccepted(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/example", "./...")

	gaps := between(out, "## Gaps", "## Requirements")
	if !strings.Contains(gaps, "R-DEC-CUSTOMER-STATE") {
		t.Fatalf("the waived requirement is not listed as a gap:\n%s", gaps)
	}
	if !strings.Contains(gaps, "accepted:") || !strings.Contains(gaps, "stored as state rather than as facts") {
		t.Errorf("the gap does not say why it was accepted:\n%s", gaps)
	}
}

// A section of a document that became nothing has to be visible as such. It is
// the defect no other view can show, and the table is where a person sees it.
func TestSourceTableShowsWhatBecameNothing(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	appendTo(t, dir, "requirements/_sources/sales/quoteflow.md",
		"\n## 11. Rabatt\n\nAb 10 Stueck wird ein Rabatt gewaehrt.\n")

	out, _ := runSpeclink(t, "generate", dir, "./...")

	if !strings.Contains(out, "Source segments that became no requirement") {
		t.Errorf("the gap list does not name the skipped section:\n%s", between(out, "## Gaps", "## Requirements"))
	}
	if !strings.Contains(out, "**nothing**") {
		t.Errorf("the source table does not mark the section as unused:\n%s", between(out, "## Source documents", ""))
	}
}

// A project with open findings is exactly when somebody wants to read this.
// Refusing to render until everything is green would make it useless in the
// only situation that needs it.
func TestGenerateRendersAProjectWithFindings(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "generate", "../../testdata/arch", "./...")
	if code != 0 {
		t.Fatalf("generate refused a project that does not verify: %d\n%s", code, out)
	}
	if !strings.Contains(out, "# Specification") {
		t.Errorf("nothing was rendered:\n%s", out)
	}
}

// The review is an act, and the document is where it shows.
func TestReviewShowsInTheDocument(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")

	before, _ := runSpeclink(t, "generate", dir, "./...")
	if !strings.Contains(before, "Requirements nobody has read") {
		t.Fatalf("an unreviewed project did not say so:\n%s", between(before, "## Gaps", "## Requirements"))
	}

	if out, code := runSpeclink(t, "freeze", dir, "-reviewer", "Frau Meier", "./..."); code != 0 {
		t.Fatalf("freeze failed:\n%s", out)
	}

	after, _ := runSpeclink(t, "generate", dir, "./...")
	if strings.Contains(after, "Requirements nobody has read") {
		t.Errorf("the recorded review did not reach the document:\n%s", between(after, "## Gaps", "## Requirements"))
	}
	if !strings.Contains(after, "Read by** Frau Meier") {
		t.Errorf("the document does not name who read it")
	}
}

func TestGenerateWritesToAFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "spec.md")

	if out, code := runSpeclink(t, "generate", "../../testdata/example", "-out", path, "./..."); code != 0 {
		t.Fatalf("generate failed with %d:\n%s", code, out)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "# Specification") {
		t.Error("the file does not hold the document")
	}
}

// between returns the text between two headings, to the end when the second is
// empty.
//
// The headings are matched at the start of a line. Matching anywhere cut the
// gap section at its own first subheading, because "### Requirements no test
// claims" contains "## Requirements".
func between(text, from, to string) string {
	i := strings.Index(text, "\n"+from)
	if i < 0 {
		return ""
	}
	rest := text[i+1:]
	if to == "" {
		return rest
	}
	if j := strings.Index(rest, "\n"+to); j > 0 {
		return rest[:j]
	}
	return rest
}

func appendTo(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, rel)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, []byte(body)...), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The interface catalogue is the table somebody assembles by hand before every
// review and gets wrong, because the information lives at both ends of each
// channel and in neither place as a list.
func TestDocumentCarriesTheInterfaceCatalogue(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/bare", "./...")

	for _, want := range []string{
		"## The boundary",
		"| Vertrieb | actor |",
		"| Dateiablage | foreign system |",
		"Angebotsablage",
		// The four descriptive columns are mandatory in the model, so a blank
		// cell cannot reach this page.
		"| Dateisystem | Abgelegte Angebote als aktueller Stand | Rechte des Prozessbenutzers | entfällt, lokal |",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the document:\n%s", want, out)
		}
	}
}

// A course of business is rendered as its edges rather than as a numbered
// sequence. Where a process branches and comes back, any numbering is a lie
// about which step follows which.
func TestDocumentCarriesTheCoursesOfBusiness(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/example", "./...")

	for _, want := range []string{
		"## Courses of business",
		"### Angebot bis zur Entscheidung",
		"Answers to: R-QUOTE-SUBMIT, R-QUOTE-APPROVE",
		// The step names the construct it performs, not a caption.
		"| SubmitQuote _(activity)_ | QuoteSubmitted _(event raised)_ |",
		// The jump backwards survives into the table.
		"| pruefen _(choice)_ | SubmitQuote _(activity)_ | nachzubessern |",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the document:\n%s", want, out)
		}
	}
	if strings.Contains(out, "| 1 |") {
		t.Error("the graph was numbered as if it were a sequence")
	}
}

// TestDocumentSaysWhatIsNotDeclared reverses an earlier decision, deliberately.
//
// This used to require the opposite: a project declaring neither got neither
// section, "rather than an empty heading to wonder about". The objection was
// sound and aimed at a bare heading, which says nothing at all. It is answered
// by not leaving one — the heading states which kind of absence this is — and
// the old behaviour turned out to have a cost the objection did not foresee.
//
// A frontend that cannot read a chapter produced the same silence as a project
// that declared none, so the JVM specification came out with no boundary, no
// processes and no surface, reading exactly like a system that has none of
// those. Nothing anywhere said otherwise.
func TestDocumentSaysWhatIsNotDeclared(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/example", "./...")
	if !strings.Contains(out, "## The boundary\n\n_No topology is declared") {
		t.Errorf("a project without a topology must be told so:\n%s", out)
	}

	out, _ = runSpeclink(t, "generate", "../../testdata/bare", "./...")
	if !strings.Contains(out, "## Courses of business\n\n_No course of business is declared") {
		t.Errorf("a project without processes must be told so:\n%s", out)
	}
}

// A source document is not a backdrop: it is what somebody wrote before any of
// this existed, and how much of it became a requirement — and how much of that
// a person has since read — is what a specification is usually least able to
// answer. Every column of this table has been in the lock all along.
func TestDocumentCarriesTheMaterialItCameFrom(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/bare", "./...")

	if !strings.Contains(out, "## The material") {
		t.Fatalf("the material table is missing:\n%s", out)
	}
	// Four sections, all cited, none read: nobody is recorded as a reviewer in
	// this fixture, and the table says so rather than leaving the column out.
	if !strings.Contains(out, "| `requirements/_sources/sales/quoteflow.md` | markdown | 4 | 4 | 0 | 0 |") {
		t.Errorf("the row for the Markdown source is wrong:\n%s", out)
	}
	if !strings.Contains(out, "| `requirements/_sources/vorgaben.standard.json` | standard | 2 | 2 | 0 | 0 |") {
		t.Errorf("the row for the standard is wrong:\n%s", out)
	}
}

// Recording a review moves the column, which is the whole point of reading it
// back out of the lock.
func TestReadingIsCountedOnceItIsRecorded(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	if out, code := runSpeclink(t, "freeze", dir, "-reviewer", "TS", "./..."); code != 0 {
		t.Fatalf("freeze failed with %d:\n%s", code, out)
	}

	out, _ := runSpeclink(t, "generate", dir, "./...")
	if !strings.Contains(out, "| `requirements/_sources/sales/quoteflow.md` | markdown | 4 | 4 | 4 | 0 |") {
		t.Errorf("a recorded review did not reach the table:\n%s", out)
	}
}
