package golang

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// ReadRequirements extracts the requirement declarations of the package's
// *.spec.go files.
//
// DerivedFrom and Supersedes are recorded as qualified Go identifiers, not as
// requirement IDs. Resolving them to IDs is the job of the second pass, once
// every declaration has been collected. That is what makes forward references
// legal and the input order irrelevant.
func (p *Package) ReadRequirements(out *diag.Set) []*ir.Requirement {
	var reqs []*ir.Requirement
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
				if !ok {
					continue
				}
				if !p.isSpecType(lit, "Requirement") {
					continue
				}
				reqs = append(reqs, p.readRequirement(vs, lit, out))
			}
		}
	}
	return reqs
}

// isSpecType reports whether a composite literal constructs speclink/spec.<name>.
func (p *Package) isSpecType(lit *ast.CompositeLit, name string) bool {
	tv, ok := p.pkg.TypesInfo.Types[lit]
	if !ok {
		return false
	}
	named, ok := tv.Type.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() != nil && obj.Pkg().Path() == SpecPkgPath && obj.Name() == name
}

func (p *Package) readRequirement(vs *ast.ValueSpec, lit *ast.CompositeLit, out *diag.Set) *ir.Requirement {
	r := &ir.Requirement{
		GoIdent: p.PkgPath() + "." + vs.Names[0].Name,
		Pos:     p.pos(vs.Pos()),
	}

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
			r.ID, _ = p.stringArg(kv.Value)
		case "Title":
			r.Title, _ = p.stringArg(kv.Value)
		case "Text":
			r.Text, _ = p.stringArg(kv.Value)
		case "Detail":
			r.Detail, _ = p.stringArg(kv.Value)
		case "Rationale":
			r.Rationale, _ = p.stringArg(kv.Value)
		case "Kind":
			r.Kind = ir.Kind(p.enumValue(kv.Value))
		case "Discipline":
			r.Discipline = ir.Discipline(p.enumValue(kv.Value))
		case "Status":
			r.Status = ir.Status(p.enumValue(kv.Value))
		case "DerivedFrom":
			r.DerivedFrom = p.identList(kv.Value)
		case "Supersedes":
			r.Supersedes = p.identList(kv.Value)
		case "Sources":
			r.Sources = p.readSources(kv.Value)
		case "Attachments":
			r.Attachments = p.readAttachments(kv.Value)
		}
	}

	p.checkRequirementShape(r, vs, out)
	return r
}

// enumValue folds a spec enum constant to its integer value. Constants resolve
// through go/types, so a misspelled enum is already a compile error.
func (p *Package) enumValue(e ast.Expr) int {
	tv, ok := p.pkg.TypesInfo.Types[e]
	if !ok || tv.Value == nil {
		return 0
	}
	n, ok := constant.Int64Val(constant.ToInt(tv.Value))
	if !ok {
		return 0
	}
	return int(n)
}

// identList reads a slice literal of requirement references and returns the
// qualified Go identifiers. Resolution to IDs happens in the second pass.
func (p *Package) identList(e ast.Expr) []string {
	lit, ok := e.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var out []string
	for _, el := range lit.Elts {
		if name := p.objectName(el); name != "" {
			out = append(out, name)
		}
	}
	return out
}

func (p *Package) readSources(e ast.Expr) []ir.Source {
	lit, ok := e.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var out []ir.Source
	for _, el := range lit.Elts {
		sl, ok := el.(*ast.CompositeLit)
		if !ok {
			continue
		}
		s := ir.Source{Pos: p.pos(sl.Pos())}
		for _, f := range sl.Elts {
			kv, ok := f.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}
			switch key.Name {
			case "Doc":
				s.Doc, _ = p.stringArg(kv.Value)
			case "Anchor":
				s.Anchor, _ = p.stringArg(kv.Value)
			case "Extern":
				s.Extern, _ = p.stringArg(kv.Value)
			case "Note":
				s.Note, _ = p.stringArg(kv.Value)
			}
		}
		out = append(out, s)
	}
	return out
}

func (p *Package) readAttachments(e ast.Expr) []ir.Attachment {
	lit, ok := e.(*ast.CompositeLit)
	if !ok {
		return nil
	}
	var out []ir.Attachment
	for _, el := range lit.Elts {
		al, ok := el.(*ast.CompositeLit)
		if !ok {
			continue
		}
		var a ir.Attachment
		for _, f := range al.Elts {
			kv, ok := f.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}
			switch key.Name {
			case "Path":
				a.Path, _ = p.stringArg(kv.Value)
			case "Role":
				a.Role = ir.Role(p.enumValue(kv.Value))
			case "Note":
				a.Note, _ = p.stringArg(kv.Value)
			}
		}
		out = append(out, a)
	}
	return out
}

// checkRequirementShape validates the fields a requirement must carry. These
// are obligations the Go type system cannot express.
func (p *Package) checkRequirementShape(r *ir.Requirement, vs *ast.ValueSpec, out *diag.Set) {
	pos := p.pos(vs.Pos())

	if r.ID == "" {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 1),
			Pos:  pos,
			What: "requirement " + vs.Names[0].Name + " has no ID.",
			Why:  "The ID is the stable identity of the requirement and the key of every report.",
			How:  "Add an ID of the form R-<PREFIX>-<NAME>, matching the file name.",
		})
	}
	if r.Kind == 0 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 2),
			Pos:  pos,
			What: "requirement " + r.ID + " has no Kind.",
			Why:  "Kind decides the first directory level of the requirement tree and whether the requirement is cross cutting.",
			How:  "Set Kind to spec.Functional, spec.NonFunctional, spec.Constraint or spec.Decision.",
		})
	}
	if r.Status == 0 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 3),
			Pos:  pos,
			What: "requirement " + r.ID + " has no Status.",
			Why:  "Status drives the backward direction of the coverage check: only spec.Normative has to be covered, every other value is an explicit exemption.",
			How:  "Set Status to spec.Normative, or to spec.Planned, spec.OutOfScope, spec.Informative or spec.Abstract to exempt it.",
		})
	}
	if r.Kind == ir.Decision && r.Rationale == "" {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 4),
			Pos:  pos,
			What: "decision " + r.ID + " has no Rationale.",
			Why:  "A decision without a justification cannot be reviewed, and its cost cannot be weighed later.",
			How:  "Add a Rationale explaining why this ruling was made and what the alternative would have cost.",
		})
	}
	if r.Text == "" {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 5),
			Pos:  pos,
			What: "requirement " + r.ID + " has no Text.",
			Why:  "Text is the normative sentence; it appears in lists, traceability matrices and diagnostics.",
			How:  "Add one normative sentence. Long form belongs in a Markdown file named by Detail.",
		})
	}
}
