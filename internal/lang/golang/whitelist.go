package golang

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// Phase V1: the node whitelist.
//
// The annotation language is not Go. It is an explicitly enumerated subset of
// go/ast node types.
//
// Because annotation files are real, normally compiled Go files there is no
// syntactic barrier of any kind. The whitelist is the only thing preventing
// *.annotation.go from turning into a second codebase with helpers, constants
// and eventually logic. Enforcement is load bearing, not formal.
//
// As long as ForStmt, RangeStmt and function definitions stay out, the language
// is total: no loops, no recursion, every evaluation terminates by
// construction. Whoever adds ForStmt gives that up and needs recursion limits
// and fixpoint detection. Widening the whitelist is a deliberate decision with
// a visible price.

// fileKind distinguishes the two carriers, which permit slightly different
// nodes: only requirement files may declare composite literals.
type fileKind int

const (
	kindAnnotation fileKind = iota + 1
	kindRequirement
)

func (k fileKind) String() string {
	if k == kindRequirement {
		return "*.spec.go"
	}
	return "*.annotation.go"
}

// CheckWhitelist verifies that every annotation and requirement file of the
// package contains nothing but permitted nodes.
func (p *Package) CheckWhitelist(out *diag.Set) {
	for _, f := range p.annotationFiles {
		p.checkFile(f, kindAnnotation, out)
	}
	for _, f := range p.requirementFiles {
		p.checkFile(f, kindRequirement, out)
	}
}

func (p *Package) checkFile(f *ast.File, kind fileKind, out *diag.Set) {
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			p.checkGenDecl(d, kind, out)
		case *ast.FuncDecl:
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseWhitelist, 1),
				Pos:  p.pos(d.Pos()),
				What: "function definitions are not permitted in " + kind.String() + ".",
				Why:  "These files carry binding terms only. The whitelist is the sole barrier against a second codebase growing here (P4: the language is closed and declarative).",
				How:  "Move the function into the neighbouring source file, or delete it.",
			})
		default:
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseWhitelist, 2),
				Pos:  p.pos(decl.Pos()),
				What: "unsupported declaration in " + kind.String() + ".",
				Why:  "Only import declarations and package level var declarations are permitted.",
				How:  "Remove the declaration.",
			})
		}
	}
}

func (p *Package) checkGenDecl(d *ast.GenDecl, kind fileKind, out *diag.Set) {
	switch d.Tok {
	case token.IMPORT:
		return
	case token.VAR:
		for _, s := range d.Specs {
			vs, ok := s.(*ast.ValueSpec)
			if !ok {
				continue
			}
			p.checkValueSpec(vs, kind, out)
		}
	case token.TYPE:
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseWhitelist, 3),
			Pos:  p.pos(d.Pos()),
			What: "type declarations are not permitted in " + kind.String() + ".",
			Why:  "The annotation language declares no types; it only references them.",
			How:  "Move the type into the neighbouring source file.",
		})
	case token.CONST:
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseWhitelist, 4),
			Pos:  p.pos(d.Pos()),
			What: "constant declarations are not permitted in " + kind.String() + ".",
			Why:  "Values belong either into the requirement tree or into the neighbouring source file.",
			How:  "Move the constant into the neighbouring source file.",
		})
	}
}

func (p *Package) checkValueSpec(vs *ast.ValueSpec, kind fileKind, out *diag.Set) {
	if len(vs.Names) != 1 || len(vs.Values) != 1 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseWhitelist, 5),
			Pos:  p.pos(vs.Pos()),
			What: "a var declaration must bind exactly one name to exactly one value.",
			Why:  "Grouped or multi-valued declarations make the term ambiguous to read back.",
			How:  "Split the declaration into one var per term.",
		})
		return
	}
	if kind == kindAnnotation && vs.Names[0].Name != "_" {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseWhitelist, 6),
			Pos:  p.pos(vs.Names[0].Pos()),
			What: "binding terms must be declared as `var _ = …`.",
			Why:  "A binding has no value worth naming; a name suggests it can be referenced.",
			How:  "Rename " + vs.Names[0].Name + " to the blank identifier _.",
		})
	}
	p.checkExpr(vs.Values[0], kind, out)
}

