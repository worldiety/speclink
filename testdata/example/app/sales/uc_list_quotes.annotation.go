package sales

import (
	"example.com/erp/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[ListQuotes](
	spec.Satisfies(quote.RQuoteOverview),
)
