package billing

import (
	"example.com/bare/requirements/fun/bill"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[DraftInvoice](
	spec.Satisfies(bill.RBillDraft),
)
