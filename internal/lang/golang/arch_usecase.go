package golang

import (
	"go/ast"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

const (
	// RuleUCFile fires when a use case is not declared in its own uc_ file.
	RuleUCFile = "K5-UC-FILE"
	// RuleUCConstructor fires when the constructor is missing or misplaced.
	RuleUCConstructor = "K5-UC-CONSTRUCTOR"
	// RuleUCSignature fires for a use case with the wrong shape.
	RuleUCSignature = "K5-UC-SIGNATURE"
	// RuleUCAuthz fires when nothing in the implementation looks like an
	// authorisation check.
	RuleUCAuthz = "K5-UC-AUTHZ"
	// RuleUCPermission fires when no permission is bound to the use case.
	RuleUCPermission = "K5-UC-PERMISSION"
	// RuleUCPermissionI18n fires when permission texts are hardcoded.
	RuleUCPermissionI18n = "K5-UC-PERMISSION-I18N"
	// RuleUCDeps fires when a use case reaches for package level state instead
	// of taking its dependencies through the constructor.
	RuleUCDeps = "K5-UC-DEPS"
)

// CheckUseCases verifies the shape, the file layout, the authorisation and the
// dependency injection of every use case in a bounded context.
func CheckUseCases(pkgs []*Package, cfg config.Config, root string, out *diag.Set) {
	for _, p := range pkgs {
		rel := p.relDir(root)
		if _, inContext := cfg.InContextRoot(rel); !inContext {
			continue
		}
		if cfg.Excluded(rel) || contextRole(rel, cfg) != roleDomain {
			continue
		}

		perms := p.permissionsByUseCase()
		for _, uc := range p.useCaseTypes() {
			p.checkUseCaseFile(uc, out)
			p.checkUseCaseSignature(uc, out)
			p.checkUseCaseConstructor(uc, perms, out)
		}
	}
}

// checkUseCaseFile enforces one use case per file, named after it.
//
// The convention is what makes a context navigable without an index: the file
// list is the capability list. It also keeps diffs of unrelated use cases apart.
func (p *Package) checkUseCaseFile(uc useCase, out *diag.Set) {
	want := useCaseFileName(uc.name)
	got := filepath.Base(uc.file)
	if got == want {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 50),
		Pos:  uc.pos,
		Rule: RuleUCFile,
		What: "use case " + uc.name + " is declared in " + got + ", expected " + want + ".",
		Why:  "One use case per file, named after it. The file list of a context is then its capability list, and unrelated use cases never share a diff.",
		How:  "Move the type and its constructor into " + want + ".",
	})
}

// checkUseCaseSignature enforces the universal shape: subject first, error last.
func (p *Package) checkUseCaseSignature(uc useCase, out *diag.Set) {
	res := uc.sig.Results()
	if res.Len() > 0 && isErrorType(res.At(res.Len()-1).Type()) {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 51),
		Pos:  uc.pos,
		Rule: RuleUCSignature,
		What: "use case " + uc.name + " does not return an error as its last result.",
		Why:  "Every use case can fail authorisation, so the error is not optional. A signature without it forces callers to ignore the case that matters most.",
		How:  "Change the signature to func(subject auth.Subject, …) (…, error).",
	})
}

// checkUseCaseConstructor verifies that the constructor exists, lives beside
// its type, performs an authorisation check and takes its dependencies as
// parameters.
func (p *Package) checkUseCaseConstructor(uc useCase, perms map[string][]permDecl, out *diag.Set) {
	name := "New" + uc.name
	fn, decl, ok := p.lookupFuncDecl(name)
	if !ok {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 52),
			Pos:  uc.pos,
			Rule: RuleUCConstructor,
			What: "use case " + uc.name + " has no constructor " + name + ".",
			Why:  "The constructor is where the dependencies of a use case enter. Without it callers build the closure themselves and every call site wires it differently.",
			How:  "Add `func " + name + "(…) " + uc.name + "` in " + useCaseFileName(uc.name) + ".",
		})
		return
	}

	declFile := filepath.Base(p.pkg.Fset.Position(decl.Pos()).Filename)
	if want := useCaseFileName(uc.name); declFile != want {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 53),
			Pos:  p.pos(decl.Pos()),
			Rule: RuleUCConstructor,
			What: name + " is declared in " + declFile + ", expected " + want + ".",
			Why:  "Type and constructor of a use case belong together; split across files the pair drifts and one of them is forgotten on rename.",
			How:  "Move " + name + " into " + want + ", next to the type it builds.",
		})
	}

	if sig, ok := fn.Type().(*types.Signature); ok && sig.Results().Len() > 0 {
		if named, ok := sig.Results().At(0).Type().(*types.Named); !ok || named.Obj().Name() != uc.name {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 54),
				Pos:  p.pos(decl.Pos()),
				Rule: RuleUCConstructor,
				What: name + " does not return " + uc.name + ".",
				Why:  "The constructor is the only sanctioned way to obtain the use case; returning something else leaves the type without a producer.",
				How:  "Change the result type to " + uc.name + ".",
			})
		}
	}

	body := closureBody(decl)
	if body == nil {
		return
	}
	// Whether the body names a permission of this use case answers both
	// questions below, so it is established once.
	named := p.namesOwnPermission(body, perms[uc.name])

	p.checkAuthorisation(uc, decl, body, named, out)
	p.checkPermissionBinding(uc, decl, body, perms, named, out)
	p.checkDependencies(uc, decl, body, out)
}

