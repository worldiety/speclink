package dec

import "github.com/worldiety/speclink/spec"

// RDecInvoiceState fixes the persistence pattern of the invoicing context.
var RDecInvoiceState = spec.Requirement{
	ID:         "R-DEC-INVOICE-STATE",
	Kind:       spec.Decision,
	Discipline: spec.Technical,
	Status:     spec.Normative,
	Title:      "Invoices are kept as current state",
	Text:       "An invoice MUST be stored as its current state, not as a log of changes.",
	Rationale:  "The context answers questions about how an invoice stands now, and the shape is still being worked out; a history would fix a second model before the first one is settled.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "14-rechnung"},
	},
}
