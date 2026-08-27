package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// RuleFieldUnbound fires when a field of the domain model or of a stored shape
// traces to no requirement.
const RuleFieldUnbound = "K1-FIELD-UNBOUND"

// CoverFields carries the forward coverage down to the field.
//
// Binding a type is not the same as accounting for what is in it. A command
// bound to "quotes must pass an approval gate" says nothing about the
// RetentionClass somebody added to it a year later, and that is where the drift
// actually is: types are created deliberately and reviewed, fields accrete.
//
// Two sets are covered, and they are not substitutes for one another.
//
// The domain model carries the meaning. It is what somebody reads to find out
// what the system believes a quote is, and a field there that nobody can trace
// to a requirement is either an unrecorded promise or dead weight.
//
// The stored shape carries the historical values. It is what a review or a
// migration has to work from, and it is the one thing that cannot be
// renegotiated: once messages carry a field, its meaning is fixed by the data.
// The moment to ask what it is for is while it is being added, not when a
// dispute turns on which of two timestamps was authoritative.
//
// Where a project separates the two — a domain type beside its stored entity —
// both are asked, because neither answers for the other. The names may match
// and the meanings still diverge; that divergence is exactly what a migration
// discovers too late.
//
// There is no exemption for envelope fields. An ActorRef appearing on seventy
// events is not thereby self evident: it is the audit trail, it exists because
// somebody must be able to reconstruct who did what, and that is a requirement
// like any other.
func CoverFields(schema []ir.SchemaType, constructs []ir.Construct, bindings []ir.Binding, domain map[string]bool, d ir.Dialect, out *diag.Set) {
	bound := boundFields(bindings)
	waived := ir.CollectWaivers(bindings)

	type shape struct {
		name    string
		pkg     string
		stored  bool
		fields  []ir.SchemaField
		fileSeq ir.Position
	}
	var shapes []shape
	seen := map[string]bool{}

	for _, t := range schema {
		shapes = append(shapes, shape{name: t.Name, pkg: t.Package, stored: true, fields: t.Fields, fileSeq: t.Pos})
		seen[t.Name] = true
	}
	for _, c := range constructs {
		// Only the domain model. A command or a projection is working memory
		// and can be reshaped at will, so nothing is promised by its fields.
		if c.Kind != ir.ConstructAggregate || len(c.Fields) == 0 || seen[c.Name] {
			continue
		}
		shapes = append(shapes, shape{name: c.Name, pkg: c.Package, fields: c.Fields, fileSeq: c.Pos})
	}

	sort.Slice(shapes, func(i, j int) bool {
		if shapes[i].fileSeq.File != shapes[j].fileSeq.File {
			return shapes[i].fileSeq.File < shapes[j].fileSeq.File
		}
		return shapes[i].fileSeq.Line < shapes[j].fileSeq.Line
	})

	for _, s := range shapes {
		if !domain[s.pkg] {
			continue
		}
		for _, f := range s.fields {
			key := s.name + "." + f.Name
			if bound[key] || waived.Has(key, RuleFieldUnbound) {
				continue
			}
			short := shortName(s.name)
			why := "A field of the domain model states what the system believes about the thing it describes. Binding the type does not answer what this particular field is for, and one that traces to nothing is either an unrecorded promise or dead weight."
			what := short + "." + f.Name + " carries meaning that traces to no requirement."
			if s.stored {
				why = "A stored field outlives every decision made about it: once messages carry it, its meaning is fixed by the data and cannot be renegotiated. It is also what a review or a migration has to work from."
				what = short + "." + f.Name + " is stored but traces to no requirement."
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 22),
				Pos:  f.Pos,
				Rule: RuleFieldUnbound,
				What: what,
				Why:  why,
				How:  "Add `" + d.BindField(s.name, f.Name) + "` naming the requirement it serves — the business requirement it answers, or the audit requirement it exists for when it carries who acted or when.",
			})
		}
	}
}

// boundFields collects the fields that satisfy at least one requirement.
func boundFields(bindings []ir.Binding) map[string]bool {
	out := map[string]bool{}
	for _, b := range bindings {
		if b.Target.Kind != ir.TargetField {
			continue
		}
		for _, a := range b.Assertions {
			if a.Kind == ir.AssertSatisfies && len(a.Requirements) > 0 {
				out[b.Target.String()] = true
			}
		}
	}
	return out
}
