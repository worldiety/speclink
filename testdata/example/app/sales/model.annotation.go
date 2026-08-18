package sales

import (
	"example.com/erp/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

var _ = spec.For[SubmitQuoteCmd](
	spec.Satisfies(quote.RQuoteSubmit),
)

var _ = spec.For[QuoteSubmitted](
	spec.Satisfies(quote.RQuoteSubmit),
	spec.Transition[QuoteSubmitted]("submitted"),
)

var _ = spec.ForField[SubmitQuoteCmd]("Title",
	spec.Satisfies(quote.RQuoteSubmit),
)

var _ = spec.ForDecl(Namespace,
	spec.Satisfies(quote.RQuoteApprove),
	spec.Rationale("The resource type is the anchor of every access rule on a quote."),
)
