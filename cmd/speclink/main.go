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

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/golang"
	"github.com/worldiety/speclink/internal/reqtree"
	"github.com/worldiety/speclink/internal/source"
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
	case "freeze":
		return freeze(args[1:])
	case "inventory":
		return inventory(args[1:])
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
  speclink freeze       [flags] [packages]
  speclink inventory    [flags] [packages]

commands:
  verify        check requirements, annotations and architecture rules
  requirements  check the requirement tree on its own, before any code binds to it
  freeze        record the shape of every persisted type that is no longer a draft
  inventory     list what the recognisers found, with kind, name and binding

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
		schema     []ir.SchemaType
		scope      = map[string]bool{}
	)
	for _, p := range pkgs {
		reqs = append(reqs, p.ReadRequirements(findings)...)
	}
	models := map[string]bool{}
	for _, p := range pkgs {
		bindings = append(bindings, p.ReadBindings(findings)...)
		constructs = append(constructs, p.Infer()...)
		scope[p.PkgPath()] = true
		for name := range p.PersistedModels() {
			models[name] = true
		}
	}
	// The persisted set has to be complete before any shape is read: a
	// repository is usually built in the wiring package, far from the type it
	// stores.
	for _, p := range pkgs {
		schema = append(schema, p.ReadSchema(models)...)
	}
	check.SortSchema(schema)

	// V4: reject annotations that state a fact already established elsewhere.
	// The freeze status is the first thing the language can say twice, because
	// it cascades: package, then type, then field.
	status := check.Drafts(schema, bindings, findings)

	// V5: resolve identity, layout, the derivation graph and the outer edge.
	tree := reqtree.Build(absRoot, reqs, findings)
	tree.CheckLayout(findings)
	docs, sourceDocs := loadSources(absRoot, layout, findings)
	tree.CheckSources(docs, findings)

	// V6: the specification rules, in both directions of the coverage.
	for _, p := range pkgs {
		p.CheckGenericCRUD(findings)
	}
	str := check.CoverConstructs(constructs, bindings, findings)
	cov := check.CoverRequirements(tree, bindings, findings)

	// V6: and the direction above the tree. Everything below it is already
	// held by the Go compiler; this is the only step in the chain that has no
	// formal semantics, and it decides whether the two figures above mean
	// anything at all.
	src := check.CoverSources(tree, docs, sourceDocs, ir.CollectWaivers(bindings), findings)
	reqtree.ReportDocuments(docs, findings)

	// V6: what has already been promised must still hold. A shape that outlives
	// the code declaring it cannot be checked against the code alone, so this
	// is the one rule family that reads a record of the past.
	base, err := baseline.Load(absRoot)
	if err != nil {
		return err
	}
	check.Evolution(schema, status, base, scope, bindings, findings)

	// The same rule family for the two edges above the code: a requirement
	// whose text was rewritten under its satisfiers, and a source segment
	// rewritten under the requirements derived from it. Neither is visible to
	// any other check, because in both cases every reference still resolves.
	check.Drift(tree, docs, sourceDocs, cov, src, base, ir.CollectWaivers(bindings), findings)

	// A collision is not a broken promise but a corruption in progress, so it
	// is checked for drafts too.
	check.Discriminators(schema, bindings, findings)

	// K1: why the data is shaped the way it is, which forward coverage does
	// not ask because aggregates and repositories carry no requirement of
	// their own.
	domain := golang.DomainPackages(pkgs, layout, absRoot)
	check.JustifyPersistence(tree, constructs, bindings, domain, findings)

	// K1: forward coverage down to the field. Types are reviewed when they are
	// created; fields accrete afterwards, which is where the drift is.
	check.CoverFields(schema, constructs, bindings, domain, findings)

	// V6: the architecture rules. They read the project layout, which is the
	// one thing speclink cannot infer and the only thing speclink.json states.
	golang.CheckUseCases(pkgs, layout, absRoot, ir.CollectWaivers(bindings), findings)
	golang.CheckBoundedContexts(pkgs, layout, absRoot, findings)
	golang.CheckInfrastructure(pkgs, layout, absRoot, findings)
	golang.CheckMainPackages(pkgs, layout, absRoot, findings)

	if err := report(*format, findings, cov, str, src, len(bindings)); err != nil {
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
	cfgPath := fs.String("config", "", "layout configuration; defaults to "+config.FileName+" in the root")
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := filepath.Abs(*root)
	if err != nil {
		return fmt.Errorf("resolve root: %w", err)
	}

	// The tree question now reaches above the tree, so this command needs the
	// layout too: it is what says where the raw documents live.
	layout, err := loadLayout(absRoot, *cfgPath)
	if err != nil {
		return err
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
	docs, sourceDocs := loadSources(absRoot, layout, findings)
	tree.CheckSources(docs, findings)
	src := check.CoverSources(tree, docs, sourceDocs, nil, findings)
	reqtree.ReportDocuments(docs, findings)

	// The tree question now also has an audience that is not an agent, so the
	// drift and review state have to be available here. requirements is the
	// command a collaboration surface runs: it needs no code to compile beyond
	// the tree itself, which is the point of it existing.
	base, err := baseline.Load(absRoot)
	if err != nil {
		return err
	}
	check.Drift(tree, docs, sourceDocs, check.Coverage{BySatisfier: map[string][]ir.Target{}}, src, base, nil, findings)

	if err := reportRequirements(*format, absRoot, findings, tree, docs, sourceDocs, src, base); err != nil {
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
// loadSources enumerates the raw requirement documents and reports the defects
// of the enumeration itself.
//
// Enumeration rather than collection from citations is the point: a document
// nobody cited would otherwise contribute nothing and look exactly like one
// that is fully covered.
func loadSources(absRoot string, layout config.Config, findings *diag.Set) (*source.Set, []string) {
	docs := source.NewSet(absRoot)
	found, errs := source.Walk(absRoot, layout.SourceRoots)
	for _, err := range errs {
		se, ok := err.(*source.SegmentError)
		if !ok {
			continue
		}
		// A defect of the enumeration is not a defect of a document, so it
		// does not share the code. A source root that is not there and a cited
		// file that is not there fail for different reasons and are fixed in
		// different places.
		findings.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 27),
			Pos:  ir.Position{File: se.Doc, Line: se.Line},
			What: se.Msg + ".",
			Why:  se.Why,
			How:  se.How,
		})
	}
	return docs, found
}

