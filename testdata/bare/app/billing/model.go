// Package billing is the invoicing context.
//
// It exists in this skeleton so that two contexts sit beside each other, which
// is where the rules about their boundary start to mean something.
package billing

// InvoiceID identifies an invoice.
type InvoiceID string

// DraftInvoiceCmd carries the input of a draft.
type DraftInvoiceCmd struct {
	Quote string
	Total int64
}
