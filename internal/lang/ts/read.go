package ts

import (
	"path"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/ts/syntax"
)

// Requirements reads every requirement declaration of the project.
//
// A requirement is a constant initialised by a call of the vocabulary's
// requirement function:
//
//	export const RTsvActionOverview = requirement({ id: 'R-TSV-ACTION-OVERVIEW', … });
//
// A constant rather than a class, because unlike a Java annotation an object
// literal can hold a reference: derivedFrom: [RDecHexagonal] is checked by the
// compiler, which is the whole reason the carrier is code and not a document.
// The declaration is read and never run, so every value has to be a literal.
func (p *Project) Requirements(out *diag.Set) []*ir.Requirement {
	var reqs []*ir.Requirement
	for _, m := range p.Modules {
		for _, it := range m.Items() {
			if it.Kind != syntax.ItemVar {
				continue
			}
			for _, v := range it.Vars {
				call := it.File.Value(v.Init)
				if call.Kind != syntax.ValCall || m.vocabularyCall(call.Path) != "requirement" {
					continue
				}
				if r := p.readRequirement(m, v, call, out); r != nil {
					reqs = append(reqs, r)
				}
			}
		}
	}
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].ID < reqs[j].ID })
	return reqs
}

func (p *Project) readRequirement(m *Module, v syntax.Var, call syntax.Value, out *diag.Set) *ir.Requirement {
	pos := m.Pos(v.Pos)
	if len(call.Elems) != 1 || call.Elems[0].Kind != syntax.ValObject {
		notLiteral(out, pos, "the argument of requirement", "requirement({ id: 'R-…', kind: Kind.Functional, … })")
		return nil
	}
	obj := call.Elems[0]
	r := &ir.Requirement{Symbol: m.ID + "." + v.Name, Pos: pos}

	for _, f := range obj.Fields {
		val := f.Value
		fpos := m.Pos(f.Pos)
		switch f.Key {
		case "id":
			r.ID = p.str(val, fpos, "id", out)
		case "title":
			r.Title = p.str(val, fpos, "title", out)
		case "text":
			r.Text = p.str(val, fpos, "text", out)
		case "detail":
			r.Detail = p.str(val, fpos, "detail", out)
		case "rationale":
			r.Rationale = p.str(val, fpos, "rationale", out)
		case "consequences":
			r.Consequences = p.str(val, fpos, "consequences", out)
		case "kind":
			r.Kind = kindOf(p.enum(m, val, "Kind", fpos, out))
		case "status":
			r.Status = statusOf(p.enum(m, val, "Status", fpos, out))
		case "discipline":
			r.Discipline = disciplineOf(p.enum(m, val, "Discipline", fpos, out))
		case "derivedFrom":
			r.DerivedFrom = p.refs(m, val, fpos, "derivedFrom", out)
		case "supersedes":
			r.Supersedes = p.refs(m, val, fpos, "supersedes", out)
		case "sources":
			for _, e := range p.objects(val, fpos, "sources", out) {
				r.Sources = append(r.Sources, ir.Source{
					Doc:    p.optStr(e, "doc", fpos, out),
					Anchor: p.optStr(e, "anchor", fpos, out),
					Extern: p.optStr(e, "extern", fpos, out),
					Note:   p.optStr(e, "note", fpos, out),
					Pos:    m.Pos(e.Pos),
				})
			}
		case "attachments":
			for _, e := range p.objects(val, fpos, "attachments", out) {
				role, _ := e.Get("role")
				r.Attachments = append(r.Attachments, ir.Attachment{
					Path: p.optStr(e, "path", fpos, out),
					Role: roleOf(p.enum(m, role, "Role", fpos, out)),
					Note: p.optStr(e, "note", fpos, out),
				})
			}
		default:
			key := f.Key
			if key == "" {
				key = "a computed or spread property"
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseWhitelist, 21),
				Pos:  fpos,
				What: key + " is not part of a requirement declaration.",
				Why:  "A requirement is read, never run, and only the fields of the vocabulary are read. Anything else would be silently ignored, which is worse than refused.",
				How:  "Use id, kind, status, discipline, title, text, detail, rationale, consequences, derivedFrom, supersedes, sources or attachments.",
			})
		}
	}

	if r.ID == "" {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseResolve, 1),
			Pos:  pos,
			What: v.Name + " is declared as a requirement but names no ID.",
			Why:  "The ID is the stable identity of a requirement. Everything that reports on one — a matrix, a gap list, a diff of the record — is keyed by it.",
			How:  "Set id, for example id: 'R-QUOTE-SUBMIT'.",
		})
		return nil
	}
	return r
}

