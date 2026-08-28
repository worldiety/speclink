package profile

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/lang"
	"github.com/worldiety/speclink/internal/lang/golang"
	"github.com/worldiety/speclink/internal/lang/jvm"
)

// The supported combinations.
//
// The set is closed and compiled in, because a style is approved as a whole:
// it prescribes a layout, a set of libraries and a way of writing use cases,
// and a project that assembled its own from parts would be pinning something
// nobody reviewed.
func init() {
	Register(&Profile{
		Name:      "go_nago_ddd1",
		Language:  Go,
		Framework: "nago",
		Style:     "ddd1",
		Summary:   "Go on the nago framework, DDD in three layers with a functional core",

		// The conventions this style assumes. They used to be config.Default(),
		// where they read as what speclink believes about every project rather
		// than as one style's layout.
		Layout: config.Config{
			ContextRoot: "app",
			CmdRoot:     "cmd",
			InfraRoots:  []string{"pkg", "foundation"},
			SourceRoots: []string{"requirements/_sources"},
		},
		Fields:       []string{"contextRoot", "cmdRoot", "infraRoots"},
		Architecture: true,

		Open: func(root string, layout config.Config, patterns []string, withTests bool, out *diag.Set) (lang.Model, error) {
			loader := golang.Load
			if withTests {
				loader = golang.LoadWithTests
			}
			loaded, err := loader(root, patterns...)
			if err != nil {
				return nil, err
			}
			if errs := golang.TypeErrors(loaded); len(errs) > 0 {
				return nil, buildBroken(errs)
			}
			all := golang.NonTests(loaded)
			return &golang.Model{
				All:          all,
				Measured:     golang.InScope(all, layout, root),
				TestPackages: golang.InScope(golang.Tests(loaded), layout, root),
				Layout:       layout,
				Root:         root,
				Style:        golang.DDD1,
				Framework:    golang.Nago,
			}, nil
		},
	})

	Register(&Profile{
		Name:      "go_bare_ddd1",
		Language:  Go,
		Framework: "bare",
		Style:     "ddd1",
		Summary:   "Go with no framework, DDD in three layers over a hand written foundation",

		Layout: config.Config{
			ContextRoot:    "app",
			CmdRoot:        "cmd",
			InfraRoots:     []string{"pkg", "foundation"},
			FoundationRoot: "foundation",
			SourceRoots:    []string{"requirements/_sources"},
		},
		Fields:       []string{"contextRoot", "cmdRoot", "infraRoots", "foundationRoot"},
		Architecture: true,

		Open: func(root string, layout config.Config, patterns []string, withTests bool, out *diag.Set) (lang.Model, error) {
			loader := golang.Load
			if withTests {
				loader = golang.LoadWithTests
			}
			loaded, err := loader(root, patterns...)
			if err != nil {
				return nil, err
			}
			if errs := golang.TypeErrors(loaded); len(errs) > 0 {
				return nil, buildBroken(errs)
			}
			all := golang.NonTests(loaded)

			// The foundation lives in the project's own module, so the paths
			// the recognisers match on are not knowable until it is read.
			module, err := golang.ModulePath(all)
			if err != nil {
				return nil, err
			}
			return &golang.Model{
				All:          all,
				Measured:     golang.InScope(all, layout, root),
				TestPackages: golang.InScope(golang.Tests(loaded), layout, root),
				Layout:       layout,
				Root:         root,
				Style:        golang.Bare,
				Framework:    golang.BareFoundation(module, layout.FoundationRoot),
				Layered:      true,
			}, nil
		},
	})

	Register(&Profile{
		Name:      "java_springboot_ddd1",
		Language:  JVM,
		Framework: "springboot",
		Style:     "ddd1",
		Summary:   "Java on Spring Boot, DDD in three layers; the style prescribes no rules yet",

		Layout: config.Config{
			SourceRoots: []string{"requirements/_sources"},
		},
		Fields: []string{"classRoots", "sourceCode", "reportRoots", "specPackage"},
		// The name claims the same architecture as its Go counterpart and the
		// claim is not yet redeemed: this frontend has no architecture rules.
		// Saying so is the same honesty as the capability lines — a rule family
		// that never ran must not read as one that came out clean.
		Architecture: false,

		Open: func(root string, layout config.Config, patterns []string, withTests bool, out *diag.Set) (lang.Model, error) {
			classes, errs := jvm.Load(root, layout.ClassRoots)
			for _, e := range errs {
				// A class file that cannot be read is reported and skipped.
				// Unlike a broken Go build its neighbours are unaffected, and
				// refusing the run over one unreadable file would be borrowing
				// a rule from a language this is not.
				fmt.Fprintln(os.Stderr, "  "+e.Error())
			}
			if len(classes) == 0 {
				return nil, fmt.Errorf("no compiled classes found under %s; build the project first, or set classRoots in %s",
					root, config.FileName)
			}
			return jvm.NewModel(jvm.NewReader(root, classes, layout.SourceCode, layout.SpecPackage)), nil
		},
	})
}

// buildBroken reports a failed compilation.
//
// Phase V2 is the compiler itself. When it fails there is nothing meaningful to
// say about annotations, and saying it anyway would bury the real cause under
// follow-up noise.
func buildBroken[E error](errs []E) error {
	var b strings.Builder
	b.WriteString("the build is broken; fix it before speclink can read anything:\n")
	for _, e := range errs {
		b.WriteString("  " + e.Error() + "\n")
	}
	return errors.New(strings.TrimRight(b.String(), "\n"))
}
