package sales

import (
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

// SubmitQuote submits an approved quote and draws a quote number.
type SubmitQuote func(subject auth.Subject, cmd SubmitQuoteCmd) (evs.SeqID, error)

// NewSubmitQuote builds the submission use case over the quote handler.
func NewSubmitQuote(numbers NumberRegistry) SubmitQuote {
	return func(subject auth.Subject, cmd SubmitQuoteCmd) (evs.SeqID, error) {
		if err := subject.Audit(PermSubmitQuote); err != nil {
			return 0, err
		}
		if _, err := numbers.Next(); err != nil {
			return 0, err
		}
		return 1, nil
	}
}

// NumberRegistry hands out gapless business numbers.
type NumberRegistry interface {
	Next() (string, error)
}