func notLiteral(out *diag.Set, pos ir.Position, what, how string) {
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseWhitelist, 20),
		Pos:  pos,
		What: what + " is not a literal.",
		Why:  "Declarations are read, never run. A value that would have to be evaluated to be known cannot be read, and guessing it would put something into the specification nobody wrote.",
		How:  "Write it out: " + how + ".",
	})
}

func (p *Project) str(v syntax.Value, pos ir.Position, field string, out *diag.Set) string {
	if v.Kind != syntax.ValString {
		notLiteral(out, pos, field, field+": '…' as a string, or strings joined with +")
		return ""
	}
	return v.Str
}

func (p *Project) optStr(obj syntax.Value, key string, pos ir.Position, out *diag.Set) string {
	v, ok := obj.Get(key)
	if !ok {
		return ""
	}
	return p.str(v, pos, key, out)
}

// enum reads Kind.Functional and the like, insisting that the enum is the
// vocabulary's.
func (p *Project) enum(m *Module, v syntax.Value, enum string, pos ir.Position, out *diag.Set) string {
	if v.Kind == syntax.ValMember {
		if e, member := m.vocabularyName(v.Path); e == enum {
			return member
		}
	}
	notLiteral(out, pos, "this "+strings.ToLower(enum), enum+".<Value>, with "+enum+" imported from "+VocabularyFile)
	return ""
}

// refs reads a list of requirement references.
func (p *Project) refs(m *Module, v syntax.Value, pos ir.Position, field string, out *diag.Set) []string {
	if v.Kind != syntax.ValArray {
		notLiteral(out, pos, field, field+": [RQuoteSubmit, …]")
		return nil
	}
	var refs []string
	for _, e := range v.Elems {
		if s, ok := p.reference(m, e); ok {
			refs = append(refs, s)
			continue
		}
		unresolved(out, m.Pos(e.Pos), e)
	}
	return refs
}

// reference qualifies a reference to a declaration.
func (p *Project) reference(m *Module, v syntax.Value) (string, bool) {
	if v.Kind != syntax.ValIdent && v.Kind != syntax.ValMember {
		return "", false
	}
	return p.symbol(m, v.Path)
}

func unresolved(out *diag.Set, pos ir.Position, v syntax.Value) {
	name := strings.Join(v.Path, ".")
	if name == "" {
		name = "this value"
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseBinding, 20),
		Pos:  pos,
		What: name + " does not name a declaration of this project.",
		Why:  "A reference is only worth something if it resolves to the thing it names. One that does not is a string with extra steps, which is what the carrier being code exists to rule out.",
		How:  "Import the declaration it means from the module that declares it, and name it here.",
	})
}

func (p *Project) objects(v syntax.Value, pos ir.Position, field string, out *diag.Set) []syntax.Value {
	if v.Kind != syntax.ValArray {
		notLiteral(out, pos, field, field+": [{ … }]")
		return nil
	}
	var objs []syntax.Value
	for _, e := range v.Elems {
		if e.Kind != syntax.ValObject {
			notLiteral(out, pos, "an element of "+field, field+": [{ … }]")
			continue
		}
		objs = append(objs, e)
	}
	return objs
}

func kindOf(s string) ir.Kind {
	switch s {
	case "Functional":
		return ir.Functional
	case "NonFunctional":
		return ir.NonFunctional
	case "Constraint":
		return ir.Constraint
	case "Decision":
		return ir.Decision
	}
	return 0
}

func statusOf(s string) ir.Status {
	switch s {
	case "Normative":
		return ir.Normative
	case "Abstract":
		return ir.Abstract
	case "Planned":
		return ir.Planned
	case "OutOfScope":
		return ir.OutOfScope
	case "Informative":
		return ir.Informative
	case "Superseded":
		return ir.Superseded
	}
	return 0
}

func disciplineOf(s string) ir.Discipline {
	switch s {
	case "Business":
		return ir.Business
	case "Technical":
		return ir.Technical
	case "Mixed":
		return ir.Mixed
	}
	return 0
}