// checkAuthorisation looks for anything that resembles an authorisation check.
//
// The heuristic is deliberately generous, and it does not attempt to prove that
// the check is correct — only that the author thought about it at all. A use
// case that nowhere mentions the subject it was handed is almost certainly an
// oversight, and that is the class of mistake worth catching. In the reference
// framework an unguarded read projection was a real, project wide bug.
func (p *Package) checkAuthorisation(uc useCase, decl *ast.FuncDecl, body *ast.BlockStmt, namesOwnPermission bool, out *diag.Set) {
	if p.hasAuthorisationEvidence(body) {
		return
	}

	// Naming the permission that guards this use case is evidence too, and
	// stronger than the presence of a subject: it says which right is being
	// enforced, not merely that something was consulted.
	//
	// It is the only evidence there is when the guard is applied at the type
	// level rather than in the body — `return guard(load, PermViewRisks)`
	// wraps the check around the function and never mentions a subject, which
	// is a perfectly ordinary way to write a read side.
	if namesOwnPermission {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 55),
		Pos:  p.pos(decl.Pos()),
		Rule: RuleUCAuthz,
		What: "use case " + uc.name + " contains nothing that looks like an authorisation check.",
		Why:  "Every use case receives a subject, and one that never consults it grants everyone everything. The check here is deliberately loose: it does not judge whether the check is right, only that there is one.",
		How:  "Call subject.Audit(Perm…), subject.AuditResource(…), subject.HasPermission/HasRole/HasGroup, return an error wrapping user.PermissionDeniedErr, or pass the subject on to another use case that checks.",
	})
}

// namesOwnPermission reports whether the body mentions one of the permissions
// declared for this use case.
func (p *Package) namesOwnPermission(body *ast.BlockStmt, declared []permDecl) bool {
	if body == nil || len(declared) == 0 {
		return false
	}
	want := map[string]bool{}
	for _, d := range declared {
		want[d.varName] = true
	}

	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		if id, ok := n.(*ast.Ident); ok && want[id.Name] {
			found = true
			return false
		}
		return true
	})
	return found
}

// authorisation method names accepted as evidence.
var authzMethods = map[string]bool{
	"Audit":                 true,
	"AuditResource":         true,
	"HasPermission":         true,
	"HasResourcePermission": true,
	"HasRole":               true,
	"HasGroup":              true,
}

// hasAuthorisationEvidence reports whether the body consults the subject in any
// recognised way.
func (p *Package) hasAuthorisationEvidence(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		// subject.Audit(…) and friends.
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok && authzMethods[sel.Sel.Name] {
			if p.isSubjectExpr(sel.X) {
				found = true
				return false
			}
		}

		// Delegation: the subject is handed to another use case, which is
		// then responsible for the check. Without this an orchestrating use
		// case would be a false positive.
		for _, arg := range call.Args {
			if p.isSubjectExpr(arg) {
				found = true
				return false
			}
		}
		return true
	})
	if found {
		return true
	}
	return p.returnsPermissionDenied(body)
}

// isSubjectExpr reports whether an expression is an auth subject.
func (p *Package) isSubjectExpr(e ast.Expr) bool {
	tv, ok := p.pkg.TypesInfo.Types[e]
	if !ok || tv.Type == nil {
		return false
	}
	return isNagoSubject(tv.Type)
}

// returnsPermissionDenied reports whether the body mentions the framework's
// permission denied error, which some use cases return directly instead of
// going through Audit.
func (p *Package) returnsPermissionDenied(body *ast.BlockStmt) bool {
	found := false
	ast.Inspect(body, func(n ast.Node) bool {
		if found {
			return false
		}
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		obj := p.pkg.TypesInfo.Uses[sel.Sel]
		if obj == nil || obj.Pkg() == nil {
			return true
		}
		if strings.HasSuffix(obj.Pkg().Path(), "/application/user") &&
			strings.Contains(obj.Name(), "PermissionDenied") {
			found = true
			return false
		}
		return true
	})
	return found
}

