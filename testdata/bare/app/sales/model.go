// Package sales is the quotation context: what the system does with quotes, and
// nothing about how it is stored or shown.
package sales

import "example.com/bare/foundation/data"

// QuoteID identifies a quote.
type QuoteID string

// QuoteNumber is the business number a quote carries once it is submitted.
//
// A named type rather than a bare string, so the rule about what may be in it
// has somewhere to live and every field carrying one inherits it.
type QuoteNumber string

// Quote is the domain model.
type Quote struct {
	ID     QuoteID
	Number QuoteNumber
	Status string
	Note   string
}

// Identity makes the quote an aggregate root.
func (q Quote) Identity() QuoteID { return q.ID }

// QuoteRepository stores quotes as their current state.
//
// The port is declared here because this is where it is used. Which
// implementation is behind it is decided under cmd, so that a test can put a
// different one there.
type QuoteRepository data.Repository[Quote, QuoteID]

// NumberRegistry hands out gapless business numbers.
//
// A hand written port rather than a repository: it stores nothing of the domain
// and only counts. spec.Persistence marks it so, because nothing about the
// interface says it.
type NumberRegistry interface {
	Next() (QuoteNumber, error)
}

// SubmitQuoteCmd carries the input of a submission.
type SubmitQuoteCmd struct {
	QuoteID QuoteID
	Title   string
}
