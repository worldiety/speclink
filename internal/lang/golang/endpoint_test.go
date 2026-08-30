package golang

import (
	"path/filepath"
	"testing"

	"github.com/worldiety/speclink/internal/config"
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
