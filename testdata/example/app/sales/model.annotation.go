package sales

import (
	"example.com/erp/requirements/dec"
	"example.com/erp/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

// The quote is event sourced and the customer is not, in one and the same
// context. That is the ordinary case, not a defect: the approval of a quote is
// evidence and must stay readable as it was, while a customer is read by its
// current name. Because the two mix here, the decision is recorded per
// aggregate rather than once for the package.
var _ = spec.For[QuoteAggregate](
	spec.Satisfies(dec.RDecQuoteEventSourced),
)

var _ = spec.For[QuoteRepository](
	spec.Satisfies(dec.RDecQuoteEventSourced),
	spec.Rationale("The repository serves the folded state of the log, not a second source of truth: it is rebuilt from the events and may be dropped at any time."),
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
	spec.Draft(),
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
