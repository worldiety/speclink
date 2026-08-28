package billing

import (
	"errors"

	"example.com/bare/foundation/auth"
)

// DraftInvoice turns a submitted quote into a draft invoice.
type DraftInvoice func(subject auth.Subject, cmd DraftInvoiceCmd) (InvoiceID, error)

// NewDraftInvoice builds the drafting use case.
func NewDraftInvoice(invoices InvoiceRepository) DraftInvoice {
	return func(subject auth.Subject, cmd DraftInvoiceCmd) (InvoiceID, error) {
		if err := subject.Audit(PermDraftInvoice); err != nil {
			return "", err
		}
		if cmd.Quote == "" {
			return "", errors.New("an invoice needs the quote it bills")
		}

		id := InvoiceID("INV-" + cmd.Quote)
		return id, invoices.Save(subject.Context(), Invoice{ID: id, Quote: cmd.Quote, Total: cmd.Total})
	}
}
