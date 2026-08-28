package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/reqtree"
)

// impact answers what a change reaches.
//
// Everything needed for it was already being computed and thrown away: the
// coverage keeps a map from requirement to the constructs satisfying it, the
// tree keeps the derivation graph, and the source coverage keeps a map from
// segment to the requirements extracted from it. Three edges of one graph,
// each read once for a percentage and then dropped.
//
// It exists because the loop this tool sits in has two moments where somebody
// has to decide what to look at, and both of them are currently guesswork.
// Somebody edits a paragraph of a specification and has to know which code that
// reaches. An agent is handed a diff and has to know which requirements it
// touches. Neither question is answerable by reading the code, because the
// chain runs through the requirement tree, and neither is answerable by reading
// the tree, because the chain runs through the document.
//
// It reports rather than judges. There are no findings and the exit code says
// whether the query resolved, not whether the project is sound — this is the
// one command whose answer is a fact about the graph rather than a verdict on
// it.
func impact(args []string) error {
	fs := flag.NewFlagSet("impact", flag.ExitOnError)
	root := fs.String("root", ".", "repository root, used to resolve source documents")
	cfgPath := fs.String("config", "", "layout configuration; defaults to "+config.FileName+" in the root")
	prof := fs.String("profile", "", "language, framework and architectural style; overrides "+config.FileName)
	format := fs.String("format", "text", "output format: text or json")
	patterns := fs.String("packages", "./...", "packages to load")
	if err := fs.Parse(args); err != nil {
		return err
	}
	targets := fs.Args()
	if len(targets) == 0 {
		return fmt.Errorf("nothing to trace; name a requirement ID, a source segment as doc.md#anchor, or a file path")
	}

	absRoot, err := absRootOf(*root)
	if err != nil {
		return err
	}

	model, layout, _, err := open(absRoot, *cfgPath, *prof, strings.Fields(*patterns), false)
	if err != nil {
		return err
	}

	// Diagnostics are discarded on purpose. This command asks where a change
	// reaches, and answering it while the project has unrelated defects is the
	// normal case — a project in good order rarely needs the question.
	discard := &diag.Set{}
	reqs := model.Requirements(discard)
	bindings := model.Bindings(discard)

	tree := reqtree.Build(absRoot, reqs, discard)
	docs, sourceDocs := loadSources(absRoot, layout, discard)
	srcCov := check.CoverSources(tree, docs, sourceDocs, nil, discard)
	cov := check.CoverRequirements(tree, bindings, nil, model.Dialect(), discard)

	g := &graph{
		root:     absRoot,
		tree:     tree,
		sources:  srcCov,
		cov:      cov,
		bindings: bindings,
		dialect:  model.Dialect(),
	}

	report := impactReport{Version: ImpactVersion}
	for _, target := range targets {
		got, err := g.trace(target)
		if err != nil {
			return err
		}
		report.Traced = append(report.Traced, got)
	}

	switch *format {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	case "text":
		return report.writeText(os.Stdout)
	default:
		return fmt.Errorf("unknown format %q, expected text or json", *format)
	}
}

// ImpactVersion is the schema version of the impact report.
const ImpactVersion = 1

type impactReport struct {
	Version int      `json:"version"`
	Traced  []traced `json:"traced"`
}

// traced is what one target reaches.
//
// The three lists are the three edges of the chain, kept apart rather than
// flattened. A reader needs to know whether something turned up because a
// requirement derives from another or because a construct implements it: the
// first is a question for whoever owns the specification, the second for
// whoever owns the code, and they are rarely the same person.
type traced struct {
	Target string `json:"target"`
	// Kind says how the target was understood: requirement, segment or file.
	Kind string `json:"kind"`
	// Requirements are the requirement IDs reached, including the target when
	// it is one.
	Requirements []string `json:"requirements,omitempty"`
	// Derived are the requirements that derive from those, transitively. A
	// change to a parent reaches every child, which is what the derivation
	// graph is for and what nothing so far has ever walked.
	Derived []string `json:"derived,omitempty"`
	// Constructs are the implementations that will have to be re-read.
	Constructs []string `json:"constructs,omitempty"`
	// Sources are the segments the requirements came from, so a code change can
	// be traced back to the paragraph that asked for it.
	Sources []string `json:"sources,omitempty"`
	// Unresolved says the target matched nothing, which is an answer and not an
	// error: asking about a file that carries no construct is a fair question
	// with the answer "nothing".
	Unresolved bool `json:"unresolved,omitempty"`
}

type graph struct {
	root     string
	tree     *reqtree.Tree
	sources  check.SourceCoverage
	cov      check.Coverage
	bindings []ir.Binding
	dialect  ir.Dialect
}

func (g *graph) trace(target string) (traced, error) {
	switch {
	case g.tree.ByID[target] != nil:
		return g.fromRequirements(target, "requirement", []string{target}), nil
	case strings.Contains(target, "#"):
		return g.fromRequirements(target, "segment", g.sources.ByCiter[target]), nil
	default:
		return g.fromFile(target), nil
	}
}

