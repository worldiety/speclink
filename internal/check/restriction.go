package check

import (
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

const (
	// RuleRestrictionUnproven fires when a restriction states a rule in prose
	// and gives no examples that decide it.
	RuleRestrictionUnproven = "K21-RESTRICTION-UNPROVEN"
	// RuleVectorsUnrestricted fires when examples are given for a type that
	// states no rule.
	RuleVectorsUnrestricted = "K21-VECTORS-UNRESTRICTED"
	// RuleVectorDuplicate fires when one example is given twice for a type.
	RuleVectorDuplicate = "K21-VECTOR-DUPLICATE"
)

// Restriction is what one type promises about the values it may hold.
type Restriction struct {
	// Type is the fully qualified name of the restricted type.
	Type string
	// Text is the rule, in the words it was stated in.
	Text string
	// Valid and Invalid are the examples that decide it.
	Valid   []string
	Invalid []string
	Pos     ir.Position
}

// Restrictions collects what the annotations say about the values types may
// hold, and holds each statement to being decidable.
//
// # Why a rule in prose is refused on its own
//
// A Go type carries what the type system carries. That a string holds at most
// 64 characters, never a control character and never a prefix this system
// reserves is not in the type, and cannot be put there. So it is written in
// words — and words are exactly what a schema generator drops without a sound.
// Both ends then agree about the shape of a message and disagree about what
// may be inside it, which is the same failure as having no agreement at all,
// arrived at with more ceremony.
//
// The examples are what makes the rule a fact rather than an intention. They
// are also the only part of it that survives translation into another language,
// another generator and another team: a case that must be accepted and a case
// that must be refused mean the same thing everywhere.
//
// Both directions are required, and the second is the one that matters.
// Accepting what should be accepted is what an implementation does by
// accident. Refusing what must be refused is the part nobody gets right
// unprompted, and the part that decides whether a reserved prefix is reserved
// or merely mentioned.
func Restrictions(bindings []ir.Binding, out *diag.Set) []Restriction {
	byType := map[string]*Restriction{}

	for _, b := range bindings {
		if b.Target.Kind != ir.TargetType {
			continue
		}
		for _, a := range b.Assertions {
			switch a.Kind {
			case ir.AssertRestrict, ir.AssertValid, ir.AssertInvalid:
			default:
				continue
			}
			r := byType[b.Target.Name]
			if r == nil {
				r = &Restriction{Type: b.Target.Name, Pos: a.Pos}
				byType[b.Target.Name] = r
			}
			switch a.Kind {
			case ir.AssertRestrict:
				r.Text = a.Text
				r.Pos = a.Pos
			case ir.AssertValid:
				r.Valid = append(r.Valid, a.Vectors...)
			case ir.AssertInvalid:
				r.Invalid = append(r.Invalid, a.Vectors...)
			}
		}
	}

	out2 := make([]Restriction, 0, len(byType))
	for _, r := range byType {
		out2 = append(out2, *r)
	}
	sort.Slice(out2, func(i, j int) bool { return out2[i].Pos.Less(out2[j].Pos) })

	for _, r := range out2 {
		checkRestriction(r, out)
	}
	return out2
}

func checkRestriction(r Restriction, out *diag.Set) {
	if r.Text == "" {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 190),
			Pos:  r.Pos,
			Rule: RuleVectorsUnrestricted,
			What: shortName(r.Type) + " gives examples but states no rule.",
			Why:  "An example decides a rule. Without one written down, a reader can see that a value was refused and not why, and the next case nobody thought of has nothing to be judged against.",
			How:  "Add spec.Restrict naming what a value of this type must satisfy.",
		})
		return
	}

	var missing string
	switch {
	case len(r.Valid) == 0 && len(r.Invalid) == 0:
		missing = "no examples"
	case len(r.Valid) == 0:
		missing = "no example it must accept"
	case len(r.Invalid) == 0:
		missing = "no example it must reject"
	}
	if missing != "" {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 191),
			Pos:  r.Pos,
			Rule: RuleRestrictionUnproven,
			What: shortName(r.Type) + " states a rule and gives " + missing + ".",
			Why:  "A rule written only in prose is one a schema generator drops without a sound, and both ends then agree about the shape of a message while disagreeing about what may be in it. The example a conforming implementation must reject is the half nobody gets right unprompted, and the half that decides whether a reserved prefix is reserved or merely mentioned.",
			How:  "Add spec.Valid and spec.Invalid with the cases that decide this rule.",
		})
	}
	checkVectorDuplicates(r, out)
}

// checkVectorDuplicates reports one example given twice.
//
// Cheap to check and worth checking, because a set of vectors is grown a line
// at a time over months, and the same tricky case gets added twice by two
// people who both remembered it. A duplicate is not wrong, but it inflates the
// count that a reader takes as the measure of how well the rule is pinned down.
func checkVectorDuplicates(r Restriction, out *diag.Set) {
	for _, group := range []struct {
		name string
		list []string
	}{{"accept", r.Valid}, {"reject", r.Invalid}} {
		seen := map[string]bool{}
		for _, v := range group.list {
			if !seen[v] {
				seen[v] = true
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 192),
				Pos:  r.Pos,
				Rule: RuleVectorDuplicate,
				What: shortName(r.Type) + " gives the same example to " + group.name + " twice: " + quote(v) + ".",
				Why:  "The number of cases a rule is pinned down by is what a reader takes as the measure of it. A repeated case inflates that number without deciding anything more.",
				How:  "Remove the duplicate.",
			})
		}
	}
}
