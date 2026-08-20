package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestInventoryListsEveryKind guards the output the acceptance of the
// inference layer depends on.
//
// verify reports only what it objects to, so a construct recognised correctly
// produces no output at all. Comparing the recognisers against an independent
// model is impossible from that; this command exists to make it possible.
func TestInventoryLists(t *testing.T) {
	out, code := runSpeclink(t, "inventory", "../../testdata/example", "./...")
	if code != 0 {
		t.Fatalf("expected a clean run, got exit %d:\n%s", code, out)
	}

	for _, kind := range []string{"aggregate", "command", "event", "permission", "projection", "query", "repository", "use case"} {
		if !strings.Contains(out, kind) {
			t.Errorf("kind %q missing from the inventory:\n%s", kind, out)
		}
	}
	// The summary counts bound constructs separately, which is the number a
	// migration is steered by and the one verify cannot show.
	if !strings.Contains(out, "bound") {
		t.Errorf("the summary must report how many are bound:\n%s", out)
	}
}

// TestInventoryJSON guards the machine readable form, which is what an
// acceptance script consumes.
func TestInventoryJSON(t *testing.T) {
	out, code := runSpeclink(t, "inventory", "../../testdata/example", "-format", "json", "./...")
	if code != 0 {
		t.Fatalf("expected a clean run, got exit %d:\n%s", code, out)
	}

	var got struct {
		Version    int `json:"version"`
		Constructs []struct {
			Kind     string `json:"kind"`
			Name     string `json:"name"`
			Evidence string `json:"evidence"`
			Bound    bool   `json:"bound"`
			Line     int    `json:"line"`
		} `json:"constructs"`
	}
	// stdout carries the document; the text summary goes to stderr, so the
	// combined output starts with the JSON.
	dec := json.NewDecoder(strings.NewReader(out))
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("decode inventory: %v\n%s", err, out)
	}
	if got.Version != 1 {
		t.Errorf("version = %d, want 1", got.Version)
	}
	if len(got.Constructs) == 0 {
		t.Fatal("no constructs reported")
	}

	var events, bound int
	for _, c := range got.Constructs {
		if c.Evidence == "" || c.Line == 0 {
			t.Errorf("%s carries no evidence or position: %+v", c.Name, c)
		}
		if c.Kind == "event" {
			events++
		}
		if c.Bound {
			bound++
		}
	}
	if events != 2 {
		t.Errorf("expected two events, got %d", events)
	}
	if bound == 0 {
		t.Error("the fixture binds constructs to requirements; none is reported as bound")
	}
}

// TestInventoryKindFilter guards the flag the acceptance uses to compare one
// kind at a time against the reference model.
func TestInventoryKindFilter(t *testing.T) {
	out, code := runSpeclink(t, "inventory", "../../testdata/example", "-kind", "event", "./...")
	if code != 0 {
		t.Fatalf("expected a clean run, got exit %d:\n%s", code, out)
	}
	if strings.Contains(out, "use case") || strings.Contains(out, "permission") {
		t.Errorf("the filter must restrict the listing:\n%s", out)
	}
}