// fromRequirements expands a set of requirements into everything downstream.
func (g *graph) fromRequirements(target, kind string, seeds []string) traced {
	out := traced{Target: target, Kind: kind}

	reached := map[string]bool{}
	for _, id := range seeds {
		reached[id] = true
	}

	// Derivation runs child-to-parent in the model, so reaching the children of
	// a requirement means walking the edges backwards. Doing it transitively is
	// the point: a decision three levels up reaches everything under it, and
	// that is exactly the change nobody traces by hand.
	children := map[string][]string{}
	for _, r := range g.tree.All() {
		for _, parent := range r.DerivedFrom {
			children[parent] = append(children[parent], r.ID)
		}
	}
	derived := map[string]bool{}
	var walk func(id string)
	walk = func(id string) {
		for _, child := range children[id] {
			if derived[child] || reached[child] {
				continue
			}
			derived[child] = true
			walk(child)
		}
	}
	for _, id := range seeds {
		walk(id)
	}

	out.Requirements = keys(reached)
	out.Derived = keys(derived)

	constructs := map[string]bool{}
	segments := map[string]bool{}
	for _, id := range append(append([]string{}, out.Requirements...), out.Derived...) {
		for _, t := range g.cov.BySatisfier[id] {
			constructs[t.String()] = true
		}
		if r := g.tree.ByID[id]; r != nil {
			for _, s := range r.Sources {
				if s.Doc != "" && s.Anchor != "" {
					segments[s.Doc+"#"+s.Anchor] = true
				}
			}
		}
	}
	out.Constructs = keys(constructs)
	out.Sources = keys(segments)
	out.Unresolved = len(out.Requirements) == 0 && len(out.Derived) == 0
	return out
}

// fromFile answers the other direction: a path out of a diff, and what
// specification it touches.
//
// The match is per file, not per package, and that distinction is what decides
// whether the answer is worth having. Constructs carry a qualified name rather
// than a position, so the obvious implementation matches on the package — and
// in the reference project that returns every requirement the package has,
// which is not an answer to anything.
//
// What makes it exact is the sidecar convention. A binding lives in
// <base>.annotation.go beside <base>.go, and a binding does carry a position,
// so the annotation file belonging to a changed file names precisely the
// constructs declared in it. The convention was introduced to keep an
// annotation next to the thing it describes; that it also makes this question
// answerable is the kind of thing a convention earns.
func (g *graph) fromFile(target string) traced {
	out := traced{Target: target, Kind: "file"}

	want := filepath.ToSlash(strings.TrimPrefix(target, "./"))
	// The sidecar name is the frontend's convention, not this command's.
	sidecar := filepath.ToSlash(filepath.Join(filepath.Dir(want), g.dialect.AnnotationFile(want)))

	reached := map[string]bool{}
	constructs := map[string]bool{}

	for _, b := range g.bindings {
		if !matchesPath(g.root, b.Pos.File, want) && !matchesPath(g.root, b.Pos.File, sidecar) {
			continue
		}
		for _, a := range b.Assertions {
			if a.Kind != ir.AssertSatisfies {
				continue
			}
			for _, ref := range a.Requirements {
				r := g.tree.ByGoIdent(ref)
				if r == nil {
					r = g.tree.ByID[ref]
				}
				if r != nil {
					reached[r.ID] = true
					constructs[b.Target.String()] = true
				}
			}
		}
	}

	// A requirement file names itself.
	for _, r := range g.tree.All() {
		if matchesPath(g.root, r.Pos.File, want) {
			reached[r.ID] = true
		}
	}

	if len(reached) == 0 {
		out.Unresolved = true
		return out
	}
	expanded := g.fromRequirements(target, "file", keys(reached))
	expanded.Kind = "file"
	for c := range constructs {
		if !contains(expanded.Constructs, c) {
			expanded.Constructs = append(expanded.Constructs, c)
		}
	}
	sort.Strings(expanded.Constructs)
	return expanded
}

func matchesPath(root, have, want string) bool {
	have = filepath.ToSlash(have)
	if rel, err := filepath.Rel(root, have); err == nil {
		have = filepath.ToSlash(rel)
	}
	return have == want || strings.HasSuffix(have, "/"+want)
}

func (r impactReport) writeText(w *os.File) error {
	for _, t := range r.Traced {
		if t.Unresolved {
			fmt.Fprintf(w, "%s (%s): reaches nothing\n\n", t.Target, t.Kind)
			continue
		}
		fmt.Fprintf(w, "%s (%s)\n", t.Target, t.Kind)
		section(w, "requirements", t.Requirements)
		section(w, "derived", t.Derived)
		section(w, "constructs to re-read", t.Constructs)
		section(w, "asked for in", t.Sources)
		fmt.Fprintln(w)
	}
	return nil
}

func section(w *os.File, label string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(w, "  %s\n", label)
	for _, s := range items {
		fmt.Fprintf(w, "    %s\n", s)
	}
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
