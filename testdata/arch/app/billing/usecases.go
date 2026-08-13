// Package billing violates every use case rule on purpose.
package billing

import (
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
)

// IssueInvoice is declared in usecases.go instead of uc_issue_invoice.go, and
// returns no error.
type IssueInvoice func(subject auth.Subject, id string) string

// NewIssueInvoice never consults the subject and reads package level state.
func NewIssueInvoice() IssueInvoice {
	return func(subject auth.Subject, id string) string {
		return defaultPrefix + id
	}
}

// DropInvoice has a permission that is declared but never checked.
type DropInvoice func(subject auth.Subject, id string) error

// NewDropInvoice ignores its permission entirely.
func NewDropInvoice() DropInvoice {
	return func(subject auth.Subject, id string) error {
		return nil
	}
}

// defaultPrefix is package level mutable state a use case must not reach for.
var defaultPrefix = "INV-"

// PermDropInvoice carries hardcoded texts instead of translated ones.
var PermDropInvoice = permission.Declare[DropInvoice](
	"billing.invoice.drop", "Rechnung verwerfen", "Träger dürfen Rechnungen verwerfen.")
