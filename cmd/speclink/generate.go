package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
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

// generate derives the specification document from the code.
//
// This is the part that decides whether speclink is worth having at all. As
// long as it exists *beside* a hand written specification, a generator and a
// knowledge graph, it makes the situation worse: one more thing to keep in
// step. The point was always the removal of the others, and nothing can be
// removed until the document comes out of here.
//
// So this writes what a person would otherwise write by hand and then fail to
// maintain: every requirement with the words it was given, where those words
// came from, what implements them, what demonstrates them, and who has read
// them. All of it was already known and none of it was rendered.
//
// Markdown, and deliberately nothing cleverer. It renders everywhere, it diffs,
// and a diff is the form in which this document is actually reviewed. An HTML
// backend with an asset pipeline would be a second thing to maintain, which is
// the failure mode this command exists to end.
func generate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	root := fs.String("root", ".", "repository root, used to resolve source documents")
	cfgPath := fs.String("config", "", "layout configuration; defaults to "+config.FileName+" in the root")
	prof := fs.String("profile", "", "language, framework and architectural style; overrides "+config.FileName)
	out := fs.String("out", "", "write to this `file` instead of standard output")
	if err := fs.Parse(args); err != nil {
		return err
	}

	absRoot, err := absRootOf(*root)
	if err != nil {
		return err
	}

	model, err := readModel(absRoot, *cfgPath, *prof, fs.Args())
	if err != nil {
		return err
	}

	w := io.Writer(os.Stdout)
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	return model.writeMarkdown(w)
}

// specModel is everything the document needs, gathered once.
type specModel struct {
	root     string
	tree     *reqtree.Tree
	docs     *source.Set
	segments []string
	src      check.SourceCoverage
	cov      check.Coverage
	ver      check.Verification
	base     *baseline.File
	waived   ir.Waivers
	skipped  int

	topo      ir.Topology
	processes []*ir.Process
}

// readModel assembles the graph without judging it.
//
// Diagnostics are discarded. The document describes what is there, and a
// project with open findings is exactly when somebody wants to read it —
// refusing to render until everything is green would make it useless in the
// only situation that needs it.
func readModel(absRoot, cfgPath, prof string, patterns []string) (*specModel, error) {
	frontend, layout, _, err := open(absRoot, cfgPath, prof, patterns, true)
	if err != nil {
		return nil, err
	}

	discard := &diag.Set{}
	reqs := frontend.Requirements(discard)
	bindings := frontend.Bindings(discard)

	var verifications []ir.Binding
	if vr, ok := frontend.(lang.VerificationReader); ok {
		verifications = vr.Verifications(discard)
	}

	m := &specModel{root: absRoot, skipped: skippedPackages(frontend)}
	m.tree = reqtree.Build(absRoot, reqs, discard)
	m.docs, m.segments = loadSources(absRoot, layout, discard)
	m.src = check.CoverSources(m.tree, m.docs, m.segments, nil, discard)

	measured := measuredRequirements(m.tree, layout, absRoot)
	m.cov = check.CoverRequirements(m.tree, bindings, measured, frontend.Dialect(), discard)
	m.ver = check.CoverVerification(m.tree, verifications, m.cov, measured, ir.CollectWaivers(bindings), frontend.Dialect(), discard)

	if tr, ok := frontend.(lang.TopologyReader); ok {
		m.topo = tr.Topology(discard)
	}
	if pr, ok := frontend.(lang.ProcessReader); ok {
		m.processes = pr.Processes(discard)
	}

	m.waived = ir.CollectWaivers(bindings)
	m.base, err = baseline.Load(absRoot)
	if err != nil {
		return nil, err
	}
	m.ver.Shown = check.Demonstrated(m.tree, m.ver, m.cov, measured, m.base, ir.CollectWaivers(bindings), discard)
	return m, nil
}

