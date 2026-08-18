// Package reqtree assembles the requirement tree and checks everything about it
// that the Go compiler cannot: identity, layout consistency, the derivation
// graph and the outer edge to the raw sources.
//
// The outer edge is the point of it. Inside a specification model references are
// usually typed and checked; the step out to the original requirement document
// is traditionally a free text note — the single most expensive manual step is
// the only one without verification (error pattern M2). Here it is a compile
// time dependency plus the checks below.
package reqtree

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// Tree is the resolved requirement graph.
type Tree struct {
	// ByID holds every requirement, keyed by its ID.
	ByID map[string]*ir.Requirement
	// byGoIdent maps the qualified Go identifier to the requirement, used to
	// resolve DerivedFrom and Supersedes in the second pass.
	byGoIdent map[string]*ir.Requirement
	// root is the repository root against which Source.Doc is resolved.
	root string
}

// Build performs the second pass: it indexes the collected declarations and
// rewrites the Go identifier references of DerivedFrom and Supersedes into
// requirement IDs.
//
// Collecting first and resolving afterwards is what makes forward references
// legal and the input order irrelevant. No requirement has to be declared
// before it is referenced.
func Build(root string, reqs []*ir.Requirement, out *diag.Set) *Tree {
	t := &Tree{
		ByID:      make(map[string]*ir.Requirement, len(reqs)),
		byGoIdent: make(map[string]*ir.Requirement, len(reqs)),
		root:      root,
	}

	for _, r := range reqs {
		if r.ID == "" {
			continue // already reported when the declaration was read
		}
		if prev, dup := t.ByID[r.ID]; dup {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseResolve, 10),
				Pos:  r.Pos,
				What: "requirement ID " + r.ID + " is declared twice.",
				Why:  "The ID is the stable identity of a requirement; two declarations make every report ambiguous. First declaration: " + prev.Pos.String() + ".",
				How:  "Rename one of them, or delete the duplicate. If one supersedes the other, give the successor a new ID and list the predecessor in Supersedes.",
			})
			continue
		}
		t.ByID[r.ID] = r
		t.byGoIdent[r.GoIdent] = r
	}

	for _, r := range t.ByID {
		r.DerivedFrom = t.resolveRefs(r, r.DerivedFrom, "DerivedFrom", out)
		r.Supersedes = t.resolveRefs(r, r.Supersedes, "Supersedes", out)
	}

	t.checkCycles(out)
	return t
}

// resolveRefs turns qualified Go identifiers into requirement IDs.
func (t *Tree) resolveRefs(from *ir.Requirement, refs []string, field string, out *diag.Set) []string {
	if len(refs) == 0 {
		return nil
	}
	ids := make([]string, 0, len(refs))
	for _, ref := range refs {
		target, ok := t.byGoIdent[ref]
		if !ok {
			// Unreachable for a well formed build: an unknown identifier is a
			// Go compile error long before speclink runs. It can only happen
			// when the referenced package was not part of the load.
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseResolve, 11),
				Pos:  from.Pos,
				What: field + " of " + from.ID + " references " + ref + ", which is not a known requirement.",
				Why:  "Every derivation must point at a declared requirement, otherwise the graph has dangling edges.",
				How:  "Include the declaring package in the analysed patterns, or remove the reference.",
			})
			continue
		}
		if target.ID == from.ID {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseResolve, 12),
				Pos:  from.Pos,
				What: from.ID + " lists itself in " + field + ".",
				Why:  "A requirement cannot be derived from itself.",
				How:  "Remove the self reference.",
			})
			continue
		}
		ids = append(ids, target.ID)
	}
	return ids
}

// checkCycles rejects cycles in the derivation graph. DerivedFrom spans a
// directed acyclic graph; the directory tree is merely storage order.
func (t *Tree) checkCycles(out *diag.Set) {
	const (
		white = 0
		grey  = 1
		black = 2
	)
	state := make(map[string]int, len(t.ByID))

	var walk func(id string, path []string) bool
	walk = func(id string, path []string) bool {
		switch state[id] {
		case grey:
			r := t.ByID[id]
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseResolve, 13),
				Pos:  r.Pos,
				What: "cycle in DerivedFrom: " + strings.Join(append(path, id), " -> ") + ".",
				Why:  "Derivation must form a directed acyclic graph; a cycle means no requirement in it can be justified.",
				How:  "Break the cycle by removing one DerivedFrom edge, or introduce a common ancestor both derive from.",
			})
			return true
		case black:
			return false
		}
		state[id] = grey
		for _, next := range t.ByID[id].DerivedFrom {
			if _, ok := t.ByID[next]; !ok {
				continue
			}
			if walk(next, append(path, id)) {
				break
			}
		}
		state[id] = black
		return false
	}

	for _, id := range t.sortedIDs() {
		if state[id] == white {
			walk(id, nil)
		}
	}
}

