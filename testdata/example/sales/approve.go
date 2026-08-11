package sales

// ApproveQuoteUC approves a quote including legal sign-off.
type ApproveQuoteUC func(cmd ApproveQuoteCmd) error

// ApproveQuoteCmd carries the input of an approval.
type ApproveQuoteCmd struct {
	QuoteID       string
	LegalApproved bool
}

// QuoteApproved is the fact recorded on approval.
type QuoteApproved struct {
	QuoteID string
}
