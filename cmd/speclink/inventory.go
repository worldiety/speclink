package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/golang"
)

// inventory lists what the recognisers found.
//
// verify reports what it objects to, which is the right output for a build
// step and the wrong one for a question like "does the tool see the same system
// the specification describes?". Accepting the inference layer means comparing
// its result against an independent model, and that cannot be done from a list
// of complaints: a construct recognised correctly produces no output at all.
//
// It is deliberately not part of verify. Mixing an inventory into a diagnostic
// stream would make the stream unreadable for the loop that consumes it, and
// the two answer different questions.
func inventory(args []string) error {
	fs := flag.NewFlagSet("inventory", flag.ExitOnError)
	format := fs.String("format", "text", "output format: text or json")
	root := fs.String("root", ".", "repository root")
	kindFilter := fs.String("kind", "", "restrict to one kind, e.g. event or use case")
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
		fmt.Fprintln(os.Stderr, "the Go build is broken; nothing can be recognised until it compiles:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "  "+e.Error())
		}
		return errFindings
	}

	// Bindings are read so the listing can say which constructs already name a
	// requirement. That is the number a migration is steered by, and it is
	// invisible in a list of findings, which shows only the ones that do not.
	var (
		constructs []ir.Construct
		bound      = map[string]bool{}
	)
	discard := &diag.Set{}
	for _, p := range pkgs {
		constructs = append(constructs, p.Infer()...)
		for _, b := range p.ReadBindings(discard) {
			for _, a := range b.Assertions {
				if a.Kind == ir.AssertSatisfies && len(a.Requirements) > 0 {
					bound[b.Target.Name] = true
				}
			}
		}
	}

	if *kindFilter != "" {
		var kept []ir.Construct
		for _, c := range constructs {
			if c.Kind.String() == *kindFilter {
				kept = append(kept, c)
			}
		}
		constructs = kept
	}

	sort.Slice(constructs, func(i, j int) bool {
		if constructs[i].Kind != constructs[j].Kind {
			return constructs[i].Kind.String() < constructs[j].Kind.String()
		}
		return constructs[i].Name < constructs[j].Name
	})

	switch *format {
	case "json":
		return writeInventoryJSON(constructs, bound)
	case "text":
		return writeInventoryText(constructs, bound)
	default:
		return fmt.Errorf("unknown format %q, expected text or json", *format)
	}
}

// inventoryEntry is the machine readable form. It is a separate type rather
// than ir.Construct so that the shape of the output is a decision rather than a
// consequence of an internal model.
type inventoryEntry struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Package  string `json:"package"`
	Evidence string `json:"evidence"`
	Bound    bool   `json:"bound"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

func writeInventoryJSON(constructs []ir.Construct, bound map[string]bool) error {
	out := struct {
		Version    int              `json:"version"`
		Constructs []inventoryEntry `json:"constructs"`
	}{Version: 1, Constructs: make([]inventoryEntry, 0, len(constructs))}

	for _, c := range constructs {
		out.Constructs = append(out.Constructs, inventoryEntry{
			Kind:     c.Kind.String(),
			Name:     c.Name,
			Package:  c.Package,
			Evidence: c.Evidence,
			Bound:    bound[c.Name],
			File:     c.Pos.File,
			Line:     c.Pos.Line,
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func writeInventoryText(constructs []ir.Construct, bound map[string]bool) error {
	for _, c := range constructs {
		mark := " "
		if bound[c.Name] {
			mark = "+"
		}
		if _, err := fmt.Printf("%s %-12s %s\n", mark, c.Kind, c.Name); err != nil {
			return err
		}
	}

	// The summary counts by kind, because that is the shape of the question
	// the acceptance asks: the reference model states how many aggregates,
	// events and use cases it expects, not which.
	counts := map[string][2]int{}
	var kinds []string
	for _, c := range constructs {
		k := c.Kind.String()
		if _, seen := counts[k]; !seen {
			kinds = append(kinds, k)
		}
		n := counts[k]
		n[0]++
		if bound[c.Name] {
			n[1]++
		}
		counts[k] = n
	}
	sort.Strings(kinds)

	fmt.Fprintln(os.Stderr)
	for _, k := range kinds {
		n := counts[k]
		fmt.Fprintf(os.Stderr, "%-12s %4d  %d bound\n", k, n[0], n[1])
	}
	fmt.Fprintf(os.Stderr, "%-12s %4d\n", "total", len(constructs))
	return nil
}
