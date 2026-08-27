package check

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoHostLanguageInTheRules guards the split this package rests on.
//
// internal/check decides; a frontend phrases. That was always the design and
// for a long time only half true: the rules were language neutral in what they
// concluded and wrote Go in every How line, which is the half a reader never
// sees until they try to add a second language.
//
// The temptation to slip one back is real, because a fix that names the actual
// syntax is genuinely more useful than one that does not — which is exactly why
// ir.Dialect exists and why this is a test rather than a note.
func TestNoHostLanguageInTheRules(t *testing.T) {
	// Spellings that only make sense in one language. The list is short on
	// purpose: it names what has been wrong here before rather than trying to
	// describe Go.
	banned := []string{
		"spec.For", "spec.Satisfies", "spec.Waive", "spec.Draft",
		"spec.Optional", "spec.Verified", "spec.ForField", "spec.ForDecl",
		"var _ =", ".annotation.go", ".spec.go",
	}

	for _, dir := range []string{".", "../reqtree", "../baseline", "../diag", "../source"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			path := filepath.Join(dir, name)
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, b := range banned {
				if strings.Contains(string(data), b) {
					t.Errorf("%s spells %q; a rule decides, a frontend phrases — add it to ir.Dialect instead", path, b)
				}
			}
		}
	}
}

// TestRulesDoNotImportAFrontend is the same boundary one level up. A rule that
// imports a frontend cannot be reused by a second one, whatever its strings say.
func TestRulesDoNotImportAFrontend(t *testing.T) {
	for _, dir := range []string{".", "../reqtree", "../baseline", "../diag", "../source", "../ir"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(data), "internal/lang/") {
				t.Errorf("%s imports a language frontend", path)
			}
		}
	}
}
