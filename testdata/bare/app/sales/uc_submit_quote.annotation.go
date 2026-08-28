package sales

import (
	"example.com/bare/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[SubmitQuote](
	spec.Satisfies(quote.RQuoteSubmit),
	spec.Help(`Submit an approved quote. The system draws the next quote number
from the central registry.`),
)
