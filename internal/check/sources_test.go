package check

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
	"github.com/worldiety/speclink/internal/source"
)

// The defect this rule exists for: a tree that is internally perfect while a
// section of what was asked for never became a requirement. Nothing else in
// speclink can see it — there is no requirement to report as uncovered and no
// construct to report as unbound.
func TestSkippedSectionIsReported(t *testing.T) {
	root := t.TempDir()
	write(t, root, "src/flow.md", "# Abgabe\n\nEine Nummer wird gezogen.\n\n# Versand\n\nDas Angebot geht raus.\n")

	got, cov := cover(t, root, cite("R-A", "src/flow.md", "abgabe"))

	if !bytes.Contains(got, []byte("became no requirement")) {
		t.Fatalf("skipped section not reported:\n%s", got)
	}
	if !bytes.Contains(got, []byte("versand")) {
		t.Errorf("the finding does not name the section:\n%s", got)
	}
	if cov.Total != 2 || cov.Covered != 1 {
		t.Errorf("got %d/%d covered, want 1/2", cov.Covered, cov.Total)
	}
}

// Enumeration rather than collection from citations is the whole rule. A
// document nobody cited must not look like one that is fully covered.
func TestUncitedDocumentIsStillMeasured(t *testing.T) {
	root := t.TempDir()
	write(t, root, "src/gelesen.md", "# Abgabe\n\nText.\n")
	write(t, root, "src/vergessen.md", "# Rabatt\n\nText.\n")

	got, cov := cover(t, root, cite("R-A", "src/gelesen.md", "abgabe"))

	if !bytes.Contains(got, []byte("vergessen.md")) {
		t.Fatalf("a document nobody cited was not measured:\n%s", got)
	}
	if cov.Total != 2 {
		t.Errorf("got %d segments, want 2", cov.Total)
	}
}

func TestInformativeSectionCarriesNoObligation(t *testing.T) {
	root := t.TempDir()
	write(t, root, "src/flow.md", "# Einleitung\n\n<!-- speclink:informative -->\n\nKontext.\n\n# Abgabe\n\nText.\n")

	got, cov := cover(t, root, cite("R-A", "src/flow.md", "abgabe"))

	if len(got) != 0 {
		t.Fatalf("informative section reported:\n%s", got)
	}
	if cov.Total != 1 {
		t.Errorf("got %d segments, want 1", cov.Total)
	}
}

// The undirected waiver switches the rule off, so a project can adopt the rest
// of speclink before its source documents are in shape. There is deliberately
// no per section waiver: spec.Waive attaches to a Go construct and a section
// has none, so the exemption is the marker in the document instead.
func TestUndirectedWaiverSwitchesTheRuleOff(t *testing.T) {
	root := t.TempDir()
	write(t, root, "src/flow.md", "# Beispiel\n\nEin durchgerechnetes Beispiel.\n")

	waivers := ir.Waivers{}
	waivers.Waive("", RuleSourceUncovered, "source documents are not under speclink yet")
	got, cov := coverWith(t, root, waivers)

	if len(got) != 0 {
		t.Fatalf("waived rule still reported:\n%s", got)
	}
	if cov.Covered != 1 {
		t.Errorf("a waived segment must count as accounted for")
	}
}

// An empty set is complete. A project with no source documents has nothing
// outstanding, and reporting zero percent would be an accusation rather than a
// measurement.
func TestEmptySetIsComplete(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	_, cov := cover(t, root)

	if cov.Ratio() != 1 {
		t.Errorf("got %v, want 1", cov.Ratio())
	}
}

// A document that could not be segmented has no segments to hold accountable.
// Counting it as covered would let an unreadable file hide a whole
// specification; counting its segments would be counting nothing.
func TestUnreadableDocumentDoesNotCountAsCovered(t *testing.T) {
	root := t.TempDir()
	write(t, root, "src/screen.png", "not an image")

	_, cov := cover(t, root)

	if cov.Total != 0 || cov.Covered != 0 {
		t.Errorf("got %d/%d, want 0/0", cov.Covered, cov.Total)
	}
}

func cover(t *testing.T, root string, reqs ...*ir.Requirement) ([]byte, SourceCoverage) {
	t.Helper()
	return run(t, root, nil, reqs...)
}

func coverWith(t *testing.T, root string, waivers ir.Waivers, reqs ...*ir.Requirement) ([]byte, SourceCoverage) {
	t.Helper()
	return run(t, root, waivers, reqs...)
}

func run(t *testing.T, root string, waivers ir.Waivers, reqs ...*ir.Requirement) ([]byte, SourceCoverage) {
	t.Helper()

	out := &diag.Set{}
	tree := reqtree.Build(root, reqs, out)
	set := source.NewSet(root)
	docs, _ := source.Walk(root, []string{"src"})
	cov := CoverSources(tree, set, docs, waivers, out)

	var buf bytes.Buffer
	if err := out.WriteText(&buf); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes(), cov
}

func cite(id, doc, anchor string) *ir.Requirement {
	return &ir.Requirement{
		ID:      id,
		GoIdent: "m/q." + id,
		Kind:    ir.Functional,
		Status:  ir.Normative,
		Text:    "text of " + id,
		Sources: []ir.Source{{Doc: doc, Anchor: anchor}},
		Pos:     ir.Position{File: "mem://" + id, Line: 1, Col: 1},
	}
}

func write(t *testing.T, root, rel, body string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
