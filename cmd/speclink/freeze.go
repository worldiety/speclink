package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/golang"
	"github.com/worldiety/speclink/internal/reqtree"
)

// freeze records the state of every edge the Go compiler cannot check: the
// shape of a frozen persisted type, the text of a requirement, and the content
// of a source segment.
//
// It is the act of promising, and the only thing in speclink that writes to the
// repository. What it writes is not intent — that stays in the code and in the
// documents — but the record of what has been read and accepted, so that a
// later run can tell a rename from a break, and a reformat from a rewrite.
//
// Its diff is where the review actually happens. That is the reason for doing
// it this way rather than with a status field somebody maintains: a moment of
// review that is a diff in a pull request survives contact with a project,
// where a discipline does not.
//
// It refuses to record anything that would violate an evolution rule. Without
// that refusal it would be the one command that launders a breaking change into
// the baseline, and the guard would be worth nothing. The escape stays where
// every other escape in this tool is: spec.Waive, with a reason.
func freeze(args []string) error {
	fs := flag.NewFlagSet("freeze", flag.ExitOnError)
	root := fs.String("root", ".", "repository root, holding "+baseline.FileName)
	cfgPath := fs.String("config", "", "layout configuration; defaults to "+config.FileName+" in the root")
	dry := fs.Bool("n", false, "report what would be recorded, write nothing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}

	layout, err := loadLayout(absRoot, *cfgPath)
	if err != nil {
		return err
	}

	pkgs, err := golang.Load(absRoot, fs.Args()...)
	if err != nil {
		return err
	}
	if errs := golang.TypeErrors(pkgs); len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "the Go build is broken; fix it before speclink can record anything:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "  "+e.Error())
		}
		return errFindings
	}

	var (
		bindings []ir.Binding
		schema   []ir.SchemaType
		reqs     []*ir.Requirement
		scope    = map[string]bool{}
	)
	discard := &diag.Set{}
	models := map[string]bool{}
	for _, p := range pkgs {
		bindings = append(bindings, p.ReadBindings(discard)...)
		reqs = append(reqs, p.ReadRequirements(discard)...)
		scope[p.PkgPath()] = true
		for name := range p.PersistedModels() {
			models[name] = true
		}
	}
	for _, p := range pkgs {
		schema = append(schema, p.ReadSchema(models)...)
	}
	check.SortSchema(schema)
	status := check.Drafts(schema, bindings, discard)

	base, err := baseline.Load(absRoot)
	if err != nil {
		return err
	}

	// A refusal is reported through the ordinary diagnostic channel, because it
	// is the same finding the next verify would produce anyway.
	blocking := &diag.Set{}
	check.Evolution(schema, status, base, scope, bindings, blocking)
	if broken := withoutMissing(blocking); !broken.Empty() {
		if err := broken.WriteText(os.Stdout); err != nil {
			return err
		}
		fmt.Fprintln(os.Stderr, "\nnothing was recorded: freezing a shape that already broke its promise would only make the break official.")
		return errFindings
	}

	added, updated := merge(base, schema, status)

	// The requirement texts and source segments are recorded without a guard of
	// their own. There is nothing to refuse: unlike a shape, which can break a
	// promise made to data already written, a rewritten requirement breaks
	// nothing by being recorded. Recording it is precisely the statement that
	// somebody has now read it.
	tree := reqtree.Build(absRoot, reqs, discard)
	docs, sourceDocs := loadSources(absRoot, layout, discard)
	changedReqs, changedSegs := check.Record(base, tree, docs, sourceDocs)

	if len(added) == 0 && len(updated) == 0 && changedReqs == 0 && changedSegs == 0 {
		fmt.Fprintln(os.Stderr, "nothing to record; every frozen shape, requirement and source segment is already recorded.")
		return nil
	}

	for _, name := range added {
		fmt.Fprintln(os.Stderr, "promise  "+name)
	}
	for _, name := range updated {
		fmt.Fprintln(os.Stderr, "extend   "+name)
	}
	if changedReqs > 0 {
		fmt.Fprintf(os.Stderr, "read     %s\n", plural(changedReqs, "requirement", "requirements"))
	}
	if changedSegs > 0 {
		fmt.Fprintf(os.Stderr, "read     %s\n", plural(changedSegs, "source segment", "source segments"))
	}
	if *dry {
		fmt.Fprintln(os.Stderr, "\nnothing written (-n).")
		return nil
	}
	if err := base.Save(absRoot); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "\nrecorded in %s. Review the diff: this is the point where a shape stops being an experiment.\n", baseline.FileName)
	return nil
}

// merge folds the current shapes into the baseline and reports what changed.
//
// Optionality is an annotation rather than something the type states, so it is
// applied here on the way into the record. Everything else comes from the type
// itself.
func merge(base *baseline.File, schema []ir.SchemaType, status map[string]check.Freeze) (added, updated []string) {
	for _, t := range schema {
		f, known := status[t.Name]
		if known && f.Draft {
			continue
		}
		if known {
			for i := range t.Fields {
				t.Fields[i].Optional = f.OptionalFields[t.Fields[i].Name]
			}
		}
		entry := baseline.EntryOf(t)
		previous, existed := base.Types[t.Name]
		switch {
		case !existed:
			added = append(added, t.Name)
		case !sameEntry(previous, entry):
			updated = append(updated, t.Name)
		default:
			continue
		}
		base.Types[t.Name] = entry
	}
	sort.Strings(added)
	sort.Strings(updated)
	return added, updated
}

// sameEntry reports whether two records describe the same promise. Comparing
// the field count alone would miss a newly optional field, and an update that
// is not written is a promise the next run cannot see.
func sameEntry(a, b baseline.Entry) bool {
	if a.Discriminator != b.Discriminator || len(a.Fields) != len(b.Fields) {
		return false
	}
	for i := range a.Fields {
		if a.Fields[i] != b.Fields[i] {
			return false
		}
	}
	return true
}

// withoutMissing drops the findings that freeze is there to answer. A shape
// that has never been recorded is exactly what this command records; the rest
// are breaks and must stop it.
func withoutMissing(in *diag.Set) *diag.Set {
	out := &diag.Set{}
	for _, f := range in.Findings() {
		if f.Rule == check.RuleBaselineMissing {
			continue
		}
		out.Add(f)
	}
	return out
}
