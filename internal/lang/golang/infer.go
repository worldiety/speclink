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
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// nago package paths the recognisers match on.
const (
	nagoAuth       = "go.wdy.de/nago/auth"
	nagoUser       = "go.wdy.de/nago/application/user"
	nagoPermission = "go.wdy.de/nago/application/permission"
	nagoEvs        = "go.wdy.de/nago/application/evs"
	nagoData       = "go.wdy.de/nago/pkg/data"
	nagoNdb        = "go.wdy.de/nago/pkg/ndb"
	nagoEnt        = "go.wdy.de/nago/application/ent"
	nagoEntCfg     = "go.wdy.de/nago/application/ent/cfg"
	// nagoUIEnt is the generic CRUD user interface. It lives under the ent
	// module rather than under presentation/, which is easy to get wrong: the
	// path was presentation/ui/ent in earlier versions and a stale constant
	// here would silently disable the rule instead of failing loudly.
	nagoUIEnt = "go.wdy.de/nago/application/ent/ui"
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
	out = append(out, p.inferProjections()...)
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

	// A repository is a named type standing for the framework interface, the
	// idiom being `type Repository data.Repository[Quote, ID]`. Both the
	// defined type and the alias form are matched, because both occur.
	if inst, ok := p.repositoryInstance(ts, named); ok {
		base.Kind = ir.ConstructRepository
		base.Evidence = "stands for data." + inst
		return base, true
	}

	switch {
	case p.hasMethods(named, "Evolve", "Discriminator"):
		base.Kind = ir.ConstructEvent
		base.Evidence = "implements evs.Evt, that is Evolve plus Discriminator"
		return base, true

	case p.hasMethods(named, "Decide"):
		base.Kind = ir.ConstructCommand
		base.Evidence = "implements evs.Cmd, that is Decide"
		return base, true

	case p.hasMethods(named, "Identity"):
		base.Kind = ir.ConstructAggregate
		base.Evidence = "implements data.Aggregate, that is Identity"
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
	base.Evidence = "is a named func type with auth.Subject as its first parameter"
	if isReadOnly(sig) {
		base.Kind = ir.ConstructQuery
		base.Evidence = "is a named func type returning data, with auth.Subject as its first parameter"
	}
	return base, true
}

// repositoryInstance reports whether the type declaration stands for an
// instantiation of data.Repository or data.ReadRepository, and returns the name
// of that interface.
//
// The framework idiom is to name the repository of an aggregate once and pass
// that name around:
//
//	type Repository data.Repository[Quote, ID]
//	type StagingRepository = data.Repository[Staging, SID]
//
// Both forms are recognised. The defined type is the common one; the alias is
// used where the interface must stay assignable from the framework's own.
//
// The right hand side is read from the syntax rather than from the underlying
// type, because a defined interface type erases where it came from: its
// underlying type is a plain *types.Interface with no trace of data.Repository.
func (p *Package) repositoryInstance(ts *ast.TypeSpec, named *types.Named) (string, bool) {
	idx, ok := ts.Type.(*ast.IndexListExpr)
	if !ok {
		// A single type argument arrives as IndexExpr rather than
		// IndexListExpr; data.Repository always takes two, but a project may
		// wrap a one parameter alias, so both shapes are accepted.
		if one, isOne := ts.Type.(*ast.IndexExpr); isOne {
			return p.repositoryName(one.X)
		}
		return "", false
	}
	return p.repositoryName(idx.X)
}

// repositoryName resolves the generic type being instantiated and reports
// whether it is one of the repository interfaces of the framework.
func (p *Package) repositoryName(x ast.Expr) (string, bool) {
	sel, ok := x.(*ast.SelectorExpr)
	if !ok {
		return "", false
	}
	obj := p.pkg.TypesInfo.Uses[sel.Sel]
	if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != nagoData {
		return "", false
	}
	switch obj.Name() {
	case "Repository", "ReadRepository":
		return obj.Name(), true
	}
	return "", false
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

// isNagoSubject matches the shapes the framework accepts as an actor.
//
// auth.Subject is a type alias for user.Subject, so the resolved type reports
// the user package, not auth. Matching only auth would silently recognise no
// use case at all in a real project — which is exactly what happened until a
// run against the reference codebase found it.
func isNagoSubject(t types.Type) bool {
	// Since Go 1.23 an alias is its own type node rather than the aliased
	// named type, so auth.Subject arrives here as *types.Alias and a plain
	// type assertion on *types.Named silently fails.
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj.Pkg() == nil {
		return false
	}
	switch obj.Pkg().Path() {
	case nagoAuth, nagoUser:
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
		// A commit sequence is not data; it identifies the write.
		if isCommitSequence(res.At(i).Type()) {
			continue
		}
		return true
	}
	return false
}

// isCommitSequence reports whether the type is the sequence number a write
// returns.
//
// evs.SeqID is an alias of the engine's ndb.Seq, so the resolved type reports
// the ndb package and matching evs alone recognises nothing. That is the same
// trap auth.Subject sets, and it is worse here: an unmatched sequence makes
// every writing use case look like a query, silently and without a finding.
func isCommitSequence(t types.Type) bool {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	if obj.Pkg() == nil {
		return false
	}
	switch obj.Pkg().Path() {
	case nagoEvs:
		return obj.Name() == "SeqID"
	case nagoNdb:
		return obj.Name() == "Seq"
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
			// Every Declare variant counts, not just the plain one: the CRUD
			// helpers such as DeclareCreate declare exactly one permission too,
			// they merely derive its translated texts from an entity name.
			fnName := p.specFuncNameIn(call.Fun, nagoPermission)
			if !strings.HasPrefix(fnName, "Declare") {
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
				Evidence: "is declared by permission." + fnName + ", bound to " + guarded,
				Pos:      p.pos(call.Pos()),
			})
			return true
		})
	}
	return out
}

