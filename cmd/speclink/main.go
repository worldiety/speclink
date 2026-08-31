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
	"strings"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang"
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
	case "init":
		return initialise(args[1:])
	case "verify":
		return verify(args[1:])
	case "requirements":
		return requirements(args[1:])
	case "freeze":
		return freeze(args[1:])
	case "inventory":
		return inventory(args[1:])
	case "impact":
		return impact(args[1:])
	case "evidence":
		return evidence(args[1:])
	case "generate":
		return generate(args[1:])
	case "diagrams":
		return diagrams(args[1:])
	case "attest":
		return attest(args[1:])
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
  speclink init         [flags]
  speclink verify       [flags] [packages]
  speclink requirements [flags] [packages]
  speclink freeze       [flags] [packages]
  speclink inventory    [flags] [packages]
  speclink impact       [flags] <requirement|doc.md#anchor|path>...
  speclink evidence     [flags] [packages]
  speclink generate     [flags] [packages]
  speclink diagrams     [flags] [packages]
  speclink attest       [flags] [packages|constructs]

commands:
  init          write a starting point for a new project, from a profile's
                template; -describe lists what there is
  verify        check requirements, annotations and architecture rules
  requirements  check the requirement tree on its own, before any code binds to it
  freeze        record the shape of every persisted type that is no longer a draft
  inventory     list what the recognisers found, with kind, name and binding
  impact        report what a change to a requirement, a source segment or a
                file reaches
  evidence      record which tests demonstrated which requirements, reading
                the output of "go test -json"
  generate      derive the specification document from the source
  diagrams      write the PlantUML sources of the context, the building blocks
                and every process; renders nothing itself
  attest        record who wrote a declaration and who has read it

