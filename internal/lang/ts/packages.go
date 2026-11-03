package ts

import (
	"github.com/worldiety/speclink/internal/ir"
)

// PackageGraph implements lang.PackageGrapher: the directories below the
// source root and the imports between them.
//
// A directory is the unit, as a package is in Go. A module is too fine to draw
// — a Vue application has one per component — and the rules speak about
// directories anyway: every layer of the style is one.
func (p *Project) PackageGraph() ir.PackageGraph {
	var g ir.PackageGraph
	layers := map[string]ir.PackageLayer{}
	spec := map[string]bool{}

	for _, m := range p.Modules {
		l := p.Layer(m)
		switch l {
		case LayerTest:
			continue
		case LayerSpec, LayerVocabulary:
			if _, ok := layers[m.Dir]; !ok {
				spec[m.Dir] = true
			}
			continue
		case LayerOutside:
			continue
		}
		delete(spec, m.Dir)
		layers[m.Dir] = packageLayer(l, layers[m.Dir])
	}
	for dir := range spec {
		if _, code := layers[dir]; !code {
			layers[dir] = ir.LayerSpec
		}
	}

	for dir, layer := range layers {
		g.Nodes = append(g.Nodes, ir.PackageNode{
			Path:     dir,
			Dir:      dir,
			Layer:    layer,
			Measured: p.Layout.InScope(dir),
		})
	}

	seen := map[ir.PackageEdge]bool{}
	for _, m := range p.Modules {
		switch p.Layer(m) {
		case LayerTest, LayerOutside, LayerSpec, LayerVocabulary:
			continue
		}
		for _, ref := range m.Imports {
			t := ref.Target
			if t == nil || t.Dir == m.Dir {
				continue
			}
			if _, drawn := layers[t.Dir]; !drawn {
				continue
			}
			e := ir.PackageEdge{From: m.Dir, To: t.Dir}
			if !seen[e] {
				seen[e] = true
				g.Edges = append(g.Edges, e)
			}
		}
	}
	g.Sort()
	return g
}

// packageLayer places a directory by the modules in it. The entry point shares
// its directory with nothing else of the style, so the first classification
// that is not "other" wins.
func packageLayer(l Layer, have ir.PackageLayer) ir.PackageLayer {
	var mine ir.PackageLayer
	switch l {
	case LayerUseCase, LayerService, LayerLocator, LayerPort, LayerType:
		mine = ir.LayerDomain
	case LayerUI:
		mine = ir.LayerPresentation
	case LayerAdapter:
		mine = ir.LayerAdapter
	case LayerEntry:
		mine = ir.LayerEntry
	default:
		mine = ir.LayerOther
	}
	if have != "" && have != ir.LayerOther {
		return have
	}
	return mine
}
