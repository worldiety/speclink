package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVerifyExample is the happy path: a small but complete target project must
// verify clean. It is the regression guard for the whole pipeline — loading,
// whitelist, reading, resolving and coverage.
func TestVerifyExample(t *testing.T) {
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("expected a clean run, got exit %d:\n%s", code, out)
	}
	for _, want := range []string{"100% bound", "100% covered"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in the summary, got:\n%s", want, out)
		}
	}
}

// TestInference pins what the nago recognisers must find. Without them
// speclink cannot tell which constructs carry business meaning, and forward
// coverage would measure nothing.
func TestInference(t *testing.T) {
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("expected a clean run, got exit %d:\n%s", code, out)
	}
	// two use cases, one query, one command, two events, one aggregate,
	// one projection, one repository, three permissions
	if !strings.Contains(out, "12 constructs") {
		t.Errorf("expected 12 recognised constructs, got:\n%s", out)
	}
}

// TestBindConstant guards what the previous ForVar could not do at all: taking
// the address of a constant is a compile error, so a constant such as a ReBAC
// namespace was unbindable while the documentation claimed otherwise.
func TestBindConstant(t *testing.T) {
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("binding a constant must verify clean, got exit %d:\n%s", code, out)
	}
}

// TestVerifyBad pins the diagnostics. Message texts are treated as public API:
// they are consumed by an LLM, and their quality decides how fast the loop
// converges. A silent change of wording is a change of interface.
func TestVerifyBad(t *testing.T) {
	out, code := runVerify(t, "../../testdata/bad")
	if code == 0 {
		t.Fatalf("expected findings, got a clean run:\n%s", out)
	}

	// Every rule the fixture violates must be reported exactly once.
	want := []string{
		"SPEC-V1-001", // function definition in an annotation file
		"SPEC-V1-006", // binding not declared as var _
		"SPEC-V3-001", // orphaned annotation file
		"SPEC-V3-003", // waiver without a reason
		"SPEC-V1-011", // address-of operator
		"SPEC-V3-005", // unknown struct field in ForField
		"SPEC-V3-006", // assertion at an illegal target kind
		"SPEC-V3-007", // non repeatable assertion given twice
		"SPEC-V3-008", // ForDecl argument names no declaration
		"SPEC-V5-004", // decision without a rationale
		"SPEC-V5-020", // normative requirement without a source
		"SPEC-V5-023", // source document does not exist
		"SPEC-V5-030", // file name does not match the requirement ID
		"SPEC-V5-032", // ID prefix contradicts Kind
		"SPEC-V5-033", // directory contradicts Kind
		"SPEC-V5-035", // domain directory contradicts the ID prefix
		"SPEC-V4-001", // proposal on a type whose package already is one
		"SPEC-V6-001", // normative requirement covered by nothing
		"SPEC-V6-011", // import of the generic CRUD user interface
		"SPEC-V6-090", // frozen shape that was never recorded
		"SPEC-V6-091", // discriminator changed
		"SPEC-V6-092", // promised field removed
		"SPEC-V6-093", // promised field renamed on the wire
		"SPEC-V6-094", // promised type gone from the source
		"SPEC-V6-095", // promise taken back by marking it a proposal
	}
	// The redundant field term fires once per covering level: the package in
	// one fixture, the type in the other.
	if n := strings.Count(out, "[SPEC-V4-002]"); n != 2 {
		t.Errorf("expected the redundant field proposal twice, got %d", n)
	}
	// The generic CRUD ban fires once per forbidden factory.
	if n := strings.Count(out, "[SPEC-V6-010]"); n != 3 {
		t.Errorf("expected the generic CRUD ban to fire 3 times, got %d", n)
	}
	// One projection and six events, none of them naming a requirement. A
	// proposal is free to change its shape, not free to exist for no reason.
	if n := strings.Count(out, "[SPEC-V6-020]"); n != 7 {
		t.Errorf("expected seven unbound constructs, got %d", n)
	}
	for _, code := range want {
		if n := strings.Count(out, "["+code+"]"); n != 1 {
			t.Errorf("expected %s exactly once, found %d times", code, n)
		}
	}

	// Diagnostics are prescriptive: every finding says what, why and how.
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "[SPEC-") {
			continue
		}
		if !strings.HasSuffix(strings.TrimSpace(line), ".") {
			t.Errorf("finding should read as a sentence: %q", line)
		}
	}
}

