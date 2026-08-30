package fs

import (
	"fmt"
	"sync/atomic"

	"example.com/bare/app/sales"
)

// Numbers hands out sequential numbers.
type Numbers struct{ last atomic.Int64 }

// NewNumbers returns a registry starting after the given number.
func NewNumbers(last int64) *Numbers {
	n := &Numbers{}
	n.last.Store(last)
	return n
}

func (n *Numbers) Next() (sales.QuoteNumber, error) {
	return sales.QuoteNumber(fmt.Sprintf("Q-%d", n.last.Add(1))), nil
}
