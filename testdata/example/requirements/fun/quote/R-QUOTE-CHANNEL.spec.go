package quote

import "github.com/worldiety/speclink/spec"

var RQuoteChannel = spec.Requirement{
	ID:         "R-QUOTE-CHANNEL",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Dispatch channel",
	Text:       "The channel a quote was sent through MUST be recorded with the submission.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "9-versand"},
	},
}
