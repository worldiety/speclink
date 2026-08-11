// Package diag renders findings for two audiences from one source.
//
// The error output is consumed mostly by an LLM (principle P7 of the concept:
// diagnostics are the interface, not decoration). It is therefore prescriptive
// rather than descriptive: every finding says what is wrong, why it is wrong,
// and what to do about it. The quality of these messages determines how fast
// the loop converges and thus whether the whole approach is practical.
//
// There are no severities. A finding is an error and the run fails. See
// docs/annotations.md §1.8 for the reasoning; in short: the Go compiler behaves
// the same way, warnings get ignored, and because annotations are part of the
// normal build a compile error cannot be warned past anyway.
package diag

import (
	"github.com/worldiety/speclink/internal/ir"
)

// Phase groups findings by the stage that produced them. Each stage only runs
// when the previous one is clean.
type Phase string

const (
	// PhaseWhitelist rejects Go constructs that the annotation subset forbids.
	PhaseWhitelist Phase = "V1"
	// PhaseBinding checks that a binding attaches to a legal target.
	PhaseBinding Phase = "V3"
	// PhaseRedundant rejects annotations of facts that are already inferable.
	PhaseRedundant Phase = "V4"
	// PhaseResolve resolves sources, anchors, rule IDs and the requirement DAG.
	PhaseResolve Phase = "V5"
	// PhaseSemantic covers the specification rules K1 to K4.
	PhaseSemantic Phase = "V6"
)

// Phase V2 has no constant on purpose: it is the Go compilation itself. It runs
// before speclink and its findings are produced by the Go compiler.

// Finding is a single diagnostic.
//
// The three text fields are separate rather than one prose blob so that the
// JSON consumer can use them individually.
type Finding struct {
	// Code is stable and greppable, e.g. "SPEC-V1-001".
	Code string
	Pos  ir.Position
	// What states the violation in one line.
	What string
	// Why names the rule or principle behind it.
	Why string
	// How gives the concrete next action. This is the field that makes the
	// difference between a diagnostic and a prompt.
	How string
	// Rule is the waivable rule ID, empty when the finding cannot be waived.
	Rule string
}

// Code builds a diagnostic code from a phase and a number, e.g. SPEC-V1-003.
func Code(p Phase, n int) string {
	return "SPEC-" + string(p) + "-" + pad3(n)
}

func pad3(n int) string {
	switch {
	case n < 10:
		return "00" + itoa(n)
	case n < 100:
		return "0" + itoa(n)
	default:
		return itoa(n)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
