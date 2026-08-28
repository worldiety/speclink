package sales

import (
	"example.com/bare/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[FindQuote](
	spec.Satisfies(quote.RQuoteLookup),
)
