package doc_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// TestEveryMarkdownReferenceHasATarget gives Markdown the guarantee Typst gets
// from its compiler.
//
// A reference is rendered as a link to a fragment, and a fragment that lands
// nowhere is invisible: the reader clicks, the page does not move, and they
// assume they missed. Typst refuses to build that. Markdown will never refuse,
// so the check has to live here — and it is the same tree, so if the anchors
// agree once they agree for every document.
func TestEveryMarkdownReferenceHasATarget(t *testing.T) {
	t.Parallel()
	d := doc.New("t")
	d.HID(2, "R-1", "R-1 — a title somebody will later edit")
	d.P(doc.T("See "), doc.Ref{ID: "R-1", Text: "R-1"})
	out := doc.Markdown{}.Render(d)

	link := regexp.MustCompile(`\]\(#([^)]+)\)`)
	anchors := regexp.MustCompile(`<a id="([^"]+)">`)
	have := map[string]bool{}
	for _, m := range anchors.FindAllStringSubmatch(out, -1) {
		have[m[1]] = true
	}
	found := link.FindAllStringSubmatch(out, -1)
	if len(found) == 0 {
		t.Fatalf("no reference was rendered at all:\n%s", out)
	}
	for _, m := range found {
		if !have[m[1]] {
			t.Errorf("reference to #%s lands nowhere; anchors are %v\n%s", m[1], have, out)
		}
	}
}

// TestAnchorsDoNotFollowTheTitle is the specific way this used to be wrong.
//
// The fragment was derived from the heading text, so a requirement titled
// "R-1 — Ledger" answered to #r-1--ledger, and every link to it broke the day
// the title gained a word. Deriving it from the identifier instead means a
// title is prose and nothing depends on it.
func TestAnchorsDoNotFollowTheTitle(t *testing.T) {
	t.Parallel()
	one := doc.New("t")
	one.HID(2, "R-1", "R-1 — Ledger")
	two := doc.New("t")
	two.HID(2, "R-1", "R-1 — Ledger balances, reworded entirely")

	a := regexp.MustCompile(`<a id="([^"]+)">`)
	first := a.FindStringSubmatch(doc.Markdown{}.Render(one))
	second := a.FindStringSubmatch(doc.Markdown{}.Render(two))
	if first == nil || second == nil {
		t.Fatal("no anchor was emitted")
	}
	if first[1] != second[1] {
		t.Errorf("rewording a title moved its anchor: %q became %q", first[1], second[1])
	}
}

// TestAReferenceKeepsItsWords is the defect that made the register useless.
//
// Typst renders a bare reference as the number of the section it points at, so
// a table whose entire job is to name requirements came out as a column of
// "Section 15.1", "Section 15.2". Every row was a link to something the reader
// could not identify without following it — in a printed document, not at all.
//
// The words have to survive, and the guarantee has to survive with them: a
// reference that lands nowhere must still refuse to compile. Both are checked,
// because the obvious fix for the first quietly gives up the second.
func TestAReferenceKeepsItsWords(t *testing.T) {
	t.Parallel()
	d := doc.New("t")
	d.HID(2, "R-QUOTE-SUBMIT", "R-QUOTE-SUBMIT — Submitting a quote")
	d.P(doc.Ref{ID: "R-QUOTE-SUBMIT", Text: "R-QUOTE-SUBMIT"})

	for _, r := range []doc.Renderer{doc.Markdown{}, doc.Typst{}} {
		out := r.Render(d)
		if !strings.Contains(out, "R-QUOTE-SUBMIT") {
			t.Errorf("%s dropped the words of a reference:\n%s", r.Ext(), out)
		}
	}
	// Specifically: the Typst output must not fall back on the numbering.
	if out := (doc.Typst{}).Render(d); strings.Contains(out, "@req-R-QUOTE-SUBMIT") {
		t.Errorf("the reference renders as a section number rather than its words:\n%s", out)
	}
}

// TestADanglingReferenceStillFailsAfterKeepingTheWords guards the trade the
// previous test invites.
//
// Rendering a reference as plain text keeps the words and throws away the only
// thing that made a Markdown anchor better than a hope. This checks that the
// spelling chosen keeps both.
func TestADanglingReferenceStillFailsAfterKeepingTheWords(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("typst"); err != nil {
		t.Skip("typst is not installed")
	}
	d := doc.New("t")
	d.HID(2, "R-1", "R-1 — real")
	d.P(doc.Ref{ID: "R-GONE", Text: "R-GONE"})

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

// TestAWideFigureGetsThePageTurned is the layout half of the same problem.
//
// A drawing five times wider than it is tall, set at the width of a column,
// comes out a few millimetres high with labels that do not resolve in print.
// That is worse than no figure: the reader believes they have seen it. Typst
// can turn the page under it, which buys the long side of the sheet instead of
// the short one.
func TestAWideFigureGetsThePageTurned(t *testing.T) {
	t.Parallel()

	d := doc.New("m")
	d.WideFig("fig-wide", "wide.svg", "A broad drawing.")
	d.Fig("fig-normal", "normal.svg", "An ordinary one.")

	out := doc.Typst{}.Render(d)
	if !strings.Contains(out, "#page(flipped: true)[") {
		t.Errorf("the broad figure did not get the page turned:\n%s", out)
	}
	if strings.Count(out, "#page(flipped: true)[") != 1 {
		t.Errorf("the ordinary figure was turned as well:\n%s", out)
	}
	// Markdown has no pages, so there is nothing to turn and nothing to say
	// about it. The same tree renders either way.
	if md := (doc.Markdown{}).Render(d); !strings.Contains(md, "wide.svg") || strings.Contains(md, "flipped") {
		t.Errorf("markdown should carry the figure and no page instruction:\n%s", md)
	}
}
