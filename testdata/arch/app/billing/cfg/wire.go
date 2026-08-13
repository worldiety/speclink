// Package cfgbilling is the wiring layer and may see both sides.
package cfgbilling

import (
	"example.com/arch/app/billing"
	_ "go.wdy.de/nago/presentation/ui"
)

// Enable wires the billing module.
func Enable() billing.IssueInvoice { return billing.NewIssueInvoice() }
