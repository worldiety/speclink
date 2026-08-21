package sales

import (
	"sort"

	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
)

// ListQuotes reads the whole quotation list, ordered by identity.
type ListQuotes func(auth.Subject) ([]QuoteOverview, error)

// NewListQuotes builds the list query out of two combinators rather than
// writing a closure of its own.
//
// The permission is applied by wrapping, not inside a body, and the only
// function literal in sight is the comparator handed to the sort — which is an
// argument of a helper and has nothing to do with the use case.
func NewListQuotes(all QuoteOverviewLister) ListQuotes {
	return guard(listAll(all, func(a, b QuoteOverview) bool {
		return a.LastQuote < b.LastQuote
	}), PermListQuotes)
}

// QuoteOverviewLister is the read side the list depends on.
type QuoteOverviewLister interface {
	All() []QuoteOverview
}

// listAll reads every row of a projection and orders it.
func listAll[T any, L interface{ All() []T }](l L, less func(a, b T) bool) func() []T {
	return func() []T {
		out := l.All()
		sort.Slice(out, func(i, j int) bool { return less(out[i], out[j]) })
		return out
	}
}

// guard wraps an unguarded reader into a query that enforces a permission.
func guard[T any](load func() T, perm permission.ID) func(auth.Subject) (T, error) {
	return func(s auth.Subject) (T, error) {
		if err := s.Audit(perm); err != nil {
			var zero T
			return zero, err
		}
		return load(), nil
	}
}
