package dec

import "github.com/worldiety/speclink/spec"

// RDecQuoteEventSourced fixes the persistence pattern of the quotation write
// side.
var RDecQuoteEventSourced = spec.Requirement{
	ID:         "R-DEC-QUOTE-EVENTSOURCED",
	Kind:       spec.Decision,
	Discipline: spec.Technical,
	Status:     spec.Normative,
	Title:      "Quotations are kept as a log of facts",
	Text:       "The write side of a quotation MUST be stored as an append only sequence of events.",
	Rationale:  "What a quote said when it was approved is itself the record, and an approval that can be edited afterwards proves nothing. Keeping only the current state would answer what a quote says today, which is not the question the approval gate asks.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "9-versand"},
	},
}
