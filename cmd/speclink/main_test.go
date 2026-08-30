package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// TestVerifyExample is the happy path: a small but complete target project must
// verify clean. It is the regression guard for the whole pipeline — loading,
// whitelist, reading, resolving and coverage.
func TestVerifyExample(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("binding a constant must verify clean, got exit %d:\n%s", code, out)
	}
}

// TestVerifyBad pins the diagnostics. Message texts are treated as public API:
// they are consumed by an LLM, and their quality decides how fast the loop
// converges. A silent change of wording is a change of interface.
func TestVerifyBad(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
//
// Runs against an unmodified source fixture are memoised. Two thirds of this
// package's runtime is go/packages, once per invocation, and half the
// invocations ask one of five fixtures the same question: 62 runs load
// testdata/example, 42 load testdata/bare, and every one of them re-parses the
// same tree to produce the same bytes.
//
// Three things make it sound rather than merely fast. The output is
// deterministic, which was measured rather than assumed. No test writes into a
// source fixture, which is why `git status` is clean after a full run — the two
// freeze runs pointed at one are both no-ops by assertion. And anything a test
// does write goes to a TempDir whose path is part of the key, so a command with
// an -out flag can never collide with another test's.
//
// A copied fixture is never cached: it is unique by construction, and its path
// is not under testdata.
func runSpeclink(t *testing.T, command, root string, extra ...string) (string, int) {
	t.Helper()

	if !shareableFixture(root) {
		out, code, err := execSpeclink(command, root, extra)
		if err != nil {
			t.Fatalf("run speclink: %v\n%s", err, out)
		}
		return out, code
	}

	key := strings.Join(append([]string{command, root}, extra...), "\x00")

	runCacheMu.Lock()
	entry, known := runCache[key]
	if !known {
		entry = &runResult{}
		runCache[key] = entry
	}
	runCacheMu.Unlock()

	// Once rather than a plain check, so that ten parallel tests asking the
	// same question run it once between them instead of ten times at once,
	// which is the case the cache exists for.
	entry.once.Do(func() {
		entry.out, entry.code, entry.err = execSpeclink(command, root, extra)
	})
	if entry.err != nil {
		t.Fatalf("run speclink: %v\n%s", entry.err, entry.out)
	}
	return entry.out, entry.code
}

// runResult is one memoised invocation.
//
// The error is kept rather than reported where it happened. A failure to start
// the binary belongs to whichever test asked, not to whichever test happened to
// win the race to run it first.
type runResult struct {
	once sync.Once
	out  string
	code int
	err  error
}

var (
	runCacheMu sync.Mutex
	runCache   = map[string]*runResult{}
)

// shareableFixture reports whether a root is one of the checked in fixtures,
// which no test modifies and every test therefore sees alike.
func shareableFixture(root string) bool {
	return strings.HasPrefix(filepath.ToSlash(root), "../../testdata/")
}

// execSpeclink drives the real binary. An exit code is a result rather than an
// error: it is part of the contract a loop runner depends on.
func execSpeclink(command, root string, extra []string) (string, int, error) {
	args := append([]string{command, "-root", root}, extra...)
	cmd := exec.Command(speclinkBin, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()

	var exit *exec.ExitError
	if err != nil {
		if ok := asExitError(err, &exit); ok {
			return string(out), exit.ExitCode(), nil
		}
		return string(out), 0, err
	}
	return string(out), 0, nil
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("a project that records its persistence choice must pass, got exit %d:\n%s", code, out)
	}
	if strings.Contains(out, "SPEC-V6-021") {
		t.Errorf("a recorded decision was not accepted:\n%s", out)
	}
}

// TestFieldCoverageReachesEveryField pins that forward coverage goes down to
// the field on both sides, and that the conformant fixture can satisfy it.
//
// Binding a type is not the same as accounting for what is in it. Types are
// created deliberately and reviewed; fields accrete afterwards.
//
// Both the domain model and the stored shape are asked, and the fixture is
// what showed why that is not pedantry. Customer.Notes existed in the domain
// and nowhere in CustomerEntity, and the mapping functions did not mention it:
// every note was dropped on save and read back empty. Both types compiled and
// both round tripped, so nothing else in the project could see it. The loss
// showed only as a domain field that traced to no requirement.
//
// Envelope fields are bound rather than exempted. QuoteID answers no business
// question on its own — it is what makes the sequence of events on a quote
// reconstructable, which is a requirement like any other.
func TestFieldCoverageReachesEveryField(t *testing.T) {
	t.Parallel()
	out, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("a project that binds its stored fields must pass, got exit %d:\n%s", code, out)
	}
	if strings.Contains(out, "SPEC-V6-022") {
		t.Errorf("a bound field was still reported:\n%s", out)
	}
}

// TestUINamingAppliesToTheUIDirectoryOnly pins that the rule stops at the ui
// directory and does not reach the packages below it.
//
// The presentation layer of a context is regularly more than one package: an
// editor for one widget, a shared table renderer. Those are ordinary Go
// packages named after what they do. Demanding uibilling of every one of them
// would force a dozen identically named packages across the system, which
// could then only be imported through aliases — the very confusion the naming
// rule exists to prevent.
//
// The fixture holds both: a ui directory with the wrong name, which must be
// reported, and a package below it with an ordinary name, which must not.
func TestUINamingAppliesToTheUIDirectoryOnly(t *testing.T) {
	t.Parallel()
	out, _ := runVerify(t, "../../testdata/arch")
	if n := strings.Count(out, "[SPEC-V6-040]"); n != 1 {
		t.Errorf("expected SPEC-V6-040 exactly once, found %d times:\n%s", n, out)
	}
	if strings.Contains(out, "amounteditor") {
		t.Errorf("a package below ui was reported:\n%s", out)
	}
}

