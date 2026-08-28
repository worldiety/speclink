package fs

import (
	"example.com/bare/app/sales"
	"example.com/bare/requirements/dec"
	"example.com/bare/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

// The adapter keeps a shape of its own and maps to it, so what is promised is
// this struct rather than the domain model.
//
// Without this the promise would sit on sales.Quote, which is the stricter
// reading and the right default: two structs in two packages look identical
// whether one maps to the other or neither knows the other exists, and nothing
// but a statement can tell them apart.
var _ = spec.For[QuoteStore](
	spec.StoredAs[sales.Quote](),
	spec.Rationale("The domain model may be restructured without touching a byte on disk; what is written down is the half that cannot be changed after the fact."),
)

// The stored fields trace to requirements in their own right. What is on disk
// outlives the domain model that produced it, so it answers for itself rather
// than through the type it happens to mirror today.
var _ = spec.ForField[QuoteStore]("ID", spec.Satisfies(dec.RDecQuoteState))
var _ = spec.ForField[QuoteStore]("Number", spec.Satisfies(quote.RQuoteSubmit))
var _ = spec.ForField[QuoteStore]("Status", spec.Satisfies(dec.RDecQuoteState))

// Note arrived after the shape was promised. Optional is the statement that
// records written earlier do not carry it, and it cannot be withdrawn later.
var _ = spec.ForField[QuoteStore]("Note", spec.Optional(), spec.Satisfies(quote.RQuoteLookup))
