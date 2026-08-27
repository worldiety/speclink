package sales

import (
	"encoding/json"
	"errors"
	"testing"

	"example.com/erp/requirements/dec"
	"example.com/erp/requirements/fun/customer"
	"example.com/erp/requirements/fun/quote"
	"example.com/erp/requirements/nfr"
	spec "github.com/worldiety/speclink/spec"
	"go.wdy.de/nago/application/permission"
	"go.wdy.de/nago/auth"
)

// spec.Verified stands at the end of each test, because that is where it means
// something: the line is written when control reaches it, so a test that fails
// halfway leaves no record behind.

// --- R-QUOTE-SUBMIT ---

// countingRegistry stands in for the central number registry and records what
// it handed out, so the test can check the sequence rather than the call.
type countingRegistry struct {
	issued []string
	err    error
}

func (r *countingRegistry) Next() (string, error) {
	if r.err != nil {
		return "", r.err
	}
	n := string(rune('A' + len(r.issued)))
	r.issued = append(r.issued, n)
	return n, nil
}

func TestSubmitQuoteDrawsAGaplessNumber(t *testing.T) {
	registry := &countingRegistry{}
	submit := NewSubmitQuote(registry)

	for i := 0; i < 3; i++ {
		if _, err := submit(allowAll{}, SubmitQuoteCmd{QuoteID: "q"}); err != nil {
			t.Fatalf("submission %d: %v", i, err)
		}
	}

	// One number per submission, none repeated: that is what "sequential,
	// duplicate free" has to mean at this level.
	if len(registry.issued) != 3 {
		t.Fatalf("drew %d numbers for 3 submissions: %v", len(registry.issued), registry.issued)
	}
	seen := map[string]bool{}
	for _, n := range registry.issued {
		if seen[n] {
			t.Fatalf("number %q drawn twice: %v", n, registry.issued)
		}
		seen[n] = true
	}

	spec.Verified(t, quote.RQuoteSubmit)
}

// A submission that cannot draw a number must not report success, or the quote
// exists without one.
func TestSubmitQuoteFailsWhenNoNumberCanBeDrawn(t *testing.T) {
	submit := NewSubmitQuote(&countingRegistry{err: errors.New("registry down")})

	if _, err := submit(allowAll{}, SubmitQuoteCmd{QuoteID: "q"}); err == nil {
		t.Fatal("submission succeeded without a number")
	}

	spec.Verified(t, quote.RQuoteSubmit)
}

// --- R-QUOTE-APPROVE ---

type gate struct {
	err     error
	asked   int
	subject auth.Subject
}

func (g *gate) Check(subject auth.Subject, _ string) error {
	g.asked++
	g.subject = subject
	return g.err
}

func TestApproveQuoteRequiresLegalSignOff(t *testing.T) {
	refusing := &gate{err: errors.New("legal did not sign off")}
	if _, err := NewApproveQuote(refusing)(allowAll{}, ApproveQuoteCmd{QuoteID: "q"}); err == nil {
		t.Fatal("a quote passed the gate without legal sign-off")
	}

	accepting := &gate{}
	if _, err := NewApproveQuote(accepting)(allowAll{}, ApproveQuoteCmd{QuoteID: "q"}); err != nil {
		t.Fatalf("a signed-off quote was rejected: %v", err)
	}
	if accepting.asked != 1 {
		t.Errorf("the gate was consulted %d times, want 1", accepting.asked)
	}
	// The gate is where the permission is checked, so it has to be handed the
	// subject rather than a copy of the command.
	if accepting.subject == nil {
		t.Error("the gate was not given the subject and cannot check anything")
	}

	spec.Verified(t, quote.RQuoteApprove)
}

// --- R-QUOTE-CHANNEL ---

// The channel has to survive being written down. It was added after the shape
// was promised and is therefore optional, which is exactly the situation in
// which a field quietly stops being stored.
func TestSubmissionRecordsTheChannel(t *testing.T) {
	stored, err := json.Marshal(QuoteSubmitted{QuoteID: "q", QuoteNumber: "A", Channel: "email"})
	if err != nil {
		t.Fatal(err)
	}

	var read QuoteSubmitted
	if err := json.Unmarshal(stored, &read); err != nil {
		t.Fatal(err)
	}
	if read.Channel != "email" {
		t.Fatalf("the channel did not survive storage: %q", read.Channel)
	}

	spec.Verified(t, quote.RQuoteChannel)
}

// --- R-QUOTE-OVERVIEW ---

