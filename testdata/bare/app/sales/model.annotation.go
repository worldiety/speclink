package sales

import (
	"example.com/bare/requirements/dec"
	"example.com/bare/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

// The quote is kept as state, and the repository is what that ruling looks like
// in code.
var _ = spec.For[Quote](
	spec.Satisfies(dec.RDecQuoteState),
)

var _ = spec.For[QuoteRepository](
	spec.Satisfies(dec.RDecQuoteState),
	spec.Rationale("The port is declared where it is used; which store is behind it is decided under cmd, so a test can put a different one there."),
)

// A hand written port. Nothing about the interface says it is storage, so it is
// marked — which is the one fact this architecture annotates rather than infers.
var _ = spec.For[NumberRegistry](
	spec.Persistence(),
	spec.Satisfies(dec.RDecNumbering),
	spec.Rationale("The registry is the gapless counter the ruling asks for, and it outlives any one process."),
)

var _ = spec.ForField[Quote]("ID", spec.Satisfies(dec.RDecQuoteState))
var _ = spec.ForField[Quote]("Number", spec.Satisfies(quote.RQuoteSubmit))
var _ = spec.ForField[Quote]("Status", spec.Satisfies(dec.RDecQuoteState))
