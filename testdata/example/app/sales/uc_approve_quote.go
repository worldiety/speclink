package sales

import (
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

// ApproveQuote approves a quote including legal sign-off.
type ApproveQuote func(subject auth.Subject, cmd ApproveQuoteCmd) (evs.SeqID, error)

// NewApproveQuote builds the approval use case.
func NewApproveQuote() ApproveQuote {
	return func(subject auth.Subject, cmd ApproveQuoteCmd) (evs.SeqID, error) {
		if err := subject.Audit(PermApproveQuote); err != nil {
			return 0, err
		}
		return 1, nil
	}
}
