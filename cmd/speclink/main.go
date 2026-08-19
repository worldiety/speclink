// Command speclink verifies that implementation and requirements agree, and
// derives documentation from that single source.
//
// Build order matters and is not negotiable: the Go compiler runs first,
// speclink second, tests third. Binding presupposes compilable source, so when
// the Go build is broken there is no annotation feedback at all — a loop runner
// consuming the JSON output has to prioritise accordingly.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/golang"
	"github.com/worldiety/speclink/internal/reqtree"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		if !errors.Is(err, errFindings) {
			fmt.Fprintln(os.Stderr, "speclink: "+err.Error())
		}
		os.Exit(1)
	}
}

// errFindings signals a failed verification. The findings themselves have
// already been printed, so main must not print the error again.
var errFindings = errors.New("verification failed")

func run(args []string) error {
	if len(args) == 0 {
		usage()
		return errors.New("no command given")
	}
	switch args[0] {
	case "verify":
		return verify(args[1:])
	case "requirements":
		return requirements(args[1:])
	case "help", "-h", "--help":
		usage()
		return nil
	default:
		usage()
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `speclink - annotation compiler

usage:
  speclink verify       [flags] [packages]
  speclink requirements [flags] [packages]

commands:
  verify        check requirements, annotations and architecture rules
  requirements  check the requirement tree on its own, before any code binds to it

run "speclink <command> -h" for the flags of a command.
`)
}

func verify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	format := fs.String("format", "text", "output format: text or json")
	root := fs.String("root", ".", "repository root, used to resolve source documents")
	cfgPath := fs.String("config", "", "layout configuration; defaults to "+config.FileName+" in the root")
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}

	// The layout is the only project knowledge speclink accepts. Without a
	// speclink.json the convention applies.
	//
	// The explicit path exists so a project can be measured without being
	// modified. Trying speclink on an unfamiliar codebase should not require a
	// commit to it, and the first run is exactly where the layout is least
	// likely to match the convention.
	layout, err := loadLayout(absRoot, *cfgPath)
	if err != nil {
		return err
	}

	pkgs, err := golang.Load(absRoot, fs.Args()...)
	if err != nil {
		return err
	}

	// Phase V2 is the Go compilation itself. If it failed there is nothing
	// meaningful to say about annotations, and saying it anyway would bury the
	// real cause under follow-up noise.
	if errs := golang.TypeErrors(pkgs); len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "the Go build is broken; fix it before speclink can check anything:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "  "+e.Error())
		}
		return errFindings
	}

	findings := &diag.Set{}

	// V1: the node whitelist, plus orphaned annotation files.
	for _, p := range pkgs {
		p.CheckWhitelist(findings)
		p.CheckOrphans(findings)
	}

	// V3: read the model. Declarations first, then assertions, so forward
	// references are legal and the input order is irrelevant.
	var (
		reqs       []*ir.Requirement
		bindings   []ir.Binding
		constructs []ir.Construct
	)
	for _, p := range pkgs {
		reqs = append(reqs, p.ReadRequirements(findings)...)
	}
	for _, p := range pkgs {
		bindings = append(bindings, p.ReadBindings(findings)...)
		constructs = append(constructs, p.Infer()...)
	}

	// V5: resolve identity, layout, the derivation graph and the outer edge.
	tree := reqtree.Build(absRoot, reqs, findings)
	tree.CheckLayout(findings)
	tree.CheckSources(findings)

	// V6: the specification rules, in both directions of the coverage.
	for _, p := range pkgs {
		p.CheckGenericCRUD(findings)
	}
	str := check.CoverConstructs(constructs, bindings, findings)
	cov := check.CoverRequirements(tree, bindings, findings)

	// V6: the architecture rules. They read the project layout, which is the
	// one thing speclink cannot infer and the only thing speclink.json states.
	golang.CheckUseCases(pkgs, layout, absRoot, findings)
	golang.CheckBoundedContexts(pkgs, layout, absRoot, findings)
	golang.CheckInfrastructure(pkgs, layout, absRoot, findings)
	golang.CheckMainPackages(pkgs, layout, absRoot, findings)

	if err := report(*format, findings, cov, str, len(bindings)); err != nil {
		return err
	}
	if !findings.Empty() {
		return errFindings
	}
	return nil
}

