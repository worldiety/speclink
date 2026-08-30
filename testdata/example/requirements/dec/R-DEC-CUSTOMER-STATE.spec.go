package dec

import "github.com/worldiety/speclink/spec"

// RDecCustomerState fixes the persistence pattern of the reference data.
var RDecCustomerState = spec.Requirement{
	ID:           "R-DEC-CUSTOMER-STATE",
	Kind:         spec.Decision,
	Discipline:   spec.Technical,
	Status:       spec.Normative,
	Title:        "Customers are kept as current state",
	Text:         "Customer master data MUST be stored as its current state in a repository.",
	Rationale:    "For master data the history is not evidence but noise: a corrected typo in a name is not a business event, and the address a quote was sent to is already carried by the quote. Keeping every correction as a fact would make the common read — who is this customer now — the expensive one.",
	Consequences: "A correction overwrites what was there, so an address entered by mistake leaves no trace, and there is no way to tell a fixed typo from a customer who really moved.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "2-kunde"},
	},
}