func roleOf(s string) ir.Role {
	switch s {
	case "Mockup":
		return ir.Mockup
	case "Scribble":
		return ir.Scribble
	case "Diagram":
		return ir.Diagram
	case "AcceptanceCriteria":
		return ir.AcceptanceCriteria
	case "Protocol":
		return ir.Protocol
	case "Document":
		return ir.Document
	}
	return 0
}

// Bindings reads the bind and bindModule statements of the project.
//
//	bind(LoadActionOverviewUseCase, satisfies(RTsvActionOverview));
//	bindModule(waive('K6-DOMAIN-NO-FRAMEWORK', '…'));
//
// The first names a declaration by value, so the compiler checks it exists.
// The second annotates the module the annotation file sits beside, which is
// how a rule about a whole module is waived.
func (p *Project) Bindings(out *diag.Set) []ir.Binding {
	var bindings []ir.Binding
	for _, m := range p.Modules {
		for _, it := range m.Items() {
			if it.Kind != syntax.ItemExpr {
				continue
			}
			call := it.File.Value(it.Body)
			if call.Kind != syntax.ValCall {
				continue
			}
			var b ir.Binding
			args := call.Elems
			switch m.vocabularyCall(call.Path) {
			case "bind":
				if len(args) == 0 {
					unresolved(out, m.Pos(call.Pos), syntax.Value{})
					continue
				}
				sym, ok := p.reference(m, args[0])
				if !ok {
					unresolved(out, m.Pos(args[0].Pos), args[0])
					continue
				}
				owner := sym[:strings.LastIndexByte(sym, '.')]
				b.Target = ir.Target{Kind: ir.TargetType, Package: path.Dir(owner), Name: sym}
				args = args[1:]
			case "bindModule":
				annotated := p.annotated(m)
				if annotated == nil {
					out.Add(diag.Finding{
						Code: diag.Code(diag.PhaseBinding, 21),
						Pos:  m.Pos(call.Pos),
						What: "bindModule in " + m.Path + " annotates no module.",
						Why:  "bindModule speaks about the module its annotation file sits beside: Foo.annotation.ts about Foo.ts or Foo.vue. There is none here, so the statement is about nothing.",
						How:  "Move it into <module>.annotation.ts next to the module it is about.",
					})
					continue
				}
				b.Target = ir.Target{Kind: ir.TargetPackage, Package: annotated.Dir, Name: annotated.ID}
			default:
				continue
			}
			b.Pos = m.Pos(call.Pos)
			for _, a := range args {
				if as, ok := p.assertion(m, a, out); ok {
					b.Assertions = append(b.Assertions, as)
				}
			}
			bindings = append(bindings, b)
		}
	}
	return bindings
}

// annotated returns the module an annotation file is about.
func (p *Project) annotated(m *Module) *Module {
	base, ok := strings.CutSuffix(m.Path, ".annotation.ts")
	if !ok {
		return nil
	}
	for _, ext := range []string{".ts", ".vue"} {
		if t := p.byPath[base+ext]; t != nil {
			return t
		}
	}
	return nil
}

func (p *Project) assertion(m *Module, v syntax.Value, out *diag.Set) (ir.Assertion, bool) {
	pos := m.Pos(v.Pos)
	name := ""
	if v.Kind == syntax.ValCall {
		name = m.vocabularyCall(v.Path)
	}
	switch name {
	case "satisfies":
		a := ir.Assertion{Kind: ir.AssertSatisfies, Pos: pos}
		for _, e := range v.Elems {
			if s, ok := p.reference(m, e); ok {
				a.Requirements = append(a.Requirements, s)
				continue
			}
			unresolved(out, m.Pos(e.Pos), e)
		}
		return a, true
	case "waive":
		var rule, reason string
		if len(v.Elems) > 0 {
			rule = p.str(v.Elems[0], pos, "the rule of waive", out)
		}
		if len(v.Elems) > 1 {
			reason = p.str(v.Elems[1], pos, "the reason of waive", out)
		}
		if reason == "" {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseBinding, 3),
				Pos:  pos,
				What: "waive requires a reason.",
				Why:  "A waiver without justification is an invisible gap. The reason is carried into the gap report so the exemption leaves a trace.",
				How:  "State why this rule cannot hold here, e.g. waive('" + rule + "', 'logging out has to work without access to this application').",
			})
		}
		return ir.Assertion{Kind: ir.AssertWaive, Pos: pos, Rule: rule, Text: reason}, true
	case "rationale":
		text := ""
		if len(v.Elems) > 0 {
			text = p.str(v.Elems[0], pos, "the text of rationale", out)
		}
		return ir.Assertion{Kind: ir.AssertRationale, Pos: pos, Text: text}, true
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseBinding, 22),
		Pos:  pos,
		What: "this is not an assertion of the vocabulary.",
		Why:  "A binding is read, never run, and only satisfies, waive and rationale are understood by this profile. Anything else would be silently ignored, which is worse than refused.",
		How:  "Use satisfies(…), waive('<rule>', '<reason>') or rationale('…'), imported from " + VocabularyFile + ".",
	})
	return ir.Assertion{}, false
}

