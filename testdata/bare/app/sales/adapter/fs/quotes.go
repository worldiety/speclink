// Package fs stores quotes as files.
//
// An adapter implements a port and is named after the technology behind it, so
// that a second one can sit beside it. Nothing outside cmd imports this: the
// context declares what it needs, this provides it, and the entry point is
// where the two meet.
package fs

import (
	"context"
	"encoding/json"
	"iter"
	"os"
	"path/filepath"
	"sync"

	"example.com/bare/app/sales"
)

// QuoteStore is the stored shape.
//
// It is a separate type from the domain model on purpose: the domain may be
// restructured without touching a byte on disk, and what is written down is the
// half that cannot be changed after the fact.
type QuoteStore struct {
	ID     string `json:"id"`
	Number string `json:"number"`
	Status string `json:"status"`
	// Note was added after the shape was promised, so every quote written
	// until then lacks it. That is not something a later release can undo,
	// which is why it is declared optional rather than simply added.
	Note string `json:"note,omitempty"`
}

// Quotes stores quotes under a directory.
type Quotes struct {
	dir string
	mu  sync.Mutex
}

// NewQuotes returns a repository over a directory.
func NewQuotes(dir string) *Quotes { return &Quotes{dir: dir} }

func (q *Quotes) Save(_ context.Context, quote sales.Quote) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	raw, err := json.Marshal(QuoteStore{ID: string(quote.ID), Number: quote.Number, Status: quote.Status, Note: quote.Note})
	if err != nil {
		return err
	}
	if err := os.MkdirAll(q.dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(q.path(quote.ID), raw, 0o644)
}

func (q *Quotes) FindByID(_ context.Context, id sales.QuoteID) (sales.Quote, bool, error) {
	raw, err := os.ReadFile(q.path(id))
	if os.IsNotExist(err) {
		return sales.Quote{}, false, nil
	}
	if err != nil {
		return sales.Quote{}, false, err
	}

	var stored QuoteStore
	if err := json.Unmarshal(raw, &stored); err != nil {
		return sales.Quote{}, false, err
	}
	return sales.Quote{ID: sales.QuoteID(stored.ID), Number: stored.Number, Status: stored.Status, Note: stored.Note}, true, nil
}

func (q *Quotes) DeleteByID(_ context.Context, id sales.QuoteID) error {
	err := os.Remove(q.path(id))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (q *Quotes) All(ctx context.Context) iter.Seq2[sales.Quote, error] {
	return func(yield func(sales.Quote, error) bool) {
		entries, err := os.ReadDir(q.dir)
		if err != nil {
			yield(sales.Quote{}, err)
			return
		}
		for _, e := range entries {
			quote, _, err := q.FindByID(ctx, sales.QuoteID(e.Name()))
			if !yield(quote, err) {
				return
			}
		}
	}
}

func (q *Quotes) path(id sales.QuoteID) string { return filepath.Join(q.dir, string(id)) }
