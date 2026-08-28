package billing

import (
	"testing"

	"example.com/bare/foundation/auth"
	"example.com/bare/requirements/fun/bill"
	"github.com/worldiety/speclink/spec"
)

func TestDraftNamesTheQuoteItBills(t *testing.T) {
	draft := NewDraftInvoice()
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
