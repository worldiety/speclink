package spec

import (
	"encoding/json"
	"strings"
	"testing"
)

// recorder stands in for *testing.T, which cannot be used to observe its own
// output.
type recorder struct{ lines []string }

func (r *recorder) Log(args ...any) {
	parts := make([]string, 0, len(args))
	for _, a := range args {
		parts = append(parts, a.(string))
	}
	r.lines = append(r.lines, strings.Join(parts, ""))
}

// *testing.T must satisfy Logger without anybody having to think about it. The
// interface exists only so this package does not import testing; if it drifted
// away from the method it is modelled on, every call site would need an
// adapter.
var _ Logger = (*testing.T)(nil)

func TestVerifiedWritesOneParsableLine(t *testing.T) {
	r := &recorder{}
	Verified(r, Requirement{ID: "R-QUOTE-SUBMIT"}, Requirement{ID: "R-QUOTE-APPROVE"})

	if len(r.lines) != 1 {
		t.Fatalf("got %d lines, want 1: %v", len(r.lines), r.lines)
	}
	payload, ok := strings.CutPrefix(r.lines[0], VerifiedMarker)
	if !ok {
		t.Fatalf("line does not carry the marker: %q", r.lines[0])
	}

	var got verifiedRecord
	if err := json.Unmarshal([]byte(payload), &got); err != nil {
		t.Fatalf("payload is not JSON: %v (%q)", err, payload)
	}
	if got.Version != VerifiedVersion {
		t.Errorf("got version %d, want %d", got.Version, VerifiedVersion)
	}
	if len(got.Reqs) != 2 {
		t.Errorf("got %v, want both requirements", got.Reqs)
	}
}

// The record ends up in speclink.lock, whose diff is read by people. A line
// that reorders itself between runs would show up there as a change that is not
// one.
func TestVerifiedIsStable(t *testing.T) {
	a, b := &recorder{}, &recorder{}
	Verified(a, Requirement{ID: "R-B"}, Requirement{ID: "R-A"})
	Verified(b, Requirement{ID: "R-A"}, Requirement{ID: "R-B"})

	if a.lines[0] != b.lines[0] {
		t.Errorf("argument order changed the line:\n%s\n%s", a.lines[0], b.lines[0])
	}
}

func TestVerifiedDropsDuplicates(t *testing.T) {
	r := &recorder{}
	Verified(r, Requirement{ID: "R-A"}, Requirement{ID: "R-A"})

	if !strings.Contains(r.lines[0], `["R-A"]`) {
		t.Errorf("duplicate not collapsed: %s", r.lines[0])
	}
}

// Saying nothing must produce nothing. An empty record would be
// indistinguishable from a test that verified something and would satisfy a
// coverage obligation it never met.
func TestVerifiedSaysNothingWhenGivenNothing(t *testing.T) {
	r := &recorder{}
	Verified(r)
	Verified(r, Requirement{})
	Verified(nil, Requirement{ID: "R-A"})

	if len(r.lines) != 0 {
		t.Errorf("got %v, want no output", r.lines)
	}
}
