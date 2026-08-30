package ir

import "sort"

// PackageGraph is the module's own code and how it depends on itself.
//
// The import graph decides almost everything this tool checks — which layer
// may reach which, where a technology choice is made, whether a context is
// still separate — and it existed nowhere as a value. Each rule walked the
// imports again for its own question and threw the answer away, so nothing
// could draw the shape those rules were defending.
//
// Only this module. An edge to the standard library or to a dependency is not
// a fact about the architecture of this system, and drawing them would bury
// the handful of edges that are.
type PackageGraph struct {
	Nodes []PackageNode
	Edges []PackageEdge
}

// Declared reports whether anything is known.
func (g PackageGraph) Declared() bool { return len(g.Nodes) > 0 }

// PackageLayer is where a package sits in the architecture.
type PackageLayer string

const (
	// LayerDomain is the business rules of a context.
	LayerDomain PackageLayer = "domain"
	// LayerPresentation is a way in: a REST or CLI package.
	LayerPresentation PackageLayer = "presentation"
	// LayerAdapter is the far side of a port.
	LayerAdapter PackageLayer = "adapter"
	// LayerEntry is where a program is assembled.
	LayerEntry PackageLayer = "entry"
	// LayerInfra is a technical helper that knows nothing of the business.
	LayerInfra PackageLayer = "infrastructure"
	// LayerSpec is a package that declares the specification itself — the
	// requirements, the processes, the topology. It is part of the module and
	// no part of the system, and lumping it in with code nobody could place
	// makes a picture full of alarms about the one thing that is exactly where
	// it should be.
	LayerSpec PackageLayer = "specification"
	// LayerOther is a package the layout does not classify. It is named
	// rather than dropped: a package nobody can place is a thing a reader
	// should see, not a thing the drawing should hide.
	LayerOther PackageLayer = "other"
)

// PackageNode is one package of this module.
type PackageNode struct {
	// Path is the import path.
	Path string
	// Dir is the path relative to the module root, which is what a reader
	// recognises and what every rule reports against.
	Dir string
	// Context is the bounded context it belongs to, empty outside one.
	Context string
	Layer   PackageLayer
	// Measured says the run reported on this package. An unmeasured node is
	// drawn, because a dependency on code nobody looked at is exactly the edge
	// a reviewer wants to see, but it must not be shown as though it had been
	// checked.
	Measured bool
}

// PackageEdge is one import, from one package of this module to another.
type PackageEdge struct {
	From, To string
}

// Sort puts the graph in a stable order, so two runs over one module produce
// the same drawing and a diff of it is readable.
func (g *PackageGraph) Sort() {
	sort.Slice(g.Nodes, func(i, j int) bool { return g.Nodes[i].Dir < g.Nodes[j].Dir })
	sort.Slice(g.Edges, func(i, j int) bool {
		if g.Edges[i].From != g.Edges[j].From {
			return g.Edges[i].From < g.Edges[j].From
		}
		return g.Edges[i].To < g.Edges[j].To
	})
}

// Contexts lists the bounded contexts present, in order.
func (g PackageGraph) Contexts() []string {
	seen := map[string]bool{}
	var out []string
	for _, n := range g.Nodes {
		if n.Context == "" || seen[n.Context] {
			continue
		}
		seen[n.Context] = true
		out = append(out, n.Context)
	}
	sort.Strings(out)
	return out
}

// Crossings are the edges that leave a bounded context for another one.
//
// This is the number that says whether the contexts are still separate. It is
// computed here rather than drawn from the picture, because a reader counting
// arrows on a diagram is doing arithmetic the document should have done.
func (g PackageGraph) Crossings() []PackageEdge {
	ctx := map[string]string{}
	for _, n := range g.Nodes {
		ctx[n.Path] = n.Context
	}
	var out []PackageEdge
	for _, e := range g.Edges {
		from, to := ctx[e.From], ctx[e.To]
		if from != "" && to != "" && from != to {
			out = append(out, e)
		}
	}
	return out
}
