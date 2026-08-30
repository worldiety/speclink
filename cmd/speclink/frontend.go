package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
//
// Package patterns narrow what is *measured*, never what is loaded. See
// scopeFromPatterns for why, and openNamed for the one command that means the
// other thing and says so.
func open(root, cfgPath, override string, patterns []string, withTests bool) (lang.Model, config.Config, *profile.Profile, error) {
	return read(root, cfgPath, override, patterns, withTests, false)
}

// openNamed loads only the packages the patterns name.
//
// It exists for the command that reads the requirement tree and nothing else.
// That command has to work while the implementation around the tree is still in
// pieces, so it must not load the code — and it is allowed to, because it makes
// no claim about the module: it never asks whether an entry point exists, what
// satisfies a requirement, or what the system exposes. Every question it does
// ask is answered entirely by the files it loaded.
//
// The distinction is the whole of it. A command that narrows the load and then
// makes a statement about the whole project is the defect this pair exists to
// keep from recurring; a command that narrows the load and only speaks about
// what it read is correct.
func openNamed(root, cfgPath, override string, patterns []string, withTests bool) (lang.Model, config.Config, *profile.Profile, error) {
	return read(root, cfgPath, override, patterns, withTests, true)
}

func read(root, cfgPath, override string, patterns []string, withTests, named bool) (lang.Model, config.Config, *profile.Profile, error) {
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

	if named {
		model, err := p.Open(root, layout, patterns, withTests, &diag.Set{})
		if err != nil {
			return nil, layout, p, err
		}
		return model, layout, p, nil
	}

	// A package pattern on the command line is a scope, and it is folded into
	// the one the configuration declares rather than handed to the loader.
	//
	// This is the same mistake the configured scope already made once and was
	// fixed for: filtering the loaded set makes every rule that resolves across
	// packages answer differently depending on what the operator typed. Worse
	// here, because the rules are written against an invariant the loader was
	// quietly breaking — All is the whole module and answers questions about
	// the module, Measured is the reported subset and answers questions about a
	// package. A narrowed load left K8-MAIN-EXISTS reporting that a module has
	// no entry point because cmd/ had not been read, and the whole requirement
	// tree absent, so a run announced "100% covered" having loaded no
	// requirement at all.
	//
	// Folding it here means there is one narrowing mechanism instead of two,
	// and the surviving one is the tested one.
	scoped, err := scopeFromPatterns(patterns)
	if err != nil {
		return nil, layout, p, err
	}
	layout.Scope = append(layout.Scope, scoped...)

	model, err := p.Open(root, layout, []string{wholeModule}, withTests, &diag.Set{})
	if err != nil {
		return nil, layout, p, err
	}
	return model, layout, p, nil
}

// wholeModule is the only pattern the loader is ever given.
const wholeModule = "./..."

// scopeFromPatterns turns command line package patterns into scope entries.
//
// The translation is lexical and it is allowed to be, because both sides are
// directories relative to the same root: a scope entry is a project relative
// path, and a relative Go pattern is resolved against -root. The only
// difference is spelling, "/..." against "/**".
//
// Anything else is refused rather than approximated. An import path or a meta
// pattern could be mapped onto directories with a couple of assumptions, and a
// pattern mapped slightly wrong does not fail — it quietly measures something
// other than what was asked for, which is precisely the failure this whole
// change exists to remove.
func scopeFromPatterns(patterns []string) ([]string, error) {
	var out []string
	for _, raw := range patterns {
		pattern := strings.TrimSpace(raw)
		if pattern == "" {
			continue
		}
		if pattern == wholeModule || pattern == "..." {
			// The whole module is asked for, so nothing is narrowed. Returning
			// early rather than collecting the rest: a run that names the whole
			// module and one package below it has asked for the whole module,
			// and quietly intersecting the two would answer a question nobody
			// put.
			return nil, nil
		}
		if !strings.HasPrefix(pattern, "./") && pattern != "." {
			return nil, fmt.Errorf("package pattern %q is not a path relative to the root; "+
				"speclink narrows by directory, so write it as ./dir/... rather than an import path or a meta pattern like all", raw)
		}
		if strings.ContainsAny(pattern, "*?[") {
			return nil, fmt.Errorf("package pattern %q uses a glob; "+
				"write ./dir/... to mean a directory and everything below it", raw)
		}

		rel := strings.TrimPrefix(pattern, "./")
		switch {
		case rel == "." || rel == "":
			out = append(out, ".")
		case strings.HasSuffix(rel, "/..."):
			out = append(out, strings.TrimSuffix(rel, "/...")+"/**")
		default:
			out = append(out, strings.TrimSuffix(rel, "/"))
		}
	}
	return out, nil
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
	// The frontend can read persisted shapes and this framework gives it
	// nothing to recognise them by. Missing() cannot see this: it asks the
	// frontend type, and both Go profiles share one. Without this line a
	// project whose stored data no rule guards reports 0 findings and a
	// summary that is 100% in every column.
	if _, reads := m.(lang.SchemaReader); reads && !p.Schemas {
		fmt.Fprintln(os.Stderr, "not measured: schema evolution, because profile "+p.Name+" has no persistence recogniser")
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
