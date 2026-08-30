package golang

import (
	"crypto/sha256"
	"encoding/hex"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"os"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// ReadBindings extracts every binding term of the package's annotation files.
//
// Everything here reads the *typed* AST. Type resolution and evaluation are two
// different things: speclink resolves fully (via go/types) but never executes.
// That is what makes order irrelevant — terms are collected in one pass, no
// annotation can change the meaning of another.
func (p *Package) ReadBindings(style Style, out *diag.Set) []ir.Binding {
	var bindings []ir.Binding
	for _, f := range p.annotationFiles {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, s := range gd.Specs {
				vs, ok := s.(*ast.ValueSpec)
				if !ok || len(vs.Values) != 1 {
					continue
				}
				call, ok := vs.Values[0].(*ast.CallExpr)
				if !ok {
					continue
				}
				if b, ok := p.readBinding(call, style, out); ok {
					bindings = append(bindings, b)
				}
			}
		}
	}
	return bindings
}

// readBinding turns one spec.For… call into an ir.Binding.
func (p *Package) readBinding(call *ast.CallExpr, style Style, out *diag.Set) (ir.Binding, bool) {
	name := p.specFuncName(call.Fun)
	pos := p.pos(call.Pos())

	var (
		target ir.Target
		args   []ast.Expr
	)
	switch name {
	case "For":
		t, ok := p.typeArg(call, 0)
		if !ok {
			return ir.Binding{}, false
		}
		target = ir.Target{Kind: ir.TargetType, Package: p.PkgPath(), Name: typeName(t)}
		args = call.Args

	case "ForDecl":
		if len(call.Args) == 0 {
			return ir.Binding{}, false
		}
		t, ok := p.declTarget(call.Args[0], out)
		if !ok {
			return ir.Binding{}, false
		}
		target = t
		args = call.Args[1:]

	case "ForField":
		t, ok := p.typeArg(call, 0)
		if !ok || len(call.Args) == 0 {
			return ir.Binding{}, false
		}
		field, _ := p.stringArg(call.Args[0])
		target = ir.Target{Kind: ir.TargetField, Package: p.PkgPath(), Name: typeName(t), Field: field}
		p.checkFieldExists(t, field, call.Args[0], out)
		args = call.Args[1:]

	case "ForPackage":
		target = ir.Target{Kind: ir.TargetPackage, Package: p.PkgPath()}
		args = call.Args

	default:
		// Not a binding term. An assertion at top level has no target and is
		// therefore meaningless; report it rather than drop it silently.
		if isAssertionName(name) {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseBinding, 2),
				Pos:  pos,
				What: "spec." + name + " is an assertion and needs a binding.",
				Why:  "An assertion states something about a construct. Standing alone it names no construct (locality principle).",
				How:  "Wrap it, e.g. `var _ = spec.For[MyType](spec." + name + "(…))`.",
			})
		}
		return ir.Binding{}, false
	}

	b := ir.Binding{Target: target, Pos: pos}
	for _, a := range args {
		ac, ok := a.(*ast.CallExpr)
		if !ok {
			continue
		}
		if as, ok := p.readAssertion(ac, out); ok {
			b.Assertions = append(b.Assertions, as)
		}
	}
	p.checkTargetAllowed(b, style, out)
	return b, true
}

