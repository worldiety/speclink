package golang

import (
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
)

// tracer follows a handler expression back to the work it performs.
//
// The strategy rests on one observation that makes the whole problem tractable:
// a use case is a named func type. It is therefore a *value* with a type the
// type checker knows, and it must be handed to whatever serves it. So the
// question "what does this route do" reduces to "which of the values reachable
// from this expression has a use case type" — which is a question about types,
// not about names.
//
// That is why there is no rule here for recognising middleware. A wrapper is
// `logging(auth(Submit(who, submit)))`, and descending into every argument of
// every call finds `submit` without ever needing to know that `logging` and
// `auth` were wrappers rather than the handler itself. Every heuristic that
// tried to classify wrappers by signature — takes a Handler, returns a Handler
// — would have been a rule about a convention, and conventions are what this
// tool exists to stop relying on.
type tracer struct {
	byPkg    map[string]*Package
	useCases map[string]bool

	// seen guards against following a cycle through a package level variable.
	// It is per trace rather than shared between endpoints: a constructor used
	// by two routes must be followed for both, or the second route silently
	// loses the use case the first one claimed.
	seen  map[types.Object]bool
	found map[string]bool

	truncated bool
}

func (t *tracer) names() []string {
	out := make([]string, 0, len(t.found))
	for n := range t.found {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

// expr follows one expression.
//
// The order of the cases is the whole design. A use case reached by its type is
// taken and the descent stops there, because the shallowest evidence is the
// evidence least likely to be wrong: an expression that *is* a use case is a
// fact, whereas a use case found four frames inside a shared helper is a guess
// about what that helper was doing on this route's behalf.
func (t *tracer) expr(p *Package, e ast.Expr, depth int) {
	if e == nil || p == nil {
		return
	}
	if depth <= 0 {
		t.truncated = true
		return
	}

	if tv, ok := p.pkg.TypesInfo.Types[e]; ok {
		if name := typeName(tv.Type); t.useCases[name] {
			t.found[name] = true
			return
		}
	}

	switch x := e.(type) {
	case *ast.ParenExpr:
		t.expr(p, x.X, depth)

	case *ast.CallExpr:
		// Arguments first, and the callee's body only if they yielded
		// nothing. `Submit(who, submit)` answers the question at the call
		// site; opening Submit as well would add whatever else it happens to
		// touch, for no gain.
		before := len(t.found)
		for _, arg := range x.Args {
			t.expr(p, arg, depth-1)
		}
		if len(t.found) == before {
			t.expr(p, x.Fun, depth-1)
		}

	case *ast.FuncLit:
		t.body(p, x.Body, depth-1)

	case *ast.Ident:
		t.object(p, p.pkg.TypesInfo.Uses[x], depth)

	case *ast.SelectorExpr:
		if obj := p.pkg.TypesInfo.Uses[x.Sel]; obj != nil {
			t.object(p, obj, depth)
			return
		}
		t.expr(p, x.X, depth-1)

	case *ast.IndexExpr:
		// A generic instantiation used as a value.
		t.expr(p, x.X, depth)
	case *ast.IndexListExpr:
		t.expr(p, x.X, depth)
	}
}

// object follows a named thing to wherever it was written down.
//
// A function is opened. A variable is followed to what it was assigned, which
// is what catches the handler built once and mounted later. Anything else — a
// parameter whose type was not a use case, a constant, a field — has nothing
// further to say, and saying so is the point: the trace ends where the evidence
// ends rather than where the tool ran out of patience.
func (t *tracer) object(p *Package, obj types.Object, depth int) {
	if obj == nil || t.seen[obj] {
		return
	}
	t.seen[obj] = true

	if obj.Pkg() == nil {
		return
	}
	owner, ok := t.byPkg[obj.Pkg().Path()]
	if !ok {
		// Outside the loaded set: a framework, or the standard library. There
		// is nothing to read, and nothing to claim about it.
		return
	}

	switch o := obj.(type) {
	case *types.Func:
		if decl := owner.funcDeclAt(o.Pos()); decl != nil {
			t.body(owner, decl.Body, depth-1)
		}
	case *types.Var:
		if rhs := owner.assignedTo(o); rhs != nil {
			t.expr(owner, rhs, depth-1)
		}
	}
}

// body looks through a function for the work it does.
//
// Only calls, and only their callees, not their arguments: inside a body the
// question has changed. At a registration the use case is being *passed*, so
// the arguments are where it lives; inside a handler the use case is being
// *called*, so the callee is. Reading arguments here as well would collect the
// input a use case was given as though it were another use case.
func (t *tracer) body(p *Package, b *ast.BlockStmt, depth int) {
	if b == nil {
		return
	}
	if depth <= 0 {
		t.truncated = true
		return
	}
	ast.Inspect(b, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		t.expr(p, call.Fun, depth)
		return true
	})
}

// funcDeclAt finds the declaration that starts at a position, including a
// method's.
func (p *Package) funcDeclAt(pos token.Pos) *ast.FuncDecl {
	for _, file := range p.pkg.Syntax {
		if pos < file.Pos() || pos > file.End() {
			continue
		}
		for _, decl := range file.Decls {
			if fd, ok := decl.(*ast.FuncDecl); ok && fd.Name.Pos() == pos {
				return fd
			}
		}
	}
	return nil
}

// assignedTo returns the single expression a variable was given.
//
// Single is deliberate. A variable assigned in two places is a variable whose
// value depends on which branch ran, and following either one would be
// reporting a guess as a fact. Better to find nothing and say so.
func (p *Package) assignedTo(v *types.Var) ast.Expr {
	var found ast.Expr
	var count int

	record := func(lhs, rhs []ast.Expr) {
		if len(lhs) != len(rhs) {
			return // a call spread across several results: nothing single to follow
		}
		for i, l := range lhs {
			id, ok := l.(*ast.Ident)
			if !ok {
				continue
			}
			obj := p.pkg.TypesInfo.Defs[id]
			if obj == nil {
				obj = p.pkg.TypesInfo.Uses[id]
			}
			if obj == v {
				found, count = rhs[i], count+1
			}
		}
	}

	for _, file := range p.pkg.Syntax {
		ast.Inspect(file, func(n ast.Node) bool {
			switch s := n.(type) {
			case *ast.AssignStmt:
				record(s.Lhs, s.Rhs)
			case *ast.ValueSpec:
				lhs := make([]ast.Expr, 0, len(s.Names))
				for _, n := range s.Names {
					lhs = append(lhs, n)
				}
				record(lhs, s.Values)
			}
			return true
		})
	}
	if count != 1 {
		return nil
	}
	return found
}

// render prints an expression the way it was written, for a diagnostic that
// points at something the reader recognises.
func render(p *Package, e ast.Expr) string {
	var sb strings.Builder
	if err := printer.Fprint(&sb, p.pkg.Fset, e); err != nil {
		return ""
	}
	return sb.String()
}

// discard is a sink for the phases that are re run for their result rather than
// their findings.
func discard() *diag.Set { return &diag.Set{} }
