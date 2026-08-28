package sales

// UseCases are the capabilities of this context.
//
// The bundle takes the weight out of the place a context is assembled: one
// constructor rather than a dozen. It is deliberately not what a handler takes,
// because a handler that depends on every use case cannot be tested with one.
type UseCases struct {
	SubmitQuote SubmitQuote
	FindQuote   FindQuote
}

// NewUseCases wires the context over its ports.
func NewUseCases(quotes QuoteRepository, numbers NumberRegistry) UseCases {
	return UseCases{
		SubmitQuote: NewSubmitQuote(quotes, numbers),
		FindQuote:   NewFindQuote(quotes),
	}
}
