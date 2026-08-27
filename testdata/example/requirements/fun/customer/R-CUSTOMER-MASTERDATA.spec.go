// Package customer holds the requirements of the customer master data.
package customer

import "github.com/worldiety/speclink/spec"

var RCustomerMasterdata = spec.Requirement{
	ID:         "R-CUSTOMER-MASTERDATA",
	Kind:       spec.Functional,
	Discipline: spec.Business,
	Status:     spec.Normative,
	Title:      "Customer master data",
	Text:       "A customer MUST be held with the name it currently trades under and the internal notes kept about them.",
	Sources: []spec.Source{
		{Doc: "requirements/_sources/sales/quoteflow.md", Anchor: "2-kunde"},
		{Doc: "requirements/_sources/sales/quotescreen.png", Anchor: "kundenfeld"},
	},
}
