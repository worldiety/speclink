package main

import (
	"strings"
	"testing"
)

// The conformant TypeScript fixture verifies with nothing to report, in every
// direction this frontend measures. It is the guard against a rule that no
// project could satisfy.
func TestTypeScriptFixtureIsClean(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/ts")
	if code != 0 {
		t.Fatalf("the conformant fixture did not verify:\n%s", out)
	}
	for _, want := range []string{
		"1 source segments (100% accounted)",
		"1 constructs (100% bound)",
		"2 normative requirements (100% covered, 100% verified, 100% demonstrated)",
		"1 route (1 traced to a use case)",
		"0 findings",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("summary lacks %q:\n%s", want, summary(out))
		}
	}
	// It reads no stored shapes, and must say so rather than report them sound.
	if !strings.Contains(out, "not measured: schema evolution") {
		t.Errorf("an unmeasured direction was not disclosed:\n%s", out)
	}
	if strings.Contains(out, "not measured: architecture") || strings.Contains(out, "not measured: verification") {
		t.Errorf("a direction this frontend measures was reported as a gap:\n%s", out)
	}
}

// The violating fixture breaks every rule of the hex1 style once. A missing
// code is a rule that stopped firing.
func TestTypeScriptArchitectureRules(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/tsarch")
	if code == 0 {
		t.Fatalf("expected findings, got a clean run:\n%s", out)
	}
	want := map[string]string{
		"SPEC-V6-230": "K5-UC-FILE: two classes in one use case file",
		"SPEC-V6-231": "K5-UC-FILE: a use case not named after its file",
		"SPEC-V6-232": "K5-UC-FILE: a use case outside the use case directory",
		"SPEC-V6-233": "K5-UC-SHAPE: a second public method",
		"SPEC-V6-234": "K5-UC-DEPS: a use case importing a port",
		"SPEC-V6-235": "K5-UC-STATELESS: module level state",
		"SPEC-V6-236": "K5-UC-AUTHZ: no permission consulted",
		"SPEC-V6-240": "K6-DOMAIN-NO-ADAPTER-IMPORT",
		"SPEC-V6-241": "K6-DOMAIN-NO-FRAMEWORK: a package",
		"SPEC-V6-242": "K6-DOMAIN-NO-FRAMEWORK: a browser global",
		"SPEC-V6-243": "K6-TYPES-PURE: an import",
		"SPEC-V6-244": "K6-TYPES-PURE: logic",
		"SPEC-V6-245": "K6-PORTS-PURE: a class",
		"SPEC-V6-246": "K6-PORTS-PURE: the name",
		"SPEC-V6-247": "K6-SERVICE-NO-USECASE",
		"SPEC-V6-248": "K6-ADAPTER-IMPLEMENTS-PORT: no port",
		"SPEC-V6-249": "K6-ADAPTER-IMPLEMENTS-PORT: the name",
		"SPEC-V6-250": "K6-ADAPTER-WIRED-IN-MAIN",
		"SPEC-V6-251": "K6-ADAPTER-ISOLATED",
		"SPEC-V6-252": "K6-UI-NO-PORT",
		"SPEC-V6-253": "K6-UI-VIA-USECASE: a service import",
		"SPEC-V6-254": "K6-UI-VIA-USECASE: the locator outside a use case",
		"SPEC-V6-255": "K6-LOCATOR-INIT-IN-MAIN",
		"SPEC-V6-256": "K6-LAYER-UNKNOWN",
		"SPEC-V6-257": "K7-ROUTE-TARGETS-VIEW: a route to a component",
		"SPEC-V6-258": "K7-ROUTE-TARGETS-VIEW: a stray file among the views",
		"SPEC-V6-259": "K7-I18N-KEY-PARITY",
		"SPEC-V6-260": "K7-I18N-KEY-EXISTS",
		"SPEC-V6-262": "K8-MAIN-COMPOSITION",
	}
	for code, what := range want {
		if n := strings.Count(out, "["+code+"]"); n != 1 {
			t.Errorf("%s (%s): got %d findings, want 1", code, what, n)
		}
	}
	// The waived use case consults no permission either, and says why.
	if strings.Contains(out, "WaivedUseCase.execute consults no permission") {
		t.Errorf("a waiver did not apply:\n%s", out)
	}
}

// The building blocks of a TypeScript project are its directories, drawn by
// layer. An architecture without bounded contexts must not be described as
// having zero of them.
func TestTypeScriptPackagesAreDrawn(t *testing.T) {
	t.Parallel()
	out, code := runSpeclink(t, "diagrams", "../../testdata/ts", "-out", t.TempDir())
	if code != 0 || !strings.Contains(out, "wrote 1 diagram") {
		t.Fatalf("no package diagram was written:\n%s", out)
	}
	out, code = runSpeclink(t, "generate", "../../testdata/ts")
	if code != 0 {
		t.Fatalf("generate failed:\n%s", out)
	}
	if strings.Contains(out, "bounded context") {
		t.Errorf("a project without bounded contexts is described in terms of them:\n%s", out)
	}
}
