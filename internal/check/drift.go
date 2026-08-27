// Package check, this file: the drift rules of the outer edge.
package check

import (
	"sort"
	"strconv"
	"strings"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
	"github.com/worldiety/speclink/internal/source"
)

const (
	// RuleReqChanged fires when a requirement's text no longer matches what was
	// recorded, while the constructs satisfying it were left alone.
	//
	// Everything about a satisfies link is checked by the Go compiler except
	// the only thing that matters: whether the code still does what the
	// requirement now says. Rewrite the text of R-QUOTE-SUBMIT completely and
	// the identifier is unchanged, every reference resolves, the coverage stays
	// at a hundred percent and nothing is reported. With a requirement tree
	// generated from natural language this is not a corner case but the
	// likeliest defect in the system, because the tree is rewritten routinely
	// and by a process that has no notion of which parts were load bearing.
	//
	// The literature answers this by hand: OpenFastTrace puts a revision in the
	// identifier that an author increments when the meaning changed,
	// Sphinx-Needs lets a link carry a predicate, DOORS raises a suspect flag
	// for a human to clear. All three depend on somebody deciding that a change
	// was semantic, and the last is famously cleared in bulk. Here the change is
	// computed, so nobody has to decide and nobody can forget.
	RuleReqChanged = "K10-REQ-CHANGED"

	// RuleSourceDrift fires when a source segment no longer matches what was
	// recorded, while the requirements derived from it were left alone.
	//
	// The same failure one layer further out, and the one the traceability
	// literature treats as open. The anchor still resolves, so nothing is
	// dangling; the section it resolves to now says something else.
	RuleSourceDrift = "K13-SOURCE-DRIFT"
)

// Drift reports the recorded edges that no longer match the source.
//
// scope names the requirements and segments the run actually looked at, on the
// same grounds as the schema scope: a run over one directory must not report
// everything outside it as changed, and taking the scope from what was found
// instead would let deleting the last requirement of a package hide the
// deletion.
func Drift(tree *reqtree.Tree, set *source.Set, docs []string, cov Coverage, srcCov SourceCoverage, base *baseline.File, waived ir.Waivers, out *diag.Set) {
	driftRequirements(tree, cov, base, waived, out)
	driftSources(tree, set, docs, srcCov, base, waived, out)
}

func driftRequirements(tree *reqtree.Tree, cov Coverage, base *baseline.File, waived ir.Waivers, out *diag.Set) {
	for _, r := range tree.All() {
		rec, ok := base.Requirements[r.ID]
		if !ok {
			// Nothing was promised about it yet. A requirement that has never
			// been recorded is not drifting, it is new, and the freeze that
			// records it is where it gets read.
			continue
		}
		if rec.Text == baseline.HashText(r.Text, r.Title) {
			continue
		}
		if waived.Has(r.ID, RuleReqChanged) || waived.Has("", RuleReqChanged) {
			continue
		}

		satisfiers := names(cov.BySatisfier[r.ID])
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 110),
			Rule: RuleReqChanged,
			Pos:  r.Pos,
			What: "the text of " + r.ID + " changed since it was last recorded.",
			Why: "Every reference to a requirement is checked by the Go compiler except the one thing that matters, which is whether the code still does what the requirement now says. The identifier did not change, so nothing else in this run has anything to report: " +
				satisfying(satisfiers) + " went on counting towards the coverage while the sentence it was written for was being rewritten.",
			How: "Re-read " + list(satisfiers) + " against the new text, change what has to change, then run `speclink freeze` to record the wording that was reviewed.",
		})
	}
}

