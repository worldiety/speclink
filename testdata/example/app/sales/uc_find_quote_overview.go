package sales

import (
	"go.wdy.de/nago/application/rebac"
	"go.wdy.de/nago/auth"
)

// FindQuoteOverview reads the quotation overview of one customer.
type FindQuoteOverview func(subject auth.Subject, customer string) (QuoteOverview, error)

// NewFindQuoteOverview builds the overview query over the read model.
//
// It authorises per resource rather than globally: the permission alone does
// not decide, the customer the caller asks about does.
func NewFindQuoteOverview(view QuoteOverviewReader) FindQuoteOverview {
	return func(subject auth.Subject, customer string) (QuoteOverview, error) {
		if err := subject.AuditResource(Namespace, rebac.Instance(customer), PermFindQuoteOverview); err != nil {
			return QuoteOverview{}, err
		}
		o, _ := view.Get(customer)
		return o, nil
	}
}

// QuoteOverviewReader is the read side the query depends on. Taking the reader
// rather than the projection keeps the use case testable without a running
// fold.
type QuoteOverviewReader interface {
	Get(customer string) (QuoteOverview, bool)
}
