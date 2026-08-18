package sales

import (
	"go.wdy.de/nago/application/evs"
)

// QuoteOverview is the read model of the quotation context: how many quotes
// have been submitted, and the number of the most recent one.
//
// It is folded from the event log and holds no storage of its own, so it can be
// rebuilt at any time by constructing it again. Persisting it would turn a
// derived view into a second truth.
type QuoteOverview struct {
	Submitted int
	LastQuote string
}

// Clone returns a deep copy. Readers of a projection are handed a clone, so an
// accidental mutation on a returned value can never corrupt the folded state.
func (o *QuoteOverview) Clone() *QuoteOverview {
	if o == nil {
		return nil
	}
	c := *o
	return &c
}

// newQuoteOverview builds the read model over the event source.
//
// The fold is registered here, at the target, rather than on the event: Evolve
// targets exactly one aggregate, while a read model deliberately crosses them.
func newQuoteOverview(src evs.Source) *evs.Singleton[*QuoteOverview] {
	p := evs.NewSingleton[*QuoteOverview](src, evs.ProjectionOptions{})
	evs.Project(p,
		func(QuoteSubmitted) evs.Unit { return evs.TheUnit() },
		func(s *QuoteOverview, e QuoteSubmitted) {
			s.Submitted++
			s.LastQuote = e.QuoteNumber
		},
	)
	return p
}
