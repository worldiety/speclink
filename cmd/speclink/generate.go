package main

import (
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

	topo       ir.Topology
	processes  []*ir.Process
	constructs []ir.Construct
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
	if inf, ok := frontend.(lang.ConstructInferrer); ok {
		m.constructs = inf.Constructs(discard)
	}
	m.tree = reqtree.Build(absRoot, reqs, discard)
	if tr, ok := frontend.(lang.TopicReader); ok {
		m.tree.ResolveTopics(tr.Topics(), discard)
	}
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
	m.writeDocuments(b)
	m.writeTopics(b)
	m.writeStandards(b)
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
		m.writeImplementation(b, r)
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

// writeStandards renders one chapter per external standard.
//
// It is the audit chapter, and it exists because the tool already had every
// part of it. A standard is a source document whose segments are its clauses,
// so which clause is answered by what is the ordinary citation index read from
// the other end — and which clause is answered by nothing is the ordinary
// coverage rule, reported against a catalogue instead of a Markdown file.
//
// The second table is the statement of applicability. ISO 27001 requires it as
// a document in its own right, and here it is the collected reasons of every
// clause somebody excluded — which is the form it has to take anyway, and the
// form nobody maintains by hand for long.
func (m *specModel) writeStandards(b *strings.Builder) {
	if len(m.src.Standards) == 0 {
		return
	}

	b.WriteString("## Standards\n\n")
	for _, st := range m.src.Standards {
		fmt.Fprintf(b, "### %s\n\n", or(st.Title, st.Doc))
		fmt.Fprintf(b, "%d of %s answered", st.Answered,
			plural(st.Clauses, "applicable clause", "applicable clauses"))
		if st.Excluded > 0 {
			fmt.Fprintf(b, " · %d not applicable", st.Excluded)
		}
		fmt.Fprintf(b, " · `%s`\n\n", st.Doc)

		doc := m.docs.Get(st.Doc)
		m.writeClauses(b, doc)
		m.writeApplicability(b, doc)
	}
}

func (m *specModel) writeClauses(b *strings.Builder, doc source.Document) {
	b.WriteString("| Clause | Obligation | Answered by |\n|---|---|---|\n")
	for _, seg := range doc.Segments {
		if seg.Informative {
			continue
		}
		by := m.src.ByCiter[seg.Ref()]
		answer := "**nothing**"
		if len(by) > 0 {
			answer = strings.Join(by, ", ")
		}
		fmt.Fprintf(b, "| `%s` | %s | %s |\n", seg.ID, oneLine(seg.Title), answer)
	}
	b.WriteString("\n")
}

// writeApplicability renders the exclusions with the reason each was given.
func (m *specModel) writeApplicability(b *strings.Builder, doc source.Document) {
	var excluded []source.Segment
	for _, seg := range doc.Segments {
		if seg.Informative {
			excluded = append(excluded, seg)
		}
	}
	if len(excluded) == 0 {
		return
	}

	b.WriteString("#### Statement of applicability\n\n")
	b.WriteString("Clauses excluded here, and why. An exclusion is a decision somebody\nmade; the reason is the whole of what it is worth.\n\n")
	b.WriteString("| Clause | Obligation | Does not apply because |\n|---|---|---|\n")
	for _, seg := range excluded {
		fmt.Fprintf(b, "| `%s` | %s | %s |\n", seg.ID, oneLine(seg.Title), oneLine(seg.Because))
	}
	b.WriteString("\n")
}

// writeTopics renders one chapter per theme, and one for what carries none.
//
// An index rather than a repetition: every requirement appears in full further
// down, and printing it twice would make the document longer without making it
// say more. What a chapter answers is which requirements belong together, and
// an ID with its title answers that.
//
// The last chapter is the one that matters. A requirement filed under no theme
// is not a defect — themes are optional on purpose — but leaving it out of the
// document silently would be, because an absent requirement reads as one that
// does not exist rather than as one nobody filed.
func (m *specModel) writeTopics(b *strings.Builder) {
	topics := m.tree.Topics()
	if len(topics) == 0 {
		return
	}

	filed := map[string]bool{}
	b.WriteString("## Themes\n\n")

	for _, top := range topics {
		fmt.Fprintf(b, "### %s\n\n", or(top.Title, top.ID))
		if top.Description != "" {
			fmt.Fprintf(b, "%s\n\n", oneLine(top.Description))
		}

		b.WriteString("| ID | Requirement |\n|---|---|\n")
		for _, id := range m.sortedRequirementIDs() {
			r := m.tree.ByID[id]
			if !hasTopic(r, top.ID) {
				continue
			}
			filed[id] = true
			fmt.Fprintf(b, "| `%s` | %s |\n", r.ID, oneLine(or(r.Title, r.Text)))
		}
		b.WriteString("\n")
	}

	var loose []string
	for _, id := range m.sortedRequirementIDs() {
		if !filed[id] {
			loose = append(loose, id)
		}
	}
	if len(loose) == 0 {
		return
	}

	b.WriteString("### Under no theme\n\n")
	fmt.Fprintf(b, "%s filed under no theme. Themes are optional, so this is not a\ndefect — but the count belongs here, because a requirement left out of every\nchapter reads as one that does not exist.\n\n",
		plural(len(loose), "requirement is", "requirements are"))
	b.WriteString("| ID | Requirement |\n|---|---|\n")
	for _, id := range loose {
		r := m.tree.ByID[id]
		fmt.Fprintf(b, "| `%s` | %s |\n", r.ID, oneLine(or(r.Title, r.Text)))
	}
	b.WriteString("\n")
}

func hasTopic(r *ir.Requirement, id string) bool {
	for _, t := range r.Topics {
		if t == id {
			return true
		}
	}
	return false
}

// sortedRequirementIDs is the reading order of the tree.
func (m *specModel) sortedRequirementIDs() []string {
	out := make([]string, 0, len(m.tree.ByID))
	for id := range m.tree.ByID {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

// writeDocuments renders what has become of the material the requirements were
// written from.
//
// Every column comes out of the lock, which has held all of it since review was
// recorded there and has never been read back. A source document is not a
// backdrop: it is the thing somebody wrote before any of this existed, and the
// question of how much of it has turned into a requirement — and how much of
// that a person has since read — is the one a specification is usually least
// able to answer.
//
// Reviewed is the strict reading: a segment counts as read when every
// requirement citing it has been. The weaker one, at least one, would let a
// section with four requirements and one reviewer read as accounted for.
func (m *specModel) writeDocuments(b *strings.Builder) {
	if len(m.segments) == 0 {
		return
	}

	b.WriteString("## The material\n\n")
	b.WriteString("| Document | Kind | Segments | Cited | Read | Drifted |\n")
	b.WriteString("|---|---|---:|---:|---:|---:|\n")

	for _, path := range m.segments {
		d := m.docs.Get(path)
		if d.Err != nil {
			fmt.Fprintf(b, "| `%s` | — | — | — | — | — |\n", path)
			continue
		}

		var total, cited, read, drifted int
		for _, seg := range d.Segments {
			if seg.Informative {
				continue
			}
			total++

			citers := m.src.ByCiter[seg.Ref()]
			if len(citers) == 0 {
				continue
			}
			cited++
			if m.allReviewed(citers) {
				read++
			}
			if rec, ok := m.base.Sources[seg.Ref()]; ok && rec.Fingerprint != seg.Fingerprint {
				drifted++
			}
		}
		fmt.Fprintf(b, "| `%s` | %s | %d | %d | %d | %d |\n", path, d.Kind, total, cited, read, drifted)
	}
	b.WriteString("\n")
}

// allReviewed reports whether every requirement in the list has been read.
func (m *specModel) allReviewed(ids []string) bool {
	for _, id := range ids {
		rec, ok := m.base.Requirements[id]
		if !ok || rec.ReviewedBy == "" {
			return false
		}
	}
	return len(ids) > 0
}

// writeImplementation names what implements a requirement, down to the lines.
//
// The list of satisfiers alone answers "is there code for this" and stops
// there. What a review or an audit asks next is where that code is and whether
// anything ran it, and both were knowable: the extent comes from the
// declaration, the coverage from the last profile handed to evidence.
//
// The coverage figure appears only where it was measured on the text that is
// there now. A profile taken before a rewrite is not evidence about what is
// there today, and printing it as though it were would be worse than leaving
// the column out.
func (m *specModel) writeImplementation(b *strings.Builder, r *ir.Requirement) {
	targets := m.cov.BySatisfier[r.ID]
	if len(targets) == 0 {
		return
	}
	names := make([]string, 0, len(targets))
	for _, t := range targets {
		names = append(names, t.String())
	}
	sort.Strings(names)

	byName := map[string]ir.Construct{}
	for _, c := range m.constructs {
		byName[c.Name] = c
		byName[lastSegment(c.Name)] = c
	}

	b.WriteString("- **Implemented by**\n")
	for _, name := range names {
		c, known := byName[name]
		if !known || c.EndLine == 0 {
			fmt.Fprintf(b, "  - `%s`\n", name)
			continue
		}
		fmt.Fprintf(b, "  - `%s` — `%s:%d–%d`%s\n",
			name, m.relative(c.Pos.File), c.Pos.Line, c.EndLine, m.exercised(c))
	}
}

// exercised renders the share of a declaration a run reached, or nothing.
func (m *specModel) exercised(c ir.Construct) string {
	rec, ok := m.base.Constructs[c.Name]
	if !ok || rec.Statements == 0 || rec.Fingerprint != c.Fingerprint {
		return ""
	}
	return fmt.Sprintf(", %d of %d statements exercised", rec.Covered, rec.Statements)
}

// relative shortens an absolute source path to something a reader can find.
func (m *specModel) relative(file string) string {
	if rel, err := filepath.Rel(m.root, file); err == nil {
		return filepath.ToSlash(rel)
	}
	return file
}
