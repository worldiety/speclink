package sales

import (
	"example.com/erp/anforderungen/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[SubmitQuoteUC](
	spec.Satisfies(quote.RQuoteSubmit),
	spec.Help(`Submit the approved quote. The system draws the next quote
number from the central registry.`),
)

var _ = spec.For[SubmitQuoteCmd](
	spec.Satisfies(quote.RQuoteSubmit),
)

var _ = spec.For[QuoteSubmitted](
	spec.Satisfies(quote.RQuoteSubmit),
	spec.Transition[QuoteSubmitted]("submitted"),
)

var _ = spec.ForVar(&PermSubmitQuote,
	spec.Rationale("Submission draws from a gapless registry and is therefore not repeatable without consequence."),
)

var _ = spec.ForField[SubmitQuoteCmd]("Title",
	spec.Satisfies(quote.RQuoteSubmit),
)
