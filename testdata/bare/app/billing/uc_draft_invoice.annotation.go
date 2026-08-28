package billing

import (
	"example.com/bare/requirements/dec"
	"example.com/bare/requirements/fun/bill"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[DraftInvoice](
	spec.Satisfies(bill.RBillDraft),
)

// The shape is not promised yet, so the guard that would demand a recorded
// baseline stands down. Draft is a statement that nothing has been written
// under this shape that anybody must keep reading.
var _ = spec.For[Invoice](
	spec.Draft(),
	spec.Satisfies(bill.RBillDraft, dec.RDecInvoiceState),
	spec.Rationale("The invoicing context is still being worked out; nothing outside it reads these records yet."),
	spec.Waive("K14-REQ-UNVERIFIED", "The ruling is that an invoice is held as it currently stands rather than as a history, and the repository interface with no event type is what discharges it. While the only store is in memory and the shape is a draft, a test could assert nothing beyond a map behaving like a map."),
)

var _ = spec.For[InvoiceRepository](
	spec.Satisfies(dec.RDecInvoiceState),
)

var _ = spec.ForField[Invoice]("ID", spec.Satisfies(dec.RDecInvoiceState))
var _ = spec.ForField[Invoice]("Quote", spec.Satisfies(bill.RBillDraft))
var _ = spec.ForField[Invoice]("Total", spec.Satisfies(bill.RBillDraft))
