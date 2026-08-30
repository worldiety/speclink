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
	"github.com/worldiety/speclink/internal/doc"
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
// What is built here is a document, not Markdown.
//
// This reverses what used to stand in this comment. It said Markdown and
// deliberately nothing cleverer, on the grounds that a second backend would be
// a second thing to maintain, which is the failure this command exists to end.
// That objection was right about parallel writers and is not answered by
// ignoring it: two functions walking this model and each deciding what the
// document says would drift apart on the first change.
//
// It is answered by there still being exactly one place that decides what the
// specification contains. Only the spelling moved, into internal/doc, where a
// renderer is handed a tree of headings, tables and sentences and gets no
// access to the model that produced them. Two outputs that disagree is not a
// bug that can be introduced by editing a renderer.
//
// What forced the question was the PDF. An audit document cites a requirement
// from a chapter, and Markdown can only emit an anchor nobody checks; Typst
// refuses to compile a reference that lands nowhere. Pivoting through Markdown
// would have capped the document forever at what Markdown can express.
//
// Markdown remains the default, because it renders everywhere, it diffs, and a
// diff is the form in which this document is actually reviewed.
func generate(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	root := fs.String("root", ".", "repository root, used to resolve source documents")
	cfgPath := fs.String("config", "", "layout configuration; defaults to "+config.FileName+" in the root")
	prof := fs.String("profile", "", "language, framework and architectural style; overrides "+config.FileName)
	out := fs.String("out", "", "write to this `file` instead of standard output")
	format := fs.String("format", "markdown", "markdown or typst")
	author := fs.String("author", "", "who the document is issued by; typst only, left off when empty")
	date := fs.String("date", "", "the date on the title page; typst only, left off when empty")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var r doc.Renderer
	switch *format {
	case "markdown":
		r = doc.Markdown{}
	case "typst":
		// The date is passed in and never read from the clock. Generating the
		// document twice from the same tree has to produce the same bytes, or
		// it cannot be committed and diffed, and a title page that changes at
		// midnight would put a spurious change in front of every reviewer.
		r = doc.Typst{Author: *author, Date: *date}
	default:
		return fmt.Errorf("unknown format %q: markdown or typst", *format)
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
	_, err = io.WriteString(w, r.Render(model.document()))
	return err
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
	endpoints  []ir.Endpoint

	// can records what the frontend was able to look for at all.
	//
	// A chapter missing because nothing was declared and one missing because
	// nobody could look are the same blank page, and the reader of a
	// specification is the last person in the chain who could tell them apart.
	// It is carried here so the document can say which it is.
	can lang.Capabilities
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

	m := &specModel{root: absRoot, skipped: skippedPackages(frontend), can: lang.Of(frontend)}
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
	if er, ok := frontend.(lang.EndpointReader); ok {
		m.endpoints = er.Endpoints()
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

// document builds the specification. It decides what the document says and
// nothing about how it is spelled; a renderer does the rest.
func (m *specModel) document() *doc.Doc {
	d := doc.New("Specification")

	d.P(doc.T("Derived from the source by "), doc.Code("speclink generate"),
		doc.T(". Do not edit: every sentence here is written somewhere else, and the point of this file is that there is only one such place."))

	m.writeSummary(d)
	m.writeGaps(d)
	m.writeDocuments(d)
	m.writeTopics(d)
	m.writeStandards(d)
	m.writeBoundary(d)
	m.writeSurface(d)
	m.writeProcesses(d)
	m.writeRequirements(d)
	m.writeSources(d)

	return d
}

func (m *specModel) writeSummary(d *doc.Doc) {
	d.H(2, "Where it stands")
	d.Table("", "measured", "complete").
		Aligned(doc.Left, doc.Right, doc.Right).
		Add(doc.Cell(doc.T("Source segments accounted for")), doc.Cell(doc.Tf("%d", m.src.Total)), doc.Cell(doc.Tf("%.0f%%", m.src.Ratio()*100))).
		Add(doc.Cell(doc.T("Normative requirements covered")), doc.Cell(doc.Tf("%d", m.cov.Normative)), doc.Cell(doc.Tf("%.0f%%", m.cov.Ratio()*100))).
		Add(doc.Cell(doc.T("… claimed by a test")), doc.Cell(doc.Tf("%d", m.ver.Normative)), doc.Cell(doc.Tf("%.0f%%", m.ver.Ratio()*100))).
		Add(doc.Cell(doc.T("… demonstrated by a run")), doc.Cell(doc.Tf("%d", m.ver.Normative)), doc.Cell(doc.Tf("%.0f%%", m.ver.ShownRatio()*100))).
		Add(doc.Cell(doc.T("… read by a person")), doc.Cell(doc.Tf("%d", m.cov.Normative)), doc.Cell(doc.Tf("%.0f%%", m.reviewedRatio()*100)))

	if m.skipped > 0 {
		d.Notef("%s lie outside the configured scope. Nothing above is claimed about them.",
			plural(m.skipped, "package", "packages"))
	}
}

// writeGaps is the part somebody acts on.
//
// It is separate from the findings a run produces, and says the same things in
// the opposite direction: a run tells an agent what to fix next, this tells a
// reader what the specification does not yet cover. Same graph, different
// question, and the second one has never had an answer.
func (m *specModel) writeGaps(d *doc.Doc) {
	type gap struct {
		title string
		items []doc.Bullet
	}
	gaps := []gap{
		{"Requirements nothing implements", plainItems(m.cov.Uncovered)},
		{"Requirements no test claims", m.unclaimed()},
		{"Requirements no run has demonstrated", plainItems(m.undemonstrated())},
		{"Requirements nobody has read", plainItems(m.unreviewed())},
		{"Source segments that became no requirement", plainItems(m.unaccounted())},
	}

	empty := true
	for _, g := range gaps {
		if len(g.items) > 0 {
			empty = false
		}
	}
	d.H(2, "Gaps")
	if empty {
		d.P(doc.T("None."))
		return
	}
	for _, g := range gaps {
		if len(g.items) == 0 {
			continue
		}
		d.H(3, g.title)
		d.Bullets(g.items...)
	}
}

// plainItems turns a list of identifiers the model produced into list items.
func plainItems(in []string) []doc.Bullet {
	var out []doc.Bullet
	for _, s := range in {
		out = append(out, doc.Item(doc.T(s)))
	}
	return out
}

func (m *specModel) writeRequirements(d *doc.Doc) {
	d.H(2, "Requirements")

	for _, r := range m.tree.All() {
		d.HID(3, r.ID, r.ID+" — "+or(r.Title, "(untitled)"))
		if r.Text != "" {
			d.P(doc.T(r.Text))
		}
		d.P(doc.Emph(fmt.Sprintf("%s, %s, %s.", r.Kind, r.Discipline, r.Status)))
		if r.Rationale != "" {
			d.P(doc.Strong("Why."), doc.T(" "+r.Rationale))
		}

		it := &items{}
		m.writeList(it, "Asked for in", m.origins(r))
		m.writeList(it, "Derived from", r.DerivedFrom)
		m.writeList(it, "Supersedes", r.Supersedes)
		m.writeImplementation(it, r)
		m.writeList(it, "Demonstrated by", m.demonstratedBy(r))
		if who := m.base.Requirements[r.ID]; who.ReviewedBy != "" && who.Text == baseline.HashText(r.Text, r.Title) {
			it.add(doc.Strong("Read by"), doc.T(" "+who.ReviewedBy))
		}
		d.Bullets(it.list...)
	}
}

// items collects the bullets under one requirement.
//
// They were separate writes into one Markdown buffer and so happened to form a
// single list; a document has to say that on purpose, because two lists in a
// row are two lists.
// items collects the bullets under one requirement.
//
// The list is assembled by four separate contributors — origins, derivation,
// implementation, demonstration — and they form one list, not four. An
// accumulator is passed instead of the document so that stays true.
type items struct{ list []doc.Bullet }

func (it *items) add(in ...doc.Inline) { it.list = append(it.list, doc.Item(in...)) }

// under hangs a sub item off the bullet most recently added.
func (it *items) under(in ...doc.Inline) {
	if len(it.list) == 0 {
		it.add(in...)
		return
	}
	it.list[len(it.list)-1] = it.list[len(it.list)-1].Under(in...)
}

func (m *specModel) writeSources(d *doc.Doc) {
	d.H(2, "Source documents")
	d.P(doc.T("What people wrote, and what became of each part of it."))

	var last string
	var t *doc.Table
	for _, name := range m.segments {
		dd := m.docs.Get(name)
		if dd.Err != nil {
			continue
		}
		for _, seg := range dd.Segments {
			if seg.Doc != last {
				d.H(3, seg.Doc)
				t = d.Table("section", "became")
				last = seg.Doc
			}
			became := doc.Cell(doc.T(strings.Join(m.src.ByCiter[seg.Ref()], ", ")))
			switch {
			case len(m.src.ByCiter[seg.Ref()]) > 0:
			case seg.Informative:
				became = doc.Cell(doc.Emph("nothing, and says so"))
			default:
				became = doc.Cell(doc.Strong("nothing"))
			}
			t.Add(doc.Cell(doc.T(or(seg.Title, seg.ID))), became)
		}
	}
}

func (m *specModel) writeList(it *items, label string, vals []string) {
	if len(vals) == 0 {
		return
	}
	it.add(doc.Strong(label), doc.T(" "+strings.Join(vals, ", ")))
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

func (m *specModel) unclaimed() []doc.Bullet {
	var out []doc.Bullet
	for _, r := range m.tree.All() {
		if r.Status.MustBeCovered() && len(m.ver.ByTest[r.ID]) == 0 {
			out = append(out, doc.Item(append([]doc.Inline{doc.T(r.ID)},
				m.excuse(r.ID, check.RuleRequirementUnverified)...)...))
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
func (m *specModel) excuse(id, rule string) []doc.Inline {
	for _, t := range m.cov.BySatisfier[id] {
		if reason := m.waived.Reason(t.String(), rule); reason != "" {
			return []doc.Inline{doc.T(" — "), doc.Emph("accepted:"), doc.T(" " + reason)}
		}
	}
	if reason := m.waived.Reason("", rule); reason != "" {
		return []doc.Inline{doc.T(" — "), doc.Emph("accepted for every requirement:"), doc.T(" " + reason)}
	}
	return nil
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
	for _, name := range m.segments {
		d := m.docs.Get(name)
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
func (m *specModel) writeBoundary(d *doc.Doc) {
	if !m.can.Topology {
		omitted(d, "The boundary", "Not measured: this frontend reads no topology declarations, so what this system talks to is missing from this document.")
		return
	}
	if !m.topo.Declared() {
		omitted(d, "The boundary", "No topology is declared, so what this system talks to is stated nowhere.")
		return
	}

	d.H(2, "The boundary")

	if len(m.topo.Participants) > 0 {
		t := d.Table("Outside", "Kind", "Role")
		for _, p := range m.topo.Participants {
			t.Add(
				doc.Cell(doc.T(p.Name)),
				doc.Cell(doc.Tf("%s", p.Kind)),
				doc.Cell(doc.T(oneLine(p.Role))),
			)
		}
	}

	if len(m.topo.Channels) == 0 {
		return
	}
	d.H(3, "Every way across")
	t := d.Table("Channel", "Protocol", "Data", "Authentication", "In transit")
	for _, c := range m.topo.Channels {
		t.Add(
			doc.Cell(doc.Strong(c.Label), doc.Break{}, doc.Tf("%s → %s", c.From, c.To)),
			doc.Cell(doc.Tf("%s", c.Protocol)),
			doc.Cell(doc.T(oneLine(c.Data))),
			doc.Cell(doc.T(oneLine(c.Auth))),
			doc.Cell(doc.T(oneLine(c.Crypto))),
		)
	}
}

// writeSurface renders the addresses the system answers on.
//
// It is a section of its own rather than a part of the boundary above, because
// the two halves are known in different ways and a reader is entitled to know
// which is which. A channel is declared: no code states that an object store is
// somebody else's responsibility. An endpoint is recognised: the code that
// mounts it says everything there is to say, and this table is read out of it
// rather than maintained beside it.
//
// The last column is the reason the table is worth printing at all. An address
// on its own is routing; an address with the requirement behind it is the
// answer to the question every review actually asks, which is why this system
// answers here and on whose authority.
func (m *specModel) writeSurface(d *doc.Doc) {
	if !m.can.Endpoints {
		omitted(d, "What answers from outside", "Not measured: this frontend recognises no routes, so any address this system answers on is missing from this document.")
		return
	}
	if len(m.endpoints) == 0 {
		omitted(d, "What answers from outside", "No route was recognised: this system answers on no address of its own.")
		return
	}

	byConstruct := map[string][]string{}
	for req, targets := range m.cov.BySatisfier {
		for _, t := range targets {
			byConstruct[t.String()] = append(byConstruct[t.String()], req)
		}
	}

	// The wire shapes appear only where a dialect states them. A column of
	// dashes against a router that never reported them would look like a
	// surface that carries no bodies, when it is a surface nobody asked.
	shapes := false
	for _, e := range m.endpoints {
		if e.ShapesStated {
			shapes = true
			break
		}
	}

	d.H(2, "What answers from outside")
	var t *doc.Table
	if shapes {
		t = d.Table("Address", "Takes", "Returns", "Serves", "Asked for by")
	} else {
		t = d.Table("Address", "Serves", "Asked for by")
	}
	for _, e := range m.endpoints {
		address := []doc.Inline{doc.Code(e.Ref())}
		if e.Path == "" {
			// Not a gap in the table but the most important row in it: an
			// address that only exists at run time is one no catalogue can
			// name, and printing a blank would hide exactly that.
			address = []doc.Inline{doc.Emph("computed, not readable")}
		}

		var served []doc.Inline
		var asked []string
		for _, uc := range e.UseCases {
			short := lastSegment(uc)
			if len(served) > 0 {
				served = append(served, doc.T(", "))
			}
			served = append(served, doc.Code(short))
			asked = append(asked, byConstruct[uc]...)
			asked = append(asked, byConstruct[short]...)
		}
		if len(served) == 0 {
			served = orDash("")
		}
		if e.LeftScope {
			// The trace walked out of what this run loaded, so the cell is not
			// empty for want of a use case but for want of a look. The dash
			// that means "nothing is behind this" would be a different and
			// much worse claim, and a document that cannot tell them apart is
			// the document this tool replaces.
			served = []doc.Inline{doc.Emph("outside this scope")}
		}

		if shapes {
			t.Add(address,
				wireShape(e, e.Request), wireShape(e, e.Response),
				served, orDash(strings.Join(unique(asked), ", ")))
			continue
		}
		t.Add(address, served, orDash(strings.Join(unique(asked), ", ")))
	}
}

// wireShape renders what crosses the boundary, and says which kind of nothing
// an empty cell is.
//
// A route mounted through a builder that reports its types and carrying none
// promises no shape: it writes bytes, and a dash is the whole truth. A route
// mounted on a router that reports nothing is a route whose body this never
// asked about, and the two must not print alike.
func wireShape(e ir.Endpoint, name string) []doc.Inline {
	if name != "" {
		return []doc.Inline{doc.Code(lastSegment(name))}
	}
	if e.ShapesStated {
		return []doc.Inline{doc.T("—")}
	}
	return []doc.Inline{doc.Emph("not stated here")}
}

// orDash renders an empty cell as something a reader notices.
//
// A blank cell reads as nothing to say, and here it never is: a route with no
// use case is work the architecture has no name for, and a route with no
// requirement is an address nobody asked for.
func orDash(s string) []doc.Inline {
	if s == "" {
		return []doc.Inline{doc.Strong("nothing")}
	}
	return []doc.Inline{doc.T(s)}
}

// unique keeps the first occurrence of each entry, in order.
func unique(in []string) []string {
	seen := map[string]bool{}
	out := in[:0]
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// writeProcesses renders each course of business as its edges.
//
// The edge list rather than a numbered sequence, because the model is a graph
// and a numbered list would have to pick an order that does not exist. Where a
// process branches and comes back, any numbering is a lie about which step
// follows which.
func (m *specModel) writeProcesses(d *doc.Doc) {
	if !m.can.Processes {
		omitted(d, "Courses of business", "Not measured: this frontend reads no process declarations.")
		return
	}
	if len(m.processes) == 0 {
		omitted(d, "Courses of business", "No course of business is declared, so no requirement is placed in one.")
		return
	}

	d.H(2, "Courses of business")
	for _, p := range m.processes {
		d.H(3, or(p.Title, p.ID))
		if p.Purpose != "" {
			d.P(doc.T(oneLine(p.Purpose)))
		}
		d.P(doc.Code(p.ID), doc.T(" · drawn in "), doc.Code("process-"+p.ID+".puml"))

		if ids := m.satisfiedBy(p); len(ids) > 0 {
			d.Pf("Answers to: %s", strings.Join(ids, ", "))
		}

		t := d.Table("From", "To", "When")
		for _, e := range p.Edges {
			t.Add(m.nodeLabel(p, e.From), m.nodeLabel(p, e.To), doc.Cell(doc.T(or(e.When, "—"))))
		}
	}
}

// nodeLabel renders a node for the table: the construct it names where it names
// one, and what it is where it does not.
func (m *specModel) nodeLabel(p *ir.Process, id string) []doc.Inline {
	n, ok := p.Node(id)
	if !ok {
		return doc.Cell(doc.T(id))
	}
	kind := doc.Emph("(" + n.Kind.String() + ")")
	switch {
	case n.Ref != "":
		return doc.Cell(doc.T(lastSegment(n.Ref)), doc.T(" "), kind)
	case n.Label != "":
		return doc.Cell(doc.T(n.Label), doc.T(" "), kind)
	}
	return doc.Cell(doc.T(id), doc.T(" "), kind)
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
//
// It no longer escapes the pipe: a cell is a run of text now, and the renderer
// that knows what a separator looks like in its own format is the one that has
// to protect it.
func oneLine(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
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
func (m *specModel) writeStandards(d *doc.Doc) {
	if len(m.src.Standards) == 0 {
		omitted(d, "Standards", "No standard is declared, so no external clause is answered here.")
		return
	}

	d.H(2, "Standards")
	for _, st := range m.src.Standards {
		d.H(3, or(st.Title, st.Doc))

		line := []doc.Inline{doc.Tf("%d of %s answered", st.Answered,
			plural(st.Clauses, "applicable clause", "applicable clauses"))}
		if st.Excluded > 0 {
			line = append(line, doc.Tf(" · %d not applicable", st.Excluded))
		}
		line = append(line, doc.T(" · "), doc.Code(st.Doc))
		d.P(line...)

		sd := m.docs.Get(st.Doc)
		m.writeClauses(d, sd)
		m.writeApplicability(d, sd)
	}
}

func (m *specModel) writeClauses(d *doc.Doc, sd source.Document) {
	t := d.Table("Clause", "Obligation", "Answered by")
	for _, seg := range sd.Segments {
		if seg.Informative {
			continue
		}
		by := m.src.ByCiter[seg.Ref()]
		answer := doc.Cell(doc.Strong("nothing"))
		if len(by) > 0 {
			answer = doc.Cell(doc.T(strings.Join(by, ", ")))
		}
		t.Add(doc.Cell(doc.Code(seg.ID)), doc.Cell(doc.T(oneLine(seg.Title))), answer)
	}
}

// writeApplicability renders the exclusions with the reason each was given.
func (m *specModel) writeApplicability(d *doc.Doc, sd source.Document) {
	var excluded []source.Segment
	for _, seg := range sd.Segments {
		if seg.Informative {
			excluded = append(excluded, seg)
		}
	}
	if len(excluded) == 0 {
		return
	}

	d.H(4, "Statement of applicability")
	d.P(doc.T("Clauses excluded here, and why. An exclusion is a decision somebody made; the reason is the whole of what it is worth."))
	t := d.Table("Clause", "Obligation", "Does not apply because")
	for _, seg := range excluded {
		t.Add(
			doc.Cell(doc.Code(seg.ID)),
			doc.Cell(doc.T(oneLine(seg.Title))),
			doc.Cell(doc.T(oneLine(seg.Because))),
		)
	}
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
func (m *specModel) writeTopics(d *doc.Doc) {
	topics := m.tree.Topics()
	if !m.can.Topics {
		omitted(d, "Themes", "Not measured: this frontend reads no theme declarations.")
		return
	}
	if len(topics) == 0 {
		omitted(d, "Themes", "No theme is declared, so the requirements are not grouped.")
		return
	}

	filed := map[string]bool{}
	d.H(2, "Themes")

	for _, top := range topics {
		d.H(3, or(top.Title, top.ID))
		if top.Description != "" {
			d.P(doc.T(oneLine(top.Description)))
		}

		t := d.Table("ID", "Requirement")
		for _, id := range m.sortedRequirementIDs() {
			r := m.tree.ByID[id]
			if !hasTopic(r, top.ID) {
				continue
			}
			filed[id] = true
			t.Add(doc.Cell(doc.Code(r.ID)), doc.Cell(doc.T(oneLine(or(r.Title, r.Text)))))
		}
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

	d.H(3, "Under no theme")
	d.Pf("%s filed under no theme. Themes are optional, so this is not a defect — but the count belongs here, because a requirement left out of every chapter reads as one that does not exist.",
		plural(len(loose), "requirement is", "requirements are"))
	t := d.Table("ID", "Requirement")
	for _, id := range loose {
		r := m.tree.ByID[id]
		t.Add(doc.Cell(doc.Code(r.ID)), doc.Cell(doc.T(oneLine(or(r.Title, r.Text)))))
	}
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
func (m *specModel) writeDocuments(d *doc.Doc) {
	if len(m.segments) == 0 {
		omitted(d, "The material", "No source document is configured, so nothing here is traced back to prose.")
		return
	}

	d.H(2, "The material")
	t := d.Table("Document", "Kind", "Segments", "Cited", "Read", "Drifted").
		Aligned(doc.Left, doc.Left, doc.Right, doc.Right, doc.Right, doc.Right)

	for _, path := range m.segments {
		sd := m.docs.Get(path)
		if sd.Err != nil {
			t.Add(doc.Cell(doc.Code(path)), doc.Cell(doc.T("—")), doc.Cell(doc.T("—")),
				doc.Cell(doc.T("—")), doc.Cell(doc.T("—")), doc.Cell(doc.T("—")))
			continue
		}

		var total, cited, read, drifted int
		for _, seg := range sd.Segments {
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
		t.Add(doc.Cell(doc.Code(path)), doc.Cell(doc.Tf("%s", sd.Kind)),
			doc.Cell(doc.Tf("%d", total)), doc.Cell(doc.Tf("%d", cited)),
			doc.Cell(doc.Tf("%d", read)), doc.Cell(doc.Tf("%d", drifted)))
	}
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
func (m *specModel) writeImplementation(it *items, r *ir.Requirement) {
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

	it.add(doc.Strong("Implemented by"))
	for _, name := range names {
		c, known := byName[name]
		if !known || c.EndLine == 0 {
			it.under(doc.Code(name))
			continue
		}
		it.under(append([]doc.Inline{
			doc.Code(name), doc.T(" — "),
			doc.Code(fmt.Sprintf("%s:%d–%d", m.relative(c.Pos.File), c.Pos.Line, c.EndLine)),
		}, m.exercised(c)...)...)
	}
}

// exercised renders the share of a declaration a run reached, or nothing.
func (m *specModel) exercised(c ir.Construct) []doc.Inline {
	rec, ok := m.base.Constructs[c.Name]
	if !ok || rec.Statements == 0 || rec.Fingerprint != c.Fingerprint {
		return nil
	}
	return []doc.Inline{doc.Tf(", %d of %d statements exercised", rec.Covered, rec.Statements)}
}

// relative shortens an absolute source path to something a reader can find.
func (m *specModel) relative(file string) string {
	if rel, err := filepath.Rel(m.root, file); err == nil {
		return filepath.ToSlash(rel)
	}
	return file
}

// omitted states that a chapter is not here, and which kind of absence it is.
//
// A missing chapter used to be a blank: the writer returned early and the
// document simply had no such section. Three different facts printed as the
// same nothing — nothing was declared, nothing exists, or this frontend cannot
// read them at all — and an auditor seeing no boundary chapter concludes the
// system talks to nothing.
//
// The heading stays so the document keeps its shape between runs and a diff of
// two generations stays readable.
func omitted(d *doc.Doc, title, reason string) {
	d.H(2, title)
	d.P(doc.Emph(reason))
}
