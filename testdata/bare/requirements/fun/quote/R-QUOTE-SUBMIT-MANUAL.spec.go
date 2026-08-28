package quote

import "github.com/worldiety/speclink/spec"

// RQuoteSubmitManual is what the rule used to be.
//
// It is kept rather than deleted because the successor has to point at
// something: a requirement that simply vanished leaves the records written
// under it unexplained, and anybody reading a two year old quote needs to know
// that its number was typed in by hand rather than drawn.
var RQuoteSubmitManual = spec.Requirement{
	ID:         "R-QUOTE-SUBMIT-MANUAL",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Superseded,
	Title:      "Quote number entered on submission",
	Text:       "On submitting a quote the clerk MUST enter the quote number.",
	Rationale:  "Superseded once the central registry existed; a typed number could collide and did.",
}