// requirements checks the requirement tree on its own: identity, the derivation
// graph, the layout and the outer edge to the source documents.
//
// It exists for the transition. Building a requirement tree for an existing
// system is a long piece of work, and until the last requirement is in place
// there is no point asking whether the code covers it — `verify` would drown
// the tree's own defects under one finding per unbound construct, and the tree
// is what has to be right first.
//
// It is therefore not a reduced verify but a different question: is this tree
// sound in itself? Nothing here reads an annotation, infers a construct or
// measures coverage in either direction, so the tree can be grown in a package
// that no implementation references yet.
func requirements(args []string) error {
	fs := flag.NewFlagSet("requirements", flag.ExitOnError)
	format := fs.String("format", "text", "output format: text or json")
	root := fs.String("root", ".", "repository root, used to resolve source documents")
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}

	patterns := fs.Args()
	pkgs, err := golang.Load(absRoot, patterns...)
	if err != nil {
		return err
	}

	// Only the loaded packages have to compile. That is the point of narrowing
	// the patterns: the tree can be checked while the implementation around it
	// is still in pieces.
	if errs := golang.TypeErrors(pkgs); len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "the Go build is broken; fix it before speclink can check anything:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "  "+e.Error())
		}
		return errFindings
	}

	findings := &diag.Set{}

	// V1 still applies. A requirement file is written in the same closed subset
	// as an annotation file, and the whitelist is what keeps it readable
	// without evaluating it.
	for _, p := range pkgs {
		p.CheckWhitelist(findings)
	}

	var reqs []*ir.Requirement
	for _, p := range pkgs {
		reqs = append(reqs, p.ReadRequirements(findings)...)
	}

	tree := reqtree.Build(absRoot, reqs, findings)
	tree.CheckLayout(findings)
	tree.CheckSources(findings)

	if err := reportRequirements(*format, findings, tree); err != nil {
		return err
	}
	if !findings.Empty() {
		return errFindings
	}
	return nil
}

// reportRequirements writes the tree summary. It counts by status because that
// is the number a migration is steered by: only normative requirements will
// later have to be covered, and everything else is an explicit, justified
// exemption.
func reportRequirements(format string, findings *diag.Set, tree *reqtree.Tree) error {
	switch format {
	case "json":
		return findings.WriteJSON(os.Stdout)
	case "text":
		if err := findings.WriteText(os.Stdout); err != nil {
			return err
		}
		all := tree.All()
		normative := 0
		for _, r := range all {
			if r.Status == ir.Normative {
				normative++
			}
		}
		fmt.Fprintf(os.Stderr, "\n%s (%d normative), %s\n",
			plural(len(all), "requirement", "requirements"), normative,
			plural(findings.Len(), "finding", "findings"))
		return nil
	default:
		return fmt.Errorf("unknown format %q, expected text or json", format)
	}
}

// loadLayout resolves the layout configuration, from an explicit path when one
// was given and from the conventional location otherwise.
func loadLayout(absRoot, explicit string) (config.Config, error) {
	if explicit == "" {
		return config.Load(absRoot)
	}
	return config.LoadFile(explicit, false)
}

func report(format string, findings *diag.Set, cov check.Coverage, str check.Structure, bindings int) error {
	switch format {
	case "json":
		return findings.WriteJSON(os.Stdout)
	case "text":
		if err := findings.WriteText(os.Stdout); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr,
			"\n%d constructs (%.0f%% bound), %d normative requirements (%.0f%% covered), %d bindings, %s\n",
			len(str.Constructs), str.Ratio()*100,
			cov.Normative, cov.Ratio()*100,
			bindings, plural(findings.Len(), "finding", "findings"))
		return nil
	default:
		return fmt.Errorf("unknown format %q, expected text or json", format)
	}
}

func plural(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}
