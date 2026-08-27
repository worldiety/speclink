package source

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestSegmentMarkdownSplitsAtEveryHeading(t *testing.T) {
	segs, errs := segmentMarkdown("doc.md", `# Angebot

Kopftext.

## 8.1 Angebot (Kopf)

Erster Abschnitt.

### Tiefer

Dritter Abschnitt.
`)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	got := Document{Segments: segs}.IDs()
	want := []string{"angebot", "81-angebot-kopf", "tiefer"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

// Segments must not overlap. A requirement citing a parent section must not
// discharge the coverage obligation of its children, or a whole subtree could
// be swallowed by one citation.
func TestSegmentsAreDisjoint(t *testing.T) {
	parent, _ := segmentMarkdown("a.md", "## Eltern\n\nEigener Text.\n\n### Kind\n\nKindtext.\n")
	alone, _ := segmentMarkdown("b.md", "## Eltern\n\nEigener Text.\n")

	if parent[0].Fingerprint != alone[0].Fingerprint {
		t.Fatal("parent fingerprint changed when a child section was added; segments are not disjoint")
	}
}

// The preamble carries the same obligation as any other section, otherwise a
// document has exactly one place where a requirement can be written and
// legitimately produce nothing.
func TestPreambleIsASegment(t *testing.T) {
	segs, _ := segmentMarkdown("doc.md", "Text vor jeder Überschrift.\n\n# Titel\n\nRest.\n")
	if len(segs) != 2 || segs[0].ID != preambleID {
		t.Fatalf("got %v, want a preamble segment first", Document{Segments: segs}.IDs())
	}
}

func TestEmptyPreambleIsNotASegment(t *testing.T) {
	segs, _ := segmentMarkdown("doc.md", "\n\n# Titel\n\nRest.\n")
	if len(segs) != 1 || segs[0].ID != "titel" {
		t.Fatalf("got %v, want only the heading segment", Document{Segments: segs}.IDs())
	}
}

// A rule that fires on reformatting is one that gets waived by habit. Trailing
// whitespace, line ending style and the number of blank lines are invisible to
// the reader and are rewritten by editors without anybody deciding anything.
func TestReformattingIsNotDrift(t *testing.T) {
	a, _ := segmentMarkdown("doc.md", "# T\n\nErste Zeile.\n\nZweite Zeile.\n")
	b, _ := segmentMarkdown("doc.md", "# T\r\n\r\n\r\nErste Zeile.   \r\n\r\n\r\n\r\nZweite Zeile.\t\r\n\r\n")

	if a[0].Fingerprint != b[0].Fingerprint {
		t.Fatal("reformatting reported as drift")
	}
}

// Everything the reader can see must count. A hash that forgives wording is
// quietly deciding that some rewrites of a requirement source do not matter.
func TestWordingIsDrift(t *testing.T) {
	a, _ := segmentMarkdown("doc.md", "# T\n\nDer Preis MUSS gerundet werden.\n")
	b, _ := segmentMarkdown("doc.md", "# T\n\nDer Preis SOLL gerundet werden.\n")

	if a[0].Fingerprint == b[0].Fingerprint {
		t.Fatal("a changed requirement text produced the same fingerprint")
	}
}

// Moving a section within a document does not change what a requirement derived
// from it was derived from.
func TestReorderingIsNotDrift(t *testing.T) {
	a, _ := segmentMarkdown("doc.md", "# Eins\n\nA.\n\n# Zwei\n\nB.\n")
	b, _ := segmentMarkdown("doc.md", "# Zwei\n\nB.\n\n# Eins\n\nA.\n")

	first, _ := Document{Segments: a}.Segment("eins")
	second, _ := Document{Segments: b}.Segment("eins")
	if first.Fingerprint != second.Fingerprint {
		t.Fatal("reordering reported as drift")
	}
}

func TestInformativeMarker(t *testing.T) {
	segs, _ := segmentMarkdown("doc.md", "# Einleitung\n\n"+InformativeMarker+"\n\nNur Kontext.\n")
	if !segs[0].Informative {
		t.Fatal("marker not honoured")
	}
}

func TestDuplicateSlugIsRejected(t *testing.T) {
	_, errs := segmentMarkdown("doc.md", "# Angebot\n\nA.\n\n## Angebot\n\nB.\n")
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
}

func TestHeadingInCodeFenceIsNotAHeading(t *testing.T) {
	segs, _ := segmentMarkdown("doc.md", "# T\n\n```\n# kein Heading\n```\n")
	if len(segs) != 1 {
		t.Fatalf("got %v, want one segment", Document{Segments: segs}.IDs())
	}
}

// --- images ---

func writeImage(t *testing.T, dir, name string, img image.Image, quality int) string {
	t.Helper()
	path := filepath.Join(dir, name)
	var buf bytes.Buffer
	var err error
	if filepath.Ext(name) == ".png" {
		err = png.Encode(&buf, img)
	} else {
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeManifest(t *testing.T, image string, regions ...Region) {
	t.Helper()
	data, err := json.Marshal(Manifest{Version: ManifestVersion, Regions: regions})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ManifestPath(image), data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func mockup(fill color.RGBA) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 40, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 40; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 3), G: uint8(y * 3), B: 0, A: 255})
		}
	}
	for y := 20; y < 30; y++ {
		for x := 20; x < 30; x++ {
			img.Set(x, y, fill)
		}
	}
	return img
}

