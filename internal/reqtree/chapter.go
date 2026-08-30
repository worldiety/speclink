package reqtree

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/doc"
	"github.com/worldiety/speclink/internal/ir"
)

// RuleChapterDuplicate fires when two prose chapters claim one ID.
const RuleChapterDuplicate = "K22-CHAPTER-DUPLICATE"

// RuleChapterDocMissing fires when the prose of a chapter cannot be read.
const RuleChapterDocMissing = "K22-CHAPTER-DOC-MISSING"

// RuleChapterUntitled fires when the prose of a chapter opens with no heading.
const RuleChapterUntitled = "K22-CHAPTER-UNTITLED"

// ResolveChapters indexes the prose chapters and checks that each one can
// actually be set.
//
// # Why the file is opened here and not when the document is written
//
// This is the whole reason the outline is a declaration rather than a list in
// a configuration file. A chapter naming a file that was moved or deleted is
// caught while the specification is checked, at the line that names it. Left
// until the document is written it would be found by whoever next reads the
// document — and a chapter that is missing looks exactly like a chapter nobody
// ever wrote.
func (t *Tree) ResolveChapters(chapters []*ir.Chapter, out *diag.Set) {
	t.chapters = nil
	seen := make(map[string]*ir.Chapter, len(chapters))

	for _, c := range chapters {
		if c.ID == "" || c.Doc == "" || c.At == 0 {
			continue // already reported as incomplete where it was read
		}
		if first, dup := seen[c.ID]; dup {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 180),
				Pos:  c.Pos,
				Rule: RuleChapterDuplicate,
				What: "chapter " + strconvQuote(c.ID) + " is declared twice.",
				Why:  "The ID anchors the chapter and appears in every diagnostic about it. Two declarations under one name make both ambiguous, and a reference to either lands wherever the document happened to be assembled.",
				How:  "Rename one of them; the other is at " + first.Pos.String() + ".",
			})
			continue
		}
		seen[c.ID] = c

		blocks, err := doc.LoadMarkdown(filepath.Join(t.root, filepath.FromSlash(c.Doc)), 2)
		if err != nil {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 181),
				Pos:  c.Pos,
				Rule: RuleChapterDocMissing,
				What: "the prose of chapter " + strconvQuote(c.ID) + " cannot be read: " + readErr(err) + ".",
				Why:  "The outline names a file that is not there. The document would be assembled with a chapter silently left out, which reads as a part of the system nobody described rather than as a link somebody broke.",
				How:  "Correct the path in Doc, or restore the file.",
			})
			continue
		}
		if doc.FirstHeading(blocks) == "" {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 182),
				Pos:  c.Pos,
				Rule: RuleChapterUntitled,
				What: "the prose of chapter " + strconvQuote(c.ID) + " does not open with a heading.",
				Why:  "The heading of the file is the title of the chapter: it is what stands in the table of contents and what a cross reference shows. Taking a title from the declaration instead would write one fact in two places, and the copy is the one that goes stale.",
				How:  "Begin " + c.Doc + " with a heading.",
			})
			continue
		}
		t.chapters = append(t.chapters, c)
	}

	// Ordered by place, then by where they are declared. Declaration order is
	// stable across runs and is the order somebody reading the tree expects,
	// which is why there is no field for it: a number to maintain by hand
	// would be one more thing that can disagree with itself.
	sort.Slice(t.chapters, func(i, j int) bool {
		a, b := t.chapters[i], t.chapters[j]
		if a.At != b.At {
			return a.At < b.At
		}
		return a.Pos.Less(b.Pos)
	})
}

// Chapters returns the usable prose chapters, in the order they are set.
func (t *Tree) Chapters() []*ir.Chapter { return t.chapters }

// readErr strips the path from a file error, which the finding already names.
func readErr(err error) string {
	var pe *os.PathError
	if errors.As(err, &pe) {
		return pe.Err.Error()
	}
	return err.Error()
}
