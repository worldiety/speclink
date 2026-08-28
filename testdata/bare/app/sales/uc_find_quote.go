package sales

import "example.com/bare/foundation/auth"

// FindQuote reads a single quote.
type FindQuote func(subject auth.Subject, id QuoteID) (Quote, error)

// NewFindQuote builds the read use case.
func NewFindQuote(quotes QuoteRepository) FindQuote {
	return func(subject auth.Subject, id QuoteID) (Quote, error) {
		if err := subject.Audit(PermFindQuote); err != nil {
			return Quote{}, err
		}

		quote, _, err := quotes.FindByID(subject.Context(), id)
		return quote, err
	}
}