func reportRequirements(format string, root string, findings *diag.Set, tree *reqtree.Tree, docs *source.Set, sourceDocs []string, src check.SourceCoverage, base *baseline.File) error {
	switch format {
	case "json":
		// The tree itself, not a findings list. This command asks whether the
		// tree is sound, so its machine readable answer is the tree; findings
		// are one field of it. The audience is also different from everywhere
		// else in speclink — a person who never sees Go, working through the
		// requirements a model extracted from documents they wrote — and a
		// list of what is broken says nothing about what is there.
		return writeTree(os.Stdout, root, tree, docs, sourceDocs, src, nil, base, findings)
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
		reviewed := 0
		for _, r := range all {
			rec := base.Requirements[r.ID]
			if rec.ReviewedBy != "" && rec.Text == baseline.HashText(r.Text, r.Title) {
				reviewed++
			}
		}
		fmt.Fprintf(os.Stderr, "\n%s (%d normative, %d reviewed), %d source segments (%.0f%% accounted), %s\n",
			plural(len(all), "requirement", "requirements"), normative, reviewed,
			src.Total, src.Ratio()*100,
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

// report writes the summary.
//
// Three figures, three directions. Bound asks whether every construct that
// carries business meaning names a requirement. Covered asks whether every
// normative requirement is satisfied by a construct. Accounted asks the
// question above both: whether every part of what was actually asked for
// became a requirement at all. Without the third the other two measure a tree
// against itself.
func report(format string, findings *diag.Set, cov check.Coverage, str check.Structure, src check.SourceCoverage, bindings int) error {
	switch format {
	case "json":
		return findings.WriteJSON(os.Stdout)
	case "text":
		if err := findings.WriteText(os.Stdout); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr,
			"\n%d source segments (%.0f%% accounted), %d constructs (%.0f%% bound), %d normative requirements (%.0f%% covered), %d bindings, %s\n",
			src.Total, src.Ratio()*100,
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
