// Package source segments the raw requirement documents and fingerprints the
// pieces.
//
// It closes the one gap the rest of speclink leaves open. Everything from the
// requirement tree downwards is verified hard: the Go compiler resolves both
// ends of a binding, constructs are enumerated rather than expected, and every
// persisted shape is frozen against a recorded baseline. The step *into* the
// tree — a person writes a document, a model turns it into requirements — was
// checked with nothing but "does the path exist, does the heading exist".
//
// That asymmetry is the wrong way round. The step below the tree is already
// covered by the Go compiler and the test suite. The step above it is the only
// one in the chain without formal semantics, and it is the one that decides
// whether a coverage figure means anything at all.
//
// Three defects live there, none of which any other check can see:
//
//   - A requirement that no document asked for. The code implementing it is
//     correct, bound and covered; nobody wanted it.
//   - A section that produced no requirement. The tree is internally consistent
//     and completely covered; the feature is missing.
//   - A section that changed while the requirement derived from it did not.
//     The anchor still resolves, the meaning has moved.
//
// The first two are the precision and recall failures of any natural language
// extraction. The third is the stale link that the traceability literature
// treats as unsolved.
//
// The answers are the two principles speclink already runs on, applied one
// layer further out: enumerate what must be covered instead of expecting it,
// and fingerprint every edge the compiler cannot check.
//
// # Segments
//
// A source document is not an atom. It is a sequence of addressable segments,
// and a requirement cites one of them:
//
//   - In Markdown the segments are derived from the heading structure. This is
//     free — the anchor slugs were already being computed to verify references.
//   - In a raster image the segments are declared in a sidecar file, because an
//     image is not decomposable by any deterministic rule. A vision model that
//     invents regions would only move the invented-requirement problem one
//     level down.
//
// No other document types exist. PDF was considered and rejected: its text
// extraction is stable only for a pinned library version, so an upgrade would
// silently move every segment boundary at once. Converting a PDF to Markdown
// upstream is a visible, reviewable step and leaves the source diffable in the
// pull request, which is the property the whole design rests on.
package source

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// Kind is the sort of document a segment was read from.
type Kind int

const (
	// KindMarkdown is a text document segmented by its headings.
	KindMarkdown Kind = iota + 1
	// KindImage is a raster image segmented by a declared region manifest.
	KindImage
)

func (k Kind) String() string {
	switch k {
	case KindMarkdown:
		return "markdown"
	case KindImage:
		return "image"
	}
	return "unknown"
}

// Segment is one addressable piece of a source document.
//
// ID is what a requirement cites in Source.Anchor. For Markdown it is the
// heading slug, for an image the declared region name. The two are deliberately
// the same field: from the requirement side there is no difference between
// pointing at a section and pointing at a part of a mockup, and there is no
// reason for the model to invent one.
//
// Fingerprint is the hash of the segment's content, not of the file. Whole file
// hashes are useless for both source kinds: a Markdown document is edited
// constantly in places no requirement depends on, and re-exporting an image
// changes its bytes without changing the picture. A per segment hash is what
// makes the drift report specific enough to act on — one changed button reports
// the requirements of that button, not of the whole screen.
//
// Informative marks a segment that is not expected to produce a requirement:
// an introduction, a glossary, a legend. It has to be declared, never guessed.
// A tool that decides on its own which prose carries no obligation would be
// making exactly the judgement it exists to check.
type Segment struct {
	// Doc is the repository relative path of the document.
	Doc string
	// ID addresses the segment within the document.
	ID string
	// Kind is the sort of document this came from.
	Kind Kind
	// Title is the human readable name, for diagnostics.
	Title string
	// Fingerprint is the content hash, hex encoded.
	Fingerprint string
	// Informative excludes the segment from the coverage obligation.
	Informative bool
	// Pos locates the segment in its document, for diagnostics. Line is 0 for
	// image regions, which have no line.
	Pos Pos
}

// Ref is the fully qualified address of a segment.
func (s Segment) Ref() string {
	if s.ID == "" {
		return s.Doc
	}
	return s.Doc + "#" + s.ID
}

// Pos locates a segment inside its document.
type Pos struct {
	File string
	Line int
}

func (p Pos) String() string {
	if p.Line == 0 {
		return p.File
	}
	return fmt.Sprintf("%s:%d", p.File, p.Line)
}

// Document is a segmented source document.
type Document struct {
	// Doc is the repository relative path.
	Doc string
	// Kind is the sort of document.
	Kind Kind
	// Segments are in document order.
	Segments []Segment
	// Err reports why the document could not be segmented. A document that
	// cannot be read is reported once, here, rather than once per citing
	// requirement.
	Err error
	// More carries the segmentation defects beyond the first, so a document
	// with several broken regions is fixed in one pass rather than one run per
	// defect.
	More []error
}

// Errors returns every segmentation defect of the document.
func (d Document) Errors() []error {
	if d.Err == nil {
		return nil
	}
	return append([]error{d.Err}, d.More...)
}

// Segment returns the segment with the given ID.
func (d Document) Segment(id string) (Segment, bool) {
	for _, s := range d.Segments {
		if s.ID == id {
			return s, true
		}
	}
	return Segment{}, false
}

// IDs returns the segment identifiers in document order.
func (d Document) IDs() []string {
	out := make([]string, 0, len(d.Segments))
	for _, s := range d.Segments {
		out = append(out, s.ID)
	}
	return out
}

// KindOf classifies a document by its extension.
//
// Recognition is by extension alone and the set is closed. An unknown extension
// is not a source document, and saying so is better than guessing at content:
// a silently unsegmented document would contribute nothing to the forward
// coverage and look exactly like one that is fully covered.
func KindOf(path string) (Kind, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md":
		return KindMarkdown, true
	case ".png", ".jpg", ".jpeg":
		return KindImage, true
	}
	return 0, false
}

// fingerprint hashes segment content.
//
// The hash covers content only, never position or surrounding structure.
// Moving a section within a document, or renaming a sibling, must not report
// drift: the requirement derived from it is still derived from the same words.
func fingerprint(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// sortedKeys returns the map keys in a stable order, so diagnostics never
// depend on map iteration order.
func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
