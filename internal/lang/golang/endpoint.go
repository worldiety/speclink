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

	// Which types are use cases is a question about the module, not about the
	// scope, so it is asked of every loaded package rather than of the measured
	// ones. The same distinction the entry point rule already makes: where a
	// construct lives belongs to the scope, whether a type is a use case at all
	// does not.
	//
	// Reading it from Measured was wrong in a way that only showed under a
	// narrowed run, and showed as the worst possible reading. A route in the
	// presentation layer whose use case sits one package away reported as
	// having nothing accountable behind it — a door in the wall that no drawing
	// shows — when in truth the drawing was simply not in the operator's hand.
	world := *m
	world.Measured = m.All
	useCases := map[string]bool{}
	for _, c := range world.Constructs(discard()) {
		if c.Kind.PerformsWork() {
			useCases[c.Name] = true
		}
	}

	// A failure here is not fatal to the recogniser. Without the module path
	// the trace cannot tell our own unloaded packages from the standard
	// library, so it reports nothing as out of scope — the behaviour this had
	// before the distinction existed. Losing the address catalogue entirely
	// because go.mod could not be read would be the worse trade.
	module, _ := ModulePath(m.All)

	var out []ir.Endpoint
	for _, p := range m.Measured {
		for _, site := range p.registrations(m.Framework) {
			t := &tracer{
				byPkg:    byPkg,
				useCases: useCases,
				module:   module,
				seen:     map[types.Object]bool{},
				found:    map[string]bool{},
			}
			for _, e := range site.traced() {
				t.expr(p, e, traceDepth)
			}

			e := ir.Endpoint{
				Method:        site.method,
				Path:          site.path,
				Package:       p.PkgPath(),
				Handler:       render(p, site.handler),
				UseCases:      t.names(),
				Request:       site.request,
				Response:      site.response,
				RequestShape:  site.requestShape,
				ResponseShape: site.responseShape,
				ShapesStated:  site.shapesStated,
				Truncated:     t.truncated,
				LeftScope:     t.leftScope,
				Pos:           p.pos(site.pos),
			}
			// Request and Response are whatever the dialect stated and nothing
			// more. They could be guessed from the use case's own signature,
			// and for a handler that takes the wire shape straight through the
			// guess would even be right — but a presentation layer that maps a
			// request body onto a command makes it silently wrong, and these
			// are the two fields that become a frozen promise. A wire type is
			// filled in by the dialect that says it, or not at all.
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
	// trace holds the expressions the work may be reached through. Usually
	// just the handler; a fluent builder spreads it over the arguments of
	// several calls, and they are kept apart so that each is followed on its
	// own rather than one of them shadowing the rest.
	trace []ast.Expr
	// request and response are the wire shapes, filled only by a dialect that
	// states them, and shapesStated says whether this dialect is one of those.
	request, response           string
	requestShape, responseShape *ir.WireShape
	shapesStated                bool
	pos                         token.Pos
	// pattern is the unreadable expression, kept for the diagnostic.
	pattern ast.Expr
}

// traced returns the expressions to follow, defaulting to the handler.
func (s site) traced() []ast.Expr {
	if len(s.trace) > 0 {
		return s.trace
	}
	return []ast.Expr{s.handler}
}

// registrations finds every call that mounts a route.
//
// Both dialects are asked, always, rather than one being chosen by the profile.
// A project on a framework is still free to mount a route on the standard
// library's router, and that route is exactly the one worth finding: it is the
// address that sits outside whatever the framework generates documentation for.
func (p *Package) registrations(fw Framework) []site {
	return append(p.muxSites(), p.hapiSites(fw)...)
}

// muxSites finds every call that mounts a route on the standard library's
// router.
//
// Recognised by the receiver's type rather than by the name of the variable, so
// that a router called `r`, `router` or `api` is found alike and a local named
// `mux` that is something else entirely is not.
func (p *Package) muxSites() []site {
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
