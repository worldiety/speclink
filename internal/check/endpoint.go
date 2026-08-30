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
	// RuleResponseFieldRemoved fires when a field an address promised to
	// return is gone.
	RuleResponseFieldRemoved = "K20-RESPONSE-FIELD-REMOVED"
	// RuleRequestFieldDropped fires when a field an address promised to accept
	// is no longer read.
	RuleRequestFieldDropped = "K20-REQUEST-FIELD-DROPPED"
	// RuleWireShapeChanged fires when what crosses a boundary keeps its name
	// and changes its structure.
	RuleWireShapeChanged = "K20-WIRE-SHAPE-CHANGED"
)

// EndpointReport is what the run can say about the surface the system exposes.
type EndpointReport struct {
	// Routes is how many registrations were recognised, and Traced how many of
	// them reach a use case. The pair is the figure that matters: a surface
	// with a hundred routes and sixty traced is not sixty percent documented,
	// it is forty routes nobody can account for.
	//
	// Unmeasured is the third number, and it is there so the other two cannot
	// lie. A route whose handler leads into a package this run did not load is
	// neither traced nor accountable-for-nothing; it is unexamined, and
	// folding it into either of the other counts would report a choice about
	// scope as a fact about the code.
	Routes, Traced, Unmeasured int
	Endpoints                  []ir.Endpoint
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
		case e.LeftScope:
			// The trace walked into a package of this module that the run did
			// not load, so what is behind this address was not measured. It is
			// counted and not reported: silence would break the rule that an
			// unmeasured direction may not read as clean, and a finding would
			// break its other half by reporting the operator's scope as a
			// defect in the code.
			//
			// Counted as unmeasured even when a use case was also found. "I
			// found one" says nothing about whether it was the only one, and
			// treating a partial trace as complete is what would let a
			// silently dropped second use case pass as an intended change.
			rep.Unmeasured++
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
//
// # Why a promise is never forgotten
//
// The baseline keeps a withdrawn address rather than dropping it, exactly as it
// keeps a removed persisted type. Forgetting a promise because it is no longer
// kept would make the record agree with the code by editing the record, which
// is the one thing a baseline exists to prevent. A deliberate withdrawal is
// therefore settled the way a deliberate deletion is: by waiving the rule on
// the package that used to mount it, with a reason somebody has to write down.
func EndpointEvolution(eps []ir.Endpoint, base *baseline.File, scope map[string]bool, bindings []ir.Binding, out *diag.Set) {
	if len(base.Endpoints) == 0 {
		// Nothing has been promised yet, so nothing can have been broken. This
		// is not the same as a clean surface and must not read as one: a run
		// before the first freeze has measured nothing.
		return
	}
	waived := ir.CollectWaivers(bindings)

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
			// The package that mounted it was not loaded this run, so its
			// absence from the surface says nothing about whether it is still
			// there. An entry recorded before the package was written down has
			// no such excuse and is still reported.
			if was.Package != "" && !scope[was.Package] {
				continue
			}
			if waived.Has(was.Package, RuleEndpointRemoved) {
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 154),
				Rule: RuleEndpointRemoved,
				What: "the address " + ref + " was promised and is no longer mounted.",
				Why:  "Every client already calling it breaks, and nothing in this repository will tell them. An address is the one promise whose holders are outside the build.",
				How:  "Restore it, or waive " + RuleEndpointRemoved + " on " + orPackage(was.Package) + " with a reason once the callers are known to be gone. A route that merely moved reads as this finding plus a new address, which is what it is to a caller.",
			})
			continue
		}
		if now.LeftScope {
			// The work behind it was not measured this run, so the set of use
			// cases is not a set this may compare. Reporting a change here
			// would be reporting the scope, and staying silent about a real
			// change is the price of not inventing a false one — which is why
			// the run counts the route as unmeasured rather than clean.
			continue
		}
		// What crosses the boundary is held to account whether or not the work
		// behind it changed. They are independent promises: a route can keep
		// its use case and change its JSON, which is the break a caller feels
		// and nothing else here would see.
		wireEvolution(ref, was.Request, wireOf(now.RequestShape), "request", RuleRequestFieldDropped, now.Pos, out)
		wireEvolution(ref, was.Response, wireOf(now.ResponseShape), "response", RuleResponseFieldRemoved, now.Pos, out)

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
//
// A route whose trace left the loaded set is not recorded. Freezing it would
// write down a half read set of use cases as the promise, and every later run
// with a full scope would then report the difference as drift — the tool
// manufacturing the very finding it exists to detect. A narrow run may add
// nothing rather than add something false.
func RecordEndpoints(base *baseline.File, eps []ir.Endpoint) (added, updated []string) {
	for _, e := range eps {
		if e.Path == "" || e.LeftScope {
			continue
		}
		entry := baseline.Endpoint{
			UseCases: e.UseCases,
			Package:  e.Package,
			Request:  wireOf(e.RequestShape),
			Response: wireOf(e.ResponseShape),
		}
		previous, existed := base.Endpoints[e.Ref()]
		switch {
		case !existed:
			added = append(added, e.Ref())
		case !sameStrings(previous.UseCases, entry.UseCases),
			previous.Package != entry.Package,
			!sameWire(previous.Request, entry.Request),
			!sameWire(previous.Response, entry.Response):
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

// orPackage names the package a waiver belongs on, or says plainly that this
// run does not know which it was.
func orPackage(pkg string) string {
	if pkg == "" {
		return "the package that mounted it"
	}
	return pkg
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

// wireEvolution holds the bodies of one address to what it promised.
//
// # Why three rules and not five
//
// A boundary breaks asymmetrically, and only the breaks are rules. A field
// removed from a response breaks every caller reading it. A field dropped from
// a request is worse than an error: the caller still sends it, still gets its
// two hundred, and the value is discarded in silence. A field that keeps its
// wire name and changes its structure breaks both directions at once.
//
// The other two directions are not rules and are deliberately not. A field
// added to a response is what every client is required to tolerate, and
// reporting it would train people to run freeze without reading the diff. A
// field added to a request breaks callers only when it is required, and nothing
// in a Go struct says that it is — a rule that fired on every optional
// parameter would be wrong far more often than right, and a rule that guessed
// which were required would be wrong invisibly.
//
// # Why the wire name and not the Go name
//
// A caller sees the tag. Renaming the Go field while keeping the tag changes
// nothing outside this repository, and reporting it would be reporting a
// refactor as an outage. The reverse — same Go name, new tag — is a real break
// and is caught, because the promised wire name is then found nowhere.
func wireEvolution(ref string, was, now *baseline.Wire, direction string, rule string, pos ir.Position, out *diag.Set) {
	if was == nil || now == nil {
		// One side states nothing, so there is nothing to compare. A dialect
		// that never reported a body cannot be held to one, and the address
		// having gained or lost the ability to say is not a change to what
		// crosses it.
		return
	}

	// A body that is not a struct has no fields to walk, and its whole
	// structure is the promise.
	if len(was.Fields) == 0 && len(now.Fields) == 0 {
		if was.Shape != now.Shape && was.Shape != "" {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 158),
				Pos:  pos,
				Rule: RuleWireShapeChanged,
				What: "the " + direction + " of " + ref + " was promised as " + was.Shape + " and is now " + now.Shape + ".",
				Why:  "Every caller that parses it was written against the old structure. Nothing in this repository will tell them, because the address did not change and neither did anything they can see from outside.",
				How:  "Restore the structure, or publish a new address and leave this one answering as it did. If the change is intended, record it with freeze so the diff of the lock file is where somebody reviews it.",
			})
		}
		return
	}

	for _, promised := range was.Fields {
		current, still := now.ByWire(promised.Wire)
		if !still {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, codeOf(rule)),
				Pos:  pos,
				Rule: rule,
				What: reasonFor(rule, ref, promised.Wire),
				Why:  whyFor(rule),
				How:  "Restore the field, or waive " + rule + " on the package that mounts the route once the callers are known to be gone.",
			})
			continue
		}
		if current.Shape == promised.Shape {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 158),
			Pos:  pos,
			Rule: RuleWireShapeChanged,
			What: "the field " + promised.Wire + " in the " + direction + " of " + ref + " was promised as " + promised.Shape + " and is now " + current.Shape + ".",
			Why:  "A caller that reads or writes it was compiled against the old structure, and a value of the new one either fails to parse or parses into something else. This is the break that no status code reports.",
			How:  "Restore the structure, or carry the new one under a new wire name and leave the promised one in place.",
		})
	}
}

