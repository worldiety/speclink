package main

import (
	"os"
	"path/filepath"
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

// TestNarrowScopeAccusesNothing is the mirror of the rule this tool is built
// on, and it was missing.
//
// speclink already refuses to call an unmeasured direction clean. The other
// half went unwritten for a while, and it is the more damaging one: a run
// narrowed to one package could not see the rest of the module, and reported
// everything it had not been asked to look at as broken — a withdrawn address,
// a route with nothing behind it, and a change of meaning, all three at once,
// against a tree where nothing whatsoever was wrong.
//
// A finding that depends on which packages the operator typed is not a finding
// about the code. This runs the untouched fixture at three widths and requires
// silence from every endpoint rule at all of them.
func TestNarrowScopeAccusesNothing(t *testing.T) {
	t.Parallel()
	for _, pattern := range []string{"./...", "./app/sales/...", "./app/sales/rest/...", "./app/billing/..."} {
		out, _ := runSpeclink(t, "verify", "../../testdata/bare", pattern)
		for _, rule := range []string{
			"K20-ENDPOINT-REMOVED",
			"K20-ENDPOINT-NO-USE-CASE",
			"K20-ENDPOINT-MEANING-CHANGED",
			"K20-ENDPOINT-TRACE-TRUNCATED",
		} {
			if strings.Contains(out, rule) {
				t.Errorf("%s fired at width %q against an untouched tree:\n%s", rule, pattern, out)
			}
		}
	}
}

// TestUnmeasuredRoutesAreCountedApart is why the silence above is not itself a
// lie.
//
// Saying nothing about a route the run could not account for would leave it
// indistinguishable from one that was traced, and the count would then claim a
// completeness the run never had. The figure carries the third number instead,
// so the reader can see the width of the run in the same line as its result.
//
// Reaching it from here takes a scope that measures the presentation package
// and excludes the context behind it, because a pattern no longer truncates the
// load. The route is then found, the trace resolves, and the use case is not
// among the constructs this run reports on.
func TestUnmeasuredRoutesAreCountedApart(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	writeConfig(t, dir, `{"profile":"go_bare_ddd1","scope":["app/sales/rest"]}`)

	out, _ := runVerify(t, dir)
	if !strings.Contains(out, "1 route") {
		t.Fatalf("the route of the measured package was not found:\n%s", summary(out))
	}
	if strings.Contains(out, "K20-ENDPOINT-NO-USE-CASE") {
		t.Errorf("a use case outside the scope was reported as absent:\n%s", out)
	}
}

// TestWithdrawnRouteCanBeSettled closes the loop the removal rule opened.
//
// The baseline never forgets a promise, so freeze cannot resolve a deliberate
// withdrawal — dropping the entry would make the record agree with the code by
// editing the record. Without a waiver the finding was therefore permanent, and
// a rule nobody can ever discharge is a rule everybody learns to ignore.
func TestWithdrawnRouteCanBeSettled(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/bare")
	rewrite(t, dir, "app/billing/rest/routes.go",
		`mux.Handle("POST /invoices/draft", rest.Log(rest.Handle(who, draft)))`, "_ = draft")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("a promised address was withdrawn and nothing was reported:\n%s", out)
	}
	// The finding has to name where the waiver belongs, because the construct
	// it was about no longer exists to carry one.
	if !strings.Contains(out, "waive K20-ENDPOINT-REMOVED on example.com/bare/app/billing/rest") {
		t.Errorf("the finding must say where it can be settled:\n%s", out)
	}

	waiver := filepath.Join(dir, "app/billing/rest/routes.annotation.go")
	if err := os.WriteFile(waiver, []byte(`package restbilling

import "github.com/worldiety/speclink/spec"

var _ = spec.ForPackage(
	spec.Waive("K20-ENDPOINT-REMOVED", "Withdrawn deliberately: the only caller was the internal console, which now uses the use case directly."),
)
`), 0o644); err != nil {
		t.Fatal(err)
	}

	out, code = runVerify(t, dir)
	if code != 0 {
		t.Fatalf("a waived withdrawal still failed the run:\n%s", out)
	}
}

