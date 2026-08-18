package sales

import (
	"example.com/erp/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[SubmitQuote](
	spec.Satisfies(quote.RQuoteSubmit),
	spec.Help(`Submit the approved quote. The system draws the next quote
number from the central registry.`),
)

var _ = spec.ForDecl(PermSubmitQuote,
	spec.Rationale("Submission draws from a gapless registry and is therefore not repeatable without consequence."),
)