// permDecl is a permission declaration bound to a use case type.
type permDecl struct {
	varName string
	id      string
	i18n    bool
	pos     ir.Position
}

// permissionsByUseCase collects the permission declarations of the package,
// keyed by the use case type they are bound to.
//
// The framework binds a permission to a use case through the type parameter of
// permission.Declare, and enforces at run time that it is a named func type.
// That makes the pairing statically recoverable.
func (p *Package) permissionsByUseCase() map[string][]permDecl {
	out := map[string][]permDecl{}

	for _, f := range p.pkg.Syntax {
		if p.isGeneratedByUs(f) {
			continue
		}
		ast.Inspect(f, func(n ast.Node) bool {
			vs, ok := n.(*ast.ValueSpec)
			if !ok {
				return true
			}
			for i, val := range vs.Values {
				call, ok := val.(*ast.CallExpr)
				if !ok {
					continue
				}
				fnName := p.specFuncNameIn(call.Fun, nagoPermission)
				if !strings.HasPrefix(fnName, "Declare") {
					continue
				}
				t, ok := p.typeArg(call, 0)
				if !ok {
					continue
				}
				name := ""
				if named, isNamed := t.(*types.Named); isNamed {
					name = named.Obj().Name()
				}
				id, _ := p.stringArg(argAt(call, 0))
				varName := ""
				if i < len(vs.Names) {
					varName = vs.Names[i].Name
				}
				out[name] = append(out[name], permDecl{
					varName: varName,
					id:      id,
					// The CRUD helpers generate their texts through i18n; the
					// plain Declare only does so when the caller passes i18n
					// values in.
					i18n: fnName != "Declare" || p.usesI18n(call),
					pos:  p.pos(call.Pos()),
				})
			}
			return true
		})
	}
	return out
}

// specFuncNameIn is like specFuncName but for an arbitrary package path.
func (p *Package) specFuncNameIn(fun ast.Expr, pkgPath string) string {
	id := calleeIdent(fun)
	if id == nil {
		return ""
	}
	obj := p.pkg.TypesInfo.Uses[id]
	if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != pkgPath {
		return ""
	}
	return obj.Name()
}

// usesI18n reports whether a permission declaration takes its texts from the
// translation catalogue rather than from hardcoded literals.
func (p *Package) usesI18n(call *ast.CallExpr) bool {
	found := false
	ast.Inspect(call, func(n ast.Node) bool {
		if found {
			return false
		}
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		obj := p.pkg.TypesInfo.Uses[sel.Sel]
		if obj != nil && obj.Pkg() != nil && strings.HasSuffix(obj.Pkg().Path(), "/i18n") {
			found = true
			return false
		}
		return true
	})
	return found
}

// checkPermissionBinding requires at least one permission bound to the use case
// and actually consulted by the implementation.
//
// At least one, not exactly one: a single use case can legitimately guard
// several operations, and the framework's own drive package binds two
// permissions to one use case type.
func (p *Package) checkPermissionBinding(uc useCase, decl *ast.FuncDecl, body *ast.BlockStmt, perms map[string][]permDecl, namesOwnPermission bool, out *diag.Set) {
	declared := perms[uc.name]
	if len(declared) == 0 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 56),
			Pos:  uc.pos,
			Rule: RuleUCPermission,
			What: "use case " + uc.name + " has no permission of its own.",
			Why:  "A permission per use case is what makes authorisation assignable and auditable. Sharing one across several use cases means they can never be granted apart.",
			How:  "Add `Perm" + uc.name + " = permission.Declare[" + uc.name + "](\"…\", name, description)` and check it in " + "New" + uc.name + ".",
		})
		return
	}

	// The texts are end user facing regardless of whether the implementation
	// consults the permission, so this is checked for every declaration rather
	// than only for the one that is used.
	for _, d := range declared {
		p.checkPermissionI18n(uc, d, out)
	}

	if namesOwnPermission {
		return
	}

	// Delegation counts here for the same reason it counts as authorisation
	// evidence: a use case that hands its subject to something else has moved
	// the check there, not skipped it. In the decide-evolve pattern that is the
	// normal shape — the closure forwards the command to the handler and the
	// permission is audited inside Decide, where the invariants are.
	//
	// Refusing to accept that would report every event sourced use case in a
	// system, which is how a rule teaches people to ignore it.
	if p.hasAuthorisationEvidence(body) {
		return
	}

	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 57),
		Pos:  p.pos(decl.Pos()),
		Rule: RuleUCPermission,
		What: "use case " + uc.name + " never uses its permission " + declared[0].varName + ".",
		Why:  "A declared but unchecked permission is worse than none: it appears in the role editor and suggests a protection that does not exist.",
		How:  "Check it in the implementation, e.g. `if err := subject.Audit(" + declared[0].varName + "); err != nil { … }`, or hand the subject to whatever does check.",
	})
}

