package sales

// UseCases bundles the capabilities of the quotation context.
//
// Callers depend on this bundle rather than on the internals, and NewUseCases
// is the single place where the shared dependencies are threaded through.
type UseCases struct {
	SubmitQuote       SubmitQuote
	ApproveQuote      ApproveQuote
	FindQuoteOverview FindQuoteOverview
	WithdrawQuote     WithdrawQuote
	ListQuotes        ListQuotes
}

// NewUseCases wires the quotation use cases.
func NewUseCases(view QuoteOverviewReader, all QuoteOverviewLister, gate LegalGate) UseCases {
	return UseCases{
		SubmitQuote:       NewSubmitQuote(nopRegistry{}),
		ApproveQuote:      NewApproveQuote(gate),
		FindQuoteOverview: NewFindQuoteOverview(view),
		WithdrawQuote:     NewWithdrawQuote(),
		ListQuotes:        NewListQuotes(all),
	}
}

type nopRegistry struct{}

func (nopRegistry) Next() (string, error) { return "Q-1", nil }