run "speclink <command> -h" for the flags of a command.
`)
}

func verify(args []string) error {
	fs := flag.NewFlagSet("verify", flag.ExitOnError)
	format := fs.String("format", "text", "output format: text or json")
	root := fs.String("root", ".", "repository root, used to resolve source documents")
	cfgPath := fs.String("config", "", "layout configuration; defaults to "+config.FileName+" in the root")
	prof := fs.String("profile", "", "language, framework and architectural style; overrides "+config.FileName)
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := absRootOf(*root)
	if err != nil {
		return err
	}

	// verify is the only command that asks for test variants, because only K14
	// asks a question about tests. It roughly doubles the load, so nothing else
	// pays for it.
	//
	// The explicit config path exists so a project can be measured without
	// being modified. Trying speclink on an unfamiliar codebase should not
	// require a commit to it.
	model, layout, p, err := open(absRoot, *cfgPath, *prof, fs.Args(), true)
	if err != nil {
		return err
	}

	findings := &diag.Set{}

	// V1 and the frontend's own rules: a syntax whitelist, an architecture, a
	// ban on constructs it cannot analyse. They stay with the frontend because
	// they are rules about a language and a framework, and another frontend
	// replaces them wholesale rather than picking from them.
	if c, ok := model.(lang.SyntaxChecker); ok {
		c.CheckSyntax(findings)
	}
	if c, ok := model.(lang.ArchitectureChecker); ok {
		c.CheckArchitecture(findings)
	}

	// V3: read the model. Declarations first, then assertions, so forward
	// references are legal and the input order is irrelevant.
	reqs := model.Requirements(findings)
	bindings := model.Bindings(findings)
	waived := ir.CollectWaivers(bindings)

	var constructs []ir.Construct
	if inf, ok := model.(lang.ConstructInferrer); ok {
		constructs = inf.Constructs(findings)
	}

	var (
		schema []ir.SchemaType
		scope  map[string]bool
	)
	if sr, ok := model.(lang.SchemaReader); ok {
		schema = sr.Schemas(findings)
		scope = sr.Scope()
		check.SortSchema(schema)
	}

	// V4: reject annotations that state a fact already established elsewhere.
	status := check.Drafts(schema, bindings, model.Dialect(), findings)

	// V5: resolve identity, layout, the derivation graph and the outer edge.
	tree := reqtree.Build(absRoot, reqs, findings)
	tree.CheckLayout(model.Dialect(), findings)
	docs, sourceDocs := loadSources(absRoot, layout, findings)
	// The topology is read here rather than in V6, because a channel may be the
	// only thing filed under a theme and the check for an empty theme has to
	// see it. Reading it twice would be the alternative, and two reads of one
	// declaration is how the two copies start to disagree.
	var topo ir.Topology
	if tr, ok := model.(lang.TopologyReader); ok {
		topo = tr.Topology(findings)
	}
	if tr, ok := model.(lang.TopicReader); ok {
		tree.ResolveTopics(tr.Topics(), findings)
		tree.CheckTopicsUsed(topicIdents(topo), findings)
	}
	if cr, ok := model.(lang.ChapterReader); ok {
		tree.ResolveChapters(cr.Chapters(findings), findings)
	}
	tree.CheckSources(docs, findings)

	// V6: the specification rules, in every direction the frontend can measure.
	measured := measuredRequirements(tree, layout, absRoot)
	// Likewise: a frontend that infers nothing has no set of constructs to hold
	// accountable, so "every construct names a requirement" is not a weaker
	// claim but no claim at all.
	var str check.Structure
	if _, ok := model.(lang.ConstructInferrer); ok {
		str = check.CoverConstructs(constructs, bindings, model.Dialect(), findings)
	}
	// A process satisfies requirements the way a construct does, so it has to
	// be in hand before coverage is computed: a requirement about the course of
	// business would otherwise read as covered by nothing.
	var processes []*ir.Process
	if pr, ok := model.(lang.ProcessReader); ok {
		processes = pr.Processes(findings)
	}
	satisfiers := bindings
	for _, p := range processes {
		satisfiers = append(satisfiers, p.Binding())
	}
	for _, c := range topo.Channels {
		satisfiers = append(satisfiers, c.Binding())
	}

	cov := check.CoverRequirements(tree, satisfiers, measured, model.Dialect(), findings)

	// Only asked when the frontend can answer. Running it over an empty set
	// would report every requirement as unverified, which is a different claim
	// from "nobody asked" and the more damaging of the two: it fails a project
	// for not doing something the tool never looked for.
	var ver check.Verification
	if vr, ok := model.(lang.VerificationReader); ok {
		ver = check.CoverVerification(tree, vr.Verifications(findings), cov, measured, waived, model.Dialect(), findings)
	}

	src := check.CoverSources(tree, docs, sourceDocs, waived, findings)
	reqtree.ReportDocuments(docs, findings)

	// V6: what has already been promised must still hold. A shape that outlives
	// the code declaring it cannot be checked against the code alone, so this
	// is the one rule family that reads a record of the past.
	base, err := baseline.Load(absRoot)
	if err != nil {
		return err
	}
	check.Evolution(schema, status, base, scope, bindings, model.Dialect(), findings)
	check.Discriminators(schema, bindings, findings)
	check.Drift(tree, docs, sourceDocs, cov, src, base, waived, findings)
	if _, ok := model.(lang.VerificationReader); ok {
		ver.Shown = check.Demonstrated(tree, ver, cov, measured, base, waived, findings)
	}

	// K1: why the data is shaped the way it is, and forward coverage down to
	// the field. Both need to know which packages hold the domain, because a
	// field of a request object states what a caller sent rather than what the
	// system believes.
	if ds, ok := model.(lang.DomainScoper); ok {
		domain := ds.DomainPackages()
		check.JustifyPersistence(tree, constructs, bindings, domain, model.Dialect(), findings)
		check.CoverFields(schema, constructs, bindings, domain, model.Dialect(), findings)
	}

	// K16: the course of business. It needs both halves — the declarations and
	// the constructs they name — so it is asked only where both are available.
	// K17: what surrounds the code, and where it reaches out.
	tp := check.Topology(tree, topo, bindings, model.Dialect(), findings)
	// A restriction is about the values a type may hold, which is a promise to
	// whoever is on the other end of the wire. Checked here rather than beside
	// the schema rules because it holds whether or not the type is persisted:
	// the far end of a channel is bound by it just the same.
	check.Restrictions(bindings, findings)
	// The protocol a channel carries, where it carries one. Separate from the
	// topology rules because those are about the boundary and these are about
	// what travels along it.
	check.Messages(tree, topo, findings)

	// K17 again, in the other direction: the shapes this system depends on
	// receiving. The rules above hold a channel to describing itself; this
	// holds it to the structure it was recorded as carrying.
	check.ContractEvolution(topo.Channels, base, findings)

	var proc check.ProcessReport
	if _, inf := model.(lang.ConstructInferrer); inf {
		proc = check.Processes(tree, processes, constructs, bindings, scope, model.Dialect(), findings)
	}

	// K20: what the system answers on. Asked only where the frontend can
	// recognise a registration: a run that cannot see routes must report no
	// routes, never an empty surface, because "there are none" and "nobody
	// looked" are the two answers this whole tool exists to keep apart.
	var eps check.EndpointReport
	if er, ok := model.(lang.EndpointReader); ok {
		eps = check.Endpoints(er.Endpoints(), findings)
		check.EndpointEvolution(eps.Endpoints, base, scope, bindings, findings)
	}

	// K18: who wrote each declaration and who has read it. A record of what
	// happened, so it reads the lock rather than the code.
	var prov check.ProvenanceReport
	if _, ok := model.(lang.ConstructInferrer); ok {
		prov = check.Provenance(constructs, base, findings)
	}

	// K15: which states an aggregate can be in. It rides on the same set of
	// constructs as forward coverage and is therefore asked under the same
	// condition — a frontend that infers nothing has no events to hold to it.
	if _, ok := model.(lang.ConstructInferrer); ok {
		check.Lifecycle(constructs, bindings, model.Dialect(), findings)
	}

	if err := report(*format, findings, lang.Of(model), cov, str, src, ver, proc, tp, prov, eps, len(bindings), skippedPackages(model)); err != nil {
		return err
	}
	if *format == "text" {
		reportCapabilities(model, p)
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
	prof := fs.String("profile", "", "language, framework and architectural style; overrides "+config.FileName)
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := absRootOf(*root)
	if err != nil {
		return err
	}

	// This is the command that reads the tree and nothing else, which is
	// exactly the overlap between the frontends — so it is the one that speaks
	// to them through the interface rather than to one of them by name. Only
	// the named packages have to compile, which is the point of narrowing the
	// patterns: the tree can be checked while the implementation around it is
	// still in pieces.
	//
	// It is the one command allowed to narrow the load, and it is allowed
	// because it makes no claim about the module. Everything below is answered
	// by the files it read.
	model, layout, _, err := openNamed(absRoot, *cfgPath, *prof, fs.Args(), false)
	if err != nil {
		return err
	}

	findings := &diag.Set{}
	if c, ok := model.(lang.SyntaxChecker); ok {
		// Only the syntax. A requirement file is written in the same closed
		// subset as an annotation file, and the whitelist is what keeps it
		// readable without evaluating it — but an architecture is a statement
		// about a whole project, and this command deliberately loaded a part.
		c.CheckSyntax(findings)
	}

	reqs := model.Requirements(findings)
	tree := reqtree.Build(absRoot, reqs, findings)
	tree.CheckLayout(model.Dialect(), findings)
	docs, sourceDocs := loadSources(absRoot, layout, findings)
	// The topology is read here rather than in V6, because a channel may be the
	// only thing filed under a theme and the check for an empty theme has to
	// see it. Reading it twice would be the alternative, and two reads of one
	// declaration is how the two copies start to disagree.
	var topo ir.Topology
	if tr, ok := model.(lang.TopologyReader); ok {
		topo = tr.Topology(findings)
	}
	if tr, ok := model.(lang.TopicReader); ok {
		tree.ResolveTopics(tr.Topics(), findings)
		tree.CheckTopicsUsed(topicIdents(topo), findings)
	}
	if cr, ok := model.(lang.ChapterReader); ok {
		tree.ResolveChapters(cr.Chapters(findings), findings)
	}
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

	// No capability lines here, and their absence is not an oversight. They
	// exist to stop a summary about a system from reading as complete when a
	// direction of it was never measured, and this command's summary is about
	// the tree: so many requirements, so many segments accounted for. It puts
	// no question to the code, so it has none to disclaim — and saying "this
	// frontend infers no constructs" would be false anyway, because it does,
	// and was simply not asked.
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
// measuredRequirements returns the requirement IDs the scope asks anything of,
// nil when the scope is the whole module.
//
// nil rather than a full set on purpose: the unrestricted case is the normal
// one, and a map built for it would be a lookup on every requirement to answer
// a question nobody asked.
func measuredRequirements(tree *reqtree.Tree, layout config.Config, root string) map[string]bool {
	if !layout.Restricted() {
		return nil
	}
	out := map[string]bool{}
	for _, r := range tree.All() {
		dir := filepath.Dir(r.Pos.File)
		if rel, err := filepath.Rel(root, dir); err == nil {
			dir = rel
		}
		if layout.InScope(filepath.ToSlash(dir)) {
			out[r.ID] = true
		}
	}
	return out
}

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
func report(format string, findings *diag.Set, can lang.Capabilities, cov check.Coverage, str check.Structure, src check.SourceCoverage, ver check.Verification, proc check.ProcessReport, tp check.TopologyReport, prov check.ProvenanceReport, eps check.EndpointReport, bindings, skipped int) error {
	switch format {
	case "json":
		return findings.WriteJSON(os.Stdout)
	case "text":
		if err := findings.WriteText(os.Stdout); err != nil {
			return err
		}
		// Only the directions that were measured. A figure for a question that
		// was never put is worse than a missing one, and the capability lines
		// below say which those were.
		parts := []string{fmt.Sprintf("%d source segments (%.0f%% accounted)", src.Total, src.Ratio()*100)}
		if can.Constructs {
			parts = append(parts, fmt.Sprintf("%d constructs (%.0f%% bound)", len(str.Constructs), str.Ratio()*100))
		}
		// A share of nothing is not a hundred percent, here as much as for the
		// processes below. A scope that reaches no requirement used to report
		// "0 normative requirements (100% covered, 100% verified, 100%
		// demonstrated)", which is arithmetically true of an empty set and
		// reads as a clean bill of health for a tree the run never asked
		// about. It says what it left out instead.
		var requirements string
		switch {
		case cov.Normative == 0 && cov.Excluded > 0:
			requirements = fmt.Sprintf("no normative requirement in this scope (%d outside it)", cov.Excluded)
		case cov.Normative == 0:
			requirements = "no normative requirements"
		default:
			requirements = fmt.Sprintf("%d normative requirements (%.0f%% covered", cov.Normative, cov.Ratio()*100)
			if can.Verifications {
				requirements += fmt.Sprintf(", %.0f%% verified, %.0f%% demonstrated", ver.Ratio()*100, ver.ShownRatio()*100)
			}
			if cov.Excluded > 0 {
				requirements += fmt.Sprintf(", %d outside this scope", cov.Excluded)
			}
			requirements += ")"
		}
		parts = append(parts, requirements)
		// Only where processes exist. A share of nothing is not a hundred
		// percent, it is no claim at all, and printing one would tell a
		// project that never adopted them that its courses of business are
		// accounted for.
		if proc.Declared > 0 {
			parts = append(parts, fmt.Sprintf("%s (%d sound, %d of %d steps placed)",
				plural(proc.Declared, "process", "processes"), proc.Sound, proc.Placed, proc.Work))
		}
		if prov.Total > 0 {
			line := fmt.Sprintf("%s (%d machine written, %d read by a person",
				plural(prov.Total, "declaration", "declarations"), prov.Machine, prov.Reviewed)
			if prov.Statements > 0 {
				line += fmt.Sprintf(", %.0f%% of statements exercised",
					float64(prov.Executed)/float64(prov.Statements)*100)
			}
			parts = append(parts, line+")")
		}
		// Both halves, always. A surface where every route is traced still
		// prints the total, because "12 of 12" and a silent line are different
		// claims: the first says somebody looked.
		//
		// The third half appears only when there is one, and it is the one
		// that keeps the other two honest: a run narrowed to a package cannot
		// see what the rest of the module mounts, and a figure that hid those
		// routes among the traced ones would turn a limit of the run into a
		// claim about the code.
		if can.Endpoints {
			// Printed even at zero, because this is one of the few figures
			// where nothing is a finding in itself: a module that answers on
			// no address is a library, and saying so out loud distinguishes it
			// from a service whose surface this failed to recognise. Where the
			// frontend cannot look at all, the capability line above says that
			// instead, and the two must never be confused.
			line := plural(eps.Routes, "route", "routes")
			if eps.Routes > 0 {
				line += fmt.Sprintf(" (%d traced to a use case", eps.Traced)
				if eps.Unmeasured > 0 {
					line += fmt.Sprintf(", %d outside this scope", eps.Unmeasured)
				}
				line += ")"
			}
			parts = append(parts, line)
		}
		if tp.Declared {
			parts = append(parts, fmt.Sprintf("%s (%d of %d boundaries described)",
				plural(tp.Channels, "channel", "channels"), tp.Described, tp.Adapters))
		}
		parts = append(parts,
			fmt.Sprintf("%d bindings", bindings),
			plural(findings.Len(), "finding", "findings"))

		fmt.Fprintf(os.Stderr, "\n%s\n", strings.Join(parts, ", "))

		// One line per standard, rather than another clause in the summary.
		// A title carries commas of its own, and a figure buried in a run of
		// them is a figure nobody reads.
		for _, st := range src.Standards {
			fmt.Fprintf(os.Stderr, "%s: %d of %s answered", st.Title, st.Answered,
				plural(st.Clauses, "applicable clause", "applicable clauses"))
			if st.Excluded > 0 {
				fmt.Fprintf(os.Stderr, ", %d not applicable", st.Excluded)
			}
			fmt.Fprintln(os.Stderr)
		}
		if skipped > 0 {
			// A run that measures part of a project has to say so. Without it
			// the figures above are true of what was looked at and silent
			// about what was not, which is the one way this tool could mislead
			// by telling the truth.
			fmt.Fprintf(os.Stderr, "%s outside the configured scope and not measured\n",
				plural(skipped, "package", "packages"))
		}
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

// topicIdents collects every theme named by the edge of the system.
func topicIdents(t ir.Topology) []string {
	var out []string
	for _, p := range t.Participants {
		out = append(out, p.Topics...)
	}
	for _, c := range t.Channels {
		out = append(out, c.Topics...)
	}
	return out
}
