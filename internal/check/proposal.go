package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// RuleProposalRedundant fires when something is marked as a proposal although
// it already is one through the level above it.
const RuleProposalRedundant = "K9-PROPOSAL-REDUNDANT"

// Freeze records, for one persisted type, whether the type and each of its
// fields have been promised.
//
// Everything persisted is frozen unless a proposal term says otherwise, and the
// term cascades downwards: a proposal package makes every persisted type in it
// a proposal, and a proposal type makes every one of its fields a proposal.
// Marking the level that is actually true is therefore enough, and it is what
// keeps the number of terms proportional to the number of exceptions rather
// than to the size of the system.
type Freeze struct {
	// Type is the fully qualified name of the persisted type.
	Type string
	// Proposal reports whether the type as a whole is still a proposal.
	Proposal bool
	// ProposalFields names the fields that are proposals while the type itself
	// is frozen. It is empty whenever Proposal is true, because then every
	// field is one anyway.
	ProposalFields map[string]bool
	// OptionalFields names the fields declared as possibly absent from stored
	// data, which every field added after the promise has to be.
	OptionalFields map[string]bool
}

// Frozen reports whether the field has been promised.
func (f Freeze) Frozen(field string) bool {
	return !f.Proposal && !f.ProposalFields[field]
}

// Proposals resolves the cascade for every persisted shape and reports the
// terms that state nothing new.
//
// The schema is what defines the persisted set, not the construct list. An
// event is persisted because of what it is; a persistence model is persisted
// because somewhere a repository was built over it, which is a fact about
// another package entirely. Reading both from the schema keeps that decision in
// one place.
//
// The redundancy check is phase V4, which had no rule until now. That was not
// an oversight but a property of the directive catalogue: everything inferable
// was deliberately given no directive at all, so there was nothing to be
// redundant with. The cascade is the first place where the language can state
// the same fact twice, and it is therefore the first place the phase has work.
func Proposals(schema []ir.SchemaType, bindings []ir.Binding, out *diag.Set) map[string]Freeze {
	var (
		packages = map[string]ir.Assertion{}
		types    = map[string]ir.Assertion{}
		fields   = map[string]map[string]ir.Assertion{}
		optional = map[string]map[string]bool{}
	)

	for _, b := range bindings {
		for _, a := range b.Assertions {
			switch {
			case a.Kind == ir.AssertOptional && b.Target.Kind == ir.TargetField:
				if optional[b.Target.Name] == nil {
					optional[b.Target.Name] = map[string]bool{}
				}
				optional[b.Target.Name][b.Target.Field] = true
				continue
			case a.Kind != ir.AssertProposal:
				continue
			}
			switch b.Target.Kind {
			case ir.TargetPackage:
				packages[b.Target.Package] = a
			case ir.TargetType:
				types[b.Target.Name] = a
			case ir.TargetField:
				if fields[b.Target.Name] == nil {
					fields[b.Target.Name] = map[string]ir.Assertion{}
				}
				fields[b.Target.Name][b.Target.Field] = a
			}
		}
	}

	for _, f := range redundantTypes(types, packages) {
		out.Add(f)
	}
	for _, f := range redundantFields(fields, types, packages) {
		out.Add(f)
	}

	result := map[string]Freeze{}
	for _, t := range schema {
		f := Freeze{
			Type:           t.Name,
			Proposal:       packages[t.Package].Kind == ir.AssertProposal || types[t.Name].Kind == ir.AssertProposal,
			ProposalFields: map[string]bool{},
			OptionalFields: optional[t.Name],
		}
		if f.OptionalFields == nil {
			f.OptionalFields = map[string]bool{}
		}
		if !f.Proposal {
			for field := range fields[t.Name] {
				f.ProposalFields[field] = true
			}
		}
		result[t.Name] = f
	}
	return result
}

// redundantTypes reports a proposal on a type whose package is already one.
func redundantTypes(types, packages map[string]ir.Assertion) []diag.Finding {
	var out []diag.Finding
	for _, name := range sortedKeys(types) {
		pkg := packageOf(name)
		if packages[pkg].Kind != ir.AssertProposal {
			continue
		}
		out = append(out, diag.Finding{
			Code: diag.Code(diag.PhaseRedundant, 1),
			Pos:  types[name].Pos,
			Rule: RuleProposalRedundant,
			What: shortName(name) + " is marked as a proposal, but its package already is one.",
			Why:  "A proposal cascades downwards. Stating it again says nothing new, and the two terms will be removed at different times, leaving the type promised while the package still claims it is not.",
			How:  "Remove this term. Promote the package by deleting its spec.Proposal, or mark the package as frozen first if only this type is still open.",
		})
	}
	return out
}

// redundantFields reports a proposal on a field whose type or package is
// already one.
//
// This is the case that has to be right: a field level term is only meaningful
// once the type itself is frozen. Before that it is noise, and noise in an
// exception marker is how an exception survives its reason.
func redundantFields(fields map[string]map[string]ir.Assertion, types, packages map[string]ir.Assertion) []diag.Finding {
	var out []diag.Finding
	for _, owner := range sortedKeys(fields) {
		pkg := packageOf(owner)
		var covered, level string
		switch {
		case packages[pkg].Kind == ir.AssertProposal:
			// A package is named by its path. shortName would cut at the last
			// dot, which in an import path is part of the host name.
			covered, level = pkg, "its package"
		case types[owner].Kind == ir.AssertProposal:
			covered, level = shortName(owner), "its type"
		}
		if covered == "" {
			continue
		}
		for _, field := range sortedKeys(fields[owner]) {
			out = append(out, diag.Finding{
				Code: diag.Code(diag.PhaseRedundant, 2),
				Pos:  fields[owner][field].Pos,
				Rule: RuleProposalRedundant,
				What: shortName(owner) + "." + field + " is marked as a proposal, but " + level + " already is one.",
				Why:  "A proposal cascades downwards, so a field term only means something once the type itself is frozen. Until then it is an exception without a rule to except it from.",
				How:  "Remove this term. Once " + covered + " is promoted, mark the field again if it is still open.",
			})
		}
	}
	return out
}

// packageOf strips the type name from a fully qualified name.
func packageOf(qualified string) string {
	if i := lastIndexByte(qualified, '.'); i >= 0 {
		return qualified[:i]
	}
	return qualified
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
