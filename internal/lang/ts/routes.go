package ts

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/ts/syntax"
)

// Route is one entry of a vue-router route table.
type Route struct {
	Router *Module
	// Path is the full path, the paths of the parents included.
	Path string
	// Component is the module the route shows, nil for a redirect or when it
	// could not be resolved.
	Component *Module
	// Unresolved records that a component was named and could not be found,
	// which is a different thing from a route that names none.
	Unresolved bool
	Pos        syntax.Pos
}

// Routes reads every route table built with createRouter.
//
// Only literal tables are read. A route assembled at run time is an address
// no catalogue can mention, and it is reported as such by the endpoint rules
// rather than guessed at here.
func (p *Project) Routes() []Route {
	var out []Route
	for _, m := range p.Modules {
		if p.Layer(m) != LayerUI {
			continue
		}
		for _, it := range m.Items() {
			var candidates []syntax.Range
			switch it.Kind {
			case syntax.ItemVar:
				for _, v := range it.Vars {
					candidates = append(candidates, v.Init)
				}
			case syntax.ItemExpr:
				candidates = append(candidates, it.Body)
			}
			for _, r := range candidates {
				call := it.File.Value(r)
				if call.Kind != syntax.ValCall || call.Last() != "createRouter" || len(call.Elems) == 0 {
					continue
				}
				routes, ok := call.Elems[0].Get("routes")
				if !ok {
					continue
				}
				if routes.Kind == syntax.ValIdent {
					routes = p.constValue(m, routes.Last())
				}
				out = append(out, p.readRoutes(m, it.File, routes, "")...)
			}
		}
	}
	return out
}

// constValue returns the initializer of a module level constant.
func (p *Project) constValue(m *Module, name string) syntax.Value {
	for _, it := range m.Items() {
		for _, v := range it.Vars {
			if v.Name == name {
				return it.File.Value(v.Init)
			}
		}
	}
	return syntax.Value{}
}

func (p *Project) readRoutes(m *Module, f *syntax.File, routes syntax.Value, parent string) []Route {
	var out []Route
	for _, e := range routes.Elems {
		if e.Kind != syntax.ValObject {
			out = append(out, Route{Router: m, Pos: e.Pos})
			continue
		}
		r := Route{Router: m, Pos: e.Pos}
		if pv, ok := e.Get("path"); ok && pv.Kind == syntax.ValString {
			r.Path = joinRoute(parent, pv.Str)
		}
		if cv, ok := e.Get("component"); ok {
			r.Component = p.component(m, f, cv)
			r.Unresolved = r.Component == nil
		}
		if r.Component != nil || r.Unresolved || r.Path == "" {
			out = append(out, r)
		}
		if children, ok := e.Get("children"); ok && children.Kind == syntax.ValArray {
			out = append(out, p.readRoutes(m, f, children, r.Path)...)
		}
	}
	return out
}

func joinRoute(parent, child string) string {
	if strings.HasPrefix(child, "/") || parent == "" {
		return child
	}
	return strings.TrimSuffix(parent, "/") + "/" + child
}

// component resolves the module a route shows: an imported component, or a
// lazy import('…').
func (p *Project) component(m *Module, f *syntax.File, v syntax.Value) *Module {
	if v.Kind == syntax.ValIdent {
		if b, ok := m.bindings()[v.Last()]; ok {
			return b.ref.Target
		}
		return nil
	}
	toks := f.Tokens(v.Range)
	for i := 0; i+2 < len(toks); i++ {
		if toks[i].Is("import") && toks[i+1].Is("(") && toks[i+2].Kind == syntax.String {
			ref := &Ref{Import: &syntax.Import{Spec: toks[i+2].Text}}
			p.resolve(m, ref)
			return ref.Target
		}
	}
	return nil
}

// traceDepth bounds how far a route is followed into the components it shows.
const traceDepth = 12

// Endpoints implements lang.EndpointReader: every route is an address the
// application answers on, and what stands behind it is the use cases its view
// and the components below it construct.
func (p *Project) Endpoints() []ir.Endpoint {
	var out []ir.Endpoint
	for _, r := range p.Routes() {
		e := ir.Endpoint{
			Method:  "VIEW",
			Path:    r.Path,
			Package: r.Router.Dir,
			Pos:     r.Router.Pos(r.Pos),
		}
		if r.Component != nil {
			e.Handler = r.Component.Path
			e.UseCases, e.Truncated, e.LeftScope = p.trace(r.Component)
		} else {
			e.Handler = "a component this run could not resolve"
		}
		out = append(out, e)
	}
	return out
}

// trace collects the use cases a component constructs, directly or through the
// components and modules of the user interface it imports.
func (p *Project) trace(start *Module) (useCases []string, truncated, leftScope bool) {
	seen := map[*Module]bool{start: true}
	found := map[string]bool{}
	frontier := []*Module{start}
	for depth := 0; len(frontier) > 0; depth++ {
		if depth == traceDepth {
			truncated = true
			break
		}
		var next []*Module
		for _, m := range frontier {
			if !p.Layout.InScope(m.Dir) {
				leftScope = true
			}
			for _, name := range p.constructed(m) {
				found[name] = true
			}
			for _, ref := range m.Imports {
				t := ref.Target
				if t == nil || seen[t] || p.Layer(t) != LayerUI {
					continue
				}
				seen[t] = true
				next = append(next, t)
			}
		}
		frontier = next
	}
	for name := range found {
		useCases = append(useCases, name)
	}
	sort.Strings(useCases)
	return useCases, truncated, leftScope
}

// constructed returns the use cases a module constructs with new.
func (p *Project) constructed(m *Module) []string {
	var out []string
	for _, f := range m.Scripts {
		toks := f.Toks
		for i := 0; i+1 < len(toks); i++ {
			if !toks[i].Is("new") || toks[i+1].Kind != syntax.Ident {
				continue
			}
			b, ok := m.bindings()[toks[i+1].Text]
			if !ok || b.ref.Target == nil || p.Layer(b.ref.Target) != LayerUseCase {
				continue
			}
			if sym, ok := p.symbol(m, []string{toks[i+1].Text}); ok {
				out = append(out, sym)
			}
		}
	}
	return out
}

// Locale is one translation catalogue.
type Locale struct {
	Path string
	Abs  string
	Keys map[string]bool
}

// Locales reads the JSON catalogues in every locales directory of the user
// interface.
func (p *Project) Locales() []Locale {
	var out []Locale
	var paths []string
	for rel := range p.files {
		if !strings.HasSuffix(rel, ".json") || path.Base(path.Dir(rel)) != "locales" {
			continue
		}
		if under(rel, p.srcPath(dirUI)) {
			paths = append(paths, rel)
		}
	}
	sort.Strings(paths)
	for _, rel := range paths {
		abs := filepath.Join(p.Root, filepath.FromSlash(rel))
		data, err := os.ReadFile(abs)
		if err != nil {
			continue
		}
		var tree map[string]any
		if json.Unmarshal(data, &tree) != nil {
			continue
		}
		keys := map[string]bool{}
		flatten("", tree, keys)
		out = append(out, Locale{Path: rel, Abs: abs, Keys: keys})
	}
	return out
}

func flatten(prefix string, v any, out map[string]bool) {
	obj, ok := v.(map[string]any)
	if !ok {
		out[prefix] = true
		return
	}
	for k, child := range obj {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		flatten(key, child, out)
	}
}
