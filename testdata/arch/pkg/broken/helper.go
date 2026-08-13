// Package broken is infrastructure that knows about the domain.
package broken

import (
	"example.com/arch/app/billing"
	"go.wdy.de/nago/auth"
)

// Render formats an invoice, which ties this helper to one domain.
func Render(uc billing.IssueInvoice) string { return "x" }

// Audit is a use case signature and does not belong in infrastructure.
type Audit func(subject auth.Subject, id string) error
