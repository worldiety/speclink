package fs

import (
	"fmt"
	"sync/atomic"
)

// Numbers hands out sequential numbers.
type Numbers struct{ last atomic.Int64 }

// NewNumbers returns a registry starting after the given number.
func NewNumbers(last int64) *Numbers {
	n := &Numbers{}
	n.last.Store(last)
	return n
}

func (n *Numbers) Next() (string, error) { return fmt.Sprintf("Q-%d", n.last.Add(1)), nil }
