package check

import (
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/baseline"

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
	// RuleEndpointRemoved fires when a promised address is no longer mounted.
	RuleEndpointRemoved = "K20-ENDPOINT-REMOVED"
	// RuleEndpointMeaningChanged fires when an address keeps its name and the
	// work behind it changes.
	RuleEndpointMeaningChanged = "K20-ENDPOINT-MEANING-CHANGED"
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

// EndpointEvolution holds the exposed surface to what it has already promised.
//
// Two rules and deliberately not four. A route that moved reads as one removal
// and one addition, because that is what it is to a client and because nothing
// in a snapshot could tell a move from a coincidence — a rule claiming to
// recognise a renamed path would be guessing, and guessing about a promise is
// worse than not checking it.
//
// The second rule is the one worth having. An address is a promise about
// behaviour, not about routing. A route that keeps its path while the work
// behind it changes is precisely the drift this tool exists for: the far end
// moved and the link still resolves, exactly as a rewritten requirement whose
// identifier still compiles.
//
// A route that has never been recorded is not a finding. That is what freeze is
// for, and the same reasoning already governs an unrecorded shape.
func EndpointEvolution(eps []ir.Endpoint, base *baseline.File, out *diag.Set) {
	if len(base.Endpoints) == 0 {
		// Nothing has been promised yet, so nothing can have been broken. This
		// is not the same as a clean surface and must not read as one: a run
		// before the first freeze has measured nothing.
		return
	}

	live := map[string]ir.Endpoint{}
	for _, e := range eps {
		if e.Path != "" {
			live[e.Ref()] = e
		}
	}

	refs := make([]string, 0, len(base.Endpoints))
	for ref := range base.Endpoints {
		refs = append(refs, ref)
	}
	sort.Strings(refs)

	for _, ref := range refs {
		was := base.Endpoints[ref]
		now, still := live[ref]
		if !still {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 154),
				Rule: RuleEndpointRemoved,
				What: "the address " + ref + " was promised and is no longer mounted.",
				Why:  "Every client already calling it breaks, and nothing in this repository will tell them. An address is the one promise whose holders are outside the build.",
				How:  "Restore it, or waive the rule once the callers are known to be gone. A route that merely moved reads as this finding plus a new address, which is what it is to a caller.",
			})
			continue
		}
		if sameStrings(was.UseCases, now.UseCases) {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 155),
			Pos:  now.Pos,
			Rule: RuleEndpointMeaningChanged,
			What: ref + " was recorded as serving " + join(was.UseCases) + " and now serves " + join(now.UseCases) + ".",
			Why:  "The address did not change, so nothing a caller can see says the behaviour behind it did. This is the drift the baseline exists to catch: the far end moved and the link still resolves.",
			How:  "If the change is intended, record it with freeze so the diff of the lock file is where somebody reviews it.",
		})
	}
}

// RecordEndpoints folds the current surface into the baseline and reports what
// changed.
func RecordEndpoints(base *baseline.File, eps []ir.Endpoint) (added, updated []string) {
	for _, e := range eps {
		if e.Path == "" {
			continue
		}
		entry := baseline.Endpoint{UseCases: e.UseCases}
		previous, existed := base.Endpoints[e.Ref()]
		switch {
		case !existed:
			added = append(added, e.Ref())
		case !sameStrings(previous.UseCases, entry.UseCases):
			updated = append(updated, e.Ref())
		default:
			continue
		}
		base.Endpoints[e.Ref()] = entry
	}
	sort.Strings(added)
	sort.Strings(updated)
	return added, updated
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// join renders a set of use cases for a diagnostic, naming the empty set rather
// than rendering it as a blank.
func join(names []string) string {
	if len(names) == 0 {
		return "nothing this could trace"
	}
	return strings.Join(names, ", ")
}
