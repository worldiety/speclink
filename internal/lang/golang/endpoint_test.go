package golang

import (
	"path/filepath"
	"testing"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/ir"
)

// bareModel loads the frameworkless fixture the way a run would.
func bareModel(t *testing.T) *Model {
	t.Helper()
	root, err := filepath.Abs("../../../testdata/bare")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Load(root, "./...")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	module, err := ModulePath(pkgs)
	if err != nil {
		t.Fatal(err)
	}
	return &Model{
		All: pkgs, Measured: pkgs, Root: root,
		Layout:    config.Config{},
		Style:     Bare,
		Framework: BareFoundation(module, ""),
		Layered:   true,
	}
}

// TestEndpointsTraceToUseCases is the test the whole tracer exists for.
//
// The two routes in the fixture are deliberately unlike. One mounts
// `Submit(who, submit)`: nothing about the registration says what it does,
// because the use case is an argument to a constructor and the code that
// answers is a closure two frames further in. The other mounts
// `Log(Handle(who, draft))`: the same use case, behind a wrapper.
//
// That both arrive at one use case each, by one code path, is the claim. A
// wrapper is only another call with the handler among its arguments, so nothing
// here knows what middleware is — and a wrapper nobody anticipated therefore
// costs nothing.
func TestEndpointsTraceToUseCases(t *testing.T) {
	want := map[string]string{
		"POST /quotes/submit":  "example.com/bare/app/sales.SubmitQuote",
		"POST /invoices/draft": "example.com/bare/app/billing.DraftInvoice",
	}

	got := map[string][]string{}
	for _, e := range bareModel(t).Endpoints() {
		if e.Truncated {
			t.Errorf("%s: trace hit its depth limit", e.Ref())
		}
		got[e.Ref()] = e.UseCases
	}

	if len(got) != len(want) {
		t.Errorf("recognised %d routes, want %d: %v", len(got), len(want), got)
	}
	for ref, uc := range want {
		switch g := got[ref]; {
		case g == nil:
			t.Errorf("route %q was not recognised at all", ref)
		case len(g) != 1 || g[0] != uc:
			t.Errorf("route %q serves %v, want exactly [%s]", ref, g, uc)
		}
	}
}

// TestSplitPattern pins the router's grammar.
//
// A pattern with no method answers to every method, and that is kept as "any"
// rather than filled in with a plausible one: "any" and "GET" are different
// promises to a client, and the difference is the whole reason to write the
// catalogue down.
func TestSplitPattern(t *testing.T) {
	for pattern, want := range map[string][2]string{
		"POST /quotes/submit": {"POST", "/quotes/submit"},
		"/health":             {"", "/health"},
		"GET /items/{id}":     {"GET", "/items/{id}"},
		"GET example.com/x":   {"GET", "example.com/x"},
		"example.com/x":       {"", "example.com/x"},
		"  DELETE /a/b  ":     {"DELETE", "/a/b"},
	} {
		m, p := splitPattern(pattern)
		if m != want[0] || p != want[1] {
			t.Errorf("%q split to (%q, %q), want (%q, %q)", pattern, m, p, want[0], want[1])
		}
	}
}

// TestTraceLeavingTheModuleIsNotTheSameAsLeavingTheProject pins the distinction
// the whole scope flag rests on.
//
// Every trace leaves the loaded set almost immediately — the first
// `http.HandlerFunc` is already outside it. If that counted as leaving the
// scope, every route in every project would be unmeasured and the flag would
// say nothing at all. Only our own module can hide a use case, so only our own
// module sets it.
func TestTraceLeavingTheModuleIsNotTheSameAsLeavingTheProject(t *testing.T) {
	t.Parallel()
	m := bareModel(t)
	for _, e := range m.Endpoints() {
		if e.LeftScope {
			t.Errorf("%s reports as unmeasured in a run that loaded the whole module: %s",
				e.Ref(), e.Handler)
		}
		if len(e.UseCases) == 0 {
			t.Errorf("%s traced to nothing", e.Ref())
		}
	}
}

// TestTraceIntoAnUnloadedPackageIsRecorded is the case the flag exists for.
//
// Loading only the presentation package puts the use case out of reach without
// putting it out of existence. The trace has to say so, because the alternative
// — an empty set of use cases, indistinguishable from a route with nothing
// behind it — is what turned a narrowed run into three false accusations.
func TestTraceIntoAnUnloadedPackageIsRecorded(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs("../../../testdata/bare")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Load(root, "./app/sales/rest/...")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	module, err := ModulePath(pkgs)
	if err != nil {
		t.Fatal(err)
	}
	m := &Model{
		All: pkgs, Measured: pkgs, Root: root,
		Layout:    config.Config{},
		Style:     Bare,
		Framework: BareFoundation(module, ""),
		Layered:   true,
	}

	eps := m.Endpoints()
	if len(eps) != 1 {
		t.Fatalf("expected the one route of the loaded package, got %d", len(eps))
	}
	if !eps[0].LeftScope {
		t.Error("the trace walked into an unloaded package of this module and did not record it")
	}
	if eps[0].Package != "example.com/bare/app/sales/rest" {
		t.Errorf("the mounting package was not recorded: %q", eps[0].Package)
	}
}

