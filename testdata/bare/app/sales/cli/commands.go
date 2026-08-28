// Package clisales exposes the quotation context on the command line.
//
// Each command takes the one use case it runs, for the same reason a route
// does: a command that took the bundle could not be exercised without building
// the whole context.
package clisales

import (
	stdflag "flag"
	"fmt"

	"example.com/bare/app/sales"
	"example.com/bare/foundation/auth"
	"example.com/bare/foundation/flag"
)

// Submit is the command that submits a quote.
func Submit(submit sales.SubmitQuote) flag.Command {
	return flag.Command{
		Name:  "submit-quote",
		Short: "Submit an approved quote and draw its number",
		Run: func(subject auth.Subject, args []string) error {
			fs := stdflag.NewFlagSet("submit-quote", stdflag.ContinueOnError)
			id := fs.String("id", "", "the quote to submit")
			title := fs.String("title", "", "the quote title")
			if err := fs.Parse(args); err != nil {
				return err
			}

			number, err := submit(subject, sales.SubmitQuoteCmd{QuoteID: sales.QuoteID(*id), Title: *title})
			if err != nil {
				return err
			}
			fmt.Println(number)
			return nil
		},
	}
}
