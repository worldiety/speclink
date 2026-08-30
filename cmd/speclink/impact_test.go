package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// The chain runs source segment -> requirement -> derived requirement ->
// construct, and until now every one of those edges was computed for a
// percentage and thrown away. These pin the two questions the loop actually
// asks.

// Somebody edits a paragraph. Which code does that reach?
func TestImpactFromASourceSegment(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "impact", "../../testdata/example",
		"requirements/_sources/sales/quoteflow.md#8-abgabe")
	if code != 0 {
		t.Fatalf("impact failed with %d:\n%s", code, out)
	}
	for _, want := range []string{"R-QUOTE-SUBMIT", "sales.SubmitQuote"} {
		if !strings.Contains(out, want) {
			t.Errorf("the paragraph does not reach %s:\n%s", want, out)
		}
	}
}

// An agent is handed a diff. Which requirements does it touch?
//
// The answer has to be per file. Constructs carry a qualified name rather than
// a position, so matching on the package is the obvious implementation and it
// returns every requirement the package has — eight of eight in this fixture,
// which is not an answer to anything. What makes it exact is the sidecar
// convention: the annotation file beside a changed file names precisely the
// constructs declared in it.
func TestImpactFromAFileIsPerFileNotPerPackage(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "impact", "../../testdata/example",
		"-format", "json", "app/sales/uc_submit_quote.go")
	if code != 0 {
		t.Fatalf("impact failed with %d:\n%s", code, out)
	}

	got := decodeImpact(t, out)
	if len(got.Traced) != 1 {
		t.Fatalf("got %d traces, want 1", len(got.Traced))
	}
	reqs := got.Traced[0].Requirements
	if len(reqs) != 1 || reqs[0] != "R-QUOTE-SUBMIT" {
		t.Errorf("a one file change reached %v; the package has eight requirements and this must not be all of them", reqs)
	}
}

// A decision three levels up reaches everything under it, and that is exactly
// the change nobody traces by hand. The derivation graph was built and cycle
// checked from the beginning and never once walked.
func TestImpactWalksTheDerivationGraph(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "impact", "../../testdata/example", "-format", "json", "R-DEC-NUMBERING")
	if code != 0 {
		t.Fatalf("impact failed with %d:\n%s", code, out)
	}

	got := decodeImpact(t, out)
	tr := got.Traced[0]
	if len(tr.Derived) == 0 {
		t.Fatalf("a decision reached nothing derived from it:\n%s", out)
	}
	// Reaching the child must also reach what implements the child, or the
	// answer stops exactly where it becomes useful.
	if len(tr.Constructs) == 0 {
		t.Errorf("the derived requirement's implementation was not reached:\n%s", out)
	}
	// And the target itself must not appear twice.
	for _, id := range tr.Derived {
		if id == "R-DEC-NUMBERING" {
			t.Errorf("the target is listed as derived from itself:\n%s", out)
		}
	}
}

// Asking about something that carries no construct is a fair question with the
// answer "nothing". It is not an error, and reporting it as one would make the
// command unusable in a script that walks a diff.
func TestImpactOfSomethingUnrelated(t *testing.T) {
	t.Parallel()
	// perm.go carries permission declarations, which are recognised but need no
	// requirement of their own, so it has no annotation sidecar.
	out, code := runSpeclink(t, "impact", "../../testdata/example", "app/sales/perm.go")
	if code != 0 {
		t.Fatalf("an empty answer must not be an error, got %d:\n%s", code, out)
	}
	if !strings.Contains(out, "reaches nothing") {
		t.Errorf("expected an explicit empty answer:\n%s", out)
	}
}

func decodeImpact(t *testing.T, out string) impactReport {
	t.Helper()

	var got impactReport
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("impact is not valid JSON: %v\n%s", err, out)
	}
	if got.Version != ImpactVersion {
		t.Errorf("got version %d, want %d", got.Version, ImpactVersion)
	}
	return got
}
