package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang"
)

// attest records who wrote a declaration and who has read it.
//
// # Why a command and not an annotation
//
// The code could say who wrote it, and it would be worthless: the same machine
// that writes the code writes the annotation. A claim of human authorship made
// by the author is not evidence of anything, which is the reason
// spec.Requirement has no Reviewed field either.
//
// So it is recorded from outside, exactly as freeze -reviewer already is. The
// harness that ran the generator knows it ran. The person who read a use case
// knows they read it. speclink writes down what it is told.
//
// # What this cannot do
//
// It cannot check any of it. If whatever drives the generator may also call
// -reviewer, the record is worth nothing at all. What keeps it honest is who is
// permitted to make which call, and that is not speclink's to decide — saying
// so plainly is better than a mechanism that implies a guarantee it has not got.
//
// # Why it takes targets
//
// A reviewer is usually specialised and reads a few declarations at a time.
// Recording a whole run as read because somebody looked at one use case is the
// fastest way to make the figure meaningless.
func attest(args []string) error {
	fs := flag.NewFlagSet("attest", flag.ExitOnError)
	var (
		root     = fs.String("root", ".", "repository root")
		cfgPath  = fs.String("config", "", "layout configuration; defaults to speclink.json in the root")
		prof     = fs.String("profile", "", "overrides the profile from speclink.json")
		origin   = fs.String("origin", "", "who wrote these declarations: llm or human")
		reviewer = fs.String("reviewer", "", "who has read them")
		dry      = fs.Bool("n", false, "report what would be recorded, write nothing")
	)
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *origin == "" && *reviewer == "" {
		return errors.New("nothing to record; pass -origin llm|human, -reviewer <who>, or both")
	}
	if *origin != "" && *origin != "llm" && *origin != "human" {
		return fmt.Errorf("unknown origin %q, expected llm or human", *origin)
	}

	absRoot, err := absRootOf(*root)
	if err != nil {
		return err
	}

	// The last argument may name a construct rather than a package pattern,
	// because that is how a specialised reviewer works: one declaration, by
	// name, without having to know which package it sits in.
	patterns, named := splitTargets(fs.Args())

	model, _, _, err := open(absRoot, *cfgPath, *prof, patterns, false)
	if err != nil {
		return err
	}
	inferrer, ok := model.(lang.ConstructInferrer)
	if !ok {
		return errors.New("this frontend infers no constructs, so there is nothing to attest")
	}

	quiet := &diag.Set{}
	constructs := selectConstructs(inferrer.Constructs(quiet), named)
	if len(constructs) == 0 {
		return errors.New("no construct matched. Name a package pattern such as ./app/sales/..., or a construct such as example.com/erp/app/sales.SubmitQuote")
	}

	base, err := baseline.Load(absRoot)
	if err != nil {
		return err
	}

	var changed []string
	for _, c := range constructs {
		if c.Fingerprint == "" {
			// A role with no body has nothing anybody could read. Recording a
			// review of it would be a number with no meaning behind it.
			continue
		}
		rec := base.Constructs[c.Name]
		before := rec

		// Every statement is about the text as it stands now. Carrying an
		// older review across a change is precisely what K18 exists to catch,
		// so it must not be possible to create that state here.
		if rec.Fingerprint != c.Fingerprint {
			rec = baseline.Construct{Fingerprint: c.Fingerprint, Origin: rec.Origin}
			if before.Fingerprint != "" {
				rec.Origin = ""
			}
		}
		if *origin != "" {
			rec.Origin = *origin
		}
		if *reviewer != "" {
			rec.ReviewedBy = *reviewer
		}
		if rec != before {
			base.Constructs[c.Name] = rec
			changed = append(changed, c.Name)
		}
	}

	sort.Strings(changed)
	for _, name := range changed {
		rec := base.Constructs[name]
		fmt.Fprintf(os.Stderr, "%-8s %s%s\n", or(rec.Origin, "recorded"), lastSegment(name), reviewedNote(rec))
	}
	if len(changed) == 0 {
		fmt.Fprintln(os.Stderr, "nothing to record; every declaration already says this.")
		return nil
	}
	if *dry {
		fmt.Fprintln(os.Stderr, "\nnothing written (-n).")
		return nil
	}

	if err := base.Save(absRoot); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "\n%s recorded in %s.\n",
		plural(len(changed), "declaration", "declarations"), baseline.FileName)
	return nil
}

func reviewedNote(rec baseline.Construct) string {
	if rec.ReviewedBy == "" {
		return ""
	}
	return "  read by " + rec.ReviewedBy
}

// splitTargets separates package patterns from construct names.
//
// A pattern is what the go tool understands; anything else is taken as the
// qualified name of a declaration. The two are told apart by the leading ./ or
// the trailing ..., which is what every pattern has and no construct name does.
func splitTargets(args []string) (patterns, named []string) {
	for _, a := range args {
		if strings.HasPrefix(a, "./") || strings.HasSuffix(a, "...") || a == "." {
			patterns = append(patterns, a)
			continue
		}
		named = append(named, a)
	}
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}
	return patterns, named
}

// selectConstructs narrows the set to the named declarations, or keeps all of
// them when only patterns were given.
func selectConstructs(all []ir.Construct, named []string) []ir.Construct {
	if len(named) == 0 {
		return all
	}
	want := map[string]bool{}
	for _, n := range named {
		want[n] = true
	}

	var out []ir.Construct
	for _, c := range all {
		// The short name is accepted too, because a reviewer working through
		// one package should not have to type an import path.
		if want[c.Name] || want[lastSegment(c.Name)] {
			out = append(out, c)
		}
	}
	return out
}
