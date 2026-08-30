package golang

import (
	"go/ast"
	"path"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// PackageGraph reports this module's own packages and how they depend on each
// other.
//
// The graph was always there and never assembled. Every architectural rule
// walks the imports for its own question and discards the answer, which is why
// speclink could enforce a layering it could not draw.
//
// Built over the whole module, not the measured subset. Which packages exist
// and what they import is a question about the module; answering it from a
// narrowed load would draw a system with holes in it and give no sign that the
// holes were the loader's rather than the code's. What the scope decides is
// which nodes are marked as measured.
func (m *Model) PackageGraph() ir.PackageGraph {
	module, _ := ModulePath(m.All)
	if module == "" {
		// Without the module path an import cannot be told from a dependency,
		// and a graph that mixes them is worse than none.
		return ir.PackageGraph{}
	}

	measured := m.Scope()
	own := map[string]bool{}
	for _, p := range m.All {
		if p.isTest {
			continue
		}
		own[p.PkgPath()] = true
	}

	var g ir.PackageGraph
	for _, p := range m.All {
		if p.isTest || !own[p.PkgPath()] {
			continue
		}
		rel := p.relDir(m.Root)
		ctx := m.contextOf(rel)
		layer := m.layerOfDir(rel, ctx)
		if p.declaresSpec() {
			// It is part of the module and no part of the system. Left to the
			// path rules it comes out unclassifiable, which puts a warning on
			// the one thing that is exactly where it belongs.
			layer = ir.LayerSpec
		}
		g.Nodes = append(g.Nodes, ir.PackageNode{
			Path:     p.PkgPath(),
			Dir:      rel,
			Context:  ctx,
			Layer:    layer,
			Measured: measured[p.PkgPath()],
		})
		for _, imp := range p.imports() {
			if !own[imp] || imp == p.PkgPath() {
				continue
			}
			g.Edges = append(g.Edges, ir.PackageEdge{From: p.PkgPath(), To: imp})
		}
	}
	g.Sort()
	return g
}

// declaresSpec reports whether this package holds the specification rather
// than the system: requirements, processes or topology.
func (p *Package) declaresSpec() bool {
	return len(p.requirementFiles) > 0 || len(p.processFiles) > 0 || len(p.topologyFiles) > 0
}

// declaresSpec reports whether one file is a specification declaration.
//
// The package level answer is too coarse for a rule about imports: a package
// may hold a topology file beside ordinary code, and only the declaration
// itself earns the latitude.
func (p *Package) declaresSpecFile(f *ast.File) bool {
	for _, set := range [][]*ast.File{p.requirementFiles, p.processFiles, p.topologyFiles} {
		for _, cand := range set {
			if cand == f {
				return true
			}
		}
	}
	return false
}

// contextOf names the bounded context a directory belongs to, if any.
func (m *Model) contextOf(rel string) string {
	root := m.Layout.ContextRoot
	if root == "" || !strings.HasPrefix(rel, root+"/") {
		return ""
	}
	name, _, _ := strings.Cut(strings.TrimPrefix(rel, root+"/"), "/")
	return name
}

// layerOfDir places a package, using the same path knowledge the rules use.
//
// It has to agree with them. A drawing that classifies a package one way while
// the rule enforcing its dependencies classifies it another is a drawing that
// will eventually be used to argue that a finding is wrong.
func (m *Model) layerOfDir(rel, ctx string) ir.PackageLayer {
	cfg := m.Layout
	if cfg.UnderCmdRoot(rel) {
		return ir.LayerEntry
	}
	for _, root := range append(append([]string(nil), cfg.InfraRoots...), cfg.FoundationRoot) {
		if root == "" {
			continue
		}
		if rel == root || strings.HasPrefix(rel, path.Clean(root)+"/") {
			return ir.LayerInfra
		}
	}
	if ctx == "" {
		return ir.LayerOther
	}
	if !m.Layered {
		// A profile with no presentation or adapter directories cannot tell
		// them apart, and inventing the distinction here would put a layer on
		// the page that nothing enforces.
		return ir.LayerDomain
	}
	switch layerOf(rel, ctx, cfg) {
	case layerPresentation:
		return ir.LayerPresentation
	case layerAdapter:
		return ir.LayerAdapter
	default:
		return ir.LayerDomain
	}
}
