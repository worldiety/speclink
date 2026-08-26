package check

import (
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// RuleFieldUnbound fires when a persisted field traces to no requirement.
const RuleFieldUnbound = "K1-FIELD-UNBOUND"

// CoverFields carries the forward coverage down to the field.
//
// Binding a type is not the same as accounting for what is in it. A command
// bound to "quotes must pass an approval gate" says nothing about the
// RetentionClass someone added to it a year later, and that field is where the
// drift actually happens: types are created deliberately and reviewed, fields
// accrete. The one thing a stored field cannot do is be renegotiated later —
// once messages carry it, its meaning is fixed by the data, so the moment to
// ask what it is for is now.
//
// The scope is the persisted set, the same one speclink.lock tracks: events and
// the stored form of repositories. Fields of a command or a projection are
// working memory and can be reshaped at will; these cannot.
//
// There is no exemption for envelope fields. An ActorRef appearing on seventy
// events is not thereby self evident — it is the audit trail, it exists because
// somebody must be able to reconstruct who did what, and that is a requirement
// like any other. Writing it down seventy times is cheap; discovering in a
// dispute that nobody knows which of the two timestamps is authoritative is
// not.
func CoverFields(schema []ir.SchemaType, bindings []ir.Binding, domain map[string]bool, out *diag.Set) {
	bound := boundFields(bindings)
	waived := ir.CollectWaivers(bindings)

	for _, t := range schema {
		if !domain[t.Package] {
			continue
		}
		for _, f := range t.Fields {
			key := t.Name + "." + f.Name
			if bound[key] || waived.Has(key, RuleFieldUnbound) {
				continue
			}
			short := shortName(t.Name)
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 22),
				Pos:  f.Pos,
				Rule: RuleFieldUnbound,
				What: short + "." + f.Name + " is stored but traces to no requirement.",
				Why:  "A field outlives every decision made about it: once messages carry it, its meaning is fixed by the data and cannot be renegotiated. Binding the type does not answer what this particular field is for.",
				How:  "Add `var _ = spec.ForField[" + short + "](\"" + f.Name + "\", spec.Satisfies(…))` naming the requirement it serves — the business requirement it answers, or the audit requirement it exists for when it carries who acted or when.",
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
