package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// RuleReviewStale fires when a declaration was read and then changed.
const RuleReviewStale = "K18-REVIEW-STALE"

// ProvenanceReport is what the run can say about who wrote the code and who
// has looked at it.
type ProvenanceReport struct {
	// Total is the number of declarations that have a body somebody could read.
	Total int
	// Machine is how many are recorded as written by a generator.
	Machine int
	// Reviewed is how many a person has read, in the state they are in now.
	Reviewed int
	// Statements is how many statements the measured declarations hold, and
	// Executed how many a run reached. Both zero where no coverage profile was
	// ever handed over — which is not the same as nothing being exercised, and
	// is why the figure is left out rather than printed as zero per cent.
	Statements, Executed int
}

// Provenance checks the record of who wrote what and who has read it.
//
// # Why there is no rule for unread code
//
// It would be the obvious one and it would be wrong. In a project whose code a
// machine writes and a person samples, a build that stays red until somebody has
// read every declaration is a build that is never green, and a signal that is
// always on is not a signal. So this is a figure, not a finding: the run says
// how much was machine written and how much of that a person has since read,
// and what to do about the gap is a judgement nobody here is qualified to make.
//
// # Why there is one for stale review
//
// A review that was recorded and then outlived its subject is worse than no
// review. It is somebody's name attached to text they never saw, and the person
// reading the record has no way to tell the difference. That is the same
// failure K10-REQ-CHANGED exists for, one layer down.
//
// # What unattested means
//
// Not human. A declaration nothing has said anything about is counted as
// neither written by a person nor read by one, because silence must not be
// able to pass for handwork — it is the state every declaration starts in, and
// the state most of them will stay in.
func Provenance(constructs []ir.Construct, base *baseline.File, out *diag.Set) ProvenanceReport {
	var rep ProvenanceReport
	if base == nil {
		return rep
	}

	sorted := append([]ir.Construct(nil), constructs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Pos.Less(sorted[j].Pos) })

	for _, c := range sorted {
		// A role with no body is nothing anybody reads. A permission is a
		// call; counting it would pad the figure with declarations that have
		// no content to review.
		if c.Fingerprint == "" {
			continue
		}
		rep.Total++

		rec, recorded := base.Constructs[c.Name]
		if !recorded {
			continue
		}
		if rec.Origin == "llm" {
			rep.Machine++
		}
		// Only where the figure was measured on this text. A profile taken
		// before a rewrite says nothing about what is there now.
		if rec.Fingerprint == c.Fingerprint {
			rep.Statements += rec.Statements
			rep.Executed += rec.Covered
		}
		if rec.ReviewedBy == "" {
			continue
		}
		if rec.Fingerprint == c.Fingerprint {
			rep.Reviewed++
			continue
		}

		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 140),
			Pos:  c.Pos,
			Rule: RuleReviewStale,
			What: shortName(c.Name) + " was read by " + rec.ReviewedBy + " and has changed since.",
			Why:  "A review that outlived its subject is worse than no review: it is somebody's name attached to text they never saw, and nothing in the record says which of the two it is.",
			How:  "Read it again and record that with `speclink attest -reviewer " + quote(rec.ReviewedBy) + " " + c.Name + "`, or drop the review.",
		})
	}
	return rep
}
