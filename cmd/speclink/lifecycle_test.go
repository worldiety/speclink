package main

import (
	"strings"
	"testing"
)

// The set of states a thing can be in is the first question anybody asks about
// it, and until now it was the one question the model recorded without ever
// reading. spec.Transition was parsed into the IR and consumed by nothing: an
// author could write it, get it past every binding rule, and have it mean
// nothing at all.

func TestEventMustNameTheStateItLeadsTo(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "app/sales/model.annotation.go",
		"\tspec.Transition[QuoteWithdrawn](\"withdrawn\"),\n", "")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("an event declares no state and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "QuoteWithdrawn does not say which state it leaves the aggregate in") {
		t.Errorf("expected K15-EVENT-NO-TRANSITION:\n%s", out)
	}
	// The How line carries a usable state rather than an ellipsis. The past
	// participle of the event is what a coarse state is named after, so the
	// suggestion is nearly always the right word.
	if !strings.Contains(out, `spec.Transition[QuoteWithdrawn]("withdrawn")`) {
		t.Errorf("the fix is not spelled out:\n%s", out)
	}
}

// The other direction, and the one that rots silently: the assertion accepts
// any type at all, so a transition naming a deleted or renamed event keeps
// compiling and keeps describing a lifecycle nobody can reach.
func TestTransitionMustNameAnEvent(t *testing.T) {
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "app/sales/model.annotation.go",
		`spec.Transition[QuoteWithdrawn]("withdrawn")`,
		`spec.Transition[SubmitQuoteCmd]("withdrawn")`)

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a transition names something that folds nothing:\n%s", out)
	}
	if !strings.Contains(out, "SubmitQuoteCmd is named as a transition but is not an event") {
		t.Errorf("expected K15-TRANSITION-UNKNOWN:\n%s", out)
	}
}

// An architecture without events is not an unmeasured lifecycle.
//
// This is the distinction the K9 gap turned on, so it is worth pinning rather
// than assuming: go_bare has no events because it is not event sourced, which
// its profile summary already states, and no figure claims a share of
// lifecycles that were checked. Reporting the family as unmeasured here would
// be noise on every run of every project that never wanted events.
func TestLifecycleIsSilentWhereThereAreNoEvents(t *testing.T) {
	out, code := runVerify(t, "../../testdata/bare")
	if code != 0 {
		t.Fatalf("the bare fixture did not verify:\n%s", out)
	}
	if strings.Contains(out, "K15-") {
		t.Errorf("a lifecycle rule fired on an architecture that has no events:\n%s", out)
	}
	if strings.Contains(out, "not measured: lifecycle") {
		t.Errorf("an absent role was reported as an unmeasured direction:\n%s", out)
	}
}