// checkExpr walks an expression and rejects everything outside the subset.
func (p *Package) checkExpr(e ast.Expr, kind fileKind, out *diag.Set) {
	switch x := e.(type) {
	case *ast.CallExpr:
		if p.specFuncName(x.Fun) == "" {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseWhitelist, 7),
				Pos:  p.pos(x.Pos()),
				What: "only calls into " + SpecPkgPath + " are permitted.",
				Why:  "Arbitrary calls would produce specification facts at run time that static analysis cannot see (P9: analysability before brevity).",
				How:  "Remove the call, or express the fact with a spec directive.",
			})
			return
		}
		for _, a := range x.Args {
			p.checkExpr(a, kind, out)
		}
	case *ast.CompositeLit:
		if kind != kindRequirement {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseWhitelist, 8),
				Pos:  p.pos(x.Pos()),
				What: "declarations are not permitted in an annotation file.",
				Why:  "A requirement is owned by the domain side and outlives the implementation; it must not come into being at an arbitrary code location.",
				How:  "Declare it under anforderungen/ as <ID>.spec.go and reference it here with spec.Satisfies(…).",
			})
			return
		}
		// Only struct literals must name their fields. In a slice or array
		// literal the elements are values, and Go's type elision makes
		// []spec.Source{{Doc: …}} the idiomatic spelling.
		requireKeys := p.isStructLit(x)
		for _, el := range x.Elts {
			kv, ok := el.(*ast.KeyValueExpr)
			if !ok {
				if requireKeys {
					out.Add(diag.Finding{
						Code: diag.Code(diag.PhaseWhitelist, 9),
						Pos:  p.pos(el.Pos()),
						What: "positional fields are not permitted in a struct literal.",
						Why:  "Positional fields silently change meaning when the struct grows.",
						How:  "Name the field explicitly, e.g. `ID: \"R-…\"`.",
					})
					continue
				}
				p.checkExpr(el, kind, out)
				continue
			}
			p.checkExpr(kv.Value, kind, out)
		}
	case *ast.Ident, *ast.SelectorExpr, *ast.BasicLit:
		// identifiers, qualified identifiers and literals are the leaves
	case *ast.UnaryExpr:
		if x.Op != token.AND {
			p.reject(x, "operator "+x.Op.String(), out)
			return
		}
		p.checkExpr(x.X, kind, out)
	case *ast.ArrayType:
		// element type of a slice literal
	case *ast.ParenExpr:
		p.checkExpr(x.X, kind, out)
	case *ast.FuncLit:
		p.reject(x, "function literals", out)
	case *ast.BinaryExpr:
		p.reject(x, "expressions", out)
	default:
		p.reject(e, "this construct", out)
	}
}

func (p *Package) reject(e ast.Expr, what string, out *diag.Set) {
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseWhitelist, 10),
		Pos:  p.pos(e.Pos()),
		What: what + " are not permitted in annotation files.",
		Why:  "The annotation language is closed and declarative (P4): it states what holds, not how it is computed. Anything computed would be invisible to the verifier.",
		How:  "State the fact directly, or express the condition as a requirement of its own under anforderungen/ and bind it with spec.Satisfies(…).",
	})
}

// CheckOrphans reports annotation files whose neighbouring source file is gone.
func (p *Package) CheckOrphans(out *diag.Set) {
	for _, f := range p.annotationFiles {
		file := p.pkg.Fset.Position(f.Pos()).Filename
		base := baseName(file)
		neighbour := base + ".go"
		if p.sourceNames[neighbour] {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseBinding, 1),
			Pos:  ir.Position{File: file, Line: 1, Col: 1},
			What: "orphaned annotation file: " + neighbour + " does not exist.",
			Why:  "An annotation file only ever annotates constructs of its neighbour. Without it the file is a leftover from a rename or deletion.",
			How:  "Rename this file to match its source file, or delete it.",
		})
	}
}

// baseName strips the .annotation.go suffix from a path and returns the file
// base name, e.g. "/x/commands.annotation.go" -> "commands".
func baseName(path string) string {
	name := path
	if i := lastIndexByte(name, '/'); i >= 0 {
		name = name[i+1:]
	}
	return name[:len(name)-len(AnnotationSuffix)]
}

func lastIndexByte(s string, b byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == b {
			return i
		}
	}
	return -1
}

// isStructLit reports whether a composite literal constructs a struct, as
// opposed to a slice, array or map. Decided over the resolved type, not over
// the syntax, so Go's type elision inside slice literals is handled correctly.
func (p *Package) isStructLit(lit *ast.CompositeLit) bool {
	tv, ok := p.pkg.TypesInfo.Types[lit]
	if !ok || tv.Type == nil {
		return true // unknown type: demand the stricter form
	}
	_, isStruct := tv.Type.Underlying().(*types.Struct)
	return isStruct
}
