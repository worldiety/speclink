package sales

import "go.wdy.de/nago/application/permission"

// Permissions of the quotation domain. Each is bound to the use case it guards
// through the type parameter, and each takes its texts from the translation
// catalogue so the role editor can be read in the user's language.
var (
	PermSubmitQuote  = permission.DeclareCreate[SubmitQuote]("sales.quote.submit", "Quote")
	PermApproveQuote = permission.DeclareFindByID[ApproveQuote]("sales.quote.approve", "Quote")
)
