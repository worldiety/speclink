package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFreezeRefusesToLaunderABreak is the property the whole guard rests on.
//
// freeze is the only command that writes, so it is also the only place a
// breaking change could be made official. If it simply recorded whatever the
// source says, every rule reading the baseline would be one command away from
// being worthless.
func TestFreezeRefusesToLaunderABreak(t *testing.T) {
	out, code := runSpeclink(t, "freeze", "../../testdata/bad", "./...")
	if code == 0 {
		t.Fatalf("freeze must refuse while a promise is broken:\n%s", out)
	}
	if !strings.Contains(out, "nothing was recorded") {
		t.Errorf("the refusal must say that nothing was written:\n%s", out)
	}
	for _, code := range []string{"SPEC-V6-091", "SPEC-V6-092", "SPEC-V6-093", "SPEC-V6-094", "SPEC-V6-095"} {
		if !strings.Contains(out, code) {
			t.Errorf("the refusal must name %s:\n%s", code, out)
		}
	}
	// The shape it was asked to record is not a break, so it must not be the
	// reason for the refusal.
	if strings.Contains(out, "SPEC-V6-090") {
		t.Errorf("an unrecorded shape is what freeze is for, not a reason to refuse:\n%s", out)
	}

	// And it must not have touched the file it refused to write.
	before, err := os.ReadFile(filepath.Join("..", "..", "testdata", "bad", "speclink.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(before), "settled.v2") {
		t.Error("the broken discriminator reached the baseline")
	}
}

// TestFreezeIsIdempotent guards the everyday case: once recorded, running it
// again must be a no-op rather than a churning diff.
func TestFreezeIsIdempotent(t *testing.T) {
	out, code := runSpeclink(t, "freeze", "../../testdata/example", "./...")
	if code != 0 {
		t.Fatalf("freeze over a clean fixture must succeed, got exit %d:\n%s", code, out)
	}
	if !strings.Contains(out, "nothing to record") {
		t.Errorf("expected a no-op, got:\n%s", out)
	}
}

// TestFreezeSkipsProposals pins the point of the marker: an unpromised shape
// stays out of the record until somebody decides otherwise.
func TestFreezeSkipsProposals(t *testing.T) {
	lock, err := os.ReadFile(filepath.Join("..", "..", "testdata", "example", "speclink.lock"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(lock)

	if !strings.Contains(text, "QuoteSubmitted") {
		t.Error("a frozen event must be recorded")
	}
	if strings.Contains(text, "QuoteWithdrawn") {
		t.Error("a proposal must not be recorded; nothing about it has been promised")
	}
}

// TestVerifyExampleWithBaseline guards that the conformant fixture stays clean
// with the baseline in place, which is the state a project lives in.
func TestVerifyExampleWithBaseline(t *testing.T) {
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("expected a clean run, got exit %d:\n%s", code, out)
	}
}
