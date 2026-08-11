package main

import (
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
	// use case, query-free command, event, aggregate, permission, second use case
	if !strings.Contains(out, "6 constructs") {
		t.Errorf("expected 6 recognised constructs, got:\n%s", out)
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
		"SPEC-V6-001", // normative requirement covered by nothing
	}
	// The generic CRUD ban fires once per forbidden factory.
	if n := strings.Count(out, "[SPEC-V6-010]"); n != 3 {
		t.Errorf("expected the generic CRUD ban to fire 3 times, got %d", n)
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

// runVerify builds and runs the command against a fixture, returning combined
// output and exit code.
func runVerify(t *testing.T, root string, extra ...string) (string, int) {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "speclink")
	build := exec.Command("go", "build", "-o", bin, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build speclink: %v\n%s", err, out)
	}

	args := append([]string{"verify", "-root", root}, extra...)
	args = append(args, "./...")
	cmd := exec.Command(bin, args...)
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

func asExitError(err error, target **exec.ExitError) bool {
	e, ok := err.(*exec.ExitError)
	if ok {
		*target = e
	}
	return ok
}