// readAssertion turns one spec.<Assertion> call into an ir.Assertion.
func (p *Package) readAssertion(call *ast.CallExpr, out *diag.Set) (ir.Assertion, bool) {
	name := p.specFuncName(call.Fun)
	pos := p.pos(call.Pos())

	switch name {
	case "Satisfies":
		a := ir.Assertion{Kind: ir.AssertSatisfies, Pos: pos}
		for _, arg := range call.Args {
			a.Requirements = append(a.Requirements, p.objectName(arg))
		}
		return a, true

	case "Transition":
		t, ok := p.typeArg(call, 0)
		if !ok {
			return ir.Assertion{}, false
		}
		state := ""
		if len(call.Args) > 0 {
			state, _ = p.stringArg(call.Args[0])
		}
		return ir.Assertion{Kind: ir.AssertTransition, Pos: pos, EventType: typeName(t), State: state}, true

	case "External":
		return ir.Assertion{Kind: ir.AssertExternal, Pos: pos}, true

	case "Help":
		text, _ := p.stringArg(argAt(call, 0))
		return ir.Assertion{Kind: ir.AssertHelp, Pos: pos, Text: text}, true

	case "Rationale":
		text, _ := p.stringArg(argAt(call, 0))
		return ir.Assertion{Kind: ir.AssertRationale, Pos: pos, Text: text}, true

	case "Term":
		return ir.Assertion{Kind: ir.AssertTerm, Pos: pos, Term: p.objectName(argAt(call, 0))}, true

	case "Waive":
		rule, _ := p.stringArg(argAt(call, 0))
		reason, _ := p.stringArg(argAt(call, 1))
		if reason == "" {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseBinding, 3),
				Pos:  pos,
				What: "spec.Waive requires a reason.",
				Why:  "A waiver without justification is an invisible gap. The reason is carried into the gap report so the exemption leaves a trace.",
				How:  "State why this rule cannot hold here, e.g. spec.Waive(\"" + rule + "\", \"system projection without a subject\").",
			})
		}
		return ir.Assertion{Kind: ir.AssertWaive, Pos: pos, Rule: rule, Text: reason}, true

	case "Draft":
		return ir.Assertion{Kind: ir.AssertDraft, Pos: pos}, true

	case "Optional":
		return ir.Assertion{Kind: ir.AssertOptional, Pos: pos}, true

	case "Persistence":
		return ir.Assertion{Kind: ir.AssertPersistence, Pos: pos}, true

	case "StoredAs":
		t, ok := p.typeArg(call, 0)
		if !ok {
			return ir.Assertion{}, false
		}
		return ir.Assertion{Kind: ir.AssertStoredAs, Pos: pos, DomainType: typeName(t)}, true
	}
	return ir.Assertion{}, false
}

func isAssertionName(name string) bool {
	switch name {
	case "Satisfies", "Transition", "External", "Help", "Term", "Rationale", "Waive", "Draft", "Optional", "Persistence", "StoredAs":
		return true
	}
	return false
}

func argAt(call *ast.CallExpr, i int) ast.Expr {
	if i < len(call.Args) {
		return call.Args[i]
	}
	return nil
}

// typeArg returns the i-th type argument of a generic call.
//
// This is what makes spec.For[T] work: the type parameter is resolved by the Go
// type checker, so a typo is a real compile error, not a speclink finding.
func (p *Package) typeArg(call *ast.CallExpr, i int) (types.Type, bool) {
	id := calleeIdent(call.Fun)
	if id == nil {
		return nil, false
	}
	inst, ok := p.pkg.TypesInfo.Instances[id]
	if !ok || inst.TypeArgs == nil || inst.TypeArgs.Len() <= i {
		return nil, false
	}
	return inst.TypeArgs.At(i), true
}

// objectName resolves an expression to the qualified name of the object it
// refers to, seeing through the address-of used by ForVar.
func (p *Package) objectName(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.UnaryExpr:
		return p.objectName(x.X)
	case *ast.ParenExpr:
		return p.objectName(x.X)
	}
	id := calleeIdent(e)
	if id == nil {
		return ""
	}
	obj := p.pkg.TypesInfo.Uses[id]
	if obj == nil {
		if def := p.pkg.TypesInfo.Defs[id]; def != nil {
			obj = def
		}
	}
	if obj == nil {
		return ""
	}
	if obj.Pkg() != nil {
		return obj.Pkg().Path() + "." + obj.Name()
	}
	return obj.Name()
}

