// Command erp serves the quotation context over HTTP and on the command line.
//
// This is the only place that names an adapter. Everything above it takes the
// port, which is what lets a test put a different implementation behind it.
package main

import (
	"log"
	"net/http"
	"os"

	"example.com/bare/app/billing"
	"example.com/bare/app/billing/adapter/mem"
	restbilling "example.com/bare/app/billing/rest"
	"example.com/bare/app/sales"
	"example.com/bare/app/sales/adapter/fs"
	clisales "example.com/bare/app/sales/cli"
	restsales "example.com/bare/app/sales/rest"
	"example.com/bare/foundation/auth"
	"example.com/bare/foundation/flag"
	"example.com/bare/foundation/rest"
)

func main() {
	quotes := fs.NewQuotes("var/quotes")
	numbers := fs.NewNumbers(0)
	uc := sales.NewUseCases(quotes, numbers)

	if len(os.Args) > 1 && os.Args[1] == "serve" {
		serve(uc)
		return
	}
	if err := flag.Run(os.Stdout, auth.System(), os.Args[1:], clisales.Submit(uc.SubmitQuote)); err != nil {
		os.Exit(1)
	}
}

func serve(uc sales.UseCases) {
	mux := http.NewServeMux()
	restsales.Mount(mux, anonymous, uc.SubmitQuote)
	restbilling.Mount(mux, anonymous, billing.NewUseCases(mem.NewInvoices()).DraftInvoice)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

// anonymous is the placeholder authenticator of a skeleton.
//
// A real one reads a token or a session. This one grants nothing, so that a
// project which forgets to replace it fails closed rather than open.
func anonymous(*http.Request) (auth.Subject, error) { return auth.Anonymous(), nil }

var _ rest.Authenticator = anonymous
