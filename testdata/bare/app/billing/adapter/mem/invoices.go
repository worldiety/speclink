// Package mem keeps invoices in memory.
//
// The invoicing context is still being worked out, so its store is deliberately
// throw-away: there is no shape on disk yet to promise, which is what lets
// Invoice stay a draft.
package mem

import (
	"context"
	"iter"
	"sync"

	"example.com/bare/app/billing"
)

// Invoices stores invoices in a map.
type Invoices struct {
	mu    sync.Mutex
	saved map[billing.InvoiceID]billing.Invoice
}

// NewInvoices returns an empty repository.
func NewInvoices() *Invoices {
	return &Invoices{saved: map[billing.InvoiceID]billing.Invoice{}}
}

func (s *Invoices) Save(_ context.Context, i billing.Invoice) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.saved[i.ID] = i
	return nil
}

func (s *Invoices) FindByID(_ context.Context, id billing.InvoiceID) (billing.Invoice, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	i, ok := s.saved[id]
	return i, ok, nil
}

func (s *Invoices) DeleteByID(_ context.Context, id billing.InvoiceID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.saved, id)
	return nil
}

func (s *Invoices) All(context.Context) iter.Seq2[billing.Invoice, error] {
	return func(yield func(billing.Invoice, error) bool) {
		s.mu.Lock()
		defer s.mu.Unlock()
		for _, i := range s.saved {
			if !yield(i, nil) {
				return
			}
		}
	}
}
