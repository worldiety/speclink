package golang

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// ReadChapters extracts the prose chapter declarations of the requirement tree.
//
// They sit beside the requirements for the same reason themes do: a chapter is
// part of how the specification is presented, nothing outside the tree refers
// to it, and a file suffix of its own would be a convention to learn for three
// lines of declaration.
func (p *Package) ReadChapters(out *diag.Set) []*ir.Chapter {
	var chapters []*ir.Chapter
	for _, f := range p.requirementFiles {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, s := range gd.Specs {
				vs, ok := s.(*ast.ValueSpec)
				if !ok || len(vs.Names) != 1 || len(vs.Values) != 1 {
					continue
				}
				lit, ok := vs.Values[0].(*ast.CompositeLit)
				if !ok || !p.isSpecType(lit, "Chapter") {
					continue
				}
				chapters = append(chapters, p.readChapter(vs, lit, out))
			}
		}
	}
	return chapters
}

func (p *Package) readChapter(vs *ast.ValueSpec, lit *ast.CompositeLit, out *diag.Set) *ir.Chapter {
	c := &ir.Chapter{
		GoIdent: p.PkgPath() + "." + vs.Names[0].Name,
		Pos:     p.pos(vs.Pos()),
	}
	var sawAt bool
	for _, el := range lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch key.Name {
		case "ID":
			c.ID, _ = p.stringArg(kv.Value)
		case "Doc":
			c.Doc, _ = p.stringArg(kv.Value)
		case "At":
			sawAt = true
			c.At = p.readPlace(kv.Value, out)
		}
	}
	p.checkChapterComplete(c, sawAt, out)
	return c
}

// checkChapterComplete reports a chapter that leaves out what it needs.
//
// One rule for all three fields rather than three, because the consequence is
// the same in every case: the chapter cannot be placed, and reporting it as
// three separate defects would say three times what somebody has to fix once.
func (p *Package) checkChapterComplete(c *ir.Chapter, sawAt bool, out *diag.Set) {
	var missing []string
	if c.ID == "" {
		missing = append(missing, "ID")
	}
	if c.Doc == "" {
		missing = append(missing, "Doc")
	}
	if !sawAt {
		missing = append(missing, "At")
	}
	if len(missing) == 0 {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseResolve, 61),
		Pos:  c.Pos,
		Rule: RuleChapterIncomplete,
		What: "this chapter leaves out " + strings.Join(missing, " and ") + ".",
		Why:  "A chapter needs a name to be referred to by, a file to take its prose from, and a place to stand in. Without all three it is a declaration that produces nothing, and an outline that silently omits a chapter reads exactly like an outline nobody wrote one for.",
		How:  "Give it an ID, the repository relative path of its Markdown file, and one of the spec.Place constants.",
	})
}

// readPlace resolves a spec.Place constant by the name it is written with.
//
// A computed value is refused rather than guessed at. The set of places is
// closed and small, and a document outline assembled at run time is a document
// nobody can read by reading the source.
func (p *Package) readPlace(e ast.Expr, out *diag.Set) ir.Place {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 60),
			Pos:  p.pos(e.Pos()),
			Rule: RuleChapterIncomplete,
			What: "the place of this chapter is not a spec.Place constant.",
			Why:  "The outline of the document has to be readable from the source. A place worked out at run time cannot be checked, and the chapter would land wherever the computation happened to put it.",
			How:  "Write one of the spec.Place constants, for example spec.BeforeArchitecture.",
		})
		return 0
	}
	place, ok := ir.PlaceOf(sel.Sel.Name)
	if !ok {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 60),
			Pos:  p.pos(e.Pos()),
			Rule: RuleChapterIncomplete,
			What: "spec." + sel.Sel.Name + " is not a place a chapter can go.",
			Why:  "The places are a closed set, one for each point of the generated document a chapter can be slotted into.",
			How:  "Use one of spec.Beginning, spec.BeforeArchitecture, spec.BeforeComposition, spec.BeforeBoundary, spec.BeforeSurface, spec.BeforeProcesses, spec.BeforeRequirements or spec.Appendix.",
		})
		return 0
	}
	return place
}

// RuleChapterIncomplete fires when a chapter cannot be placed in the document.
const RuleChapterIncomplete = "K22-CHAPTER-INCOMPLETE"
