package sales

import (
	"example.com/erp/anforderungen/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[FindQuoteOverview](
	spec.Satisfies(quote.RQuoteOverview),
)
