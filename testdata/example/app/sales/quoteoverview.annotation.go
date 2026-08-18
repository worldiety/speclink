package sales

import (
	"example.com/erp/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[QuoteOverview](
	spec.Satisfies(quote.RQuoteOverview),
	spec.Help(`The overview is folded from the submission events, so it is
always as current as the log and never needs a migration of its own.`),
)
