package doc_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/worldiety/speclink/internal/doc"
)

// sample is a document using every construct the specification uses, so that a
// backend cannot pass by handling only the easy ones.
func sample() *doc.Doc {
	d := doc.New("Specification")
	d.H(2, "Where it stands")
	d.Table("", "measured", "complete").
		Aligned(doc.Left, doc.Right, doc.Right).
		Add(doc.Cell(doc.T("Normative requirements covered")), doc.Cell(doc.T("8")), doc.Cell(doc.T("100%")))
	d.HID(3, "R-1", "R-1 — Ledger balance")
	d.P(doc.T("Balances "), doc.Emph("must"), doc.T(" reconcile, see "), doc.Code("erp.Ledger"), doc.T("."))
	d.Note(doc.T("Not measured: this frontend reads no topology declarations."))
	d.Bullets(
		doc.Item(doc.Strong("Read by"), doc.T(" a reviewer")),
		doc.Item(doc.T("See "), doc.Ref{ID: "R-1", Text: "R-1"}),
	)
	d.P(doc.Link{Text: "the source", URL: "https://example.invalid/spec.md"})
	return d
}

// TestBothBackendsRenderEveryConstruct is the guard that keeps the two outputs
// from being two documents.
//
// A renderer gets a tree and no access to the model that produced it, so it
// cannot state a fact the other one does not. What it can still do is drop
// one silently, which is what this catches.
func TestBothBackendsRenderEveryConstruct(t *testing.T) {
	t.Parallel()
	for _, r := range []doc.Renderer{doc.Markdown{}, doc.Typst{}} {
		out := r.Render(sample())
		for _, want := range []string{
			"Where it stands", "measured", "Normative requirements covered",
			"Ledger balance", "reconcile", "erp.Ledger",
			"Not measured", "Read by", "a reviewer", "the source",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("%s dropped %q:\n%s", r.Ext(), want, out)
			}
		}
	}
}

// TestMarkdownEscapesText is about requirement text this tool does not own.
//
// The words come from a document somebody else writes, and an underscore in an
// identifier or a pipe in a sentence has to arrive in a table cell as itself.
// An unescaped pipe does not corrupt the sentence, it silently adds a column
// and shifts every value in the row under the wrong heading, which is the kind
// of wrong a reader has no way to detect.
func TestMarkdownEscapesText(t *testing.T) {
	t.Parallel()
	d := doc.New("t")
	d.Table("a", "b").Add(doc.Cell(doc.T("total|net")), doc.Cell(doc.T("snake_case_name")))
	out := doc.Markdown{}.Render(d)

	if strings.Contains(out, "total|net") {
		t.Errorf("a pipe in text was left to split the row:\n%s", out)
	}
	if strings.Contains(out, "snake_case_name") {
		t.Errorf("underscores in an identifier were left as emphasis:\n%s", out)
	}
	// The row still has exactly the two columns it was given. Only an
	// unescaped pipe is a separator, which is the whole point of the escape.
	for _, l := range strings.Split(out, "\n") {
		if !strings.HasPrefix(l, "| total") {
			continue
		}
		if sep := strings.Count(l, "|") - strings.Count(l, `\|`); sep != 3 {
			t.Errorf("the row has %d separators, want 3: %q", sep, l)
		}
	}
}

// TestCodeIsNotEscaped separates the two kinds of string.
//
// Code is spelled exactly and both formats already treat the span as literal.
// Escaping inside it would put backslashes on the page, which is how an
// identifier in a specification stops matching the identifier in the source.
func TestCodeIsNotEscaped(t *testing.T) {
	t.Parallel()
	d := doc.New("t")
	d.P(doc.Code("map[string]*ir.Endpoint"))
	for _, r := range []doc.Renderer{doc.Markdown{}, doc.Typst{}} {
		if out := r.Render(d); !strings.Contains(out, "`map[string]*ir.Endpoint`") {
			t.Errorf("%s mangled a symbol:\n%s", r.Ext(), out)
		}
	}
}

// TestTypstCompiles is the only test here that proves anything about the PDF.
//
// A renderer that emits plausible looking Typst and does not compile is worth
// nothing, and no amount of string assertions in Go would find the missing
// bracket. Skipped rather than failed when the tool is absent: speclink writes
// and never runs, so Typst is a prerequisite of the environment that renders
// the document, not of the one that builds this tool.
func TestTypstCompiles(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst is not installed; the emitted document cannot be checked here")
	}
	dir := t.TempDir()
	src := filepath.Join(dir, "spec.typ")
	if err := os.WriteFile(src, []byte(doc.Typst{Author: "worldiety"}.Render(sample())), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("typst", "compile", src, filepath.Join(dir, "spec.pdf")).CombinedOutput()
	if err != nil {
		t.Fatalf("the generated document does not compile: %v\n%s", err, out)
	}
}

// TestTypstRefusesADanglingReference is why this backend exists next to
// Markdown at all.
//
// An audit document cites a requirement from a process and a chapter from
// another chapter. Markdown can only emit an anchor, and an anchor that lands
// nowhere reads exactly like one that lands: the reader follows it, arrives at
// the top of the page, and assumes they misclicked. Here it is a build error,
// so a citation to a requirement that was deleted cannot reach a reviewer.
func TestTypstRefusesADanglingReference(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst is not installed")
	}
	d := doc.New("t")
	d.HID(2, "R-1", "Real")
	d.P(doc.T("See "), doc.Ref{ID: "R-99", Text: "R-99"})

	dir := t.TempDir()
	src := filepath.Join(dir, "spec.typ")
	if err := os.WriteFile(src, []byte(doc.Typst{}.Render(d)), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("typst", "compile", src, filepath.Join(dir, "spec.pdf")).CombinedOutput()
	if err == nil {
		t.Fatal("a reference to a requirement that does not exist compiled cleanly")
	}
	if !strings.Contains(string(out), "does not exist") {
		t.Errorf("the build failed for some other reason:\n%s", out)
	}
}

// TestEmptyListAddsNothing keeps the guard in one place.
//
// Chapters are assembled by appending whatever the model produced, and most of
// them can legitimately produce nothing. If a caller had to check first, one
// of them eventually would not, and a stray bullet under a heading is the sort
// of thing that survives review forever.
func TestEmptyListAddsNothing(t *testing.T) {
	t.Parallel()
	d := doc.New("t")
	d.H(2, "Gaps")
	d.Bullets()
	out := doc.Markdown{}.Render(d)
	if strings.Contains(out, "-") {
		t.Errorf("an empty list left a mark:\n%q", out)
	}
}
