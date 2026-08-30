package golang

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"sort"
	"strconv"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// How far a trace will follow a handler before it gives up.
//
// Deep enough for the shapes that occur — a registration, a constructor, a
// wrapper or three, the call — and shallow enough that a cycle through a
// package level variable cannot turn a run into a hang. Exceeding it is
// recorded on the endpoint rather than swallowed, because a trace that stopped
// looking and a handler that has nothing behind it must never report alike.
const traceDepth = 8

// Endpoints recognises every address the system answers on.
//
// Registrations are looked for in every measured package rather than in a
// designated one. A route is mounted wherever somebody found it convenient,
// and a recogniser that insisted on a particular package would only be
// describing the fixture it was written against.
func (m *Model) Endpoints() []ir.Endpoint {
	byPkg := map[string]*Package{}
	for _, p := range m.All {
		byPkg[p.PkgPath()] = p
	}

	useCases := map[string]bool{}
	for _, c := range m.Constructs(discard()) {
		if c.Kind.PerformsWork() {
			useCases[c.Name] = true
		}
	}

	var out []ir.Endpoint
	for _, p := range m.Measured {
		for _, site := range p.registrations() {
			t := &tracer{byPkg: byPkg, useCases: useCases, seen: map[types.Object]bool{}, found: map[string]bool{}}
			t.expr(p, site.handler, traceDepth)

			e := ir.Endpoint{
				Method:    site.method,
				Path:      site.path,
				Handler:   render(p, site.handler),
				UseCases:  t.names(),
				Truncated: t.truncated,
				Pos:       p.pos(site.pos),
			}
			// Request and Response stay empty here. They could be guessed from
			// the use case's own signature, and for the shape this fixture
			// happens to have the guess would even be right — but a
			// presentation layer that maps a DTO onto a command would make it
			// silently wrong, and these two fields are the ones that become a
			// frozen promise. A wire type is filled in by the dialect that
			// states it or not at all.
			out = append(out, e)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Path != out[j].Path {
			return out[i].Path < out[j].Path
		}
		return out[i].Method < out[j].Method
	})
	return out
}

// site is one call that mounts a route.
type site struct {
	method, path string
	// handler is the expression that answers on it. Empty path means the
	// pattern could not be read, which is a finding rather than a skip.
	handler ast.Expr
	pos     token.Pos
	// pattern is the unreadable expression, kept for the diagnostic.
	pattern ast.Expr
}

// registrations finds every call that mounts a route on the standard library's
// router.
//
// Recognised by the receiver's type rather than by the name of the variable, so
// that a router called `r`, `router` or `api` is found alike and a local named
// `mux` that is something else entirely is not.
func (p *Package) registrations() []site {
	var out []site
	for _, file := range p.pkg.Syntax {
		if p.isGeneratedByUs(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || len(call.Args) != 2 || !p.isMuxRegistration(call) {
				return true
			}
			s := site{handler: call.Args[1], pattern: call.Args[0], pos: call.Pos()}
			if pat, ok := p.stringValue(call.Args[0]); ok {
				s.method, s.path = splitPattern(pat)
			}
			out = append(out, s)
			return true
		})
	}
	return out
}

// isMuxRegistration reports whether a call mounts a route.
//
// Both spellings: the method on a router, and the package level pair that
// mounts on the default one. The second is worth recognising precisely because
// it is the careless spelling — a route on a global router is the kind that
// gets forgotten, and leaving it unrecognised would reward it.
func (p *Package) isMuxRegistration(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	switch sel.Sel.Name {
	case "Handle", "HandleFunc":
	default:
		return false
	}

	// The package level pair, mounting on the default router.
	if id, ok := sel.X.(*ast.Ident); ok {
		if pn, ok := p.pkg.TypesInfo.Uses[id].(*types.PkgName); ok {
			return pn.Imported().Path() == "net/http"
		}
	}

	tv, ok := p.pkg.TypesInfo.Types[sel.X]
	if !ok {
		return false
	}
	return typeName(deref(tv.Type)) == "net/http.ServeMux"
}

// stringValue reads a compile time string, whether it is written as a literal
// or reached through a constant.
//
// A constant is worth following because a project that keeps its paths in one
// place is doing the right thing, and a recogniser that only understood
// literals would report the tidy project as unreadable and the sloppy one as
// fine.
func (p *Package) stringValue(e ast.Expr) (string, bool) {
	if tv, ok := p.pkg.TypesInfo.Types[e]; ok && tv.Value != nil && tv.Value.Kind() == constant.String {
		return constant.StringVal(tv.Value), true
	}
	if lit, ok := e.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		if s, err := strconv.Unquote(lit.Value); err == nil {
			return s, true
		}
	}
	return "", false
}

// splitPattern separates the method from the path.
//
// The router's own grammar: an optional method, an optional host, then the
// path. A pattern with no method answers to every method, and that is written
// down as such rather than guessed at, because "any" and "GET" are different
// promises to a client.
func splitPattern(pat string) (method, path string) {
	pat = strings.TrimSpace(pat)
	head, rest, ok := strings.Cut(pat, " ")
	if !ok || strings.Contains(head, "/") {
		return "", pat
	}
	return head, strings.TrimSpace(rest)
}

// deref removes one pointer, which is how a router is nearly always held.
func deref(t types.Type) types.Type {
	if ptr, ok := t.Underlying().(*types.Pointer); ok {
		return ptr.Elem()
	}
	return t
}
