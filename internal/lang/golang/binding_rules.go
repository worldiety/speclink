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
	// Draft is about a persisted shape, so it attaches where a shape lives:
	// the package holding the types, the type itself, or a single field of it.
	// A function or a variable has no shape on the wire.
	ir.AssertDraft: {ir.TargetType, ir.TargetField, ir.TargetPackage},
	// Optional is a statement about one field and has no meaning anywhere else:
	// a whole type cannot be absent from a message, it simply is the message.
	ir.AssertOptional: {ir.TargetField},
	// Verified is stated by a test about itself, so its only target is the
	// test function it stands in.
	ir.AssertVerified: {ir.TargetFunc},
	// A restriction and its vectors are about the values a type may hold, so
	// they attach to the type. A single field is restricted by restricting the
	// named type it is declared with, which is also what makes the rule
	// reusable instead of repeated at every field that carries it.
	ir.AssertRestrict: {ir.TargetType},
	ir.AssertValid:    {ir.TargetType},
	ir.AssertInvalid:  {ir.TargetType},
	// A claim is about one field. A whole type cannot be an assertion by the
	// sender: it is the message, and the question is which parts of it the
	// receiver may believe.
	ir.AssertClaim: {ir.TargetField},
	// Persistence is a statement about a type: an interface is a port, a
	// struct is a shape. A function has neither.
	ir.AssertPersistence: {ir.TargetType},
	// StoredAs says this struct is the written form of a domain type. Only a
	// type has a written form, and only a type can stand in for another.
	ir.AssertStoredAs: {ir.TargetType},
}

// repeatable lists the assertions that may appear more than once per target.
// The others state a single fact that cannot sensibly be given twice.
var repeatable = map[ir.AssertionKind]bool{
	ir.AssertSatisfies:  true,
	ir.AssertTransition: true,
	ir.AssertWaive:      true,
	ir.AssertTerm:       true,
	// A test may demonstrate several requirements, and may say so in more than
	// one place: the statement is about the point control reached, so two calls
	// on two paths are two different statements.
	ir.AssertVerified: true,
	// Vectors accumulate. Several calls are how a set is grown a line at a
	// time as cases are found, and merging them into one would make every
	// addition a change to an existing line.
	ir.AssertValid:   true,
	ir.AssertInvalid: true,
}

// checkTargetAllowed validates one binding against the rules above and against
// the style's vocabulary.
func (p *Package) checkTargetAllowed(b ir.Binding, style Style, out *diag.Set) {
	seen := map[ir.AssertionKind]bool{}

	for _, a := range b.Assertions {
		if !style.Admits(a.Kind) {
			// Not a silent no-op. The author expected an effect, and the
			// effect is that something else already provides it — which is
			// only useful to know if it is said.
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseBinding, 8),
				Pos:  a.Pos,
				What: "spec." + exportedName(a.Kind) + " is not available in this architecture.",
				Why:  "It exists for architectures that cannot state the fact by themselves. Here " + styleReason(style, a.Kind) + ", so the term would be a second source for something that already has one.",
				How:  "Remove it.",
			})
			continue
		}
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

// styleReason explains what already states the fact, so the refusal is
// actionable rather than merely correct.
func styleReason(style Style, k ir.AssertionKind) string {
	if why, ok := style.Why[k]; ok {
		return why
	}
	return "the framework already states it"
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
	case ir.AssertDraft:
		return "Draft"
	case ir.AssertOptional:
		return "Optional"
	case ir.AssertPersistence:
		return "Persistence"
	case ir.AssertStoredAs:
		return "StoredAs"
	case ir.AssertRestrict:
		return "Restrict"
	case ir.AssertValid:
		return "Valid"
	case ir.AssertInvalid:
		return "Invalid"
	case ir.AssertClaim:
		return "Claim"
	case ir.AssertVerified:
		return "Verified"
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
