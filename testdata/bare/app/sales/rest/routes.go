// Package restsales exposes the quotation context over HTTP.
//
// It takes the use cases it calls and nothing more. The standard library's
// router is enough: a pattern carries the method and the path variables, which
// is the whole of what a route needed a library for.
package restsales

import (
	"net/http"

	"example.com/bare/app/sales"
	"example.com/bare/foundation/auth"
	"example.com/bare/foundation/rest"
)

// Submit exposes the submission use case.
func Submit(who rest.Authenticator, submit sales.SubmitQuote) http.HandlerFunc {
	return rest.Handle(who, func(subject auth.Subject, cmd sales.SubmitQuoteCmd) (sales.QuoteNumber, error) {
		return submit(subject, cmd)
	})
}

// Mount registers the routes of this context.
//
// It takes each use case separately rather than the bundle, so that a test for
// one route builds one use case.
func Mount(mux *http.ServeMux, who rest.Authenticator, submit sales.SubmitQuote) {
	mux.Handle("POST /quotes/submit", Submit(who, submit))
}
