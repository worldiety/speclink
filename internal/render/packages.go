package render

import (
	"fmt"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// Packages draws the module's own packages and the dependencies between them.
//
// Grouped by bounded context, because that is the boundary the architecture
// actually defends: an arrow inside a group is ordinary, and one crossing
// between groups is the thing a reviewer is looking for. Drawing a flat list of
// packages would bury that distinction in the middle of the picture.
//
// Only this module's packages. An edge to the standard library or to a
// dependency is not a fact about the architecture of this system, and there are
// hundreds of them.
func Packages(g ir.PackageGraph, system string) string {
	g = systemOnly(g)
	b := &strings.Builder{}
	b.WriteString("@startuml packages\n")
	b.WriteString("skinparam componentStyle rectangle\n")
	b.WriteString("skinparam shadowing false\n")
	b.WriteString("skinparam defaultTextAlignment center\n")
	b.WriteString("skinparam package {\n  BorderColor #999999\n  BackgroundColor #FCFCFC\n}\n")
	// One colour per layer, and a legend, so the picture can be read without
	// the surrounding prose — which is how a diagram in an audit document is
	// usually met.
	for _, l := range layerOrder {
		fmt.Fprintf(b, "skinparam component<<%s>> {\n  BackgroundColor %s\n  BorderColor #777777\n}\n",
			string(l), layerFill[l])
	}
	fmt.Fprintf(b, "title %s\n\n", quoted(system))

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
	seen := map[string]bool{}
	for _, e := range g.Edges {
		line := fmt.Sprintf("%s --> %s\n", ident(e.From), ident(e.To))
		if seen[line] {
			continue
		}
		seen[line] = true
		b.WriteString(line)
	}

	b.WriteString("\nlegend right\n")
	for _, l := range layerOrder {
		if !used(g, l) {
			continue
		}
		fmt.Fprintf(b, "  <back:%s>   </back> %s\n", strings.TrimPrefix(layerFill[l], "#"), string(l))
	}
	b.WriteString("endlegend\n")
	b.WriteString("@enduml\n")
	return b.String()
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
