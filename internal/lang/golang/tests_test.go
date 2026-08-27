package golang

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Asking for test variants must not change what any existing rule sees.
//
// Both hazards here were measured on the reference project before the split was
// written, not guessed at. With a single _test.go file present, go/packages
// returns three extra entries: the package rebuilt with its in-package tests,
// the external test package, and a generated main. One of them answers PkgPath
// identically to the package it tests, and the generated main is a main package
// outside cmd/.

func TestTestVariantsAreDetectedByIDNotPkgPath(t *testing.T) {
	dir := fixtureWithTest(t)

	pkgs, err := LoadWithTests(dir, "./...")
	if err != nil {
		t.Fatal(err)
	}

	var (
		tests    = Tests(pkgs)
		nonTests = NonTests(pkgs)
	)
	if len(tests) == 0 {
		t.Fatalf("no test variant found among %d packages", len(pkgs))
	}
	if len(tests)+len(nonTests) != len(pkgs) {
		t.Fatalf("the split loses packages: %d + %d != %d", len(tests), len(nonTests), len(pkgs))
	}

	// The point of deciding on ID: at least one test variant shares its
	// PkgPath with a package proper, so a check on PkgPath would separate
	// nothing.
	byPath := map[string]int{}
	for _, p := range nonTests {
		byPath[p.PkgPath()]++
	}
	shared := 0
	for _, p := range tests {
		if byPath[p.PkgPath()] > 0 {
			shared++
		}
	}
	if shared == 0 {
		t.Error("no test variant shares a PkgPath with its subject; this test no longer proves what it was written for")
	}

	// And no two packages that survive the filter may answer PkgPath the same,
	// or every rule keyed on a package sees it twice.
	seen := map[string]string{}
	for _, p := range nonTests {
		if prev, dup := seen[p.PkgPath()]; dup {
			t.Errorf("%s is loaded twice as %s and %s", p.PkgPath(), prev, p.ID())
		}
		seen[p.PkgPath()] = p.ID()
	}
}

// The generated <pkg>.test binary is a main package outside cmd/. Left in, it
// would make K8-MAIN-LOCATION fire in every package that has a test.
func TestGeneratedTestMainIsClassifiedAsATest(t *testing.T) {
	dir := fixtureWithTest(t)

	pkgs, err := LoadWithTests(dir, "./...")
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range NonTests(pkgs) {
		if p.pkg.Name == "main" && p.PkgPath() != "example.com/erp/cmd/erp" {
			t.Errorf("%s is a main package that no rule expects", p.ID())
		}
	}
}

// Test files must not be mistaken for annotation or requirement files. Go
// requires _test.go to be the final suffix, so the classification order is what
// decides this rather than the names being disjoint.
func TestTestFilesAreTheirOwnCategory(t *testing.T) {
	dir := fixtureWithTest(t)

	pkgs, err := LoadWithTests(dir, "./...")
	if err != nil {
		t.Fatal(err)
	}

	found := 0
	for _, p := range pkgs {
		found += len(p.testFiles)
		for _, f := range p.annotationFiles {
			name := filepath.Base(p.pkg.Fset.Position(f.Pos()).Filename)
			if len(name) > len(TestSuffix) && name[len(name)-len(TestSuffix):] == TestSuffix {
				t.Errorf("%s was classified as an annotation file", name)
			}
		}
	}
	if found == 0 {
		t.Error("no test file was classified as one")
	}
}

// A load without tests must be unchanged, or the split is not a split.
func TestLoadWithoutTestsSeesNoTestPackages(t *testing.T) {
	dir := fixtureWithTest(t)

	pkgs, err := Load(dir, "./...")
	if err != nil {
		t.Fatal(err)
	}
	if got := len(Tests(pkgs)); got != 0 {
		t.Errorf("got %d test packages from a plain Load, want 0", got)
	}
	for _, p := range pkgs {
		if len(p.testFiles) != 0 {
			t.Errorf("%s carries test files after a plain Load", p.ID())
		}
	}
}

// fixtureWithTest copies the reference project and adds a test file to it.
//
// A copy rather than the fixture itself, because these tests are about what a
// test file does to the load and the fixture is shared with every other test in
// the repository. The one that eventually lives there permanently belongs to
// K14, not here.
func fixtureWithTest(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	src, err := filepath.Abs(filepath.Join("..", "..", "..", "testdata", "example"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(dir, os.DirFS(src)); err != nil {
		t.Fatal(err)
	}

	// The fixture replaces the spec module with a relative path, which stops
	// naming this repository once the module is somewhere else.
	repo, err := filepath.Abs(filepath.Join("..", "..", ".."))
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

	// Both kinds: an in-package test, which is the variant that collides on
	// PkgPath, and an external one, which does not.
	write := func(rel, body string) {
		if err := os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("app/sales/probe_test.go", "package sales\n\nimport \"testing\"\n\nfunc TestProbe(t *testing.T) { _ = SubmitQuoteCmd{} }\n")
	write("app/sales/probe_ext_test.go", "package sales_test\n\nimport \"testing\"\n\nfunc TestProbeExternal(t *testing.T) {}\n")
	return dir
}
