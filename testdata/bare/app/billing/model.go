// Package billing is the invoicing context.
//
// It exists in this skeleton so that two contexts sit beside each other, which
// is where the rules about their boundary start to mean something.
package billing

import "example.com/bare/foundation/data"

// InvoiceID identifies an invoice.
type InvoiceID string

// Invoice is the draft an accepted quote turns into.
//
// The shape is still being worked out, which is why the binding marks it
// spec.Draft(): nothing has been promised about it and it may still change
// freely. That is a claim about this repository, not about a deployment.
type Invoice struct {
	ID    InvoiceID
	Quote string
	Total int64
}

// Identity makes the invoice an aggregate root.
func (i Invoice) Identity() InvoiceID { return i.ID }

// InvoiceRepository stores invoices as their current state.
//
// No adapter maps it to a shape of its own yet, so the promise would sit on
// Invoice itself — which is exactly why it is still a draft.
type InvoiceRepository data.Repository[Invoice, InvoiceID]

// DraftInvoiceCmd carries the input of a draft.
type DraftInvoiceCmd struct {
	Quote string
	Total int64
}