// checkPermissionI18n requires the human readable texts of a permission to come
// from the translation catalogue.
//
// Permission names and descriptions are end user text: they appear in the role
// editor, where somebody who is not a developer decides who may do what. A
// hardcoded German or English sentence makes that screen untranslatable, and
// the decision it supports is exactly the kind that must be understood.
func (p *Package) checkPermissionI18n(uc useCase, d permDecl, out *diag.Set) {
	if d.i18n {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 59),
		Pos:  d.pos,
		Rule: RuleUCPermissionI18n,
		What: "permission " + d.varName + " carries hardcoded texts.",
		Why:  "Name and description appear in the role editor, where a non developer decides who may do what. Hardcoded sentences make that screen untranslatable.",
		How:  "Wrap the texts in i18n.MustString(…), or use a permission.Declare<Verb> helper, which derives translated texts from the entity name.",
	})
}

// checkDependencies enforces that a use case takes what it needs through its
// constructor instead of reaching for package level state.
//
// Package level mutable state makes a use case untestable in isolation and
// couples it to an initialisation order nobody can see at the call site.
// Constants and permission declarations are exempt: the first cannot change,
// the second is exactly what is supposed to be referenced here.
func (p *Package) checkDependencies(uc useCase, decl *ast.FuncDecl, body *ast.BlockStmt, out *diag.Set) {
	params := map[types.Object]bool{}
	if decl.Type.Params != nil {
		for _, field := range decl.Type.Params.List {
			for _, name := range field.Names {
				if obj := p.pkg.TypesInfo.Defs[name]; obj != nil {
					params[obj] = true
				}
			}
		}
	}

	scope := p.pkg.Types.Scope()
	reported := map[string]bool{}

	ast.Inspect(body, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok {
			return true
		}
		obj := p.pkg.TypesInfo.Uses[id]
		v, isVar := obj.(*types.Var)
		if !isVar || params[obj] || v.IsField() {
			return true
		}
		// Only package level variables matter; locals are fine.
		if scope.Lookup(v.Name()) != obj {
			return true
		}
		if reported[v.Name()] || isPermissionVar(v) {
			return true
		}
		reported[v.Name()] = true

		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 58),
			Pos:  p.pos(id.Pos()),
			Rule: RuleUCDeps,
			What: "use case " + uc.name + " reads the package level variable " + v.Name() + ".",
			Why:  "Package level state makes the use case untestable in isolation and ties it to an initialisation order that is invisible at the call site.",
			How:  "Pass " + v.Name() + " into New" + uc.name + " as a parameter and capture it in the closure.",
		})
		return true
	})
}

// isPermissionVar reports whether a package level variable is a permission ID,
// which a use case is supposed to reference.
func isPermissionVar(v *types.Var) bool {
	named, ok := v.Type().(*types.Named)
	if !ok {
		return false
	}
	obj := named.Obj()
	return obj.Pkg() != nil && obj.Pkg().Path() == nagoPermission && obj.Name() == "ID"
}

// lookupFuncDecl resolves a package level function and its declaration.
func (p *Package) lookupFuncDecl(name string) (*types.Func, *ast.FuncDecl, bool) {
	fn, ok := p.lookupFunc(name)
	if !ok {
		return nil, nil, false
	}
	for _, f := range p.pkg.Syntax {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if ok && fd.Name.Name == name && fd.Recv == nil {
				return fn, fd, true
			}
		}
	}
	return fn, nil, false
}

// closureBody returns the body of the function literal a constructor returns,
// which is where the actual use case implementation lives.
func closureBody(decl *ast.FuncDecl) *ast.BlockStmt {
	if decl == nil || decl.Body == nil {
		return nil
	}
	var body *ast.BlockStmt
	ast.Inspect(decl.Body, func(n ast.Node) bool {
		if body != nil {
			return false
		}
		if lit, ok := n.(*ast.FuncLit); ok {
			body = lit.Body
			return false
		}
		return true
	})
	if body == nil {
		// A constructor that delegates rather than building a closure, e.g.
		// one that decorates another use case. Its own body is what there is.
		return decl.Body
	}
	return body
}
