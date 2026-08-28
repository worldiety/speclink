package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/lang"
	"github.com/worldiety/speclink/internal/lang/golang"
	"github.com/worldiety/speclink/internal/lang/jvm"
)

// openModel picks a frontend and reads the project with it.
//
// The choice is made from what is on disk rather than from a flag, because it
// is not a preference: a project is written in a language, and asking is asking
// somebody to tell the tool something it can see. A flag exists anyway for the
// case where both are true, which is a repository holding a Go tool beside a
// Java service.
//
// The frontends are not interchangeable and the interface does not pretend they
// are. Loading is where they differ most — package patterns against directories
// of compiled classes, a build that either succeeds or says nothing against
// files that fail one at a time — so each is loaded on its own terms and only
// the reading is shared.
func openModel(root string, layout config.Config, want string, patterns []string, withTests bool, verb string) (lang.Model, error) {
	kind := want
	if kind == "" {
		kind = detectFrontend(root)
	}

	switch kind {
	case "go":
		loaded, err := load(root, withTests, verb, patterns)
		if err != nil {
			return nil, err
		}
		all := golang.NonTests(loaded)
		return &golang.Model{
			All:          all,
			Measured:     golang.InScope(all, layout, root),
			TestPackages: golang.InScope(golang.Tests(loaded), layout, root),
			Layout:       layout,
			Root:         root,
		}, nil

	case "jvm":
		classes, errs := jvm.Load(root, layout.ClassRoots)
		for _, e := range errs {
			// A class file that cannot be read is reported and skipped. Unlike
			// a broken Go build, its neighbours are unaffected, and refusing
			// the whole run over one unreadable file would be borrowing a rule
			// from a language this is not.
			fmt.Fprintln(os.Stderr, "  "+e.Error())
		}
		if len(classes) == 0 {
			return nil, fmt.Errorf("no compiled classes found under %s; build the project first, or set classRoots in %s",
				root, config.FileName)
		}
		return jvm.NewModel(jvm.NewReader(root, classes, layout.SourceCode, layout.SpecPackage)), nil

	default:
		return nil, fmt.Errorf("unknown frontend %q, expected go or jvm", kind)
	}
}

// detectFrontend guesses the language from what a project has in it.
//
// A go.mod wins, because a repository holding both is far likelier to be a Go
// project with a Java fixture in it than the other way round — which is the
// case in this repository, and the reason the flag exists.
func detectFrontend(root string) string {
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
		return "go"
	}
	for _, rel := range jvm.ClassRoots {
		if info, err := os.Stat(filepath.Join(root, rel)); err == nil && info.IsDir() {
			return "jvm"
		}
	}
	return "go"
}

// reportCapabilities says which directions a run could not measure.
//
// It exists for the same reason the scope line does. A frontend that infers no
// constructs does not measure forward coverage weakly — it does not measure it
// — and a summary that stayed silent would let "no answer" read as "clean".
func reportCapabilities(m lang.Model) {
	for _, missing := range lang.Of(m).Missing() {
		fmt.Fprintln(os.Stderr, "not measured: "+missing)
	}
}