func (m *specModel) writeMarkdown(w io.Writer) error {
	b := &strings.Builder{}

	b.WriteString("# Specification\n\n")
	b.WriteString("Derived from the source by `speclink generate`. Do not edit: every\n")
	b.WriteString("sentence here is written somewhere else, and the point of this file is\n")
	b.WriteString("that there is only one such place.\n\n")

	m.writeSummary(b)
	m.writeGaps(b)
	m.writeBoundary(b)
	m.writeProcesses(b)
	m.writeRequirements(b)
	m.writeSources(b)

	_, err := io.WriteString(w, b.String())
	return err
}

func (m *specModel) writeSummary(b *strings.Builder) {
	b.WriteString("## Where it stands\n\n")
	b.WriteString("| | measured | complete |\n|---|---:|---:|\n")
	fmt.Fprintf(b, "| Source segments accounted for | %d | %.0f%% |\n", m.src.Total, m.src.Ratio()*100)
	fmt.Fprintf(b, "| Normative requirements covered | %d | %.0f%% |\n", m.cov.Normative, m.cov.Ratio()*100)
	fmt.Fprintf(b, "| … claimed by a test | %d | %.0f%% |\n", m.ver.Normative, m.ver.Ratio()*100)
	fmt.Fprintf(b, "| … demonstrated by a run | %d | %.0f%% |\n", m.ver.Normative, m.ver.ShownRatio()*100)
	fmt.Fprintf(b, "| … read by a person | %d | %.0f%% |\n", m.cov.Normative, m.reviewedRatio()*100)
	b.WriteString("\n")

	if m.skipped > 0 {
		fmt.Fprintf(b, "> %s lie outside the configured scope. Nothing above is claimed about them.\n\n",
			plural(m.skipped, "package", "packages"))
	}
}

// writeGaps is the part somebody acts on.
//
// It is separate from the findings a run produces, and says the same things in
// the opposite direction: a run tells an agent what to fix next, this tells a
// reader what the specification does not yet cover. Same graph, different
// question, and the second one has never had an answer.
func (m *specModel) writeGaps(b *strings.Builder) {
	type gap struct {
		title string
		items []string
	}
	gaps := []gap{
		{"Requirements nothing implements", m.cov.Uncovered},
		{"Requirements no test claims", m.unclaimed()},
		{"Requirements no run has demonstrated", m.undemonstrated()},
		{"Requirements nobody has read", m.unreviewed()},
		{"Source segments that became no requirement", m.unaccounted()},
	}

	empty := true
	for _, g := range gaps {
		if len(g.items) > 0 {
			empty = false
		}
	}
	b.WriteString("## Gaps\n\n")
	if empty {
		b.WriteString("None.\n\n")
		return
	}
	for _, g := range gaps {
		if len(g.items) == 0 {
			continue
		}
		fmt.Fprintf(b, "### %s\n\n", g.title)
		for _, item := range g.items {
			fmt.Fprintf(b, "- %s\n", item)
		}
		b.WriteString("\n")
	}
}

func (m *specModel) writeRequirements(b *strings.Builder) {
	b.WriteString("## Requirements\n\n")

	for _, r := range m.tree.All() {
		fmt.Fprintf(b, "### %s — %s\n\n", r.ID, or(r.Title, "(untitled)"))
		if r.Text != "" {
			fmt.Fprintf(b, "%s\n\n", r.Text)
		}
		fmt.Fprintf(b, "*%s, %s, %s.*\n\n", r.Kind, r.Discipline, r.Status)
		if r.Rationale != "" {
			fmt.Fprintf(b, "**Why.** %s\n\n", r.Rationale)
		}

		m.writeList(b, "Asked for in", m.origins(r))
		m.writeList(b, "Derived from", r.DerivedFrom)
		m.writeList(b, "Supersedes", r.Supersedes)
		m.writeList(b, "Implemented by", targetNames(m.cov.BySatisfier[r.ID]))
		m.writeList(b, "Demonstrated by", m.demonstratedBy(r))
		if who := m.base.Requirements[r.ID]; who.ReviewedBy != "" && who.Text == baseline.HashText(r.Text, r.Title) {
			fmt.Fprintf(b, "- **Read by** %s\n", who.ReviewedBy)
		}
		b.WriteString("\n")
	}
}