// copyFixture makes a writable copy of a fixture.
//
// The tests that record a review have to write into the module they analyse,
// and a fixture is shared: one of them freezing in place would decide the
// outcome of every other test in the package depending on the order they ran.
func copyFixture(t *testing.T, src string) string {
	t.Helper()

	dir := t.TempDir()
	root, err := filepath.Abs(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(dir, os.DirFS(root)); err != nil {
		t.Fatal(err)
	}

	// The fixture replaces the spec module with a relative path, which stops
	// meaning this repository the moment the module is somewhere else.
	repo, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	gomod := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(gomod)
	if err != nil {
		t.Fatal(err)
	}
	fixed := strings.ReplaceAll(string(data),
		"replace github.com/worldiety/speclink => ../..",
		"replace github.com/worldiety/speclink => "+filepath.ToSlash(repo))
	if err := os.WriteFile(gomod, []byte(fixed), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// rewrite replaces a substring in a fixture file, failing when it is not there
// rather than silently testing nothing.
func rewrite(t *testing.T, root, rel, old, new string) {
	t.Helper()

	path := filepath.Join(root, rel)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), old) {
		t.Fatalf("%s does not contain %q", rel, old)
	}
	updated := strings.Replace(string(data), old, new, 1)
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestTestFilesDoNotChangeTheVerdict pins the split that made test loading
// possible.
//
// verify asks go/packages for test variants, and a test variant is the same
// source seen twice. Both hazards were measured before the filter was written:
// the in-package variant answers PkgPath identically to its subject, so every
// construct and every schema would be read twice, and the generated <pkg>.test
// binary is a main package outside cmd/, which K8-MAIN-LOCATION would report in
// every package that has a test.
func TestTestFilesDoNotChangeTheVerdict(t *testing.T) {
	t.Parallel()
	before, code := runVerify(t, "../../testdata/example")
	if code != 0 {
		t.Fatalf("the reference project is not clean: %s", before)
	}

	dir := copyFixture(t, "../../testdata/example")
	for name, body := range map[string]string{
		"app/sales/probe_test.go":     "package sales\n\nimport \"testing\"\n\nfunc TestProbe(t *testing.T) { _ = SubmitQuoteCmd{} }\n",
		"app/sales/probe_ext_test.go": "package sales_test\n\nimport \"testing\"\n\nfunc TestProbeExternal(t *testing.T) {}\n",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	after, code := runVerify(t, dir)
	if code != 0 {
		t.Fatalf("adding a test file broke the run:\n%s", after)
	}
	if summary(before) != summary(after) {
		t.Errorf("adding a test file changed the verdict:\nbefore: %s\nafter:  %s", summary(before), summary(after))
	}
}

// summary returns the last non-empty line, which is the counts line.
func summary(out string) string {
	lines := strings.Split(strings.TrimSpace(out), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// TestVerificationIsSeparateFromCoverage pins the question coverage never
// asked.
//
// K3 says code was written for a requirement. It has never said the code does
// what the requirement asks, and in a loop where the same model writes the
// implementation and the tests, nothing else does either. Deleting the test
// must therefore leave coverage untouched and verification broken — if both
// move together, one of them is not measuring anything of its own.
func TestVerificationIsSeparateFromCoverage(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")

	rewrite(t, dir, "app/sales/sales_test.go",
		"spec.Verified(t, quote.RQuoteSubmit)\n}\n\n// A submission that cannot draw",
		"}\n\n// A submission that cannot draw")
	rewrite(t, dir, "app/sales/sales_test.go",
		"\tspec.Verified(t, quote.RQuoteSubmit)\n}\n\n// --- R-QUOTE-APPROVE ---",
		"}\n\n// --- R-QUOTE-APPROVE ---")

	out, code := runVerify(t, dir)
	if code == 0 {
		t.Fatalf("removing the only test of a requirement was not reported:\n%s", out)
	}
	if !strings.Contains(out, "no test demonstrates R-QUOTE-SUBMIT") {
		t.Errorf("expected K14 for R-QUOTE-SUBMIT:\n%s", out)
	}
	// Coverage is a different question and must not have moved.
	if !strings.Contains(out, "100% covered") {
		t.Errorf("deleting a test changed the coverage figure:\n%s", summary(out))
	}
}

// spec.Verified outside a test claims something nothing can ever back. Left
// unreported it would look like verification that simply never gets recorded,
// which is the most expensive kind of silence in this tool.
func TestVerifiedOutsideATestIsReported(t *testing.T) {
	t.Parallel()
	dir := copyFixture(t, "../../testdata/example")

	rewrite(t, dir, "app/sales/uc_submit_quote.go",
		"import (",
		"import (\n\tspec \"github.com/worldiety/speclink/spec\"\n\t\"example.com/erp/requirements/fun/quote\"")
	rewrite(t, dir, "app/sales/uc_submit_quote.go",
		"\t\tif _, err := numbers.Next(); err != nil {",
		"\t\tspec.Verified(nil, quote.RQuoteSubmit)\n\t\tif _, err := numbers.Next(); err != nil {")

	out, _ := runVerify(t, dir)
	if !strings.Contains(out, "spec.Verified outside a test") {
		t.Errorf("expected the call in production code to be reported:\n%s", out)
	}
}
