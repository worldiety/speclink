package reqtree

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/source"
)

// The outer edge is the only step in the chain without formal semantics. These
// tests pin what it now guarantees.

func TestUnsourcedNormativeRequirement(t *testing.T) {
	root := t.TempDir()
	got := checkSources(t, root, req("R-QUOTE-X", "m/q.X", ir.Functional, ir.Normative))

	if !bytes.Contains(got, []byte("names no source")) {
		t.Errorf("a requirement no document asked for was not reported:\n%s", got)
	}
}

// Only normative requirements owe a source. An abstract node is a pure
// derivation point and has nothing of its own to trace back.
func TestUnsourcedAbstractRequirementIsFine(t *testing.T) {
	root := t.TempDir()
	got := checkSources(t, root, req("R-DEC-BASE", "m/d.Base", ir.Decision, ir.Abstract))

	if len(got) != 0 {
		t.Errorf("abstract requirement reported:\n%s", got)
	}
}

func TestAnchorResolvesAgainstMarkdownSegments(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "src/flow.md", "# Abgabe\n\nText.\n\n# Versand\n\nText.\n")

	got := checkSources(t, root, sourced(
		req("R-QUOTE-X", "m/q.X", ir.Functional, ir.Normative),
		ir.Source{Doc: "src/flow.md", Anchor: "abgabe"},
	))
	if len(got) != 0 {
		t.Errorf("valid anchor reported:\n%s", got)
	}
}

func TestStaleAnchorIsReported(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "src/flow.md", "# Abgabe\n\nText.\n")

	got := checkSources(t, root, sourced(
		req("R-QUOTE-X", "m/q.X", ir.Functional, ir.Normative),
		ir.Source{Doc: "src/flow.md", Anchor: "versand"},
	))
	if !bytes.Contains(got, []byte("does not exist in")) {
		t.Errorf("stale anchor not reported:\n%s", got)
	}
	// The message has to carry the alternatives, or acting on it means opening
	// the document.
	if !bytes.Contains(got, []byte("abgabe")) {
		t.Errorf("finding does not suggest the existing anchors:\n%s", got)
	}
}

// A whole document is not an origin. Without an anchor nothing can be said
// about which part of it a requirement came from, and neither drift nor
// completeness can be checked.
func TestCitationWithoutAnchorIsReported(t *testing.T) {
	root := t.TempDir()
	writeDoc(t, root, "src/flow.md", "# Abgabe\n\nText.\n")

	got := checkSources(t, root, sourced(
		req("R-QUOTE-X", "m/q.X", ir.Functional, ir.Normative),
		ir.Source{Doc: "src/flow.md"},
	))
	if !bytes.Contains(got, []byte("names no anchor")) {
		t.Errorf("unanchored citation not reported:\n%s", got)
	}
}

// An image region is an anchor like any other. Until now an anchor on an image
// was rejected outright and the location had to go into Note, which was free
// text and therefore the last unverified reference in the chain.
func TestImageRegionIsAnAnchor(t *testing.T) {
	root := t.TempDir()
	writeMockup(t, root, "src/screen.png", source.Region{
		Name: "abgabe-knopf",
		Rect: source.Rect{X: 0, Y: 0, W: 10, H: 10},
	})

	got := checkSources(t, root, sourced(
		req("R-QUOTE-X", "m/q.X", ir.Functional, ir.Normative),
		ir.Source{Doc: "src/screen.png", Anchor: "abgabe-knopf"},
	))
	if len(got) != 0 {
		t.Errorf("valid image region reported:\n%s", got)
	}
}

func TestStaleImageRegionIsReported(t *testing.T) {
	root := t.TempDir()
	writeMockup(t, root, "src/screen.png", source.Region{
		Name: "abgabe-knopf",
		Rect: source.Rect{X: 0, Y: 0, W: 10, H: 10},
	})

	got := checkSources(t, root, sourced(
		req("R-QUOTE-X", "m/q.X", ir.Functional, ir.Normative),
		ir.Source{Doc: "src/screen.png", Anchor: "versand-knopf"},
	))
	if !bytes.Contains(got, []byte("region")) {
		t.Errorf("the finding does not explain that an image anchor is a region:\n%s", got)
	}
}

// Ten requirements pointing at a file that was moved is one defect, not ten.
func TestMissingDocumentIsReportedOncePerDocument(t *testing.T) {
	root := t.TempDir()
	docs := source.NewSet(root)
	out := &diag.Set{}

	reqs := []*ir.Requirement{
		sourced(req("R-A", "m/q.A", ir.Functional, ir.Normative), ir.Source{Doc: "src/weg.md", Anchor: "a"}),
		sourced(req("R-B", "m/q.B", ir.Functional, ir.Normative), ir.Source{Doc: "src/weg.md", Anchor: "b"}),
		sourced(req("R-C", "m/q.C", ir.Functional, ir.Normative), ir.Source{Doc: "src/weg.md", Anchor: "c"}),
	}
	tree := Build(root, reqs, out)
	tree.CheckSources(docs, out)
	ReportDocuments(docs, out)

	if out.Len() != 1 {
		t.Fatalf("got %d findings, want 1:\n%s", out.Len(), render(t, out))
	}
}

// Extern names laws and standards that have no document here. Nothing about it
// is verifiable and nothing is claimed to be.
func TestExternNeedsNoAnchor(t *testing.T) {
	root := t.TempDir()
	got := checkSources(t, root, sourced(
		req("R-NFR-X", "m/n.X", ir.NonFunctional, ir.Normative),
		ir.Source{Extern: "GoBD Rz. 36"},
	))
	if len(got) != 0 {
		t.Errorf("external source reported:\n%s", got)
	}
}

func checkSources(t *testing.T, root string, reqs ...*ir.Requirement) []byte {
	t.Helper()

	out := &diag.Set{}
	docs := source.NewSet(root)
	tree := Build(root, reqs, out)
	tree.CheckSources(docs, out)
	ReportDocuments(docs, out)
	return render(t, out)
}

func sourced(r *ir.Requirement, s ...ir.Source) *ir.Requirement {
	for i := range s {
		s[i].Pos = r.Pos
	}
	r.Sources = s
	return r
}

func writeDoc(t *testing.T, root, rel, body string) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeMockup(t *testing.T, root, rel string, regions ...source.Region) {
	t.Helper()
	abs := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		t.Fatal(err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := 0; y < 20; y++ {
		for x := 0; x < 20; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 12), G: uint8(y * 12), A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(abs, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}

	manifest, err := json.Marshal(source.Manifest{Version: source.ManifestVersion, Regions: regions})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(source.ManifestPath(abs), manifest, 0o644); err != nil {
		t.Fatal(err)
	}
}
