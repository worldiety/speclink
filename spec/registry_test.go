package spec_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/worldiety/speclink/spec"
)

// The declarations below mimic what a target project writes in its annotation
// files. They run at package initialisation, which is exactly how the runtime
// registry gets filled.

type submitQuoteUC func() error

type quoteSubmitted struct{}

type submitQuoteCmd struct {
	Title string
}

var permSubmitQuote = "sales.quote.submit"

var rQuoteSubmit = spec.Requirement{
	ID:     "R-QUOTE-SUBMIT",
	Kind:   spec.Functional,
	Status: spec.Normative,
	Text:   "A quote number MUST be drawn on submission.",
}

var _ = spec.For[submitQuoteUC](
	spec.Satisfies(rQuoteSubmit),
	spec.Transition[quoteSubmitted]("submitted"),
	spec.Help("Submit the approved quote."),
)

var _ = spec.ForVar(&permSubmitQuote,
	spec.Rationale("Drawing from a gapless registry is not repeatable without consequence."),
)

var _ = spec.ForField[submitQuoteCmd]("Title",
	spec.Satisfies(rQuoteSubmit),
)

// TestRegistryCapturesPositions guards the property the cross check rests on:
// every binding records the source position of the term.
//
// Without it the comparison between the static and the runtime view would be
// impossible, because ForVar receives a pointer and the variable name is gone
// at run time.
func TestRegistryCapturesPositions(t *testing.T) {
	entries := spec.Entries()
	if len(entries) < 3 {
		t.Fatalf("expected at least 3 recorded bindings, got %d", len(entries))
	}

	for _, e := range entries {
		if !strings.HasSuffix(e.File, "registry_test.go") {
			t.Errorf("binding recorded from an unexpected file: %s", e.File)
		}
		if e.Line <= 0 {
			t.Errorf("binding at %s has no line number", e.File)
		}
	}
}

// TestRegistryTargets checks that each binding form identifies its target as
// far as the runtime allows.
func TestRegistryTargets(t *testing.T) {
	kinds := map[string]string{}
	for _, e := range spec.Entries() {
		kinds[e.TargetKind] = e.Target
	}

	if got := kinds["type"]; !strings.HasSuffix(got, ".submitQuoteUC") {
		t.Errorf("type binding target = %q, want …submitQuoteUC", got)
	}
	if got := kinds["field"]; !strings.HasSuffix(got, ".submitQuoteCmd.Title") {
		t.Errorf("field binding target = %q, want …submitQuoteCmd.Title", got)
	}
	// ForVar can only report the pointer type: the variable name is not
	// recoverable at run time. This is why the comparison is keyed by position.
	if got := kinds["var"]; got != "*string" {
		t.Errorf("var binding target = %q, want *string", got)
	}
}

// TestDumpJSON guards the one place in speclink with genuine version skew: the
// target project pins speclink/spec in its go.mod while the developer runs an
// arbitrary speclink binary.
func TestDumpJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := spec.DumpJSON(&buf); err != nil {
		t.Fatalf("DumpJSON: %v", err)
	}

	var dump spec.Dump
	if err := json.Unmarshal(buf.Bytes(), &dump); err != nil {
		t.Fatalf("dump is not valid JSON: %v", err)
	}
	if dump.Version != spec.DumpVersion {
		t.Errorf("dump version = %d, want %d", dump.Version, spec.DumpVersion)
	}
	if len(dump.Entries) == 0 {
		t.Fatal("dump contains no entries")
	}

	var found bool
	for _, e := range dump.Entries {
		for _, a := range e.Assertions {
			if a.Kind == "satisfies" && len(a.Requirements) == 1 && a.Requirements[0] == "R-QUOTE-SUBMIT" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("dump does not carry the satisfied requirement:\n%s", buf.String())
	}
}

// TestEntriesSorted checks that the dump is stable for diffing.
//
// Go initialises package level variables in dependency order, which is
// deterministic but unrelated to source order. The registry therefore must not
// be compared as a sequence; sorting makes the output stable regardless.
func TestEntriesSorted(t *testing.T) {
	entries := spec.Entries()
	for i := 1; i < len(entries); i++ {
		prev, cur := entries[i-1], entries[i]
		if prev.File > cur.File || (prev.File == cur.File && prev.Line > cur.Line) {
			t.Fatalf("entries are not sorted: %s:%d before %s:%d",
				prev.File, prev.Line, cur.File, cur.Line)
		}
	}
}