// TestVerifyJSON guards the machine readable contract consumed by the loop.
func TestVerifyJSON(t *testing.T) {
	out, _ := runVerify(t, "../../testdata/bad", "-format", "json")
	for _, want := range []string{`"version": 1`, `"code":`, `"what":`, `"why":`, `"how":`, `"phase":`} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON report is missing %s:\n%s", want, out)
		}
	}
}

// speclinkBin is the command under test, built once for the whole package.
// Every test drives the real binary rather than calling run() in process,
// because the exit code is part of the contract a loop runner depends on.
var speclinkBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "speclink-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, "create temp dir:", err)
		os.Exit(1)
	}
	speclinkBin = filepath.Join(dir, "speclink")
	if out, err := exec.Command("go", "build", "-o", speclinkBin, ".").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build speclink: %v\n%s", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// runSpeclink runs a command against a fixture, returning combined output and
// exit code.
func runSpeclink(t *testing.T, command, root string, extra ...string) (string, int) {
	t.Helper()

	args := append([]string{command, "-root", root}, extra...)
	cmd := exec.Command(speclinkBin, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()

	code := 0
	var exit *exec.ExitError
	if err != nil {
		if ok := asExitError(err, &exit); ok {
			code = exit.ExitCode()
		} else {
			t.Fatalf("run speclink: %v\n%s", err, out)
		}
	}
	return string(out), code
}

// runVerify is runSpeclink for the common case: verify over the whole fixture.
func runVerify(t *testing.T, root string, extra ...string) (string, int) {
	t.Helper()
	return runSpeclink(t, "verify", root, append(extra, "./...")...)
}

func asExitError(err error, target **exec.ExitError) bool {
	e, ok := err.(*exec.ExitError)
	if ok {
		*target = e
	}
	return ok
}

// TestArchitectureRules pins the five layout linters. The fixture violates each
// rule exactly once, so a missing or duplicated finding shows up immediately.
func TestArchitectureRules(t *testing.T) {
	out, code := runVerify(t, "../../testdata/arch")
	if code == 0 {
		t.Fatalf("expected findings, got a clean run:\n%s", out)
	}

	want := []string{
		"SPEC-V6-030", // main package outside cmd/
		"SPEC-V6-032", // infrastructure imports a bounded context
		"SPEC-V6-033", // infrastructure declares a use case
		"SPEC-V6-040", // ui package misnamed
		"SPEC-V6-041", // domain imports the presentation layer
		"SPEC-V6-042", // context without a UseCases bundle
		"SPEC-V6-051", // use case without a trailing error
		"SPEC-V6-056", // use case without a permission
		"SPEC-V6-057", // permission declared but never checked
		"SPEC-V6-058", // use case reads package level state
		"SPEC-V6-059", // permission with hardcoded texts
	}
	for _, code := range want {
		if n := strings.Count(out, "["+code+"]"); n != 1 {
			t.Errorf("expected %s exactly once, found %d times", code, n)
		}
	}

	// Both use cases of the fixture violate file layout and authorisation.
	for _, code := range []string{"SPEC-V6-050", "SPEC-V6-053", "SPEC-V6-055"} {
		if n := strings.Count(out, "["+code+"]"); n != 2 {
			t.Errorf("expected %s twice, found %d times", code, n)
		}
	}

	// The wiring layer connects views to use cases and must stay exempt,
	// otherwise every context becomes unfixable.
	if strings.Contains(out, "app/billing/cfg") {
		t.Errorf("the cfg wiring layer must not be reported:\n%s", out)
	}
}

// TestConformantProject guards that a project following every convention
// verifies clean. Without it the linters could drift into rules nothing can
// satisfy.
func TestConformantProject(t *testing.T) {
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("the conformant fixture must verify clean, got exit %d:\n%s", code, out)
	}
}