func driftSources(tree *reqtree.Tree, set *source.Set, docs []string, cov SourceCoverage, base *baseline.File, waived ir.Waivers, out *diag.Set) {
	for _, doc := range docs {
		d := set.Get(doc)
		if d.Err != nil {
			continue
		}
		for _, seg := range d.Segments {
			rec, ok := base.Sources[seg.Ref()]
			if !ok || rec.Fingerprint == seg.Fingerprint {
				continue
			}
			if waived.Has(seg.Ref(), RuleSourceDrift) || waived.Has("", RuleSourceDrift) {
				continue
			}

			derived := cov.ByCiter[seg.Ref()]
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 111),
				Rule: RuleSourceDrift,
				Pos:  ir.Position{File: seg.Pos.File, Line: seg.Pos.Line},
				What: describe(seg) + " changed since it was last recorded.",
				Why:  "The anchor still resolves, so nothing is dangling and no other rule has anything to say. What it resolves to " + nowShows(seg) + ", and " + derivedFrom(derived) + " unchanged.",
				How:  "Re-read " + list(derived) + " against " + theNew(seg) + ", change what has to change, then run `speclink freeze` to record the version that was reviewed.",
			})
		}
	}
}

// Record folds the current state into the baseline. It is what freeze writes.
//
// reviewer names the person the run is being made on behalf of, empty when
// nobody read anything. A run in CI records the text so the drift rules have
// something to compare against next time, and records no review, because none
// happened. Conflating the two would make the record claim a reading that never
// took place, and the record is the only thing standing behind that claim.
//
// A review is bound to the wording it was given for. When the text changed, the
// previous reviewer is dropped rather than carried over: what they read is not
// what is there any more, and a name against the wrong sentence is worse than
// no name at all.
func Record(base *baseline.File, tree *reqtree.Tree, set *source.Set, docs []string, reviewer string) (requirements, segments int) {
	for _, r := range tree.All() {
		rec := baseline.Requirement{
			Text:  baseline.HashText(r.Text, r.Title),
			Title: r.Title,
			From:  primarySource(r),
		}
		previous, existed := base.Requirements[r.ID]
		if existed && previous.Text == rec.Text {
			rec.ReviewedBy = previous.ReviewedBy
		}
		if reviewer != "" {
			rec.ReviewedBy = reviewer
		}
		if !existed || previous != rec {
			requirements++
		}
		base.Requirements[r.ID] = rec
	}

	for _, doc := range docs {
		d := set.Get(doc)
		if d.Err != nil {
			continue
		}
		for _, seg := range d.Segments {
			rec := baseline.Segment{Fingerprint: seg.Fingerprint, Kind: seg.Kind.String()}
			if base.Sources[seg.Ref()] != rec {
				segments++
			}
			base.Sources[seg.Ref()] = rec
		}
	}
	return requirements, segments
}

// primarySource is the segment a requirement is recorded against.
//
// A requirement may cite several, and the first is taken. The record exists to
// keep an identifier stable across a regeneration, so what matters is that the
// answer is deterministic, not which of several origins wins.
func primarySource(r *ir.Requirement) string {
	for _, s := range r.Sources {
		if s.Doc != "" && s.Anchor != "" {
			return s.Doc + "#" + s.Anchor
		}
	}
	return ""
}

// names renders the constructs satisfying a requirement, so the finding can be
// acted on without looking the requirement up first.
func names(targets []ir.Target) []string {
	out := make([]string, 0, len(targets))
	for _, t := range targets {
		out = append(out, t.String())
	}
	return out
}

// The two source kinds drift in different words. A region of a mockup does not
// read differently, it looks different, and a diagnostic that said otherwise
// would send the reader looking for text that is not there.
func nowShows(s source.Segment) string {
	if s.Kind == source.KindImage {
		return "now looks different"
	}
	return "now reads differently"
}

func theNew(s source.Segment) string {
	if s.Kind == source.KindImage {
		return "the new mockup"
	}
	return "the new wording"
}

func satisfying(ids []string) string {
	switch len(ids) {
	case 0:
		return "nothing"
	case 1:
		return "the one construct satisfying it"
	}
	return "the " + itoa(len(ids)) + " constructs satisfying it"
}

func itoa(n int) string { return strconv.Itoa(n) }

func derivedFrom(ids []string) string {
	switch len(ids) {
	case 0:
		return "nothing derived from it was left"
	case 1:
		return "the requirement derived from it was left"
	}
	return "the requirements derived from it were left"
}

func list(ids []string) string {
	if len(ids) == 0 {
		return "whatever was written for it"
	}
	sorted := append([]string(nil), ids...)
	sort.Strings(sorted)
	return strings.Join(sorted, ", ")
}
