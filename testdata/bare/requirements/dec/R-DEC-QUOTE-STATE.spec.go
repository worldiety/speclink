package dec

import "github.com/worldiety/speclink/spec"

// RDecQuoteState fixes the persistence pattern of the quotation context.
var RDecQuoteState = spec.Requirement{
	ID:         "R-DEC-QUOTE-STATE",
	Kind:       spec.Decision,
	Discipline: spec.Technical,
	Status:     spec.Normative,
	Title:      "Quotes are kept as current state",
	Text:       "A quote MUST be stored as its current state in a repository, not as a log of changes.",
	Rationale:  "The context answers questions about how a quote stands now; a history would be a second model to keep true.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "13-ablage"},
		{Doc: "requirements/_sources/vorgaben.standard.json", Anchor: "OPS-01"},
	},
}
