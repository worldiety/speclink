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

// The withdrawal is not promised yet: the shape may still change in any way,
// and the event may be dropped again. Promotion is the deletion of this term.
var _ = spec.For[QuoteWithdrawn](
	spec.Satisfies(quote.RQuoteApprove),
	spec.Proposal(),
)

var _ = spec.ForDecl(Namespace,
	spec.Satisfies(quote.RQuoteApprove),
	spec.Rationale("The resource type is the anchor of every access rule on a quote."),
)

// Added after the promise, so old messages lack it. Optionality cannot be
// withdrawn later, because those messages cannot be rewritten.
var _ = spec.ForField[QuoteSubmitted]("Channel",
	spec.Optional(),
)
