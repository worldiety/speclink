// Package quote holds the requirements of the quotation domain.
package quote

import (
	"example.com/erp/requirements/dec"
	"github.com/worldiety/speclink/spec"
)

var RQuoteSubmit = spec.Requirement{
	ID:          "R-QUOTE-SUBMIT",
	Kind:        spec.Functional,
	Discipline:  spec.Business,
	Status:      spec.Normative,
	Title:       "Quote number on submission",
	Text:        "On submitting an approved quote a sequential, duplicate free quote number MUST be drawn.",
	DerivedFrom: []spec.Requirement{dec.RDecNumbering},
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "8-abgabe"},
		{Doc: "requirements/_sources/sales/quotescreen.png", Anchor: "abgabeknopf"},
	},
}
