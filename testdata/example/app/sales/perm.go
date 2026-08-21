package sales

import (
	"example.com/erp/pkg/permtext"

	"go.wdy.de/nago/application/permission"
)

// Permissions of the quotation domain. Each is bound to the use case it guards
// through the type parameter, and each takes its texts from the translation
// catalogue so the role editor can be read in the user's language.
var (
	PermSubmitQuote       = permission.DeclareCreate[SubmitQuote]("sales.quote.submit", "Quote")
	PermApproveQuote      = permission.DeclareFindByID[ApproveQuote]("sales.quote.approve", "Quote")
	PermFindQuoteOverview = permission.DeclareFindAll[FindQuoteOverview]("sales.quote.overview", "Quote")
)

// PermWithdrawQuote takes its texts from a helper of the project rather than
// spelling out the catalogue call here. The helper goes through i18n, which is
// what the rule is about.
var PermWithdrawQuote = permission.Declare[WithdrawQuote](
	"sales.quote.withdraw",
	permtext.Name("sales.quote.withdraw", "Withdraw quote"),
	"Holders may take back a submitted quote.",
)

// PermListQuotes guards the list query. It is never named inside a closure of
// the use case, only handed to the combinator that wraps it.
var PermListQuotes = permission.Declare[ListQuotes](
	"sales.quote.list",
	permtext.Name("sales.quote.list", "List quotes"),
	permtext.Desc("sales.quote.list", "Holders may read the quotation list."),
)
