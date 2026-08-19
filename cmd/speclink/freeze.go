package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/golang"
)

// freeze records the current shape of every frozen persisted type.
//
// It is the act of promising, and the only thing in speclink that writes to the
// repository. What it writes is not intent — that stays in the code — but the
// record of what the code has committed to, so that a later run can tell a
// rename from a break.
//
// It refuses to record anything that would violate an evolution rule. Without
// that refusal it would be the one command that launders a breaking change into
// the baseline, and the guard would be worth nothing. The escape stays where
// every other escape in this tool is: spec.Waive, with a reason.
func freeze(args []string) error {
	fs := flag.NewFlagSet("freeze", flag.ExitOnError)
	root := fs.String("root", ".", "repository root, holding "+baseline.FileName)
	dry := fs.Bool("n", false, "report what would be recorded, write nothing")
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
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
		bindings   []ir.Binding
		constructs []ir.Construct
		schema     []ir.SchemaType
		scope      = map[string]bool{}
	)
	discard := &diag.Set{}
	for _, p := range pkgs {
		bindings = append(bindings, p.ReadBindings(discard)...)
		constructs = append(constructs, p.Infer()...)
		schema = append(schema, p.ReadSchema()...)
		scope[p.PkgPath()] = true
	}
	check.SortSchema(schema)
	status := check.Proposals(constructs, bindings, discard)

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
	if len(added) == 0 && len(updated) == 0 {
		fmt.Fprintln(os.Stderr, "nothing to record; every frozen shape is already promised.")
		return nil
	}

	for _, name := range added {
		fmt.Fprintln(os.Stderr, "promise  "+name)
	}
	for _, name := range updated {
		fmt.Fprintln(os.Stderr, "extend   "+name)
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
func merge(base *baseline.File, schema []ir.SchemaType, status map[string]check.Freeze) (added, updated []string) {
	for _, t := range schema {
		if f, ok := status[t.Name]; ok && f.Proposal {
			continue
		}
		entry := baseline.EntryOf(t)
		old, existed := base.Types[t.Name]
		switch {
		case !existed:
			added = append(added, t.Name)
		case len(old.Fields) != len(entry.Fields):
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