func codeOf(rule string) int {
	if rule == RuleResponseFieldRemoved {
		return 156
	}
	return 157
}

func reasonFor(rule, ref, field string) string {
	if rule == RuleResponseFieldRemoved {
		return "the response of " + ref + " promised the field " + field + " and no longer returns it."
	}
	return "the request of " + ref + " promised to accept the field " + field + " and no longer reads it."
}

func whyFor(rule string) string {
	if rule == RuleResponseFieldRemoved {
		return "Every caller reading that field breaks, and an address is the one promise whose holders are outside the build. Nothing here can tell them."
	}
	return "Callers still send it, still receive a success, and the value is discarded. A break that answers with two hundred is worse than one that answers with an error, because nothing anywhere reports it."
}

// wireOf converts a recognised body into the form the baseline keeps.
//
// Nil stays nil. A dialect that states no body must not be recorded as one that
// states an empty body, or the next run holds it to a promise it never made.
func wireOf(w *ir.WireShape) *baseline.Wire {
	if w == nil {
		return nil
	}
	out := &baseline.Wire{Type: w.Type, Shape: w.Shape}
	for _, f := range w.Fields {
		out.Fields = append(out.Fields, baseline.Field{
			Name:  f.Name,
			Wire:  f.Wire,
			Shape: f.Shape,
		})
	}
	return out
}

func sameWire(a, b *baseline.Wire) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Type != b.Type || a.Shape != b.Shape || len(a.Fields) != len(b.Fields) {
		return false
	}
	for i := range a.Fields {
		if a.Fields[i] != b.Fields[i] {
			return false
		}
	}
	return true
}
