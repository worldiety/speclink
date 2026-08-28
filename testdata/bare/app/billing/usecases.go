package billing

// UseCases are the capabilities of this context.
type UseCases struct {
	DraftInvoice DraftInvoice
}

// NewUseCases wires the context.
func NewUseCases(invoices InvoiceRepository) UseCases {
	return UseCases{DraftInvoice: NewDraftInvoice(invoices)}
}
