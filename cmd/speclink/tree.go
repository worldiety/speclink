package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/baseline"
	"github.com/worldiety/speclink/internal/check"
	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/golang"
	"github.com/worldiety/speclink/internal/reqtree"
	"github.com/worldiety/speclink/internal/source"
)

// TreeVersion is the schema version of the requirement export.
//
// It is a contract, not a dump of an internal type, for the same reason the
// inventory export is: the shape of the output should be a decision rather
// than a consequence of a refactoring somewhere in internal/ir.
const TreeVersion = 1

// treeReport is what a caller outside this repository reads.
//
// The audience is a collaboration surface where people who never see Go work on
// the specification. They edit the source documents, read the requirements that
// were extracted from them, and say whether those requirements are right. Every
// one of those steps needs the requirement as data — its text, where it came
// from, what implements it, whether anybody has vouched for it — and until now
// the only machine readable output speclink produced was a list of findings.
//
// A findings list is the wrong shape for that audience. It says what is broken,
// which is what an agent needs, and says nothing about what is there, which is
// what a reviewer needs.
type treeReport struct {
	Version      int                 `json:"version"`
	Requirements []treeRequirement   `json:"requirements"`
	Sources      []treeSourceSegment `json:"sources"`
	Findings     []jsonFindingRef    `json:"findings,omitempty"`
}

// treeRequirement is one requirement, as a reviewer needs to see it.
type treeRequirement struct {
	ID          string   `json:"id"`
	Title       string   `json:"title,omitempty"`
	Text        string   `json:"text,omitempty"`
	Kind        string   `json:"kind,omitempty"`
	Discipline  string   `json:"discipline,omitempty"`
	Status      string   `json:"status,omitempty"`
	Rationale   string   `json:"rationale,omitempty"`
	DerivedFrom []string `json:"derivedFrom,omitempty"`
	Supersedes  []string `json:"supersedes,omitempty"`
	// Sources are the segment references this requirement was written from,
	// in the form "path/to/doc.md#anchor". They are what lets the surface show
	// a requirement next to the paragraph or the part of the screen it came
	// from, which is the only way a reviewer can judge whether it is faithful.
	Sources []string `json:"sources,omitempty"`
	// Extern are the laws and standards cited instead of a document.
	Extern []string `json:"extern,omitempty"`
	// Satisfiers are the constructs implementing it.
	Satisfiers []string `json:"satisfiers,omitempty"`
	// ReviewedBy names who vouched for this exact wording, empty when nobody
	// has. It is the queue the surface works through.
	ReviewedBy string `json:"reviewedBy,omitempty"`
	// Reviewed is false when the current text has never been read by a person.
	Reviewed bool `json:"reviewed"`
	// File and Line locate the declaration, for an agent that has to change it.
	File string `json:"file,omitempty"`
	Line int    `json:"line,omitempty"`
}

// treeSourceSegment is one addressable piece of a source document.
type treeSourceSegment struct {
	Ref   string `json:"ref"`
	Doc   string `json:"doc"`
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Title string `json:"title,omitempty"`
	// Informative marks a segment that is not expected to produce a
	// requirement.
	Informative bool `json:"informative,omitempty"`
	// Requirements are the IDs extracted from this segment. Empty on a
	// non-informative segment is the defect K12-SOURCE-UNCOVERED reports, and
	// the surface can show it as a gap without parsing diagnostics.
	Requirements []string `json:"requirements,omitempty"`
	// Drifted says the segment changed since it was last recorded, so anything
	// derived from it is owed a second look.
	Drifted bool `json:"drifted,omitempty"`
}

// jsonFindingRef is the minimal form of a finding, so the surface can mark a
// requirement as problematic without a second invocation.
type jsonFindingRef struct {
	Code string `json:"code"`
	Rule string `json:"rule,omitempty"`
	File string `json:"file,omitempty"`
	Line int    `json:"line,omitempty"`
	What string `json:"what"`
}

