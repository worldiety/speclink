package golang

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

// TestFrameworkContractResolves is the guard that makes the string coupling
// survivable.
//
// Every name in frameworkContract is a claim about a framework speclink neither
// links nor pins, checked by nothing at run time. A wrong claim does not
// produce a wrong answer, which would be noticed; it produces no answer, which
// is not. The rule depending on it simply stops matching and goes on passing.
//
// This resolves every claim against a real framework and, when one is gone,
// says which rule it took with it. That is the difference between a version
// bump failing here and a rule quietly ceasing to exist.
func TestFrameworkContractResolves(t *testing.T) {
	// Resolved from a fixture, because that is where the framework is required.
	// speclink's own module deliberately does not depend on it.
	dir, err := filepath.Abs("../../../testdata/example")
	if err != nil {
		t.Fatal(err)
	}

	seen := map[string]bool{}
	var patterns []string
	for _, s := range frameworkContract {
		if !seen[s.Pkg] {
			seen[s.Pkg] = true
			patterns = append(patterns, s.Pkg)
		}
	}

	loaded, err := packages.Load(&packages.Config{
		// NeedFiles is not optional here. A path that names nothing still comes
		// back as a package, with an empty scope and no files, so without it a
		// package that has moved away resolves exactly like one that never had
		// the symbol — which is the very bug this guard exists for.
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedFiles,
		Dir:  dir,
	}, patterns...)
	if err != nil {
		t.Fatalf("load framework packages: %v", err)
	}

	byPath := map[string]*packages.Package{}
	for _, p := range loaded {
		if p.Types != nil && len(p.Errors) == 0 && len(p.GoFiles) > 0 {
			byPath[p.PkgPath] = p
		}
	}

	for _, s := range frameworkContract {
		pkg, ok := byPath[s.Pkg]
		if !ok {
			t.Errorf("package %s does not resolve; this disables %s", s.Pkg, s.Breaks)
			continue
		}
		if s.Name == "" {
			continue
		}
		if obj := pkg.Types.Scope().Lookup(s.Name); obj == nil || !obj.Exported() {
			t.Errorf("%s.%s does not resolve; this disables %s", s.Pkg, s.Name, s.Breaks)
		}
	}
}

// TestEveryFrameworkPathIsUnderContract pins that the contract is complete.
//
// A declared list only helps while it is the only list. The previous version of
// this guard kept its own copy of the paths beside the constants, and the copy
// had already drifted: nagoDataJSON was never in it, so the JSON repository
// recogniser was unchecked from the day it was written. A constant that no
// entry mentions is that mistake happening again.
func TestEveryFrameworkPathIsUnderContract(t *testing.T) {
	constants := map[string]string{
		"nagoAuth":       nagoAuth,
		"nagoUser":       nagoUser,
		"nagoPermission": nagoPermission,
		"nagoEvs":        nagoEvs,
		"nagoData":       nagoData,
		"nagoDataJSON":   nagoDataJSON,
		"nagoNdb":        nagoNdb,
		"nagoEnt":        nagoEnt,
		"nagoEntCfg":     nagoEntCfg,
		"nagoUIEnt":      nagoUIEnt,
	}

	covered := map[string]bool{}
	for _, s := range frameworkContract {
		covered[s.Pkg] = true
	}
	for name, path := range constants {
		if !covered[path] {
			t.Errorf("%s (%s) is matched on but named by no entry of frameworkContract, so nothing checks that it still resolves", name, path)
		}
	}
}
