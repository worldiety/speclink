// Package sales implements the quotation domain.
package sales

import (
	"context"

	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

// QuoteAggregate is the consistency boundary of a quote.
type QuoteAggregate struct {
	ID     string
	Status string
}

// Identity makes the quote an aggregate root.
func (q QuoteAggregate) Identity() string { return q.ID }

// SubmitQuoteCmd carries the input of a submission.
type SubmitQuoteCmd struct {
	QuoteID string
	Title   string
}

// Decide validates the submission and yields the resulting facts.
func (c SubmitQuoteCmd) Decide(s auth.Subject, q *QuoteAggregate) ([]QuoteSubmitted, error) {
	if err := s.Audit(PermSubmitQuote); err != nil {
		return nil, err
	}
	return []QuoteSubmitted{{QuoteID: c.QuoteID}}, nil
}

// QuoteSubmitted is the fact recorded on submission.
type QuoteSubmitted struct {
	QuoteID     string
	QuoteNumber string
}

// Evolve folds the fact into the aggregate.
func (e QuoteSubmitted) Evolve(_ context.Context, q *QuoteAggregate) error {
	q.Status = "submitted"
	return nil
}

// Discriminator is the stable serialisation tag.
func (e QuoteSubmitted) Discriminator() evs.Discriminator { return "sales.quote.submitted.v1" }

// ApproveQuoteCmd carries the input of an approval.
type ApproveQuoteCmd struct {
	QuoteID       string
	LegalApproved bool
}

// Namespace is the ReBAC resource type of the quotation domain.
const Namespace = "erp.sales.quote"

// QuoteWithdrawn is the fact recorded when a submitted quote is taken back.
//
// It is still a proposal: the shape has not been promised, so fields may be
// added, removed or retyped freely, and the whole event may be dropped again.
type QuoteWithdrawn struct {
	QuoteID string
	Reason  string
}

// Evolve folds the fact into the aggregate.
func (e QuoteWithdrawn) Evolve(_ context.Context, q *QuoteAggregate) error {
	q.Status = "withdrawn"
	return nil
}

// Discriminator is the stable serialisation tag.
func (e QuoteWithdrawn) Discriminator() evs.Discriminator { return "sales.quote.withdrawn.v1" }
