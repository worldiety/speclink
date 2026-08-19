package golang

import (
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// Binding rules: which assertion may attach to which kind of target, and which
// assertion may repeat.
//
// These are not style rules. An assertion at the wrong target is either
// meaningless (a lifecycle transition on a struct field) or a sign that the
// author meant a different construct — both worth stopping the build for.

// allowedTargets lists the target kinds each assertion accepts.
var allowedTargets = map[ir.AssertionKind][]ir.TargetKind{
	ir.AssertSatisfies:  {ir.TargetType, ir.TargetFunc, ir.TargetVar, ir.TargetConst, ir.TargetField, ir.TargetPackage},
	ir.AssertTransition: {ir.TargetType, ir.TargetFunc, ir.TargetVar, ir.TargetConst},
	ir.AssertExternal:   {ir.TargetType},
	ir.AssertHelp:       {ir.TargetType, ir.TargetFunc, ir.TargetVar, ir.TargetConst},
	ir.AssertTerm:       {ir.TargetType, ir.TargetPackage},
	ir.AssertRationale:  {ir.TargetType, ir.TargetFunc, ir.TargetVar, ir.TargetConst, ir.TargetPackage},
	ir.AssertWaive:      {ir.TargetType, ir.TargetFunc, ir.TargetVar, ir.TargetConst, ir.TargetField, ir.TargetPackage},
	// Proposal is about a persisted shape, so it attaches where a shape lives:
	// the package holding the types, the type itself, or a single field of it.
	// A function or a variable has no shape on the wire.
	ir.AssertProposal: {ir.TargetType, ir.TargetField, ir.TargetPackage},
}

// repeatable lists the assertions that may appear more than once per target.
// The others state a single fact that cannot sensibly be given twice.
var repeatable = map[ir.AssertionKind]bool{
	ir.AssertSatisfies:  true,
	ir.AssertTransition: true,
	ir.AssertWaive:      true,
	ir.AssertTerm:       true,
}

// checkTargetAllowed validates one binding against the rules above.
func (p *Package) checkTargetAllowed(b ir.Binding, out *diag.Set) {
	seen := map[ir.AssertionKind]bool{}

	for _, a := range b.Assertions {
		if !targetPermitted(a.Kind, b.Target.Kind) {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseBinding, 6),
				Pos:  a.Pos,
				What: "spec." + exportedName(a.Kind) + " cannot be attached to a " + b.Target.Kind.String() + ".",
				Why:  "The assertion states nothing meaningful about this kind of construct.",
				How:  "Attach it to " + targetHint(a.Kind) + ", or remove it.",
			})
		}
		if seen[a.Kind] && !repeatable[a.Kind] {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseBinding, 7),
				Pos:  a.Pos,
				What: "spec." + exportedName(a.Kind) + " is given more than once for " + shortTarget(b.Target) + ".",
				Why:  "This assertion states a single fact; a second one either repeats or contradicts the first.",
				How:  "Merge both into one spec." + exportedName(a.Kind) + " call.",
			})
		}
		seen[a.Kind] = true
	}
}

func targetPermitted(a ir.AssertionKind, t ir.TargetKind) bool {
	for _, allowed := range allowedTargets[a] {
		if allowed == t {
			return true
		}
	}
	return false
}

// exportedName maps an assertion kind back to the exported Go name, so
// diagnostics quote what the author actually wrote.
func exportedName(k ir.AssertionKind) string {
	switch k {
	case ir.AssertSatisfies:
		return "Satisfies"
	case ir.AssertTransition:
		return "Transition"
	case ir.AssertExternal:
		return "External"
	case ir.AssertHelp:
		return "Help"
	case ir.AssertTerm:
		return "Term"
	case ir.AssertRationale:
		return "Rationale"
	case ir.AssertWaive:
		return "Waive"
	case ir.AssertProposal:
		return "Proposal"
	}
	return "unknown"
}

// targetHint names the binding forms an assertion accepts, for the How line.
func targetHint(k ir.AssertionKind) string {
	var names []string
	for _, t := range allowedTargets[k] {
		switch t {
		case ir.TargetType:
			names = append(names, "spec.For[T]")
		case ir.TargetFunc, ir.TargetVar, ir.TargetConst:
			names = appendUnique(names, "spec.ForDecl")
		case ir.TargetField:
			names = append(names, "spec.ForField[T]")
		case ir.TargetPackage:
			names = append(names, "spec.ForPackage")
		}
	}
	return join(names, " or ")
}

// appendUnique keeps the hint list free of duplicates: func, var and const all
// map to the same binding form.
func appendUnique(list []string, s string) []string {
	for _, existing := range list {
		if existing == s {
			return list
		}
	}
	return append(list, s)
}

func join(parts []string, sep string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	}
	out := parts[0]
	for _, p := range parts[1 : len(parts)-1] {
		out += ", " + p
	}
	return out + sep + parts[len(parts)-1]
}

// shortTarget renders a target the way its author wrote it.
func shortTarget(t ir.Target) string {
	name := t.Name
	if i := lastIndexByte(name, '.'); i >= 0 {
		name = name[i+1:]
	}
	if t.Kind == ir.TargetField {
		return name + "." + t.Field
	}
	if name == "" {
		return t.Package
	}
	return name
}
