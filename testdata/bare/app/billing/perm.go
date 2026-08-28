package billing

import "example.com/bare/foundation/permission"

var PermDraftInvoice = permission.Declare[DraftInvoice]("billing.invoice.draft", "Draft invoice", "Holders may draft an invoice from a submitted quote.")