// sortedIDs returns the requirement IDs in a stable order, so diagnostics do
// not depend on map iteration order.
func (t *Tree) sortedIDs() []string {
	ids := make([]string, 0, len(t.ByID))
	for id := range t.ByID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// CheckSources verifies the outer edge: that every Doc exists and every Anchor
// resolves, and that exactly one of Doc and Extern is given.
func (t *Tree) CheckSources(out *diag.Set) {
	for _, id := range t.sortedIDs() {
		r := t.ByID[id]
		if len(r.Sources) == 0 && r.Status.MustBeCovered() {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseResolve, 20),
				Pos:  r.Pos,
				What: "normative requirement " + r.ID + " names no source.",
				Why:  "Without a source the requirement cannot be traced back to what was actually asked for. This is the outer edge that is traditionally free text and unverified.",
				How:  "Add a Source with Doc pointing into requirements/_sources/, or Extern naming the law or standard.",
			})
		}
		for _, s := range r.Sources {
			t.checkSource(r, s, out)
		}
	}
}

func (t *Tree) checkSource(r *ir.Requirement, s ir.Source, out *diag.Set) {
	switch {
	case s.Doc == "" && s.Extern == "":
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 21),
			Pos:  s.Pos,
			What: "source of " + r.ID + " names neither Doc nor Extern.",
			Why:  "A source must point somewhere; an empty one is exactly the unverified free text reference this design removes.",
			How:  "Set Doc to a path below requirements/_sources/, or Extern to the law or standard.",
		})
		return
	case s.Doc != "" && s.Extern != "":
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 22),
			Pos:  s.Pos,
			What: "source of " + r.ID + " sets both Doc and Extern.",
			Why:  "Exactly one of them applies: either the origin is a document in this repository, or it is external.",
			How:  "Split it into two Source entries.",
		})
		return
	case s.Extern != "":
		return // nothing to verify, deliberately
	}

	abs := filepath.Join(t.root, s.Doc)
	info, err := os.Stat(abs)
	if err != nil || info.IsDir() {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 23),
			Pos:  s.Pos,
			What: "source document " + s.Doc + " of " + r.ID + " does not exist.",
			Why:  "The path is repository relative and must name a file. A reference into nothing is worse than none, because it looks verified.",
			How:  "Correct the path, or move the document into requirements/_sources/.",
		})
		return
	}
	if s.Anchor == "" {
		return
	}
	t.checkAnchor(r, s, abs, out)
}

// checkAnchor verifies that a heading with the given slug exists.
func (t *Tree) checkAnchor(r *ir.Requirement, s ir.Source, abs string, out *diag.Set) {
	if !strings.EqualFold(filepath.Ext(abs), ".md") {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 24),
			Pos:  s.Pos,
			What: "source of " + r.ID + " sets an Anchor on a non text document.",
			Why:  "Anchors are headings of a Markdown document; there is no equivalent in an image or PDF.",
			How:  "Remove the Anchor and describe the location in Note instead. Note is not verifiable, which is the accepted residual gap.",
		})
		return
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		return // already reported by the existence check
	}
	slugs := Headings(string(data))
	for _, got := range slugs {
		if got == s.Anchor {
			return
		}
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseResolve, 25),
		Pos:  s.Pos,
		What: "anchor " + s.Anchor + " does not exist in " + s.Doc + ".",
		Why:  "An anchor is the slug of a heading in the target document. A stale anchor points at a section that has been renamed or removed.",
		How:  "Use one of the existing headings: " + suggest(slugs, s.Anchor) + ".",
	})
}

// suggest offers the closest few anchors, so the message can be acted upon
// without opening the document.
func suggest(slugs []string, want string) string {
	if len(slugs) == 0 {
		return "the document has no headings"
	}
	var near []string
	for _, s := range slugs {
		if strings.HasPrefix(s, firstSegment(want)) {
			near = append(near, s)
		}
	}
	if len(near) == 0 {
		near = slugs
	}
	if len(near) > 5 {
		near = near[:5]
	}
	return strings.Join(near, ", ")
}

func firstSegment(s string) string {
	if i := strings.IndexByte(s, '-'); i > 0 {
		return s[:i]
	}
	return s
}

// ByGoIdent resolves a qualified Go identifier to its requirement.
//
// Bindings record what the author wrote — a Go identifier — rather than an ID
// string, so this is the lookup that turns a reference into a requirement.
func (t *Tree) ByGoIdent(ident string) *ir.Requirement { return t.byGoIdent[ident] }

// All returns every requirement, ordered by ID.
func (t *Tree) All() []*ir.Requirement {
	out := make([]*ir.Requirement, 0, len(t.ByID))
	for _, id := range t.sortedIDs() {
		out = append(out, t.ByID[id])
	}
	return out
}
