package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// Rule IDs of the evolution checks. They appear in diagnostics and in
// spec.Waive calls, so they are public surface and must stay stable.
const (
	// RuleBaselineMissing fires when a promised type has never been recorded.
	RuleBaselineMissing = "K9-BASELINE-MISSING"
	// RuleDiscriminatorFrozen fires when the serialisation tag changed.
	RuleDiscriminatorFrozen = "K9-DISCRIMINATOR-FROZEN"
	// RuleTypeRemoved fires when a promised type disappeared from the source.
	RuleTypeRemoved = "K9-TYPE-REMOVED"
	// RuleFieldRemoved fires when a promised field disappeared.
	RuleFieldRemoved = "K9-FIELD-REMOVED"
	// RuleFieldRenamed fires when a promised field changed its wire name.
	RuleFieldRenamed = "K9-FIELD-RENAMED"
	// RuleProposalFrozen fires when something already promised is marked as a
	// proposal again.
	RuleProposalFrozen = "K9-PROPOSAL-FROZEN"
)

// Evolution compares the current shapes against what has been promised.
//
// Only frozen types take part. A proposal has promised nothing, so it is absent
// from the baseline and free to change in any way — that is the whole purpose
// of marking it.
//
// scope names the packages that were actually loaded, not the ones a shape was
// found in. Without it a run over one directory would report every type outside
// it as removed; taken from the schema instead, deleting the last event of a
// package would take the package out of scope and hide the very removal that
// has to be reported.
func Evolution(schema []ir.SchemaType, freeze map[string]Freeze, base *baseline.File, scope map[string]bool, bindings []ir.Binding, out *diag.Set) {
	waived := waivedRules(bindings)
	current := map[string]ir.SchemaType{}

	for _, t := range schema {
		// The type is in the source either way, so it is never a removal.
		current[t.Name] = t

		entry, recorded := base.Types[t.Name]
		if f, ok := freeze[t.Name]; ok && f.Proposal {
			if recorded {
				reportProposalFrozen(t, waived, out)
			}
			continue
		}
		if !recorded {
			reportBaselineMissing(t, waived, out)
			continue
		}
		compare(t, entry, waived, out)
	}

	reportRemoved(base, current, scope, waived, out)
}

// reportProposalFrozen catches an attempt to take a promise back.
//
// The marker means "nothing has been promised here yet", and once a shape is
// recorded that is simply untrue: messages may already be stored under it.
// Allowing the demotion would make the baseline a suggestion rather than a
// record, and every rule that reads it worthless.
func reportProposalFrozen(t ir.SchemaType, waived map[waiverKey]bool, out *diag.Set) {
	if waived[waiverKey{target: t.Name, rule: RuleProposalFrozen}] {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 95),
		Pos:  t.Pos,
		Rule: RuleProposalFrozen,
		What: shortName(t.Name) + " is marked as a proposal, but its shape has already been promised.",
		Why:  "A proposal says that nothing has been committed to yet. Once the shape is recorded that is no longer true — messages may already be stored under it, and no marker in the source can unwrite them.",
		How:  "Remove the spec.Proposal term. If the shape really has to change, introduce a new type beside this one and retire this one deliberately.",
	})
}

// reportBaselineMissing asks for the decision that has to be made once: is this
// shape promised, or is it still an experiment?
//
// This is what drives adoption. Without it an empty baseline would stay empty
// and the whole guard would never engage, because nothing would ever ask.
func reportBaselineMissing(t ir.SchemaType, waived map[waiverKey]bool, out *diag.Set) {
	if waived[waiverKey{target: t.Name, rule: RuleBaselineMissing}] {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 90),
		Pos:  t.Pos,
		Rule: RuleBaselineMissing,
		What: shortName(t.Name) + " is persisted and frozen, but its shape has never been recorded.",
		Why:  "Everything persisted is frozen unless it says otherwise, so this shape counts as promised — but nothing states what was promised, and a promise nobody wrote down cannot be kept.",
		How:  "Run `speclink freeze` to record it, or mark it as `spec.Proposal()` while it is still being worked out.",
	})
}

