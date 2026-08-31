package render

import (
	"fmt"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// ContextMap draws the bounded contexts and what depends on what.
//
// # Why the one picture had to become several
//
// A drawing of every package of a real system is not a drawing anybody reads.
// It has one node per package and one arrow per import, and past about thirty
// nodes the layout engine spends its effort on avoiding crossings rather than
// on showing structure. Scaled to the width of a page it becomes a grey mesh
// with unreadable labels — worse than no picture, because a reader believes
// they have seen the architecture.
//
// The split follows the boundary the architecture already defends. This map
// answers "what are the parts and which depends on which", which is the
// question somebody has on first meeting the system, and it has as many nodes
// as there are contexts. The inside of each is a picture of its own, met by
// somebody who has already decided which part they care about.
func ContextMap(g ir.PackageGraph, system string) string {
	g = systemOnly(g)

	b := &strings.Builder{}
	header(b, "context-map", system)

	// A context is drawn from the layers its packages occupy, so the map keeps
	// the colour scheme of the detailed drawings rather than inventing a
	// second one a reader has to learn.
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
		fmt.Fprintf(b, "component %s as %s <<%s>>\n",
			quoted(fmt.Sprintf("%s\\n<size:9>%s</size>", ctx,
				plural(len(byContext[ctx]), "package", "packages"))),
			contextIdent(ctx), string(ir.LayerDomain))
	}

	// Everything outside a context keeps its own node. These are the entry
	// points and the shared foundation, and collapsing them into one box would
	// hide the fact a reviewer is here for: which parts reach into which.
	for _, n := range loose {
		fmt.Fprintf(b, "%s\n", componentOf(n))
	}

	b.WriteString("\n")
	writeEdges(b, collapse(g))

	legend(b, g)
	b.WriteString("@enduml\n")
	return b.String()
}

// PackagesOf draws the inside of one bounded context.
//
// Everything it reaches outside itself is drawn as a single box per foreign
// context, because at this magnification the question is what this context
// depends on, not how that context is built. Drawing the far side in full
// would put the whole system back on the page one context at a time.
func PackagesOf(g ir.PackageGraph, context, system string) string {
	g = systemOnly(g)

	b := &strings.Builder{}
	header(b, "packages-"+ident(context), system+" — "+context)

	inside := map[string]bool{}
	for _, n := range g.Nodes {
		if n.Context == context {
			inside[n.Path] = true
		}
	}

	fmt.Fprintf(b, "package %s {\n", quoted(context))
	for _, n := range g.Nodes {
		if n.Context == context {
			fmt.Fprintf(b, "  %s\n", componentOf(n))
		}
	}
	b.WriteString("}\n")

	// The neighbours, each as one box, and only the ones actually touched.
	// A box for a context this one never reaches would be an arrow that is not
	// there, drawn as a thing that is.
	neighbours := map[string]bool{}
	for _, e := range g.Edges {
		switch {
		case inside[e.From] && !inside[e.To]:
			neighbours[outsideName(g, e.To)] = true
		case inside[e.To] && !inside[e.From]:
			neighbours[outsideName(g, e.From)] = true
		}
	}
	for _, name := range sortedKeys(neighbours) {
		fmt.Fprintf(b, "component %s as %s <<%s>>\n", quoted(name), ident(name), string(ir.LayerOther))
	}

	b.WriteString("\n")
	var edges []ir.PackageEdge
	for _, e := range g.Edges {
		switch {
		case inside[e.From] && inside[e.To]:
			edges = append(edges, e)
		case inside[e.From]:
			edges = append(edges, ir.PackageEdge{From: e.From, To: outsideName(g, e.To)})
		case inside[e.To]:
			edges = append(edges, ir.PackageEdge{From: outsideName(g, e.From), To: e.To})
		}
	}
	writeEdges(b, edges)

	legend(b, g)
	b.WriteString("@enduml\n")
	return b.String()
}

// outsideName is what a package outside the drawn context is collapsed into:
// its context where it has one, and its own directory where it has not.
//
// A package with no context is not collapsed, because those are the entry
// points and the foundation — the ones whose identity is the point.
func outsideName(g ir.PackageGraph, path string) string {
	for _, n := range g.Nodes {
		if n.Path != path {
			continue
		}
		if n.Context != "" {
			return n.Context
		}
		return n.Dir
	}
	return path
}

// collapse rewrites every edge onto the contexts its ends belong to and drops
// the ones that then point at themselves.
//
// The self edges are the ordinary case — a presentation importing its own
// domain — and on a map of the whole system they are noise: they say only that
// a context has an inside, which the reader assumes.
func collapse(g ir.PackageGraph) []ir.PackageEdge {
	name := map[string]string{}
	for _, n := range g.Nodes {
		if n.Context != "" {
			name[n.Path] = contextIdent(n.Context)
			continue
		}
		name[n.Path] = ident(n.Path)
	}

	var out []ir.PackageEdge
	for _, e := range g.Edges {
		from, to := name[e.From], name[e.To]
		if from == "" || to == "" || from == to {
			continue
		}
		out = append(out, ir.PackageEdge{From: from, To: to})
	}
	return out
}

// writeEdges draws each arrow once. A graph of imports has the same pair many
// times over, and the picture only ever needs to say it once.
func writeEdges(b *strings.Builder, edges []ir.PackageEdge) {
	seen := map[string]bool{}
	var lines []string
	for _, e := range edges {
		line := fmt.Sprintf("%s --> %s\n", ident(e.From), ident(e.To))
		if seen[line] {
			continue
		}
		seen[line] = true
		lines = append(lines, line)
	}
	sort.Strings(lines)
	for _, line := range lines {
		b.WriteString(line)
	}
}

// header is the styling every package drawing shares.
func header(b *strings.Builder, name, title string) {
	fmt.Fprintf(b, "@startuml %s\n", name)
	b.WriteString("skinparam componentStyle rectangle\n")
	b.WriteString("skinparam shadowing false\n")
	b.WriteString("skinparam defaultTextAlignment center\n")
	b.WriteString("skinparam package {\n  BorderColor #999999\n  BackgroundColor #FCFCFC\n}\n")
	for _, l := range layerOrder {
		fmt.Fprintf(b, "skinparam component<<%s>> {\n  BackgroundColor %s\n  BorderColor #777777\n}\n",
			string(l), layerFill[l])
	}
	fmt.Fprintf(b, "title %s\n\n", quoted(title))
}

// legend prints the colour key, and only for the layers on this page.
func legend(b *strings.Builder, g ir.PackageGraph) {
	b.WriteString("\nlegend right\n")
	for _, l := range layerOrder {
		if !used(g, l) {
			continue
		}
		fmt.Fprintf(b, "  <back:%s>   </back> %s\n", strings.TrimPrefix(layerFill[l], "#"), string(l))
	}
	b.WriteString("endlegend\n")
}

func contextIdent(ctx string) string { return ident("ctx/" + ctx) }

func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}
