package quote

import "github.com/worldiety/speclink/spec"

// RQuoteLookup is the reason the read use case exists.
var RQuoteLookup = spec.Requirement{
	ID:         "R-QUOTE-LOOKUP",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Look up a quote",
	Text:       "Sales MUST be able to look up a single quote together with its number.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "12-auskunft"},
	},
}