func TestOverviewCountsSubmissionsAndKeepsTheLastNumber(t *testing.T) {
	view := &QuoteOverview{}
	applyQuoteSubmitted(view, QuoteSubmitted{QuoteID: "q1", QuoteNumber: "A"})
	applyQuoteSubmitted(view, QuoteSubmitted{QuoteID: "q2", QuoteNumber: "B"})

	if view.Submitted != 2 {
		t.Errorf("counted %d submissions, want 2", view.Submitted)
	}
	if view.LastQuote != "B" {
		t.Errorf("last quote is %q, want B", view.LastQuote)
	}

	// Readers are handed a clone, so a caller mutating what it was given cannot
	// corrupt the folded state.
	clone := view.Clone()
	clone.Submitted = 99
	if view.Submitted != 2 {
		t.Error("a mutation on a returned value reached the projection")
	}

	spec.Verified(t, quote.RQuoteOverview)
}

// --- R-CUSTOMER-MASTERDATA ---

// Both directions together, because a mapping that drops a field on the way out
// and never reads it back looks correct from either side alone. That is how
// Notes was once lost: written, stored empty, read back empty, with nothing in
// any type to show for it.
func TestCustomerMappingKeepsEveryField(t *testing.T) {
	want := Customer{ID: "c1", Name: "Contoso GmbH", Notes: []string{"pays late", "prefers email"}}

	entity, err := customerToEntity(want)
	if err != nil {
		t.Fatal(err)
	}
	got, err := customerFromEntity(entity)
	if err != nil {
		t.Fatal(err)
	}

	if got.ID != want.ID || got.Name != want.Name {
		t.Errorf("identity or name lost: %+v", got)
	}
	if len(got.Notes) != len(want.Notes) {
		t.Fatalf("notes lost in the mapping: %v", got.Notes)
	}
	for i := range want.Notes {
		if got.Notes[i] != want.Notes[i] {
			t.Errorf("note %d changed: %q", i, got.Notes[i])
		}
	}

	spec.Verified(t, customer.RCustomerMasterdata)
}

// --- R-DEC-QUOTE-EVENTSOURCED ---

// The write side is a sequence of facts, so the aggregate must be derivable by
// replaying them rather than being stored as a state somebody edits.
func TestQuoteStateIsFoldedFromItsFacts(t *testing.T) {
	aggregate := QuoteAggregate{ID: "q"}
	if err := (QuoteSubmitted{QuoteID: "q"}).Evolve(t.Context(), &aggregate); err != nil {
		t.Fatal(err)
	}
	if aggregate.Status != "submitted" {
		t.Fatalf("status after submission is %q", aggregate.Status)
	}

	// Replaying the same facts in the same order must reach the same state, or
	// the log is not the source of truth.
	replayed := QuoteAggregate{ID: "q"}
	if err := (QuoteSubmitted{QuoteID: "q"}).Evolve(t.Context(), &replayed); err != nil {
		t.Fatal(err)
	}
	if replayed != aggregate {
		t.Errorf("replay produced %+v, want %+v", replayed, aggregate)
	}

	spec.Verified(t, dec.RDecQuoteEventSourced)
}

// --- R-NFR-TRACEABILITY ---

// Every stored fact has to say what it is about, or a sequence of them cannot
// be attributed to anything after the fact.
func TestEveryStoredFactIdentifiesItsSubject(t *testing.T) {
	facts := []struct {
		name string
		data any
	}{
		{"QuoteSubmitted", QuoteSubmitted{QuoteID: "q", QuoteNumber: "A"}},
		{"QuoteWithdrawn", QuoteWithdrawn{QuoteID: "q", Reason: "price"}},
	}

	for _, f := range facts {
		raw, err := json.Marshal(f.data)
		if err != nil {
			t.Fatalf("%s: %v", f.name, err)
		}
		var fields map[string]any
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatalf("%s: %v", f.name, err)
		}
		if id, ok := fields["QuoteID"]; !ok || id == "" {
			t.Errorf("%s does not carry the quote it is about: %s", f.name, raw)
		}
	}

	spec.Verified(t, nfr.RNfrTraceability)
}

// allowAll is a subject that grants everything.
//
// Authorisation is checked by K5-UC-AUTHZ over the source; these tests are
// about what the use cases do once the subject is past the gate. Embedding the
// interface keeps the stub to the one method that is exercised, and leaves the
// rest to panic loudly if a test ever depends on it by accident.
type allowAll struct{ auth.Subject }

func (allowAll) Audit(_ permission.ID) error { return nil }
