package sales

import (
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

// ApproveQuoteUC approves a quote including legal sign-off.
type ApproveQuoteUC func(auth.Subject, ApproveQuoteCmd) (evs.SeqID, error)

// ApproveQuoteCmd carries the input of an approval.
type ApproveQuoteCmd struct {
	QuoteID       string
	LegalApproved bool
}
