package golang

// Recognisers for the nago framework.
//
// speclink knows the framework, never the project (docs/annotations.md §1.7).
// A recogniser must know its framework, and a framework is shared by many
// projects, so that knowledge amortises. Project knowledge never does.
//
// Everything here decides over resolved types, never over the spelling of an
// identifier: an alias import, a dot import or a shadowed name cannot fool it.

import (
	"go/ast"
	"go/types"

	"github.com/worldiety/speclink/internal/ir"
)

// nago package paths the recognisers match on.
const (
	nagoAuth       = "go.wdy.de/nago/auth"
	nagoPermission = "go.wdy.de/nago/application/permission"
	nagoEvs        = "go.wdy.de/nago/application/evs"
	nagoData       = "go.wdy.de/nago/pkg/data"
	nagoEnt        = "go.wdy.de/nago/application/ent"
	nagoEntCfg     = "go.wdy.de/nago/application/ent/cfg"
	nagoUIEnt      = "go.wdy.de/nago/presentation/ui/ent"
)

// Infer walks the ordinary source files of the package and returns the
// architectural constructs it recognises.
//
// Annotation and requirement files are skipped: they carry the statements
// about constructs, not constructs themselves.
func (p *Package) Infer() []ir.Construct {
	var out []ir.Construct
	for _, f := range p.pkg.Syntax {
		if p.isGeneratedByUs(f) {
			continue
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, s := range gd.Specs {
				ts, ok := s.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if c, ok := p.inferType(ts); ok {
					out = append(out, c)
				}
			}
		}
	}
	out = append(out, p.inferPermissions()...)
	return out
}

// isGeneratedByUs reports whether the file is an annotation or requirement
// file rather than ordinary source.
func (p *Package) isGeneratedByUs(f *ast.File) bool {
	for _, a := range p.annotationFiles {
		if a == f {
			return true
		}
	}
	for _, r := range p.requirementFiles {
		if r == f {
			return true
		}
	}
	return false
}

// inferType classifies a type declaration.
//
// Order matters: a command is also a struct, and an event is also a struct, so
// the more specific interface satisfaction is tested before the shape rules.
func (p *Package) inferType(ts *ast.TypeSpec) (ir.Construct, bool) {
	obj := p.pkg.TypesInfo.Defs[ts.Name]
	if obj == nil {
		return ir.Construct{}, false
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return ir.Construct{}, false
	}

	base := ir.Construct{
		Name:    p.PkgPath() + "." + ts.Name.Name,
		Package: p.PkgPath(),
		Pos:     p.pos(ts.Pos()),
	}

	switch {
	case p.hasMethods(named, "Evolve", "Discriminator"):
		base.Kind = ir.ConstructEvent
		base.Evidence = "implements evs.Evt (Evolve plus Discriminator)"
		return base, true

	case p.hasMethods(named, "Decide"):
		base.Kind = ir.ConstructCommand
		base.Evidence = "implements evs.Cmd (Decide)"
		return base, true

	case p.hasMethods(named, "Identity"):
		base.Kind = ir.ConstructAggregate
		base.Evidence = "implements data.Aggregate (Identity)"
		return base, true
	}

	// A named func type whose first parameter is an auth subject is the
	// universal use case signature of the framework.
	sig, ok := named.Underlying().(*types.Signature)
	if !ok {
		return ir.Construct{}, false
	}
	if !p.firstParamIsSubject(sig) {
		return ir.Construct{}, false
	}
	base.Kind = ir.ConstructUseCase
	base.Evidence = "named func type with auth.Subject as first parameter"
	if isReadOnly(sig) {
		base.Kind = ir.ConstructQuery
		base.Evidence = "named func type returning data, with auth.Subject as first parameter"
	}
	return base, true
}

// hasMethods reports whether the named type carries all the given methods,
// on the value or the pointer receiver.
func (p *Package) hasMethods(named *types.Named, names ...string) bool {
	for _, name := range names {
		if !hasMethod(named, name) {
			return false
		}
	}
	return true
}

func hasMethod(named *types.Named, name string) bool {
	for i := 0; i < named.NumMethods(); i++ {
		if named.Method(i).Name() == name {
			return true
		}
	}
	ptr := types.NewPointer(named)
	ms := types.NewMethodSet(ptr)
	for i := 0; i < ms.Len(); i++ {
		if ms.At(i).Obj().Name() == name {
			return true
		}
	}
	return false
}

// firstParamIsSubject reports whether the signature takes an auth subject or a
// permission auditable as its first parameter.
func (p *Package) firstParamIsSubject(sig *types.Signature) bool {
	if sig.Params().Len() == 0 {
		return false
	}
	return isNagoSubject(sig.Params().At(0).Type())
}

// isNagoSubject matches auth.Subject and permission.Auditable, the two shapes
// the framework accepts as an actor.
func isNagoSubject(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj.Pkg() == nil {
		return false
	}
	switch obj.Pkg().Path() {
	case nagoAuth:
		return obj.Name() == "Subject"
	case nagoPermission:
		return obj.Name() == "Auditable"
	}
	return false
}

// isReadOnly distinguishes a query from a writing use case: a query returns
// data, a command returns only an error or a commit sequence.
func isReadOnly(sig *types.Signature) bool {
	res := sig.Results()
	if res.Len() < 2 {
		return false
	}
	for i := 0; i < res.Len(); i++ {
		if isErrorType(res.At(i).Type()) {
			continue
		}
		if named, ok := res.At(i).Type().(*types.Named); ok {
			obj := named.Obj()
			// A commit sequence is not data; it identifies the write.
			if obj.Pkg() != nil && obj.Pkg().Path() == nagoEvs && obj.Name() == "SeqID" {
				continue
			}
		}
		return true
	}
	return false
}

func isErrorType(t types.Type) bool {
	named, ok := t.(*types.Named)
	if !ok {
		return false
	}
	return named.Obj().Name() == "error" && named.Obj().Pkg() == nil
}

// inferPermissions finds permission.Declare calls and binds each permission to
// the use case type it guards.
//
// The framework enforces at run time that the type parameter is a named func
// type. That makes the binding permission <-> use case recoverable statically,
// with the ID and the human readable name as constant literals.
func (p *Package) inferPermissions() []ir.Construct {
	var out []ir.Construct
	for _, f := range p.pkg.Syntax {
		if p.isGeneratedByUs(f) {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if !p.callsInto(call.Fun, nagoPermission, "Declare") {
				return true
			}
			id, _ := p.stringArg(argAt(call, 0))
			guarded := ""
			if t, ok := p.typeArg(call, 0); ok {
				guarded = typeName(t)
			}
			out = append(out, ir.Construct{
				Kind:     ir.ConstructPermission,
				Name:     id,
				Package:  p.PkgPath(),
				Evidence: "permission.Declare bound to " + guarded,
				Pos:      p.pos(call.Pos()),
			})
			return true
		})
	}
	return out
}

// callsInto reports whether fun resolves to the named function of the given
// package.
func (p *Package) callsInto(fun ast.Expr, pkgPath, name string) bool {
	id := calleeIdent(fun)
	if id == nil {
		return false
	}
	obj := p.pkg.TypesInfo.Uses[id]
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	return obj.Pkg().Path() == pkgPath && obj.Name() == name
}
