package quote

import "github.com/worldiety/speclink/spec"

var RQuoteApprove = spec.Requirement{
	ID:         "R-QUOTE-APPROVE",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Approval gate",
	Text:       "A quote MUST pass an approval gate including legal sign-off before it can be submitted.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "9-versand"},
	},
}