// inferProjections finds the read models built with evs.NewProjection or
// evs.NewSingleton and reports the state type each of them folds into.
//
// The construct is the state type, not the projection value: the state is what
// a requirement talks about ("the overview shows open quotes per customer"),
// and it is the thing a binding can name with spec.For[T].
//
// A projection cannot be recognised from its type declaration the way an
// aggregate can. The state type carries only Clone, which every cloneable value
// carries, so the fact that it is a read model exists solely at the point of
// construction. That is why this is call driven.
//
// Each state type is reported once even when several constructors or several
// evs.Project registrations mention it, because the construct is the read
// model, not the number of ways it is fed.
func (p *Package) inferProjections() []ir.Construct {
	var out []ir.Construct
	seen := map[string]bool{}

	for _, f := range p.pkg.Syntax {
		if p.isGeneratedByUs(f) {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			fnName := p.specFuncNameIn(call.Fun, nagoEvs)

			// NewProjection[K, S] carries the state second, NewSingleton[S]
			// first: the singleton fixes the key type and drops it.
			var arg int
			switch fnName {
			case "NewProjection":
				arg = 1
			case "NewSingleton":
				arg = 0
			default:
				return true
			}

			t, ok := p.typeArg(call, arg)
			if !ok {
				return true
			}
			named, ok := stateNamed(t)
			if !ok {
				return true
			}
			// A read model folded in one package may well be declared in
			// another; only the local one is this package's construct.
			if named.Obj().Pkg() == nil || named.Obj().Pkg().Path() != p.PkgPath() {
				return true
			}
			name := p.PkgPath() + "." + named.Obj().Name()
			if seen[name] {
				return true
			}
			seen[name] = true

			out = append(out, ir.Construct{
				Kind:     ir.ConstructProjection,
				Name:     name,
				Package:  p.PkgPath(),
				Evidence: "is the state of an evs." + fnName + " read model",
				Pos:      p.pos(named.Obj().Pos()),
			})
			return true
		})
	}
	return out
}

// stateNamed unwraps the projection state type, which is always a pointer: the
// fold mutates it in place, so a value type could not be folded at all.
func stateNamed(t types.Type) (*types.Named, bool) {
	if ptr, ok := types.Unalias(t).(*types.Pointer); ok {
		t = ptr.Elem()
	}
	named, ok := types.Unalias(t).(*types.Named)
	return named, ok
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