// CheckSyntax keeps the specification files in the subset they are read in.
//
// A requirement file holds imports and requirement declarations, an
// annotation file imports and bindings. Anything else in them would be code
// that runs, which is the one thing a declaration read without running it
// cannot account for.
func (p *Project) CheckSyntax(out *diag.Set) {
	for _, m := range p.Modules {
		req := strings.HasSuffix(m.Path, ".req.ts")
		ann := strings.HasSuffix(m.Path, ".annotation.ts")
		test := p.Layer(m) == LayerTest
		for _, it := range m.Items() {
			pos := m.Pos(it.Pos)
			call := syntax.Value{}
			switch it.Kind {
			case syntax.ItemExpr:
				call = it.File.Value(it.Body)
			case syntax.ItemVar:
				if len(it.Vars) == 1 {
					call = it.File.Value(it.Vars[0].Init)
				}
			}
			vocab := ""
			if call.Kind == syntax.ValCall {
				vocab = m.vocabularyCall(call.Path)
			}

			switch {
			case it.Kind == syntax.ItemImport:
			case req && it.Kind == syntax.ItemVar && it.VarKind == "const" && vocab == "requirement":
			case ann && it.Kind == syntax.ItemExpr && (vocab == "bind" || vocab == "bindModule"):
			case req:
				out.Add(diag.Finding{
					Code: diag.Code(diag.PhaseWhitelist, 22),
					Pos:  pos,
					What: "a requirement file holds a " + it.Kind.String() + ".",
					Why:  "A requirement file is read, never run. It holds imports and const declarations of requirement(…), so that what it says can be known without executing it.",
					How:  "Move this into a module of its own and keep only the requirement here.",
				})
			case ann:
				out.Add(diag.Finding{
					Code: diag.Code(diag.PhaseWhitelist, 23),
					Pos:  pos,
					What: "an annotation file holds a " + it.Kind.String() + ".",
					Why:  "An annotation file is read, never run. It holds imports and bind(…) or bindModule(…) statements, so that what it says can be known without executing it.",
					How:  "Move this into the module it belongs to and keep only bindings here.",
				})
			case it.Kind == syntax.ItemExpr && (vocab == "bind" || vocab == "bindModule"):
				out.Add(diag.Finding{
					Code: diag.Code(diag.PhaseWhitelist, 24),
					Pos:  pos,
					What: vocab + " is called in " + path.Base(m.Path) + ", which is not an annotation file.",
					Why:  "Bindings live beside the code they are about, in a file of their own, so that the code stays free of them and a reader finds them in one place.",
					How:  "Move it into " + annotationFileFor(m.Path) + ".",
				})
			}
			if !test {
				p.checkVerifiedOutsideTests(m, it, out)
			}
		}
	}
}

// checkVerifiedOutsideTests refuses a verification claim in production code,
// where no test result can ever stand behind it.
func (p *Project) checkVerifiedOutsideTests(m *Module, it itemOf, out *diag.Set) {
	toks := it.File.Tokens(it.Range)
	for i, t := range toks {
		if t.Kind != syntax.Ident || i+1 >= len(toks) || !toks[i+1].Is("(") || i > 0 && toks[i-1].Is(".") {
			continue
		}
		if m.vocabularyCall([]string{t.Text}) != "verified" {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseWhitelist, 25),
			Pos:  m.Pos(t.Pos),
			What: "verified is called outside a test.",
			Why:  "verified says that a test demonstrated a requirement. Outside a test there is no result that could ever confirm it.",
			How:  "Move the call into the test that demonstrates the requirement.",
		})
	}
}

func annotationFileFor(file string) string {
	base := path.Base(file)
	for _, ext := range []string{".ts", ".vue", ".js"} {
		if b, ok := strings.CutSuffix(base, ext); ok {
			return b + ".annotation.ts"
		}
	}
	return base + ".annotation.ts"
}