func (m *specModel) writeSources(b *strings.Builder) {
	b.WriteString("## Source documents\n\n")
	b.WriteString("What people wrote, and what became of each part of it.\n\n")

	var last string
	for _, doc := range m.segments {
		d := m.docs.Get(doc)
		if d.Err != nil {
			continue
		}
		for _, seg := range d.Segments {
			if seg.Doc != last {
				fmt.Fprintf(b, "### %s\n\n| section | became |\n|---|---|\n", seg.Doc)
				last = seg.Doc
			}
			became := strings.Join(m.src.ByCiter[seg.Ref()], ", ")
			switch {
			case became != "":
			case seg.Informative:
				became = "*nothing, and says so*"
			default:
				became = "**nothing**"
			}
			fmt.Fprintf(b, "| %s | %s |\n", or(seg.Title, seg.ID), became)
		}
	}
	b.WriteString("\n")
}

func (m *specModel) writeList(b *strings.Builder, label string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "- **%s** %s\n", label, strings.Join(items, ", "))
}

func (m *specModel) origins(r *ir.Requirement) []string {
	var out []string
	for _, s := range r.Sources {
		switch {
		case s.Extern != "":
			out = append(out, s.Extern)
		case s.Doc != "" && s.Anchor != "":
			out = append(out, s.Doc+"#"+s.Anchor)
		case s.Doc != "":
			out = append(out, s.Doc)
		}
	}
	return out
}

func (m *specModel) demonstratedBy(r *ir.Requirement) []string {
	return m.base.VerifiedBy(r.ID, baseline.HashText(r.Text, r.Title))
}

func (m *specModel) unclaimed() []string {
	var out []string
	for _, r := range m.tree.All() {
		if r.Status.MustBeCovered() && len(m.ver.ByTest[r.ID]) == 0 {
			out = append(out, r.ID+m.excuse(r.ID, check.RuleRequirementUnverified))
		}
	}
	return out
}

// excuse renders the justification for a gap that was accepted on purpose.
//
// A waiver without its reason in this document would be the worst of both: the
// gap looks like an oversight to whoever reads it, and the sentence somebody
// was made to write disappears. The reason is mandatory precisely so that it
// can be read here.
func (m *specModel) excuse(id, rule string) string {
	for _, t := range m.cov.BySatisfier[id] {
		if reason := m.waived.Reason(t.String(), rule); reason != "" {
			return " — *accepted:* " + reason
		}
	}
	if reason := m.waived.Reason("", rule); reason != "" {
		return " — *accepted for every requirement:* " + reason
	}
	return ""
}

func (m *specModel) undemonstrated() []string {
	var out []string
	for _, r := range m.tree.All() {
		if !r.Status.MustBeCovered() || len(m.ver.ByTest[r.ID]) == 0 {
			continue
		}
		if len(m.demonstratedBy(r)) == 0 {
			out = append(out, r.ID)
		}
	}
	return out
}

func (m *specModel) unreviewed() []string {
	var out []string
	for _, r := range m.tree.All() {
		if !r.Status.MustBeCovered() {
			continue
		}
		rec := m.base.Requirements[r.ID]
		if rec.ReviewedBy == "" || rec.Text != baseline.HashText(r.Text, r.Title) {
			out = append(out, r.ID)
		}
	}
	return out
}

func (m *specModel) unaccounted() []string {
	var out []string
	for _, doc := range m.segments {
		d := m.docs.Get(doc)
		if d.Err != nil {
			continue
		}
		for _, seg := range d.Segments {
			if !seg.Informative && len(m.src.ByCiter[seg.Ref()]) == 0 {
				out = append(out, seg.Ref())
			}
		}
	}
	return out
}

func (m *specModel) reviewedRatio() float64 {
	if m.cov.Normative == 0 {
		return 1
	}
	read := 0
	for _, r := range m.tree.All() {
		if !r.Status.MustBeCovered() {
			continue
		}
		rec := m.base.Requirements[r.ID]
		if rec.ReviewedBy != "" && rec.Text == baseline.HashText(r.Text, r.Title) {
			read++
		}
	}
	return float64(read) / float64(m.cov.Normative)
}

