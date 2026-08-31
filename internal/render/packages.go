package render

import (
	"fmt"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// Packages is kept as the whole picture, for a module small enough to have
// one.
//
// The threshold is not a matter of taste. Past about thirty nodes a layout
// engine spends its effort on avoiding crossings rather than on showing
// structure, and scaled to the width of a page the result is a grey mesh with
// unreadable labels — worse than no picture, because a reader believes they
// have seen the architecture. Above it, the document draws a context map and
// one picture per context instead; see ContextMap and PackagesOf.
func Packages(g ir.PackageGraph, system string) string {
	g = systemOnly(g)
	b := &strings.Builder{}
	header(b, "packages", system)

	byContext := map[string][]ir.PackageNode{}
	var loose []ir.PackageNode
	for _, n := range g.Nodes {
		if n.Context == "" {
			loose = append(loose, n)
			continue
		}
		byContext[n.Context] = append(byContext[n.Context], n)
	}

	for _, ctx := range g.Contexts() {
		fmt.Fprintf(b, "package %s {\n", quoted(ctx))
		for _, n := range byContext[ctx] {
			fmt.Fprintf(b, "  %s\n", componentOf(n))
		}
		b.WriteString("}\n")
	}
	for _, n := range loose {
		fmt.Fprintf(b, "%s\n", componentOf(n))
	}

	b.WriteString("\n")
	writeEdges(b, g.Edges)

	legend(b, g)
	b.WriteString("@enduml\n")
	return b.String()
}

// Crowded reports whether the whole picture has stopped being readable.
//
// Measured on the module's own packages, because those are what gets drawn.
// The number is a judgement and is written down here rather than left in
// somebody's head: thirty nodes is about where a page-width drawing stops
// being legible at print size.
func Crowded(g ir.PackageGraph) bool {
	return len(systemOnly(g).Nodes) > 30
}

// systemOnly drops the packages that declare the specification.
//
// They are part of the module and no part of the system, and in a project that
// takes this tool seriously there are a lot of them — every requirement file
// is a package, each importing several others. Drawn together with the code
// they describe, they are most of the nodes and most of the arrows, and the
// architecture disappears underneath its own documentation.
//
// The count is reported in the prose instead, so nothing is hidden: a reader
// is told they exist and told why they are not on the page.
func systemOnly(g ir.PackageGraph) ir.PackageGraph {
	keep := map[string]bool{}
	var out ir.PackageGraph
	for _, n := range g.Nodes {
		if n.Layer == ir.LayerSpec {
			continue
		}
		keep[n.Path] = true
		out.Nodes = append(out.Nodes, n)
	}
	for _, e := range g.Edges {
		if keep[e.From] && keep[e.To] {
			out.Edges = append(out.Edges, e)
		}
	}
	return out
}

var layerOrder = []ir.PackageLayer{
	ir.LayerEntry, ir.LayerPresentation, ir.LayerDomain,
	ir.LayerAdapter, ir.LayerInfra, ir.LayerSpec, ir.LayerOther,
}

var layerFill = map[ir.PackageLayer]string{
	ir.LayerEntry:        "#E8EEF7",
	ir.LayerPresentation: "#EAF3E8",
	ir.LayerDomain:       "#FFFFFF",
	ir.LayerAdapter:      "#FBF0E4",
	ir.LayerInfra:        "#F2F2F2",
	ir.LayerSpec:         "#EFEAF5",
	// Only genuinely unplaceable code is coloured as a warning, because a
	// colour that fires on the ordinary case is a colour nobody reads.
	ir.LayerOther: "#FCE8E8",
}

func used(g ir.PackageGraph, l ir.PackageLayer) bool {
	for _, n := range g.Nodes {
		if n.Layer == l {
			return true
		}
	}
	return false
}

// componentOf draws one package.
//
// A package the run did not measure is marked, because a dependency on code
// nobody looked at is precisely the edge a reviewer wants to find, and a
// picture that draws it identically to a checked one hides that.
func componentOf(n ir.PackageNode) string {
	name := n.Dir
	if !n.Measured {
		name += "\\n<size:9><i>not measured</i></size>"
	}
	return fmt.Sprintf("component %s as %s <<%s>>", quoted(name), ident(n.Path), string(n.Layer))
}
