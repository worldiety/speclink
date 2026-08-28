// Package restbilling exposes the invoicing context over HTTP.
package restbilling

import (
	"net/http"

	"example.com/bare/app/billing"
	"example.com/bare/foundation/rest"
)

// Mount registers the routes of this context.
func Mount(mux *http.ServeMux, who rest.Authenticator, draft billing.DraftInvoice) {
	mux.Handle("POST /invoices/draft", rest.Handle(who, draft))
}
