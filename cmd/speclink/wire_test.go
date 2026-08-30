package main

import (
	"strings"
	"testing"
)

// What crosses an address is the promise with the most holders and the fewest
// ways of finding out. A caller is outside the build: no compiler sees it, no
// test here runs it, and a field that changed its name is discovered by a
// support ticket.
//
// The name of the type is not that promise, which is why none of this is keyed
// on it. The shape is, and it is read the way a persisted shape is read —
// expanded through named types, compared field by field on the wire name a
// caller actually sees.

// TestRemovedResponseFieldIsReported is the plain break.
func TestRemovedResponseFieldIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "cmd/erp/api.go",
		"\tSequence uint64 `json:\"sequence\"`", "\tOk bool `json:\"ok\"`")
	rewrite(t, dir, "cmd/erp/api.go",
		"return SubmitQuoteResponse{Sequence: uint64(seq)}, nil",
		"_ = seq\n\t\t\t\treturn SubmitQuoteResponse{Ok: true}, nil")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a promised response field was dropped and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "promised the field sequence and no longer returns it") {
		t.Errorf("expected K20-RESPONSE-FIELD-REMOVED:\n%s", out)
	}
}

// TestDroppedRequestFieldIsReported is the break that answers with a success.
//
// The caller still sends the field, still gets its two hundred, and the value
// is discarded. Nothing anywhere reports it, which is what makes it worth a
// rule of its own rather than folding it in with the response.
func TestDroppedRequestFieldIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "cmd/erp/api.go",
		"type SubmitQuoteBody struct {\n\tQuoteID string `json:\"quoteId\"`\n\tTitle   string `json:\"title\"`\n}",
		"type SubmitQuoteBody struct {\n\tQuoteID string `json:\"quoteId\"`\n}")
	rewrite(t, dir, "cmd/erp/api.go", "dst.Title = body.Title", "_ = body")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a promised request field is no longer read and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "promised to accept the field title and no longer reads it") {
		t.Errorf("expected K20-REQUEST-FIELD-DROPPED:\n%s", out)
	}
}

// TestChangedFieldShapeIsReported is the one that breaks both directions at
// once.
func TestChangedFieldShapeIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "cmd/erp/api.go",
		"\tSequence uint64 `json:\"sequence\"`", "\tSequence string `json:\"sequence\"`")
	rewrite(t, dir, "cmd/erp/api.go",
		"return SubmitQuoteResponse{Sequence: uint64(seq)}, nil",
		"return SubmitQuoteResponse{Sequence: fmt.Sprint(seq)}, nil")
	rewrite(t, dir, "cmd/erp/api.go", "import (\n\t\"io\"", "import (\n\t\"fmt\"\n\t\"io\"")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a promised field changed type and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "was promised as int and is now string") {
		t.Errorf("expected K20-WIRE-SHAPE-CHANGED:\n%s", out)
	}
}

// TestNestedShapeChangeReachesTheRoute is the reason the shape is expanded
// through named types rather than compared as a name.
//
// The edit here is two packages away from anything that mentions HTTP: a field
// of a projection gains a JSON tag. Nobody making that change is thinking about
// an API, and nothing else in this repository would connect the two.
func TestNestedShapeChangeReachesTheRoute(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")
	rewrite(t, dir, "app/sales/quoteoverview.go",
		"\tSubmitted int\n", "\tSubmitted int `json:\"submittedCount\"`\n")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a nested wire change did not reach the address that returns it:\n%s", out)
	}
	if !strings.Contains(out, "the field quotes in the response of GET /api/v1/quotes") {
		t.Errorf("expected the finding on the route, not on the projection:\n%s", out)
	}
}

// TestHarmlessWireChangesStaySilent is half the value of the rules and all of
// their credibility.
//
// A field added to a response is what every client is required to tolerate, and
// a Go field renamed behind an unchanged tag is invisible outside this
// repository. A tool that reported either would teach people to run freeze
// without reading the diff, which is the only thing that makes freeze worth
// having.
func TestHarmlessWireChangesStaySilent(t *testing.T) {
	t.Parallel()

	t.Run("a field added to a response", func(t *testing.T) {
		t.Parallel()
		dir := copyFixture(t, "../../testdata/example")
		rewrite(t, dir, "cmd/erp/api.go",
			"\tSequence uint64 `json:\"sequence\"`",
			"\tSequence uint64 `json:\"sequence\"`\n\tNote     string `json:\"note\"`")

		if out, code := runVerify(t, dir); code != 0 {
			t.Errorf("adding a response field was reported as a break:\n%s", out)
		}
	})

	t.Run("a Go field renamed behind its tag", func(t *testing.T) {
		t.Parallel()
		dir := copyFixture(t, "../../testdata/example")
		rewrite(t, dir, "cmd/erp/api.go",
			"\tSequence uint64 `json:\"sequence\"`", "\tSeq uint64 `json:\"sequence\"`")
		rewrite(t, dir, "cmd/erp/api.go",
			"return SubmitQuoteResponse{Sequence: uint64(seq)}, nil",
			"return SubmitQuoteResponse{Seq: uint64(seq)}, nil")

		if out, code := runVerify(t, dir); code != 0 {
			t.Errorf("a rename that no caller can see was reported as a break:\n%s", out)
		}
	})

	t.Run("an integer widened", func(t *testing.T) {
		t.Parallel()
		dir := copyFixture(t, "../../testdata/example")
		rewrite(t, dir, "app/sales/quoteoverview.go", "\tSubmitted int\n", "\tSubmitted int64\n")

		if out, code := runVerify(t, dir); code != 0 {
			t.Errorf("JSON has one number type, so widening one is not a wire change:\n%s", out)
		}
	})
}

// TestWireShapesAreNotPromisedWhereNoneWereStated guards the pointer.
//
// The frameworkless fixture mounts on a router that reports no bodies. Recording
// an empty body rather than none would hold every one of its routes to a
// promise it never made, and the first field anybody added would read as a
// break.
func TestWireShapesAreNotPromisedWhereNoneWereStated(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/bare")
	if code != 0 {
		t.Fatalf("the frameworkless fixture did not verify:\n%s", out)
	}
	for _, rule := range []string{
		"K20-RESPONSE-FIELD-REMOVED",
		"K20-REQUEST-FIELD-DROPPED",
		"K20-WIRE-SHAPE-CHANGED",
	} {
		if strings.Contains(out, rule) {
			t.Errorf("%s fired on a dialect that states no bodies:\n%s", rule, out)
		}
	}
}