// nagoModel loads the reference project, which mounts its routes through the
// fluent builder rather than on the standard library's router.
func nagoModel(t *testing.T) *Model {
	t.Helper()
	root, err := filepath.Abs("../../../testdata/example")
	if err != nil {
		t.Fatal(err)
	}
	pkgs, err := Load(root, "./...")
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	return &Model{
		All: pkgs, Measured: pkgs, Root: root,
		Layout:    config.Config{},
		Style:     DDD1,
		Framework: Nago,
	}
}

func endpointAt(t *testing.T, m *Model, ref string) ir.Endpoint {
	t.Helper()
	for _, e := range m.Endpoints() {
		if e.Ref() == ref {
			return e
		}
	}
	t.Fatalf("no route %s among %v", ref, refsOf(m.Endpoints()))
	return ir.Endpoint{}
}

func refsOf(eps []ir.Endpoint) []string {
	out := make([]string, 0, len(eps))
	for _, e := range eps {
		out = append(out, e.Ref())
	}
	return out
}

// TestHapiRoutesCarryTheirWireShapes is what this dialect buys over a bare
// router.
//
// The builder states what crosses the boundary in type arguments the compiler
// has already resolved, so the request and the response are read rather than
// guessed. On a mux they stay empty, because the only way to obtain them there
// is to assume the use case's parameters are the wire shape — and this fixture
// maps a request body onto a command precisely so that the assumption would be
// wrong if anybody made it.
func TestHapiRoutesCarryTheirWireShapes(t *testing.T) {
	t.Parallel()
	m := nagoModel(t)

	e := endpointAt(t, m, "POST /api/v1/quotes")
	if e.Request != "example.com/erp/cmd/erp.SubmitQuoteRequest" {
		t.Errorf("request shape: %q", e.Request)
	}
	if e.Response != "example.com/erp/cmd/erp.SubmitQuoteResponse" {
		t.Errorf("response shape: %q", e.Response)
	}
	if !e.ShapesStated {
		t.Error("a dialect that reports its wire types must say that it does")
	}
	if len(e.UseCases) != 1 || e.UseCases[0] != "example.com/erp/app/sales.SubmitQuote" {
		t.Errorf("the work behind the route was not traced: %v", e.UseCases)
	}
}

// TestHapiVerbComesFromTheCallNotADefault guards the general form.
//
// hapi.Endpoint states no method of its own and the framework falls back to
// GET. A recogniser that took the fallback would print a promise the code never
// made, so the method is read from the operation or not at all.
func TestHapiVerbComesFromTheCallNotADefault(t *testing.T) {
	t.Parallel()
	m := nagoModel(t)

	e := endpointAt(t, m, "DELETE /api/v1/quotes/{quoteId}")
	if len(e.UseCases) != 1 || e.UseCases[0] != "example.com/erp/app/sales.WithdrawQuote" {
		t.Errorf("the work behind the general form was not traced: %v", e.UseCases)
	}
	// The response is written as bytes, so the route promises no shape. That
	// is a dash in the catalogue and not a gap, which is why the dialect has
	// to say it reports shapes even when this one has none.
	if e.Response != "" {
		t.Errorf("a binary response must state no type, got %q", e.Response)
	}
	if !e.ShapesStated {
		t.Error("an empty response on this dialect means none, not unknown")
	}
}

// TestHapiChainIsOneRoute is the reason the chain is walked rather than the
// calls counted.
//
// A registration here is three calls deep. Reporting one route per link would
// produce two phantom addresses per real one and then report them as duplicates
// of each other.
func TestHapiChainIsOneRoute(t *testing.T) {
	t.Parallel()
	m := nagoModel(t)

	seen := map[string]int{}
	for _, e := range m.Endpoints() {
		seen[e.Ref()]++
	}
	for ref, n := range seen {
		if n != 1 {
			t.Errorf("%s was recognised %d times", ref, n)
		}
	}
	if len(seen) != 3 {
		t.Errorf("expected the three routes of the fixture, got %v", refsOf(m.Endpoints()))
	}
}

// TestBareRoutesStateNoWireShapes is the other half of the same honesty.
//
// The standard library's router says nothing about bodies, so speclink says
// nothing either — and records that it was never asked, so an empty cell in the
// catalogue cannot be read as "this route carries nothing".
func TestBareRoutesStateNoWireShapes(t *testing.T) {
	t.Parallel()
	for _, e := range bareModel(t).Endpoints() {
		if e.ShapesStated {
			t.Errorf("%s claims its dialect reports wire types", e.Ref())
		}
		if e.Request != "" || e.Response != "" {
			t.Errorf("%s invented a wire shape: %q -> %q", e.Ref(), e.Request, e.Response)
		}
	}
}
