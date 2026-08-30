// Package check, this file: the forward direction of the outer edge.
package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
	"github.com/worldiety/speclink/internal/source"
)

// RuleSourceUncovered fires on a section of a requirement document that became
// no requirement.
//
// This is the defect nothing else in speclink can see, and the one with the
// worst failure mode. The requirement tree can be perfectly consistent, every
// construct bound, every normative requirement covered, the summary reporting
// a hundred percent in both directions — and a whole section of what was asked
// for is simply absent. Backward coverage cannot find it, because there is no
// requirement to be uncovered. Forward coverage of the code cannot find it,
// because there is no construct to be unbound.
//
// It is the recall failure of any natural language extraction, and it is
// systematically likelier than the precision failure: an invented requirement
// leaves code behind that somebody eventually reads, whereas a skipped section
// leaves nothing at all.
//
// The exemption is not a waiver but a marker in the document.
//
// A waiver attaches to a construct of the host language, and a source segment
// has none, so a
// waiver narrowed to one section could not be written down at all. Rather than
// invent a second escape hatch, the statement is made where it belongs: a
// section that carries no obligation says so in the document, in a comment no
// renderer shows, and the person writing the specification is the one who
// decides. That also satisfies P2 — the fact is stated once, in one place.
//
// The undirected waiver is honoured, the same way K3-REQ-UNCOVERED honours it,
// so a project can adopt the rest of speclink before its source documents are
// in shape. It switches the whole rule off and leaves a trace saying so.
const RuleSourceUncovered = "K12-SOURCE-UNCOVERED"

// SourceCoverage is the result of the forward direction.
type SourceCoverage struct {
	// Total is the number of segments that had to produce a requirement.
	Total int
	// Covered is how many did.
	Covered int
	// ByCiter maps a segment reference to the requirements citing it.
	ByCiter map[string][]string

	// Standards holds the clause catalogues, one entry per standard.
	//
	// Kept apart from the figure above rather than folded into it, because a
	// mean of the two hides the thing the split exists to show: a project can
	// have accounted for every line of its own material and answered half a
	// standard, and one number saying seventy five per cent tells nobody that.
	Standards []StandardCoverage
}

// StandardCoverage is what a run can say about one external standard.
type StandardCoverage struct {
	// Doc is the repository relative path of the catalogue.
	Doc string
	// Title is the name of the standard, as a reader refers to it.
	Title string
	// Clauses is how many clauses apply here, and Answered how many of those a
	// requirement cites.
	Clauses, Answered int
	// Excluded is how many were declared not to apply. Reported beside the
	// others because an exclusion is a decision, not an absence, and the count
	// of them is the first thing an auditor weighs.
	Excluded int
}

// Ratio is the share of applicable clauses that a requirement answers.
func (s StandardCoverage) Ratio() float64 {
	if s.Clauses == 0 {
		return 1
	}
	return float64(s.Answered) / float64(s.Clauses)
}

// Ratio is the covered share, 1 for an empty set. An empty set is complete: a
// project with no source documents has nothing outstanding, and reporting zero
// percent would be an accusation rather than a measurement.
func (c SourceCoverage) Ratio() float64 {
	if c.Total == 0 {
		return 1
	}
	return float64(c.Covered) / float64(c.Total)
}

// CoverSources checks that every segment of every source document produced at
// least one requirement.
//
// docs are the enumerated documents, not the cited ones. The distinction is the
// rule: measuring over what is cited could never report the document nobody
// read.
func CoverSources(tree *reqtree.Tree, set *source.Set, docs []string, waived ir.Waivers, out *diag.Set) SourceCoverage {
	cov := SourceCoverage{ByCiter: map[string][]string{}}

	for _, r := range tree.All() {
		for _, s := range r.Sources {
			if s.Doc == "" {
				continue
			}
			ref := s.Doc
			if s.Anchor != "" {
				ref += "#" + s.Anchor
			}
			cov.ByCiter[ref] = append(cov.ByCiter[ref], r.ID)
		}
	}
	for ref := range cov.ByCiter {
		sort.Strings(cov.ByCiter[ref])
	}

	for _, doc := range docs {
		d := set.Get(doc)
		if d.Kind == source.KindStandard && d.Err == nil {
			cov.Standards = append(cov.Standards, standardCoverage(d, cov.ByCiter))
		}
		if d.Err != nil {
			// Reported against the document itself. A document that could not
			// be segmented has no segments to hold accountable, and counting
			// it as fully covered would let an unreadable file hide a whole
			// specification.
			continue
		}
		for _, seg := range d.Segments {
			if seg.Informative {
				continue
			}
			// A clause is counted in its own standard, never in the figure for
			// the project's own material. The two answer different questions
			// and a project is routinely in a different place on each.
			if d.Kind != source.KindStandard {
				cov.Total++
			}
			counted := d.Kind != source.KindStandard
			if len(cov.ByCiter[seg.Ref()]) > 0 {
				if counted {
					cov.Covered++
				}
				continue
			}
			if waived.Has("", RuleSourceUncovered) {
				if counted {
					cov.Covered++
				}
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 100),
				Rule: RuleSourceUncovered,
				Pos:  ir.Position{File: seg.Pos.File, Line: seg.Pos.Line},
				What: describe(seg) + " became no requirement.",
				Why:  whyUncovered(seg),
				How:  howUncovered(seg),
			})
		}
	}
	return cov
}

// markInformative names the exemption in the terms of the document it applies
// to. The statement always belongs in the source, but a Markdown section and an
// image region say it in different places, and a diagnostic that named the
// wrong one would be worse than none.
// standardCoverage counts one catalogue on its own terms.
func standardCoverage(d source.Document, byCiter map[string][]string) StandardCoverage {
	sc := StandardCoverage{Doc: d.Doc, Title: d.Title}
	for _, seg := range d.Segments {
		if seg.Informative {
			sc.Excluded++
			continue
		}
		sc.Clauses++
		if len(byCiter[seg.Ref()]) > 0 {
			sc.Answered++
		}
	}
	return sc
}

// howUncovered spells the two ways out in the terms of the document.
//
// The three read differently on purpose. A section of prose carries no
// obligation; a region of a mockup is decoration; a clause of a standard does
// not apply here — and that last one is a decision somebody has to justify,
// which is why the catalogue demands a reason where the other two do not.
func howUncovered(s source.Segment) string {
	write := "Write a requirement citing " + s.Ref()
	switch s.Kind {
	case source.KindImage:
		return write + `, or set "informative": true on the region in ` + source.ManifestPath(s.Doc) + " if it states no obligation."
	case source.KindStandard:
		return write + `, or exclude the clause with "applicable": false and a "because" saying why it does not apply here.`
	}
	return write + ", or mark the section informative with " + source.InformativeMarker + " if it states no obligation."
}

// whyUncovered says what is lost, in the terms of the document.
func whyUncovered(s source.Segment) string {
	if s.Kind == source.KindStandard {
		return "A clause of a standard is an obligation somebody accepted. One that became no requirement is not covered by anything and is not excluded either — it is simply unaccounted for, which is the state an audit exists to find."
	}
	return "Every part of a requirement document has to turn into something. A section that did not is the one defect the rest of speclink cannot see: there is no requirement to report as uncovered and no construct to report as unbound, so the run stays green while a piece of what was asked for is missing."
}

func describe(s source.Segment) string {
	switch s.Kind {
	case source.KindImage:
		return "region " + s.ID + " of " + s.Doc
	case source.KindStandard:
		return "clause " + s.ID + " of " + s.Doc
	}
	return "section " + s.ID + " of " + s.Doc
}
