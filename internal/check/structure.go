package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// RuleConstructUnbound fires when a recognised construct carries no
// requirement binding.
const RuleConstructUnbound = "K1-CONSTRUCT-UNBOUND"

// Structure is the result of the forward coverage analysis.
type Structure struct {
	// Constructs is every recognised construct.
	Constructs []ir.Construct
	// Unbound lists the constructs that need a requirement but have none.
	Unbound []ir.Construct
	// Required is the number of constructs that had to be bound.
	Required int
}

// Ratio returns the bound share of constructs that require a requirement.
func (s Structure) Ratio() float64 {
	if s.Required == 0 {
		return 1
	}
	return float64(s.Required-len(s.Unbound)) / float64(s.Required)
}

// CoverConstructs performs the forward direction: is every construct that
// carries business meaning tied to a requirement?
//
// This is where inference earns its keep. Without recognisers speclink cannot
// tell which constructs are use cases, commands or events, and would either
// have to demand an annotation on everything — unbearable — or on nothing,
// which measures nothing. The framework already states the architectural role
// unambiguously; the annotation only adds what the code cannot say, namely
// which requirement the construct exists for.
func CoverConstructs(constructs []ir.Construct, bindings []ir.Binding, d ir.Dialect, out *diag.Set) Structure {
	s := Structure{Constructs: constructs}

	bound := map[string]bool{}
	for _, b := range bindings {
		// A field binding satisfies a requirement for that field, not for the
		// type it belongs to. Counting it would let one annotated field make a
		// whole command look accounted for.
		if b.Target.Kind == ir.TargetField {
			continue
		}
		for _, a := range b.Assertions {
			if a.Kind == ir.AssertSatisfies && len(a.Requirements) > 0 {
				bound[b.Target.Name] = true
			}
		}
	}
	waived := ir.CollectWaivers(bindings)

	sorted := append([]ir.Construct(nil), constructs...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Pos.File != sorted[j].Pos.File {
			return sorted[i].Pos.File < sorted[j].Pos.File
		}
		return sorted[i].Pos.Line < sorted[j].Pos.Line
	})

	for _, c := range sorted {
		if !c.Kind.NeedsRequirement() {
			continue
		}
		s.Required++
		if bound[c.Name] {
			continue
		}
		s.Unbound = append(s.Unbound, c)

		if waived.Has(c.Name, RuleConstructUnbound) {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 20),
			Pos:  c.Pos,
			Rule: RuleConstructUnbound,
			What: shortName(c.Name) + " is " + c.Kind.WithArticle() + " but is bound to no requirement.",
			Why:  "Recognised because it " + c.Evidence + ". A construct carrying business meaning must trace back to something that was asked for, otherwise nobody can tell whether it should exist.",
			How:  "Add `" + d.BindConstruct(c.Name) + "` in " + d.AnnotationFile(c.Pos.File) + ", or waive the rule with a reason.",
		})
	}
	return s
}

func shortName(qualified string) string {
	if i := lastIndexByte(qualified, '.'); i >= 0 {
		return qualified[i+1:]
	}
	return qualified
}

func lastIndexByte(s string, b byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == b {
			return i
		}
	}
	return -1
}
