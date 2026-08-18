// Command erp assembles the bounded contexts into a running application.
package main

import "example.com/erp/app/sales"

func main() {
	_ = sales.NewUseCases(emptyOverview{})
}

// emptyOverview stands in for the running projection while the wiring of the
// event store is out of scope for this fixture.
type emptyOverview struct{}

func (emptyOverview) Get(customer string) (sales.QuoteOverview, bool) {
	return sales.QuoteOverview{}, false
}
