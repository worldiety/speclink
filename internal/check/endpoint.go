package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// Rule IDs of the endpoint checks.
const (
	// RuleEndpointPatternUnreadable fires when a route is mounted under a
	// pattern this run cannot evaluate.
	RuleEndpointPatternUnreadable = "K20-ENDPOINT-PATTERN-UNREADABLE"
	// RuleEndpointNoUseCase fires when nothing accountable was found behind a
	// route.
	RuleEndpointNoUseCase = "K20-ENDPOINT-NO-USE-CASE"
	// RuleEndpointTraceTruncated fires when the trace gave up before it ran
	// out of places to look.
	RuleEndpointTraceTruncated = "K20-ENDPOINT-TRACE-TRUNCATED"
	// RuleEndpointDuplicate fires when two registrations claim one address.
	RuleEndpointDuplicate = "K20-ENDPOINT-DUPLICATE"
)

// EndpointReport is what the run can say about the surface the system exposes.
type EndpointReport struct {
	// Routes is how many registrations were recognised, and Traced how many of
	// them reach a use case. The pair is the figure that matters: a surface
	// with a hundred routes and sixty traced is not sixty percent documented,
	// it is forty routes nobody can account for.
	Routes, Traced int
	Endpoints      []ir.Endpoint
}

// Endpoints holds the exposed surface accountable.
//
// # Why this is recognised rather than declared
//
// The same reason a stored shape is. The code that mounts a route already
// states the method, the path and the handler, and a second declaration beside
// it would be a second source of one fact — free to disagree with the first,
// and certain to eventually. What makes a route a promise rather than an
// implementation detail is not being written down twice; it is being frozen.
//
// # Why a trace that failed is louder than one that found nothing interesting
//
// An endpoint is an address the world can reach. A route whose pattern cannot
// be read, or whose handler leads nowhere this can follow, is not a gap in the
// documentation — it is a door in the wall that no drawing shows. That is the
// most dangerous thing this tool can encounter, so it is reported, and reported
// as three distinct facts: the pattern was not constant, nothing accountable
// was behind it, or the tool stopped looking. The third is a defect in
// speclink and must never be able to hide inside the second.
func Endpoints(eps []ir.Endpoint, out *diag.Set) EndpointReport {
	rep := EndpointReport{Routes: len(eps), Endpoints: eps}

	seen := map[string]ir.Position{}
	for _, e := range eps {
		if e.Path == "" {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 150),
				Pos:  e.Pos,
				Rule: RuleEndpointPatternUnreadable,
				What: "this route is mounted under a pattern that is not a compile time constant, so the address it answers on cannot be recorded.",
				Why:  "An address that only exists at run time is one no catalogue, no diagram and no client contract can ever mention. It is a door in the wall that no drawing shows.",
				How:  "Build the pattern from constants, or mount the route where its path is written literally. The handler is " + e.Handler + ".",
			})
			continue
		}

		if prev, dup := seen[e.Ref()]; dup {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 151),
				Pos:  e.Pos,
				Rule: RuleEndpointDuplicate,
				What: "the address " + e.Ref() + " is mounted twice.",
				Why:  "Which of the two answers depends on the order the routers were built in, so one of them is dead and nothing in the code says which.",
				How:  "Remove one, or give them distinct patterns. The other is at " + prev.String() + ".",
			})
		} else {
			seen[e.Ref()] = e.Pos
		}

		switch {
		case e.Truncated && len(e.UseCases) == 0:
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 152),
				Pos:  e.Pos,
				Rule: RuleEndpointTraceTruncated,
				What: "the trace from " + e.Ref() + " gave up before it reached anything accountable.",
				Why:  "This is a limit of speclink rather than a defect in the code, and it is reported separately so that it cannot be mistaken for one. The route may be perfectly well founded and this simply could not see it.",
				How:  "Shorten the path from the registration to the use case, or report the handler shape so the tracer can learn it. The handler is " + e.Handler + ".",
			})
		case len(e.UseCases) == 0:
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 153),
				Pos:  e.Pos,
				Rule: RuleEndpointNoUseCase,
				What: "nothing accountable was found behind " + e.Ref() + ".",
				Why:  "Either the work belongs in a use case and is currently sitting in the presentation layer, or the route does something the architecture has no name for. Both are worth knowing about an address the world can reach.",
				How:  "Move the work into a use case and hand it to the handler, so the route inherits the requirement the use case already names. The handler is " + e.Handler + ".",
			})
		default:
			rep.Traced++
		}
	}
	return rep
}

// SortEndpoints orders a surface the way a catalogue reads it.
func SortEndpoints(eps []ir.Endpoint) {
	sort.Slice(eps, func(i, j int) bool {
		if eps[i].Path != eps[j].Path {
			return eps[i].Path < eps[j].Path
		}
		return eps[i].Method < eps[j].Method
	})
}
