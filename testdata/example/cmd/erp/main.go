// Command erp assembles the bounded contexts into a running application.
package main

import (
	"example.com/erp/app/sales"

	"go.wdy.de/nago/auth"
)

func main() {
	_ = sales.NewUseCases(emptyOverview{}, emptyOverview{}, openGate{})
}

// emptyOverview stands in for the running projection while the wiring of the
// event store is out of scope for this fixture.
type emptyOverview struct{}

func (emptyOverview) Get(customer string) (sales.QuoteOverview, bool) {
	return sales.QuoteOverview{}, false
}

func (emptyOverview) All() []sales.QuoteOverview { return nil }

// openGate stands in for the legal review while the wiring of it is out of
// scope for this fixture.
type openGate struct{}

func (openGate) Check(subject auth.Subject, quote string) error { return nil }