func targetNames(targets []ir.Target) []string {
	out := make([]string, 0, len(targets))
	for _, t := range targets {
		out = append(out, "`"+t.String()+"`")
	}
	sort.Strings(out)
	return out
}

func or(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// writeBoundary renders the interface catalogue.
//
// It is the table somebody assembles by hand before every review and gets
// wrong, because the information lives at both ends of each channel and in
// neither place as a list. Here it is one row per declared channel, and the
// four descriptive columns are mandatory in the model, so a blank cell cannot
// reach this page.
func (m *specModel) writeBoundary(b *strings.Builder) {
	if !m.topo.Declared() {
		return
	}

	b.WriteString("## The boundary\n\n")

	if len(m.topo.Participants) > 0 {
		b.WriteString("| Outside | Kind | Role |\n|---|---|---|\n")
		for _, p := range m.topo.Participants {
			fmt.Fprintf(b, "| %s | %s | %s |\n", p.Name, p.Kind, oneLine(p.Role))
		}
		b.WriteString("\n")
	}

	if len(m.topo.Channels) == 0 {
		return
	}
	b.WriteString("### Every way across\n\n")
	b.WriteString("| Channel | Protocol | Data | Authentication | In transit |\n")
	b.WriteString("|---|---|---|---|---|\n")
	for _, c := range m.topo.Channels {
		fmt.Fprintf(b, "| **%s**<br>%s → %s | %s | %s | %s | %s |\n",
			c.Label, c.From, c.To, c.Protocol, oneLine(c.Data), oneLine(c.Auth), oneLine(c.Crypto))
	}
	b.WriteString("\n")
}

// writeProcesses renders each course of business as its edges.
//
// The edge list rather than a numbered sequence, because the model is a graph
// and a numbered list would have to pick an order that does not exist. Where a
// process branches and comes back, any numbering is a lie about which step
// follows which.
func (m *specModel) writeProcesses(b *strings.Builder) {
	if len(m.processes) == 0 {
		return
	}

	b.WriteString("## Courses of business\n\n")
	for _, p := range m.processes {
		fmt.Fprintf(b, "### %s\n\n", or(p.Title, p.ID))
		if p.Purpose != "" {
			fmt.Fprintf(b, "%s\n\n", oneLine(p.Purpose))
		}
		fmt.Fprintf(b, "`%s` · drawn in `process-%s.puml`\n\n", p.ID, p.ID)

		if ids := m.satisfiedBy(p); len(ids) > 0 {
			fmt.Fprintf(b, "Answers to: %s\n\n", strings.Join(ids, ", "))
		}

		b.WriteString("| From | To | When |\n|---|---|---|\n")
		for _, e := range p.Edges {
			fmt.Fprintf(b, "| %s | %s | %s |\n", m.nodeLabel(p, e.From), m.nodeLabel(p, e.To), or(e.When, "—"))
		}
		b.WriteString("\n")
	}
}

// nodeLabel renders a node for the table: the construct it names where it names
// one, and what it is where it does not.
func (m *specModel) nodeLabel(p *ir.Process, id string) string {
	n, ok := p.Node(id)
	if !ok {
		return id
	}
	switch {
	case n.Ref != "":
		return lastSegment(n.Ref) + " *(" + n.Kind.String() + ")*"
	case n.Label != "":
		return n.Label + " *(" + n.Kind.String() + ")*"
	}
	return id + " *(" + n.Kind.String() + ")*"
}

// satisfiedBy resolves the requirement identifiers a process names.
func (m *specModel) satisfiedBy(p *ir.Process) []string {
	var out []string
	for _, ref := range p.Satisfies {
		if r := m.tree.ByGoIdent(ref); r != nil {
			out = append(out, r.ID)
		}
	}
	return out
}

func lastSegment(name string) string {
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		return name[i+1:]
	}
	return name
}

// oneLine keeps a cell from breaking the table it sits in.
func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "|", "\\|")
	return strings.TrimSpace(s)
}