// stringArg folds a constant string expression. Constants resolve, computed
// values do not — and must not, since the whitelist forbids expressions.
func (p *Package) stringArg(e ast.Expr) (string, bool) {
	if e == nil {
		return "", false
	}
	tv, ok := p.pkg.TypesInfo.Types[e]
	if !ok || tv.Value == nil || tv.Value.Kind() != constant.String {
		return "", false
	}
	return constant.StringVal(tv.Value), true
}

// checkFieldExists verifies the string field name of ForField against the type.
//
// This is the only reference in the language that the Go compiler does not
// check, so speclink has to. The check is exact: it walks the struct fields.
func (p *Package) checkFieldExists(t types.Type, field string, at ast.Expr, out *diag.Set) {
	if field == "" {
		return
	}
	st, ok := underlyingStruct(t)
	if !ok {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseBinding, 4),
			Pos:  p.pos(at.Pos()),
			What: shortName(t) + " is not a struct type.",
			Why:  "spec.ForField binds to a struct field.",
			How:  "Use spec.For[" + shortName(t) + "](…) to bind to the type itself.",
		})
		return
	}
	for i := 0; i < st.NumFields(); i++ {
		if st.Field(i).Name() == field {
			return
		}
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseBinding, 5),
		Pos:  p.pos(at.Pos()),
		What: shortName(t) + " has no field " + field + ".",
		Why:  "The field name of spec.ForField is a string and is the only reference the Go compiler cannot check, so speclink checks it.",
		How:  "Correct the field name, or bind to the type with spec.For[" + shortName(t) + "](…).",
	})
}

func underlyingStruct(t types.Type) (*types.Struct, bool) {
	if t == nil {
		return nil, false
	}
	st, ok := t.Underlying().(*types.Struct)
	return st, ok
}

func typeName(t types.Type) string {
	if t == nil {
		return ""
	}
	if named, ok := t.(*types.Named); ok {
		obj := named.Obj()
		if obj.Pkg() != nil {
			return obj.Pkg().Path() + "." + obj.Name()
		}
		return obj.Name()
	}
	return t.String()
}

// shortName renders a type as a reader would write it in this package, e.g.
// "Thing" instead of "example.com/bad/sales.Thing". Diagnostics quote what the
// author would type, not what the type checker prints.
func shortName(t types.Type) string {
	full := typeName(t)
	if i := lastIndexByte(full, '.'); i >= 0 {
		return full[i+1:]
	}
	return full
}

// declTarget resolves the argument of spec.ForDecl to its declaration and
// derives the target kind from the type checker.
//
// The kind is never taken from the directive the author picked. types.Info
// already knows whether the identifier denotes a function, a variable or a
// constant, so there is no way for the annotation to contradict the code.
func (p *Package) declTarget(e ast.Expr, out *diag.Set) (ir.Target, bool) {
	obj := p.objectOf(e)
	if obj == nil {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseBinding, 8),
			Pos:  p.pos(e.Pos()),
			What: "spec.ForDecl requires a declared function, variable or constant.",
			Why:  "The argument names the construct being annotated. A literal or an expression names nothing, so the binding would have no target at all.",
			How:  "Pass a package level identifier, e.g. spec.ForDecl(PermSubmitQuote, …).",
		})
		return ir.Target{}, false
	}

	kind, ok := targetKindOf(obj)
	if !ok {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseBinding, 9),
			Pos:  p.pos(e.Pos()),
			What: obj.Name() + " is not a function, variable or constant.",
			Why:  "spec.ForDecl binds to a declaration. A type is bound with spec.For[T], a struct field with spec.ForField[T].",
			How:  "Use spec.For[" + obj.Name() + "](…) if you meant the type.",
		})
		return ir.Target{}, false
	}

	pkgPath := p.PkgPath()
	if obj.Pkg() != nil {
		pkgPath = obj.Pkg().Path()
	}
	return ir.Target{Kind: kind, Package: pkgPath, Name: qualified(obj)}, true
}