func TestImageRegionsAreFingerprinted(t *testing.T) {
	dir := t.TempDir()
	abs := writeImage(t, dir, "screen.png", mockup(color.RGBA{A: 255}), 0)
	writeManifest(t, abs,
		Region{Name: "kopf", Rect: Rect{X: 0, Y: 0, W: 40, H: 10}},
		Region{Name: "knopf", Rect: Rect{X: 20, Y: 20, W: 10, H: 10}},
	)

	segs, errs := segmentImage("screen.png", abs)
	if len(errs) != 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	if len(segs) != 2 || segs[0].ID != "kopf" || segs[1].ID != "knopf" {
		t.Fatalf("got %v", Document{Segments: segs}.IDs())
	}
	if segs[0].Fingerprint == segs[1].Fingerprint {
		t.Fatal("distinct regions share a fingerprint")
	}
}

// Changing one control must report the requirements of that control, not every
// requirement on the screen. A report too coarse to act on is one that gets
// ignored.
func TestDriftIsPerRegion(t *testing.T) {
	dir := t.TempDir()
	regions := []Region{
		{Name: "kopf", Rect: Rect{X: 0, Y: 0, W: 40, H: 10}},
		{Name: "knopf", Rect: Rect{X: 20, Y: 20, W: 10, H: 10}},
	}

	before := writeImage(t, dir, "a.png", mockup(color.RGBA{A: 255}), 0)
	writeManifest(t, before, regions...)
	after := writeImage(t, dir, "b.png", mockup(color.RGBA{R: 255, A: 255}), 0)
	writeManifest(t, after, regions...)

	a, _ := segmentImage("a.png", before)
	b, _ := segmentImage("b.png", after)

	if a[0].Fingerprint != b[0].Fingerprint {
		t.Error("untouched region reported as drifted")
	}
	if a[1].Fingerprint == b[1].Fingerprint {
		t.Error("changed region not reported as drifted")
	}
}

// Re-exporting the same mockup writes different bytes while the picture is
// unchanged. A file hash would fire on every export, so the fingerprint is
// taken over decoded pixels.
func TestReExportIsNotDrift(t *testing.T) {
	dir := t.TempDir()
	img := mockup(color.RGBA{A: 255})
	region := Region{Name: "knopf", Rect: Rect{X: 20, Y: 20, W: 10, H: 10}}

	first := writeImage(t, dir, "a.png", img, 0)
	writeManifest(t, first, region)

	// A second encode of the same pixels, written as a distinct file to stand
	// in for an export that differs in metadata and compression.
	second := filepath.Join(dir, "b.png")
	var buf bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if err := enc.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	writeManifest(t, second, region)

	firstBytes, _ := os.ReadFile(first)
	secondBytes, _ := os.ReadFile(second)
	if bytes.Equal(firstBytes, secondBytes) {
		t.Skip("the two encodings happened to be byte identical; the test cannot show anything")
	}

	a, _ := segmentImage("a.png", first)
	b, _ := segmentImage("b.png", second)
	if a[0].Fingerprint != b[0].Fingerprint {
		t.Fatal("re-export reported as drift")
	}
}

func TestImageWithoutManifestIsAnError(t *testing.T) {
	dir := t.TempDir()
	abs := writeImage(t, dir, "screen.png", mockup(color.RGBA{A: 255}), 0)

	_, errs := segmentImage("screen.png", abs)
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
}

func TestRegionOutsideImageIsRejected(t *testing.T) {
	dir := t.TempDir()
	abs := writeImage(t, dir, "screen.png", mockup(color.RGBA{A: 255}), 0)
	writeManifest(t, abs, Region{Name: "weg", Rect: Rect{X: 30, Y: 30, W: 40, H: 40}})

	_, errs := segmentImage("screen.png", abs)
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
}

func TestManifestVersionMismatchRefuses(t *testing.T) {
	dir := t.TempDir()
	abs := writeImage(t, dir, "screen.png", mockup(color.RGBA{A: 255}), 0)
	if err := os.WriteFile(ManifestPath(abs), []byte(`{"version":99,"regions":[]}`), 0o644); err != nil {
		t.Fatal(err)
	}

	_, errs := segmentImage("screen.png", abs)
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
}

func TestKindOf(t *testing.T) {
	for _, tc := range []struct {
		path string
		kind Kind
		ok   bool
	}{
		{"a.md", KindMarkdown, true},
		{"a.MD", KindMarkdown, true},
		{"a.png", KindImage, true},
		{"a.jpg", KindImage, true},
		{"a.jpeg", KindImage, true},
		{"a.pdf", 0, false},
		{"a.txt", 0, false},
	} {
		kind, ok := KindOf(tc.path)
		if ok != tc.ok || kind != tc.kind {
			t.Errorf("%s: got (%v, %v), want (%v, %v)", tc.path, kind, ok, tc.kind, tc.ok)
		}
	}
}

func TestLoadReportsMissingDocumentOnce(t *testing.T) {
	d := Load(t.TempDir(), "fehlt.md")
	if len(d.Errors()) != 1 {
		t.Fatalf("got %d errors, want 1", len(d.Errors()))
	}
}