// compare checks one type against its record.
func compare(t ir.SchemaType, e baseline.Entry, waived map[waiverKey]bool, out *diag.Set) {
	if t.Discriminator != e.Discriminator && !waived[waiverKey{target: t.Name, rule: RuleDiscriminatorFrozen}] {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 91),
			Pos:  t.Pos,
			Rule: RuleDiscriminatorFrozen,
			What: "the discriminator of " + shortName(t.Name) + " changed from " + quote(e.Discriminator) + " to " + quote(t.Discriminator) + ".",
			Why:  "The discriminator is the key stored messages are decoded by. Changing it does not rename anything; it orphans every message written under the old tag, silently and irreversibly.",
			How:  "Restore " + quote(e.Discriminator) + ". A genuinely new shape is a new type with its own tag, next to the old one.",
		})
	}

	for _, f := range e.Fields {
		cur, ok := t.Field(f.Name)
		if !ok {
			// A field may be renamed in Go without touching the wire, so a
			// missing Go name is only a removal when the wire name is gone too.
			if _, byWire := wireField(t, f.Wire); byWire {
				continue
			}
			if waived[waiverKey{target: t.Name + "." + f.Name, rule: RuleFieldRemoved}] {
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 92),
				Pos:  t.Pos,
				Rule: RuleFieldRemoved,
				What: shortName(t.Name) + " no longer carries the promised field " + quote(f.Wire) + ".",
				Why:  "Stored messages still contain it, and a reader that stops expecting it cannot tell an absent value from one that was never written.",
				How:  "Put the field back. If it is genuinely obsolete, keep it and stop writing it, or waive " + RuleFieldRemoved + " with a reason.",
			})
			continue
		}
		if cur.Wire != f.Wire && !waived[waiverKey{target: t.Name + "." + f.Name, rule: RuleFieldRenamed}] {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 93),
				Pos:  t.Pos,
				Rule: RuleFieldRenamed,
				What: "field " + f.Name + " of " + shortName(t.Name) + " changed its stored name from " + quote(f.Wire) + " to " + quote(cur.Wire) + ".",
				Why:  "The stored name is how a value is found again. Renaming it makes every value written so far unreachable, while the code keeps compiling.",
				How:  "Restore the json tag `json:\"" + f.Wire + "\"`. The Go field name may be changed freely; only the stored name is promised.",
			})
		}
	}
}

// wireField looks a field up by the name it carries in stored data.
func wireField(t ir.SchemaType, wire string) (ir.SchemaField, bool) {
	for _, f := range t.Fields {
		if f.Wire == wire {
			return f, true
		}
	}
	return ir.SchemaField{}, false
}

// reportRemoved names promised types that are gone from the source.
//
// The waiver sits on the package, because the type it would otherwise sit on no
// longer exists. Retiring a promised type is deliberate work: the stored
// messages have to be purged first, and saying so out loud is the point.
func reportRemoved(base *baseline.File, current map[string]ir.SchemaType, scope map[string]bool, waived map[waiverKey]bool, out *diag.Set) {
	for _, name := range base.Names() {
		if _, ok := current[name]; ok {
			continue
		}
		pkg := packageOf(name)
		if !scope[pkg] {
			continue // not loaded in this run; absence says nothing
		}
		if waived[waiverKey{target: name, rule: RuleTypeRemoved}] ||
			waived[waiverKey{target: pkg, rule: RuleTypeRemoved}] {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 94),
			Pos:  ir.Position{File: pkg, Line: 1, Col: 1},
			Rule: RuleTypeRemoved,
			What: shortName(name) + " was promised but no longer exists.",
			Why:  "Messages written under its discriminator are still in the log and can no longer be decoded, so a replay fails on data that was perfectly valid when it was written.",
			How:  "Restore the type, or purge its messages and waive " + RuleTypeRemoved + " on the package with a reason.",
		})
	}
}

func quote(s string) string {
	if s == "" {
		return "nothing"
	}
	return `"` + s + `"`
}

// SortSchema orders types by name so every run reports in the same sequence.
func SortSchema(schema []ir.SchemaType) {
	sort.Slice(schema, func(i, j int) bool { return schema[i].Name < schema[j].Name })
}
