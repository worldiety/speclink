package sales

import "go.wdy.de/nago/pkg/data"

// QuoteRepository stores the quotes that are kept as current state rather than
// derived from the log.
//
// Naming the instantiation once is the framework idiom: callers depend on this
// name instead of repeating the type arguments, and the aggregate it serves
// stays visible in one place.
type QuoteRepository data.Repository[QuoteAggregate, string]