// targetKindOf maps a resolved object to its target kind.
func targetKindOf(obj types.Object) (ir.TargetKind, bool) {
	switch o := obj.(type) {
	case *types.Func:
		return ir.TargetFunc, true
	case *types.Const:
		return ir.TargetConst, true
	case *types.Var:
		// Struct fields are also *types.Var but are never package level, and
		// they are bound with spec.ForField instead.
		if o.IsField() {
			return 0, false
		}
		return ir.TargetVar, true
	}
	return 0, false
}

// objectOf resolves an expression to the object it denotes, or nil when the
// expression denotes no declaration.
func (p *Package) objectOf(e ast.Expr) types.Object {
	switch x := e.(type) {
	case *ast.ParenExpr:
		return p.objectOf(x.X)
	case *ast.UnaryExpr:
		// The address-of operator is rejected by the whitelist. Resolving
		// through it anyway keeps that one mistake to one message instead of
		// letting the unresolvable target pile a second finding on top.
		return p.objectOf(x.X)
	case *ast.Ident:
		return p.lookup(x)
	case *ast.SelectorExpr:
		return p.lookup(x.Sel)
	case *ast.IndexExpr:
		return p.objectOf(x.X)
	case *ast.IndexListExpr:
		return p.objectOf(x.X)
	}
	return nil
}

func (p *Package) lookup(id *ast.Ident) types.Object {
	if obj := p.pkg.TypesInfo.Uses[id]; obj != nil {
		return obj
	}
	return p.pkg.TypesInfo.Defs[id]
}

func qualified(obj types.Object) string {
	if obj.Pkg() != nil {
		return obj.Pkg().Path() + "." + obj.Name()
	}
	return obj.Name()
}

// extent records where a declaration begins and ends, and what it says.
//
// The text is hashed rather than kept: nothing downstream wants to print a
// declaration, and everything wants to know whether it is the one somebody
// looked at.
func (p *Package) extent(c *ir.Construct, start, end token.Pos) {
	fset := p.pkg.Fset
	from, to := fset.Position(start), fset.Position(end)
	if from.Filename != to.Filename {
		return
	}
	c.EndLine = to.Line

	body, err := p.fileBytes(from.Filename)
	if err != nil || to.Offset > len(body) || from.Offset > to.Offset {
		return
	}
	sum := sha256.Sum256(body[from.Offset:to.Offset])
	c.Fingerprint = hex.EncodeToString(sum[:])
}

// fileBytes reads a source file once per run.
//
// A package holds many declarations and each would otherwise re-read the whole
// file. The cache is per package, which is where the loop is.
func (p *Package) fileBytes(name string) ([]byte, error) {
	if p.files == nil {
		p.files = map[string][]byte{}
	}
	if body, ok := p.files[name]; ok {
		return body, nil
	}
	body, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	p.files[name] = body
	return body, nil
}

// FuncExtent returns the source range of a top level function of this package.
func (p *Package) FuncExtent(name string) (start, end token.Pos, ok bool) {
	for _, f := range p.pkg.Syntax {
		if p.isGeneratedByUs(f) {
			continue
		}
		for _, decl := range f.Decls {
			fd, isFunc := decl.(*ast.FuncDecl)
			if !isFunc || fd.Recv != nil || fd.Name.Name != name {
				continue
			}
			return fd.Pos(), fd.End(), true
		}
	}
	return 0, 0, false
}

// FoldInto adds a second range to an existing fingerprint.
//
// Order matters and is the caller's: the same two ranges hashed the other way
// round would be a different fingerprint for the same code.
func (p *Package) FoldInto(c *ir.Construct, start, end token.Pos) {
	fset := p.pkg.Fset
	from, to := fset.Position(start), fset.Position(end)
	body, err := p.fileBytes(from.Filename)
	if err != nil || to.Offset > len(body) || from.Offset > to.Offset {
		return
	}
	sum := sha256.Sum256(append([]byte(c.Fingerprint+"\x00"), body[from.Offset:to.Offset]...))
	c.Fingerprint = hex.EncodeToString(sum[:])
	if to.Line > c.EndLine {
		c.EndLine = to.Line
	}
}
