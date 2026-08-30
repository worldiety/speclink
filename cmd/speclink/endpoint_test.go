package main

import (
	"strings"
	"testing"
)

// The surface a system exposes is the one promise whose holders are outside the
// build. Nothing in this repository can tell a caller that an address went
// away, so the record of what was promised is the only thing that can.

func TestRoutesAreCountedAndTraced(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/bare")
	if code != 0 {
		t.Fatalf("the bare fixture did not verify:\n%s", out)
	}
	// Both halves, always. "2 of 2" and a silent line are different claims,
	// and only the first says somebody looked.
	if !strings.Contains(out, "2 routes (2 traced to a use case)") {
		t.Errorf("the exposed surface was not measured:\n%s", summary(out))
	}
}

// TestRemovedRouteIsReported is the rule the baseline exists for here.
//
// A withdrawn address breaks every client already calling it, and the working
// tree afterwards looks exactly like one that never had the route. Only the
// record can tell the difference.
func TestRemovedRouteIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "app/billing/rest/routes.go",
		`mux.Handle("POST /invoices/draft", rest.Log(rest.Handle(who, draft)))`, "_ = draft")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a promised address was withdrawn and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "the address POST /invoices/draft was promised and is no longer mounted") {
		t.Errorf("expected K20-ENDPOINT-REMOVED:\n%s", out)
	}
}

// TestRouteThatChangesMeaningIsReported is the more interesting half.
//
// The address does not move, so nothing a caller can see says the behaviour
// behind it changed, and nothing the compiler checks says so either. This is
// the same failure as a rewritten requirement whose identifier still compiles:
// the far end moved and the link still resolves.
func TestRouteThatChangesMeaningIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "app/sales/rest/routes.go",
		"func Mount(mux *http.ServeMux, who rest.Authenticator, submit sales.SubmitQuote) {\n\tmux.Handle(\"POST /quotes/submit\", Submit(who, submit))",
		"func Mount(mux *http.ServeMux, who rest.Authenticator, submit sales.SubmitQuote, find sales.FindQuote) {\n\tmux.Handle(\"POST /quotes/submit\", rest.Handle(who, find))")
	rewrite(t, dir, "cmd/erp/main.go",
		"restsales.Mount(mux, anonymous, uc.SubmitQuote)",
		"restsales.Mount(mux, anonymous, uc.SubmitQuote, uc.FindQuote)")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("the work behind a promised address changed and nothing was reported:\n%s", out)
	}
	if !strings.Contains(out, "POST /quotes/submit was recorded as serving") {
		t.Errorf("expected K20-ENDPOINT-MEANING-CHANGED:\n%s", out)
	}
	if !strings.Contains(out, "now serves example.com/bare/app/sales.FindQuote") {
		t.Errorf("the finding must name what it serves now:\n%s", out)
	}
}

// TestRouteWithNothingBehindItIsReported guards the case that motivated the
// whole trace.
//
// An address the world can reach with no use case behind it is either work
// sitting in the presentation layer or something the architecture has no name
// for. Both are worth knowing; neither is visible from the registration.
func TestRouteWithNothingBehindItIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "app/billing/rest/routes.go",
		`mux.Handle("POST /invoices/draft", rest.Log(rest.Handle(who, draft)))`,
		"_ = draft\n\tmux.Handle(\"GET /invoices/health\", http.NotFoundHandler())")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a route with nothing accountable behind it was accepted:\n%s", out)
	}
	if !strings.Contains(out, "nothing accountable was found behind GET /invoices/health") {
		t.Errorf("expected K20-ENDPOINT-NO-USE-CASE:\n%s", out)
	}
}

// TestComputedRouteIsReportedRatherThanSkipped is the lesson the storage rules
// already taught once.
//
// A route mounted under a pattern speclink cannot evaluate is the dangerous
// case, not the harmless one: it is an address the system answers on that no
// catalogue will ever list. Dropping it for being inconvenient would be the
// same silence that once let a whole profile report unmeasured shapes as clean.
func TestComputedRouteIsReportedRatherThanSkipped(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "app/billing/rest/routes.go",
		`mux.Handle("POST /invoices/draft", rest.Log(rest.Handle(who, draft)))`,
		"mux.Handle(\"POST \"+prefix()+\"/invoices/draft\", rest.Log(rest.Handle(who, draft)))")
	rewrite(t, dir, "app/billing/rest/routes.go",
		"// Mount registers the routes of this context.",
		"func prefix() string { return \"/v1\" }\n\n// Mount registers the routes of this context.")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a computed pattern was silently skipped:\n%s", out)
	}
	if !strings.Contains(out, "not a compile time constant") {
		t.Errorf("expected K20-ENDPOINT-PATTERN-UNREADABLE:\n%s", out)
	}
}

// TestSurfaceCatalogueCarriesTheChain is what the recognition is for.
//
// An address on its own is routing. An address with the requirement behind it
// answers the question every review actually asks — why does this system
// answer here, and on whose authority — and it is a chain no one has to
// maintain: the route is read from the code that mounts it, the use case from
// the trace, and the requirement from the binding the use case already carries.
func TestSurfaceCatalogueCarriesTheChain(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/bare", "./...")

	if !strings.Contains(out, "## What answers from outside") {
		t.Fatalf("the exposed surface is missing from the document:\n%s", out)
	}
	for _, want := range []string{
		"| `POST /quotes/submit` | `SubmitQuote` | R-QUOTE-SUBMIT |",
		"| `POST /invoices/draft` | `DraftInvoice` | R-BILL-DRAFT |",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected the row %q in the catalogue", want)
		}
	}
}

// TestSurfaceCatalogueNamesWhatItCouldNotRead guards the row that matters most.
//
// A blank cell reads as nothing to say, and here it never is. An address that
// only exists at run time is one no catalogue can name, and a document that
// quietly left it out would be worse than one that never had the table.
func TestSurfaceCatalogueNamesWhatItCouldNotRead(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "app/billing/rest/routes.go",
		`mux.Handle("POST /invoices/draft", rest.Log(rest.Handle(who, draft)))`,
		"mux.Handle(\"POST \"+prefix()+\"/invoices/draft\", rest.Log(rest.Handle(who, draft)))")
	rewrite(t, dir, "app/billing/rest/routes.go",
		"// Mount registers the routes of this context.",
		"func prefix() string { return \"/v1\" }\n\n// Mount registers the routes of this context.")

	out, _ := runSpeclink(t, "generate", dir, "./...")
	if !strings.Contains(out, "_computed, not readable_") {
		t.Errorf("an unreadable address must be named in the catalogue, not omitted:\n%s", out)
	}
}
