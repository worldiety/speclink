package quote

import "github.com/worldiety/speclink/spec"

var RQuoteOverview = spec.Requirement{
	ID:         "R-QUOTE-OVERVIEW",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Quotation overview per customer",
	Text:       "Sales MUST see per customer how many quotes have been submitted and which number the last one carries.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "10-übersicht"},
		{Doc: "requirements/_sources/sales/quotescreen.png", Anchor: "angebotsliste"},
	},
}
