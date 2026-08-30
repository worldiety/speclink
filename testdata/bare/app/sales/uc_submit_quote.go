package sales

import "example.com/bare/foundation/auth"

// SubmitQuote submits an approved quote and draws its number.
type SubmitQuote func(subject auth.Subject, cmd SubmitQuoteCmd) (QuoteNumber, error)

// NewSubmitQuote builds the submission use case.
func NewSubmitQuote(quotes QuoteRepository, numbers NumberRegistry) SubmitQuote {
	return func(subject auth.Subject, cmd SubmitQuoteCmd) (QuoteNumber, error) {
		if err := subject.Audit(PermSubmitQuote); err != nil {
			return "", err
		}

		number, err := numbers.Next()
		if err != nil {
			return "", err
		}
		return number, quotes.Save(subject.Context(), Quote{ID: cmd.QuoteID, Number: number, Status: "submitted"})
	}
}
