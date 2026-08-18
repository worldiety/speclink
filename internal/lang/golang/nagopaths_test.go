package golang

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

// TestNagoPathsResolve verifies that every framework path a recogniser matches
// on still names a real package.
//
// This is the one coupling speclink has to the framework, and it is a coupling
// by string. That is deliberate: one binary serves projects on different
// framework versions, so it must not link the framework and must not pin it.
// The price is that a path is only a claim, checked by nothing — and a stale
// claim does not fail, it silently disables the rule that depends on it.
//
// That is not hypothetical. The generic CRUD user interface moved from
// presentation/ui/ent to application/ent/ui, and the constant kept pointing at
// the old location; the rule went on passing because it never matched anything
// again. This test is what turns such a move into a failure at the next version
// bump instead of a rule that quietly stopped working.
func TestNagoPathsResolve(t *testing.T) {
	paths := map[string]string{
		"nagoAuth":       nagoAuth,
		"nagoUser":       nagoUser,
		"nagoPermission": nagoPermission,
		"nagoEvs":        nagoEvs,
		"nagoData":       nagoData,
		"nagoNdb":        nagoNdb,
		"nagoEnt":        nagoEnt,
		"nagoEntCfg":     nagoEntCfg,
		"nagoUIEnt":      nagoUIEnt,
	}

	// Resolved from a fixture, because that is where the framework is required.
	// speclink's own module deliberately does not depend on it.
	dir, err := filepath.Abs("../../../testdata/example")
	if err != nil {
		t.Fatal(err)
	}

	patterns := make([]string, 0, len(paths))
	for _, p := range paths {
		patterns = append(patterns, p)
	}

	loaded, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles,
		Dir:  dir,
	}, patterns...)
	if err != nil {
		t.Fatalf("load framework packages: %v", err)
	}

	found := map[string]bool{}
	for _, p := range loaded {
		if len(p.Errors) > 0 || len(p.GoFiles) == 0 {
			continue
		}
		found[p.PkgPath] = true
	}

	for name, path := range paths {
		if !found[path] {
			t.Errorf("%s = %q names no package; the recogniser that uses it matches nothing.\n"+
				"The framework has moved or removed it. Find the new path and update the constant, "+
				"or drop the recogniser if the concept is gone.", name, path)
		}
	}
}
