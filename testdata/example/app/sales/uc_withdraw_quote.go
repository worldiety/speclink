package sales

import (
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

// WithdrawQuote takes a submitted quote back.
type WithdrawQuote func(subject auth.Subject, cmd WithdrawQuoteCmd) (evs.SeqID, error)

// WithdrawQuoteCmd carries the input of a withdrawal.
type WithdrawQuoteCmd struct {
	QuoteID string
	Reason  string
}

// NewWithdrawQuote builds the withdrawal use case.
func NewWithdrawQuote() WithdrawQuote {
	return func(subject auth.Subject, cmd WithdrawQuoteCmd) (evs.SeqID, error) {
		if err := subject.Audit(PermWithdrawQuote); err != nil {
			return 0, err
		}
		return 1, nil
	}
}
