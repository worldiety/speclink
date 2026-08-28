package billing

import (
	"context"
	"iter"
	"testing"

	"example.com/bare/foundation/auth"
	"example.com/bare/requirements/fun/bill"
	"github.com/worldiety/speclink/spec"
)

// memInvoices is the port, in memory, so the test builds no adapter.
type memInvoices struct{ saved map[InvoiceID]Invoice }

func newMemInvoices() *memInvoices { return &memInvoices{saved: map[InvoiceID]Invoice{}} }

func (m *memInvoices) Save(_ context.Context, i Invoice) error { m.saved[i.ID] = i; return nil }

func (m *memInvoices) FindByID(_ context.Context, id InvoiceID) (Invoice, bool, error) {
	i, ok := m.saved[id]
	return i, ok, nil
}

func (m *memInvoices) DeleteByID(_ context.Context, id InvoiceID) error {
	delete(m.saved, id)
	return nil
}

func (m *memInvoices) All(context.Context) iter.Seq2[Invoice, error] {
	return func(yield func(Invoice, error) bool) {
		for _, i := range m.saved {
			if !yield(i, nil) {
				return
			}
		}
	}
}

func TestDraftNamesTheQuoteItBills(t *testing.T) {
	draft := NewDraftInvoice(newMemInvoices())
	subject := auth.Static("tester", PermDraftInvoice)

	id, err := draft(subject, DraftInvoiceCmd{Quote: "Q-1", Total: 100})
	if err != nil {
		t.Fatal(err)
	}
	if id != "INV-Q-1" {
		t.Fatalf("the invoice does not name its quote: %q", id)
	}
	if _, err := draft(subject, DraftInvoiceCmd{}); err == nil {
		t.Error("an invoice was drafted without a quote")
	}

	spec.Verified(t, bill.RBillDraft)
}
