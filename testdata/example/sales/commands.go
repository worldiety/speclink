// Package sales implements the quotation domain.
package sales

// SubmitQuoteUC submits an approved quote and draws a quote number.
type SubmitQuoteUC func(cmd SubmitQuoteCmd) error

// SubmitQuoteCmd carries the input of a submission.
type SubmitQuoteCmd struct {
	QuoteID string
	Title   string
}

// QuoteSubmitted is the fact recorded on submission.
type QuoteSubmitted struct {
	QuoteID     string
	QuoteNumber string
}

// PermSubmitQuote guards the submission use case.
var PermSubmitQuote = "sales.quote.submit"

// NewSubmitQuote wires the submission use case.
func NewSubmitQuote() SubmitQuoteUC { return func(SubmitQuoteCmd) error { return nil } }
