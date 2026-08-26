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
	// four use cases, two queries, one command, two events, three aggregates,
	// one projection, two repositories, five permissions
	if !strings.Contains(out, "19 constructs") {
		t.Errorf("expected 19 recognised constructs, got:\n%s", out)
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
		"SPEC-V4-001", // draft on a type whose package already is one
		"SPEC-V6-001", // normative requirement covered by nothing
		"SPEC-V6-011", // import of the generic CRUD user interface
		"SPEC-V6-091", // discriminator changed
		"SPEC-V6-092", // promised field removed
		"SPEC-V6-093", // promised field renamed on the wire
		"SPEC-V6-094", // promised type gone from the source
		"SPEC-V6-095", // promise taken back by marking it a draft
		"SPEC-V6-096", // promised field changed its stored shape
		"SPEC-V6-097", // optionality withdrawn
		"SPEC-V6-098", // field added without being declared optional
		"SPEC-V6-099", // two persisted types claiming one serialisation tag
	}
	// The redundant field term fires once per covering level: the package in
	// one fixture, the type in the other.
	if n := strings.Count(out, "[SPEC-V4-002]"); n != 2 {
		t.Errorf("expected the redundant field draft twice, got %d", n)
	}
	// The generic CRUD ban fires once per forbidden factory.
	if n := strings.Count(out, "[SPEC-V6-010]"); n != 3 {
		t.Errorf("expected the generic CRUD ban to fire 3 times, got %d", n)
	}
	// Two shapes are persisted without ever having been recorded: an event and
	// a domain model tied to the wire by the sloppy repository.
	if n := strings.Count(out, "[SPEC-V6-090]"); n != 2 {
		t.Errorf("expected two unrecorded shapes, got %d", n)
	}
	// One projection and eight events, none of them naming a requirement. A
	// draft is free to change its shape, not free to exist for no reason.
	if n := strings.Count(out, "[SPEC-V6-020]"); n != 9 {
		t.Errorf("expected nine unbound constructs, got %d", n)
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
		"SPEC-V6-021", // aggregate without a persistence decision
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

	// Both use cases of the fixture violate file layout.
	for _, code := range []string{"SPEC-V6-050", "SPEC-V6-053"} {
		if n := strings.Count(out, "["+code+"]"); n != 2 {
			t.Errorf("expected %s twice, found %d times", code, n)
		}
	}

	// Both use cases violate authorisation as well, but one of them waives the
	// rule. That the count is one rather than two is the whole assertion: the
	// rule still fires, and the waiver suppressed exactly the other instance.
	// spec.Waive did not reach the architecture rules at all until it was
	// wired, and nothing would have said so.
	if n := strings.Count(out, "[SPEC-V6-055]"); n != 1 {
		t.Errorf("expected SPEC-V6-055 once, the other being waived, found %d times", n)
	}
	if strings.Contains(out, "app/billing/usecases.go") && strings.Contains(out, "IssueInvoice contains nothing") {
		t.Error("the waived use case was still reported for K5-UC-AUTHZ")
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

// TestPermissionTextsThroughAHelper guards a rule that would otherwise reward
// the worse structure.
//
// Permission texts have to come from the translation catalogue, and writing the
// catalogue call out at every declaration is a dozen lines each — hundreds
// across a system. Any project will factor that into a helper. If the check
// only accepted the inlined form it would report the factored one, and the way
// to satisfy it would be to undo the refactoring.
//
// The conformant fixture therefore declares one permission through a helper of
// its own, in another package. That the helper sits elsewhere is the point: a
// lookup that resolved identifiers through the calling package's type
// information would find nothing there and answer no, silently.
func TestPermissionTextsThroughAHelper(t *testing.T) {
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("a helper that goes through the catalogue must satisfy the rule, got exit %d:\n%s", code, out)
	}
	if strings.Contains(out, "SPEC-V6-059") {
		t.Errorf("the factored form was reported:\n%s", out)
	}
}

// TestUseCaseBuiltFromCombinators guards against reading the wrong function.
//
// A constructor need not return a closure. It may build the use case out of
// combinators — guard(listAll(rows, less), PermX) — where the permission is
// applied by wrapping and the only function literal in sight is the comparator
// handed to the sort. That comparator is an argument of a helper, not the use
// case, and it names neither a subject nor a permission. Taking the first
// literal found anywhere lands on it, and every check downstream then examines
// a body that could not possibly hold what it is looking for.
//
// The fixture builds one use case exactly that way. Both checks that read the
// body have to see through it.
func TestUseCaseBuiltFromCombinators(t *testing.T) {
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("a use case assembled from combinators must pass, got exit %d:\n%s", code, out)
	}
	for _, code := range []string{"SPEC-V6-055", "SPEC-V6-057"} {
		if strings.Contains(out, code) {
			t.Errorf("%s misfired on the combinator form:\n%s", code, out)
		}
	}
}

// TestPersistenceDecisionAcceptsEveryReasonableForm guards the rule against the
// mistake its own family keeps making: reporting the projects that did the
// right thing.
//
// The choice between keeping facts and keeping state is written down in three
// places depending on the project, and all three are ordinary. On the type,
// because that is where the shape is. On the constructor, because a repository
// is reached through New… and that is where the mapping between domain and
// stored form actually lives. On the package, when a whole context made one
// choice.
//
// The conformant fixture uses two of them, and mixes the strategies inside one
// context on purpose: the quote is event sourced because its approval is
// evidence, the customer next to it is kept as current state because a
// corrected typo in a name is not a business event. That mixture is why the
// rule asks per aggregate and not per package.
func TestPersistenceDecisionAcceptsEveryReasonableForm(t *testing.T) {
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("a project that records its persistence choice must pass, got exit %d:\n%s", code, out)
	}
	if strings.Contains(out, "SPEC-V6-021") {
		t.Errorf("a recorded decision was not accepted:\n%s", out)
	}
}

// TestFieldCoverageReachesEveryStoredField pins that forward coverage goes down
// to the field, and that the conformant fixture can actually satisfy it.
//
// Binding a type is not the same as accounting for what is in it. Types are
// created deliberately and reviewed; fields accrete afterwards, and a field is
// the one thing that cannot be renegotiated once messages carry it.
//
// The fixture binds every stored field, including the ones that answer no
// business question on their own. QuoteID is bound to the traceability
// requirement rather than exempted, because an identifier on a stored fact is
// not self evident — it is what makes the sequence of events reconstructable,
// which is a requirement like any other.
func TestFieldCoverageReachesEveryStoredField(t *testing.T) {
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("a project that binds its stored fields must pass, got exit %d:\n%s", code, out)
	}
	if strings.Contains(out, "SPEC-V6-022") {
		t.Errorf("a bound field was still reported:\n%s", out)
	}
}
