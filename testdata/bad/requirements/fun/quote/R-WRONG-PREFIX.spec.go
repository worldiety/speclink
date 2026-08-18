// Package quote deliberately violates layout and shape rules.
package quote

import "github.com/worldiety/speclink/spec"

// Prefix WRONG does not match the directory quote/.
var RWrongPrefix = spec.Requirement{
	ID:         "R-WRONG-PREFIX",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Text:       "Something is required.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/missing.md", Anchor: "nope"},
	},
}

// A decision without a rationale, in the wrong directory, with the wrong prefix.
var RQuoteDecision = spec.Requirement{
	ID:         "R-QUOTE-DECISION",
	Kind:       spec.Decision,
	Discipline: spec.Technical,
	Status:     spec.Normative,
	Text:       "We decided something.",
}
