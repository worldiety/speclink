package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/lang"
	"github.com/worldiety/speclink/internal/lang/golang"
	"github.com/worldiety/speclink/internal/profile"
)

// open resolves the profile, applies the project's deviations to it, and reads
// the project with the frontend it names.
//
// There is no detection and no default, and that is the change. Language could
// be guessed from a go.mod and framework from an import, but style cannot be
// guessed from anything — and guessing it wrongly is expensive in a particular
// way. It reports dozens of findings about a convention the project never meant
// to follow, which teaches the reader that the tool is wrong rather than that
// the project is. Better to ask once.
func open(root, cfgPath, override string, patterns []string, withTests bool) (lang.Model, config.Config, *profile.Profile, error) {
	declared, err := loadLayout(root, cfgPath)
	if err != nil {
		return nil, config.Config{}, nil, err
	}

	name := declared.Profile
	if override != "" {
		name = override
	}
	p, err := profile.Get(name)
	if err != nil {
		return nil, config.Config{}, nil, err
	}
	// A key the profile does not use is refused rather than ignored. It is
	// almost always a profile that changed without the configuration following,
	// and the symptom otherwise is a setting that quietly has no effect.
	if err := p.CheckConfig(declared.Keys()); err != nil {
		return nil, config.Config{}, nil, err
	}

	// The profile carries the conventions, the project states its deviations.
	layout := declared.Over(p.Layout)

	model, err := p.Open(root, layout, patterns, withTests, &diag.Set{})
	if err != nil {
		return nil, layout, p, err
	}
	return model, layout, p, nil
}

// absRootOf resolves the repository root.
func absRootOf(root string) (string, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("resolve root: %w", err)
	}
	return abs, nil
}

// reportCapabilities says which directions a run could not measure.
//
// It exists for the same reason the scope line does. A frontend that infers no
// constructs does not measure forward coverage weakly — it does not measure it
// — and a summary that stayed silent would let "no answer" read as "clean".
func reportCapabilities(m lang.Model, p *profile.Profile) {
	for _, missing := range lang.Of(m).Missing() {
		fmt.Fprintln(os.Stderr, "not measured: "+missing)
	}
	if !p.Architecture {
		// The profile's name claims an architecture and the rules for it do not
		// exist yet. Passing quietly would make an unwritten rule family read
		// as one that came out clean.
		fmt.Fprintln(os.Stderr, "not measured: architecture, because profile "+p.Name+" prescribes no rules yet")
	}
}

// skippedPackages reports how much of a project the scope left out.
//
// Only a frontend with a notion of packages can answer, which is why it is
// asked rather than assumed. A frontend that has none skips nothing, and
// reporting a zero is the truthful answer rather than a placeholder.
func skippedPackages(m lang.Model) int {
	if g, ok := m.(*golang.Model); ok {
		return len(golang.OutOfScope(g.All, g.Layout, g.Root))
	}
	return 0
}

// profileLanguage returns which reader a run is using, for the few places that
// still have to know — evidence, where a Go test stream and a JVM test report
// are genuinely different artefacts and neither is a special case of the other.
func profileLanguage(override string, layout config.Config) profile.Language {
	name := layout.Profile
	if override != "" {
		name = override
	}
	if p, err := profile.Get(name); err == nil {
		return p.Language
	}
	return profile.Go
}
