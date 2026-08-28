package sales

import (
	"context"
	"iter"
	"testing"

	"example.com/bare/foundation/auth"
	"example.com/bare/foundation/permission"
	"example.com/bare/requirements/dec"
	"example.com/bare/requirements/fun/quote"
	"github.com/worldiety/speclink/spec"
)

// countingNumbers records what it handed out, so a test can check the sequence
// rather than the call.
type countingNumbers struct{ issued []string }

func (n *countingNumbers) Next() (string, error) {
	s := string(rune('A' + len(n.issued)))
	n.issued = append(n.issued, s)
	return s, nil
}

// memQuotes is the port, in memory. The whole point of declaring the port in the
// context is that a test can put this behind it.
type memQuotes struct{ saved map[QuoteID]Quote }

func newMemQuotes() *memQuotes { return &memQuotes{saved: map[QuoteID]Quote{}} }

func (m *memQuotes) Save(_ context.Context, q Quote) error { m.saved[q.ID] = q; return nil }

func (m *memQuotes) FindByID(_ context.Context, id QuoteID) (Quote, bool, error) {
	q, ok := m.saved[id]
	return q, ok, nil
}

func (m *memQuotes) DeleteByID(_ context.Context, id QuoteID) error {
	delete(m.saved, id)
	return nil
}

func (m *memQuotes) All(context.Context) iter.Seq2[Quote, error] {
	return func(yield func(Quote, error) bool) {
		for _, q := range m.saved {
			if !yield(q, nil) {
				return
			}
		}
	}
}

func allowed() auth.Subject {
	return auth.Static("tester", PermSubmitQuote, PermFindQuote)
}

func TestSubmitDrawsAGaplessNumber(t *testing.T) {
	numbers := &countingNumbers{}
	submit := NewSubmitQuote(newMemQuotes(), numbers)

	for i := range 3 {
		if _, err := submit(allowed(), SubmitQuoteCmd{QuoteID: QuoteID(rune('1' + i))}); err != nil {
			t.Fatalf("submission %d: %v", i, err)
		}
	}

	seen := map[string]bool{}
	for _, n := range numbers.issued {
		if seen[n] {
			t.Fatalf("number %q drawn twice: %v", n, numbers.issued)
		}
		seen[n] = true
	}
	if len(numbers.issued) != 3 {
		t.Fatalf("drew %d numbers for 3 submissions", len(numbers.issued))
	}

	spec.Verified(t, quote.RQuoteSubmit, dec.RDecNumbering)
}

// A subject without the permission must not get through, and the check has to
// happen before anything is written.
func TestSubmitRefusesWithoutPermission(t *testing.T) {
	quotes := newMemQuotes()
	submit := NewSubmitQuote(quotes, &countingNumbers{})

	if _, err := submit(auth.Anonymous(), SubmitQuoteCmd{QuoteID: "q1"}); err == nil {
		t.Fatal("an anonymous subject submitted a quote")
	}
	if len(quotes.saved) != 0 {
		t.Error("a refused submission still wrote a quote")
	}

	spec.Verified(t, quote.RQuoteSubmit)
}

// A quote is stored as its current state, so what goes in comes back out — and
// what comes back is the state, not a history to fold.
func TestQuoteIsKeptAsCurrentState(t *testing.T) {
	quotes := newMemQuotes()
	submit := NewSubmitQuote(quotes, &countingNumbers{})
	find := NewFindQuote(quotes)

	number, err := submit(allowed(), SubmitQuoteCmd{QuoteID: "q1", Title: "First"})
	if err != nil {
		t.Fatal(err)
	}

	got, err := find(allowed(), "q1")
	if err != nil {
		t.Fatal(err)
	}
	if got.Number != number || got.Status != "submitted" {
		t.Fatalf("read back %+v", got)
	}

	spec.Verified(t, quote.RQuoteLookup, dec.RDecQuoteState)
}

func TestPermissionsAreDeclaredOnce(t *testing.T) {
	seen := map[permission.ID]bool{}
	for _, p := range permission.All() {
		if seen[p.ID] {
			t.Errorf("permission %q is declared twice", p.ID)
		}
		seen[p.ID] = true
		if p.UseCase == "" {
			t.Errorf("permission %q names no use case", p.ID)
		}
	}
}