// writeTree renders the requirement tree and its sources.
func writeTree(w io.Writer, root string, tree *reqtree.Tree, set *source.Set, docs []string, cov check.SourceCoverage, sat map[string][]ir.Target, base *baseline.File, findings *diag.Set) error {
	report := treeReport{Version: TreeVersion}

	for _, r := range tree.All() {
		rec := base.Requirements[r.ID]
		out := treeRequirement{
			ID:          r.ID,
			Title:       r.Title,
			Text:        r.Text,
			Kind:        r.Kind.String(),
			Discipline:  r.Discipline.String(),
			Status:      r.Status.String(),
			Rationale:   r.Rationale,
			DerivedFrom: r.DerivedFrom,
			Supersedes:  r.Supersedes,
			ReviewedBy:  rec.ReviewedBy,
			// A review is bound to the wording it was given for. A recorded
			// name against a text that has since changed is not a review of
			// what is there now.
			Reviewed: rec.ReviewedBy != "" && rec.Text == baseline.HashText(r.Text, r.Title),
			File:     relTo(root, r.Pos.File),
			Line:     r.Pos.Line,
		}
		for _, s := range r.Sources {
			switch {
			case s.Extern != "":
				out.Extern = append(out.Extern, s.Extern)
			case s.Doc != "" && s.Anchor != "":
				out.Sources = append(out.Sources, s.Doc+"#"+s.Anchor)
			case s.Doc != "":
				out.Sources = append(out.Sources, s.Doc)
			}
		}
		for _, t := range sat[r.ID] {
			out.Satisfiers = append(out.Satisfiers, t.String())
		}
		// Sorted, because the export is diffed and cached by whoever reads it.
		// Binding order is stable within a run but is a consequence of package
		// order, which is not a property anybody should have to rely on.
		sort.Strings(out.Satisfiers)
		report.Requirements = append(report.Requirements, out)
	}

	for _, doc := range docs {
		d := set.Get(doc)
		if d.Err != nil {
			continue
		}
		for _, seg := range d.Segments {
			rec, recorded := base.Sources[seg.Ref()]
			report.Sources = append(report.Sources, treeSourceSegment{
				Ref:          seg.Ref(),
				Doc:          seg.Doc,
				ID:           seg.ID,
				Kind:         seg.Kind.String(),
				Title:        seg.Title,
				Informative:  seg.Informative,
				Requirements: cov.ByCiter[seg.Ref()],
				Drifted:      recorded && rec.Fingerprint != seg.Fingerprint,
			})
		}
	}

	for _, f := range findings.Findings() {
		report.Findings = append(report.Findings, jsonFindingRef{
			Code: string(f.Code),
			Rule: f.Rule,
			File: relTo(root, f.Pos.File),
			Line: f.Pos.Line,
			What: f.What,
		})
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// export writes the requirement tree and its sources as JSON.
//
// It is the read surface of the collaboration side, and it exists because that
// side has an audience the rest of speclink does not. Everywhere else the
// consumer is an agent, which needs to know what is broken; here the consumer
// is a person who never sees Go, working through the requirements that were
// extracted from documents they wrote, deciding whether those requirements say
// what they meant. A list of findings is the wrong shape for that: it says what
// is wrong and nothing about what is there.
//
// It is a separate command rather than a format of `requirements` because
// `-format json` is already a contract, and the thing reading it is an agent
// that would break. Adding an output is cheaper than redefining one.
//
// Nothing is verified beyond what the tree needs. Like `requirements`, only the
// named packages have to compile, so the surface stays usable while the
// implementation is in pieces — which is the normal state of affairs when
// somebody is still deciding what the requirements are.
func export(args []string) error {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	root := fs.String("root", ".", "repository root, used to resolve source documents")
	cfgPath := fs.String("config", "", "layout configuration; defaults to "+config.FileName+" in the root")
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
		fmt.Fprintln(os.Stderr, "the Go build is broken; fix it before speclink can read anything:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, "  "+e.Error())
		}
		return errFindings
	}

	findings := &diag.Set{}
	var (
		reqs     []*ir.Requirement
		bindings []ir.Binding
	)
	for _, p := range pkgs {
		reqs = append(reqs, p.ReadRequirements(findings)...)
		bindings = append(bindings, p.ReadBindings(findings)...)
	}

	tree := reqtree.Build(absRoot, reqs, findings)
	tree.CheckLayout(findings)
	docs, sourceDocs := loadSources(absRoot, layout, findings)
	tree.CheckSources(docs, findings)
	src := check.CoverSources(tree, docs, sourceDocs, nil, findings)
	reqtree.ReportDocuments(docs, findings)

	base, err := baseline.Load(absRoot)
	if err != nil {
		return err
	}

	// Satisfiers are read from the bindings directly rather than through
	// CoverRequirements, because the coverage check would add findings about
	// code this command is not asking about.
	sat := map[string][]ir.Target{}
	for _, b := range bindings {
		for _, a := range b.Assertions {
			if a.Kind != ir.AssertSatisfies {
				continue
			}
			for _, ref := range a.Requirements {
				r := tree.ByGoIdent(ref)
				if r == nil {
					r = tree.ByID[ref]
				}
				if r != nil {
					sat[r.ID] = append(sat[r.ID], b.Target)
				}
			}
		}
	}

	check.Drift(tree, docs, sourceDocs, check.Coverage{BySatisfier: sat}, src, base, nil, findings)
	return writeTree(os.Stdout, absRoot, tree, docs, sourceDocs, src, sat, base, findings)
}

// relTo makes a path repository relative. The export is read by a surface that
// has no notion of the machine the run happened on, so an absolute path there
// is at best noise and at worst something it tries to open.
func relTo(root, path string) string {
	if path == "" || root == "" {
		return path
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return filepath.ToSlash(rel)
}
