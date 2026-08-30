package source

import (
	"os"
	"path/filepath"
)

// SegmentError is a defect of a source document itself, as opposed to a defect
// of the requirements citing it.
//
// It carries the same three fields as a diagnostic — what, why, what to do —
// because it ends up as one. The package does not import diag: segmentation is
// a reader, and a reader that formats diagnostics could not be reused by
// anything that wants the segments for another purpose.
type SegmentError struct {
	Doc  string
	Line int
	Msg  string
	Why  string
	How  string
}

func (e *SegmentError) Error() string {
	if e.Line > 0 {
		return e.Doc + ":" + itoa(e.Line) + ": " + e.Msg
	}
	return e.Doc + ": " + e.Msg
}

// Pos locates the defect.
func (e *SegmentError) Pos() Pos { return Pos{File: e.Doc, Line: e.Line} }

// Load segments one document. doc is repository relative, root is the
// repository root.
//
// A document that cannot be read is reported once, in Document.Err, rather than
// once per requirement citing it. Ten requirements pointing at a file that was
// moved is one defect, not ten.
func Load(root, doc string) Document {
	kind, ok := KindOf(doc)
	if !ok {
		return Document{Doc: doc, Err: &SegmentError{
			Doc: doc,
			Msg: "unsupported source document type",
			Why: "speclink segments Markdown by its headings, raster images by a declared region manifest, and a standard by its clauses. A type it cannot segment contributes nothing to the forward coverage while looking fully covered.",
			How: "Convert the document to Markdown, supply it as PNG or JPEG with a region manifest, or as a clause catalogue named " + StandardSuffix + ".",
		}}
	}

	abs := filepath.Join(root, doc)
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		return Document{Doc: doc, Kind: kind, Err: &SegmentError{
			Doc: doc,
			Msg: "source document does not exist",
			Why: "The path is repository relative and must name a file. A reference into nothing is worse than no reference, because it looks verified.",
			How: "Correct the path, or add the document to the repository.",
		}}
	}

	var (
		segs  []Segment
		errs  []error
		title string
	)
	switch kind {
	case KindMarkdown:
		data, readErr := os.ReadFile(abs)
		if readErr != nil {
			return Document{Doc: doc, Kind: kind, Err: &SegmentError{
				Doc: doc,
				Msg: "source document cannot be read: " + readErr.Error(),
				How: "Check the file permissions.",
			}}
		}
		segs, errs = segmentMarkdown(doc, string(data))
	case KindImage:
		segs, errs = segmentImage(doc, abs)
	case KindStandard:
		title, segs, errs = segmentStandard(doc, abs)
	}

	d := Document{Doc: doc, Kind: kind, Title: title, Segments: segs}
	if len(errs) > 0 {
		d.Err = errs[0]
		d.More = errs[1:]
	}
	return d
}

// Set is the collection of source documents a run has looked at.
//
// Documents are loaded once and cached. A document is normally cited by many
// requirements, and both segmenting a large Markdown file and decoding an image
// are expensive enough that doing it per citation would be noticeable.
type Set struct {
	root string
	docs map[string]Document
}

// NewSet returns an empty set rooted at the repository root.
func NewSet(root string) *Set {
	return &Set{root: root, docs: map[string]Document{}}
}

// Get segments a document, or returns the cached result.
func (s *Set) Get(doc string) Document {
	if d, ok := s.docs[doc]; ok {
		return d
	}
	d := Load(s.root, doc)
	s.docs[doc] = d
	return d
}

// Loaded returns the documents seen so far, ordered by path.
func (s *Set) Loaded() []Document {
	out := make([]Document, 0, len(s.docs))
	for _, doc := range sortedKeys(s.docs) {
		out = append(out, s.docs[doc])
	}
	return out
}
