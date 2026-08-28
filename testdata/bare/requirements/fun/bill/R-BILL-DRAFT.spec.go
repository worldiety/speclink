// Package bill holds the requirements of the invoicing domain.
package bill

import "github.com/worldiety/speclink/spec"

// RBillDraft is the reason the drafting use case exists.
var RBillDraft = spec.Requirement{
	ID:         "R-BILL-DRAFT",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Draft an invoice from a quote",
	Text:       "A submitted quote MUST be turnable into a draft invoice that names the quote it bills.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "14-rechnung"},
	},
}