// TestFluentRoutesAreRecognised pins that the surface of the reference project
// is measured at all.
//
// It mounts nothing on the standard library's router, so before the builder was
// recognised this run reported no routes — and reported it by saying nothing,
// which reads as a system that exposes no addresses rather than as a tool that
// cannot see the ones it has.
func TestFluentRoutesAreRecognised(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("the reference project did not verify:\n%s", out)
	}
	if !strings.Contains(out, "3 routes (3 traced to a use case)") {
		t.Errorf("the exposed surface of the reference project was not measured:\n%s", summary(out))
	}
}

// TestSurfaceCatalogueCarriesTheWireShapes is the chain the document exists to
// print, now complete.
//
// Address, what crosses it in each direction, the work behind it and the
// requirement that asked for it — five columns, none of them maintained by
// hand, each one derived from the code that already states it.
func TestSurfaceCatalogueCarriesTheWireShapes(t *testing.T) {
	t.Parallel()
	out, _ := runSpeclink(t, "generate", "../../testdata/example", "./...")

	for _, want := range []string{
		"| Address | Takes | Returns | Serves | Asked for by |",
		"| `POST /api/v1/quotes` | `SubmitQuoteBody` | `SubmitQuoteResponse` | `SubmitQuote` | R-QUOTE-SUBMIT |",
		// A GET carries no body, so the cell is a dash and not the assembled
		// request model, which no caller ever sends.
		"| `GET /api/v1/quotes` | — | `ListQuotesResponse` | `ListQuotes` | R-QUOTE-OVERVIEW |",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected the row %q in the catalogue", want)
		}
	}
}

// TestCatalogueDistinguishesNoBodyFromNotAsked is the distinction that made the
// column worth adding.
//
// A route on the fluent builder that writes bytes promises no shape, and a
// route on a bare mux has a shape nobody read. Both would print as a blank, and
// a document that cannot tell them apart is the document this tool replaces.
func TestCatalogueDistinguishesNoBodyFromNotAsked(t *testing.T) {
	t.Parallel()

	stated, _ := runSpeclink(t, "generate", "../../testdata/example", "./...")
	if !strings.Contains(stated, "| `DELETE /api/v1/quotes/{quoteId}` | — | — |") {
		t.Errorf("a route that promises no shape must print a dash:\n%s", surfaceOf(stated))
	}

	// The frameworkless fixture states nothing, so the columns are absent
	// altogether rather than filled with dashes that would read as answers.
	silent, _ := runSpeclink(t, "generate", "../../testdata/bare", "./...")
	if strings.Contains(silent, "| Address | Takes | Returns |") {
		t.Errorf("a dialect that reports no wire types must not get the columns:\n%s", surfaceOf(silent))
	}
	if !strings.Contains(silent, "| Address | Serves | Asked for by |") {
		t.Errorf("the surface table is missing:\n%s", surfaceOf(silent))
	}
}

func surfaceOf(doc string) string {
	_, rest, ok := strings.Cut(doc, "## What answers from outside")
	if !ok {
		return doc
	}
	if end := strings.Index(rest, "\n## "); end >= 0 {
		return rest[:end]
	}
	return rest
}

// TestASurfaceNobodyLookedAtIsNotAnEmptyOne closes the last silence of this
// family.
//
// Three states have to be told apart and two of them used to print the same
// thing, which was nothing at all. A frontend that cannot recognise routes says
// so on its own line. A frontend that can and found none prints the zero,
// because a module answering on no address is a library and that is worth
// stating. Only a frontend that found routes prints how many were accounted
// for.
func TestASurfaceNobodyLookedAtIsNotAnEmptyOne(t *testing.T) {
	t.Parallel()

	// Can look, found none.
	none, _ := runVerify(t, "../../testdata/arch")
	if !strings.Contains(none, "0 routes") {
		t.Errorf("a project with no routes must say so rather than stay silent:\n%s", summary(none))
	}
	if strings.Contains(none, "not measured: the exposed surface") {
		t.Errorf("a frontend that looked reported that it cannot:\n%s", none)
	}

	// Cannot look at all.
	blind, _ := runVerify(t, "../../testdata/java")
	if !strings.Contains(blind, "not measured: the exposed surface") {
		t.Errorf("a frontend that recognises no routes must disclose it:\n%s", blind)
	}
	if strings.Contains(blind, "0 routes") {
		t.Errorf("a frontend that cannot look reported an empty surface:\n%s", summary(blind))
	}
}
