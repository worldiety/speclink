package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// Rule IDs of the contract checks.
const (
	// RuleContractFieldRemoved fires when a field this system relied on
	// receiving is no longer part of the shape that crosses.
	RuleContractFieldRemoved = "K17-CONTRACT-FIELD-REMOVED"
	// RuleContractFieldChanged fires when a field kept its name and changed
	// its structure.
	RuleContractFieldChanged = "K17-CONTRACT-FIELD-CHANGED"
	// RuleContractShapeChanged fires when a contract that is not a struct
	// changed structure.
	RuleContractShapeChanged = "K17-CONTRACT-SHAPE-CHANGED"
)

// ContractEvolution compares the contracts crossing the boundary against what
// was recorded.
//
// # Why this is the other half of the endpoint rules, and the sharper half
//
// K20 holds this system to the surfaces it offers: a field removed from a
// response is a promise broken to somebody who parsed it. This holds it to the
// surfaces it depends on, and the asymmetry matters. A promise this system
// makes is broken by people who can be told — they are in this repository, and
// the rule tells them. A promise it relies on is broken by somebody who has
// never heard of this repository, and the first sign is a field arriving empty
// in production.
//
// # Why a change here is reported rather than refused
//
// It is nearly always the far end that moved, and this system has to follow.
// The finding is not "you broke it" but "what you relied on is not what you
// recorded", which is a thing somebody must look at and often then accept with
// freeze. That is why every message ends in the same place: decide, and record
// the decision where a diff shows it.
func ContractEvolution(channels []ir.Channel, base *baseline.File, out *diag.Set) {
	if base == nil {
		return
	}
	sorted := append([]ir.Channel(nil), channels...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name() < sorted[j].Name() })

	for _, ch := range sorted {
		rec, known := base.Channels[ch.Name()]
		if !known || rec.Contract == nil {
			// Never recorded. That is what freeze is for, not a finding: a
			// contract nobody has promised yet cannot have been broken.
			continue
		}
		if ch.Contract == nil {
			// The declaration dropped its contract. The recorded shape still
			// describes something this system depends on, and now nothing
			// checks it.
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 170),
				Pos:  ch.Pos,
				Rule: RuleContractShapeChanged,
				What: "the channel " + ch.Name() + " recorded a contract of " + rec.Contract.Shape + " and now names none.",
				Why:  "The shape is still crossing; the only thing that changed is that nothing compares it any more. A dependency that stops being checked does not stop being a dependency.",
				How:  "Name the type again in the Contract field, or record the removal with freeze so the lock file is where somebody reviews it.",
			})
			continue
		}
		contractEvolution(ch, rec.Contract, wireOf(ch.Contract), out)
	}
}

// contractEvolution compares one channel's promised shape with the current one.
func contractEvolution(ch ir.Channel, was, now *baseline.Wire, out *diag.Set) {
	ref := ch.Name()

	if len(was.Fields) == 0 && len(now.Fields) == 0 {
		if was.Shape != now.Shape && was.Shape != "" {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 170),
				Pos:  ch.Pos,
				Rule: RuleContractShapeChanged,
				What: "what crosses " + ref + " was recorded as " + was.Shape + " and is now " + now.Shape + ".",
				Why:  "This system reads that structure. Nothing at the far end knows this repository exists, so the first report of a mismatch is whatever goes wrong in production.",
				How:  "Follow the change here, or hold the far end to the old shape. Either way record the decision with freeze.",
			})
		}
		return
	}

	for _, promised := range was.Fields {
		current, still := now.ByWire(promised.Wire)
		if !still {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 171),
				Pos:  ch.Pos,
				Rule: RuleContractFieldRemoved,
				What: ref + " relied on the field " + promised.Wire + ", which the shape crossing it no longer has.",
				Why:  "Code in this module reads it. Where the far end dropped it, every read now yields a zero value, and a zero value is indistinguishable from a legitimately empty one at the point it is read.",
				How:  "Stop reading the field, or restore it at the far end. If its loss is accepted, record it with freeze so the lock file carries the decision.",
			})
			continue
		}
		if current.Shape != promised.Shape {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 172),
				Pos:  ch.Pos,
				Rule: RuleContractFieldChanged,
				What: "the field " + promised.Wire + " of " + ref + " was recorded as " + promised.Shape + " and is now " + current.Shape + ".",
				Why:  "It kept its name, so nothing at either end will fail to find it. It will simply be read as the wrong thing, which is the failure that surfaces furthest from its cause.",
				How:  "Follow the change, or hold the far end to the recorded structure. Record whichever with freeze.",
			})
		}
	}
}

// RecordContracts writes the contracts into the baseline.
//
// Only channels that state one. A channel with no contract is left out rather
// than recorded as empty, because unstated and empty are different facts and
// the record exists to keep them apart.
func RecordContracts(channels []ir.Channel, base *baseline.File) int {
	changed := 0
	for _, ch := range channels {
		if ch.Contract == nil {
			continue
		}
		if base.Channels == nil {
			base.Channels = map[string]baseline.Channel{}
		}
		now := wireOf(ch.Contract)
		if rec, ok := base.Channels[ch.Name()]; ok && sameWire(rec.Contract, now) {
			continue
		}
		base.Channels[ch.Name()] = baseline.Channel{Contract: now}
		changed++
	}
	return changed
}
