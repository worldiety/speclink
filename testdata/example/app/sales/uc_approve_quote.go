package sales

import (
	"go.wdy.de/nago/application/evs"
	"go.wdy.de/nago/auth"
)

// ApproveQuote approves a quote including legal sign-off.
type ApproveQuote func(subject auth.Subject, cmd ApproveQuoteCmd) (evs.SeqID, error)

// LegalGate decides whether a quote may pass legal review. It receives the
// subject and is where the permission for this capability is checked.
type LegalGate interface {
	Check(subject auth.Subject, quote string) error
}

// NewApproveQuote builds the approval use case.
//
// It never mentions PermApproveQuote itself: the check lives in the gate, which
// is handed the subject. That is the ordinary shape in an event sourced system,
// where authorisation sits next to the invariants in Decide rather than in the
// closure that forwards to it. Reporting it would mean reporting every use case
// of that shape, which is how a rule teaches people to switch it off.
func NewApproveQuote(gate LegalGate) ApproveQuote {
	return func(subject auth.Subject, cmd ApproveQuoteCmd) (evs.SeqID, error) {
		if err := gate.Check(subject, cmd.QuoteID); err != nil {
			return 0, err
		}
		return 1, nil
	}
}
