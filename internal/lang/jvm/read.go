package jvm

import (
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/jvm/classfile"
)

// SpecPackage is where the annotation vocabulary is expected to live.
//
// There is no library to depend on and nothing to publish. A project declares
// these annotations itself — some forty lines of them — and speclink recognises
// them by their fully qualified name, exactly as it recognises a framework by
// its import path rather than by linking it. The reasoning is the same in both
// places: one binary serves projects that are not all on the same version of
// anything, so it must not be part of what they depend on.
//
// The cost is that nothing checks a project's annotations have the shape this
// expects. A missing argument reads as a missing value, which is why every read
// below reports what it did not find rather than defaulting quietly.
const SpecPackage = "speclink"

// Reader turns compiled classes into the language neutral model.
type Reader struct {
	classes []*Class
	pos     *positions
	pkg     string
}

// NewReader prepares a read over the given classes.
//
// specPackage overrides where the annotations live, for a project that would
// rather keep them under its own namespace.
func NewReader(root string, classes []*Class, sourceRoots []string, specPackage string) *Reader {
	if specPackage == "" {
		specPackage = SpecPackage
	}
	return &Reader{classes: classes, pos: newPositions(root, sourceRoots), pkg: specPackage}
}

func (r *Reader) annotationType(name string) string { return r.pkg + "." + name }

// ReadRequirements collects the requirement declarations.
//
// A requirement is a class carrying @Requirement. It is a class rather than a
// record or an enum constant because a Java annotation may hold only compile
// time constants, and only a class literal lets one requirement name another
// with the compiler checking the reference. With a string there, the derivation
// graph would be the one part of the model speclink had to verify itself — and
// an unverified reference is what this whole design exists to remove.
func (r *Reader) ReadRequirements(out *diag.Set) []*ir.Requirement {
	var reqs []*ir.Requirement

	for _, c := range r.classes {
		a, ok := find(c.Annotations, r.annotationType("Requirement"))
		if !ok {
			continue
		}
		file, line := r.pos.Of(c, c.Simple())
		pos := ir.Position{File: file, Line: line, Col: 1}

		req := &ir.Requirement{
			// The class is the identity a reference resolves to, the same role
			// the qualified Go identifier plays in the other frontend.
			GoIdent:   c.Name,
			ID:        a.Values["id"].String(),
			Title:     a.Values["title"].String(),
			Text:      a.Values["text"].String(),
			Rationale: a.Values["rationale"].String(),
			Pos:       pos,
		}
		req.Kind = readKind(a.Values["kind"])
		req.Status = readStatus(a.Values["status"])
		req.DerivedFrom = classRefs(a.Values["derivedFrom"])
		req.Supersedes = classRefs(a.Values["supersedes"])
		req.Sources = readSources(a.Values["sources"], pos)

		if req.ID == "" {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseResolve, 1),
				Pos:  pos,
				What: c.Simple() + " is annotated as a requirement but names no ID.",
				Why:  "The ID is the stable identity of a requirement. Everything that reports on one — a matrix, a gap list, a diff of the record — is keyed by it.",
				How:  `Set id, for example id = "R-QUOTE-SUBMIT".`,
			})
			continue
		}
		reqs = append(reqs, req)
	}

	sort.Slice(reqs, func(i, j int) bool { return reqs[i].ID < reqs[j].ID })
	return reqs
}

// ReadBindings collects the @Satisfies annotations.
//
// They may sit on a type or on a method, and the target follows: a type says
// the whole construct was written for the requirement, a method says one
// operation of it was.
func (r *Reader) ReadBindings(out *diag.Set) []ir.Binding {
	var bindings []ir.Binding

	for _, c := range r.classes {
		if c.IsSynthetic() {
			continue
		}
		if a, ok := find(c.Annotations, r.annotationType("Satisfies")); ok {
			file, line := r.pos.Of(c, c.Simple())
			bindings = append(bindings, ir.Binding{
				Target: ir.Target{Kind: ir.TargetType, Package: c.Package(), Name: c.Name},
				Assertions: []ir.Assertion{{
					Kind:         ir.AssertSatisfies,
					Requirements: classRefs(a.Values["value"]),
					Pos:          ir.Position{File: file, Line: line, Col: 1},
				}},
				Pos: ir.Position{File: file, Line: line, Col: 1},
			})
		}

		for _, m := range c.Methods {
			if m.IsSynthetic() {
				continue
			}
			a, ok := find(m.Annotations, r.annotationType("Satisfies"))
			if !ok {
				continue
			}
			file, line := r.pos.Of(c, m.Name)
			if m.Line > 0 {
				// A method does have a line in the class file, and it beats a
				// text search: the search finds the first mention of the name,
				// the table records where the code actually is.
				line = m.Line
			}
			bindings = append(bindings, ir.Binding{
				Target: ir.Target{Kind: ir.TargetFunc, Package: c.Package(), Name: c.Name + "." + m.Name},
				Assertions: []ir.Assertion{{
					Kind:         ir.AssertSatisfies,
					Requirements: classRefs(a.Values["value"]),
					Pos:          ir.Position{File: file, Line: line, Col: 1},
				}},
				Pos: ir.Position{File: file, Line: line, Col: 1},
			})
		}
	}
	return bindings
}

// classRefs returns the binary names of a Class<?>[] argument.
//
// A single element written without braces arrives as a bare value rather than a
// one element array, which is Java's shorthand and not a special case worth
// making every caller handle.
func classRefs(v classfile.Value) []string {
	var out []string
	for _, e := range elements(v) {
		if e.Kind == 'c' && e.Class != "" {
			out = append(out, e.Class)
		}
	}
	return out
}

func find(in []classfile.Annotation, typ string) (classfile.Annotation, bool) {
	for _, a := range in {
		if a.Type == typ {
			return a, true
		}
	}
	return classfile.Annotation{}, false
}

func readKind(v classfile.Value) ir.Kind {
	switch v.EnumConst {
	case "FUNCTIONAL":
		return ir.Functional
	case "NON_FUNCTIONAL":
		return ir.NonFunctional
	case "CONSTRAINT":
		return ir.Constraint
	case "DECISION":
		return ir.Decision
	}
	return 0
}

func readStatus(v classfile.Value) ir.Status {
	switch v.EnumConst {
	case "NORMATIVE":
		return ir.Normative
	case "ABSTRACT":
		return ir.Abstract
	case "PLANNED":
		return ir.Planned
	case "OUT_OF_SCOPE":
		return ir.OutOfScope
	case "INFORMATIVE":
		return ir.Informative
	case "SUPERSEDED":
		return ir.Superseded
	}
	return 0
}

func readSources(v classfile.Value, pos ir.Position) []ir.Source {
	var out []ir.Source
	for _, e := range elements(v) {
		if e.Nested == nil {
			continue
		}
		out = append(out, ir.Source{
			Doc:    e.Nested.Values["doc"].String(),
			Anchor: e.Nested.Values["anchor"].String(),
			Extern: e.Nested.Values["extern"].String(),
			Note:   e.Nested.Values["note"].String(),
			Pos:    pos,
		})
	}
	return out
}

// elements normalises an annotation argument that may be an array or a single
// value, which is Java's shorthand for a one element array.
func elements(v classfile.Value) []classfile.Value {
	if v.Kind == '[' {
		return v.Array
	}
	if v.Kind == 0 {
		return nil
	}
	return []classfile.Value{v}
}

// Dialect phrases the fixes in Java.
type Dialect struct{}

var _ ir.Dialect = Dialect{}

func (Dialect) BindConstruct(construct string) string {
	return "@Satisfies(…) on " + simple(construct)
}

func (Dialect) BindField(construct, field string) string {
	return "@Satisfies(…) on the field " + field + " of " + simple(construct)
}

func (Dialect) BindFieldOptional(construct, field string) string {
	return "@Optional on the field " + field + " of " + simple(construct)
}

func (Dialect) BindDecision(construct string) string {
	return "@Satisfies(RDecEventSourcing.class) on " + simple(construct)
}

// AnnotationFile is the file the binding goes in, which in Java is the file the
// declaration is already in.
//
// There is no sidecar and there cannot be one: an annotation attaches to the
// declaration it precedes, and Java requires a public type to be alone in a
// file named after it. What the Go frontend achieves with a neighbouring file,
// this achieves with adjacency in the same one.
func (Dialect) AnnotationFile(sourceFile string) string {
	if i := strings.LastIndexByte(sourceFile, '/'); i >= 0 {
		return sourceFile[i+1:]
	}
	if sourceFile == "" {
		return "the file declaring it"
	}
	return sourceFile
}

// RequirementFile turns a requirement ID into the class file it belongs in.
//
// R-QUOTE-SUBMIT becomes RQuoteSubmit.java, because a Java identifier admits no
// hyphens and a public class must be alone in a file of its own name. The rule
// underneath is the same as the other frontend's — one requirement per file,
// named after its ID — only the spelling differs.
func (Dialect) RequirementFile(id string) string { return ClassNameOf(id) + ".java" }

// ClassNameOf renders a requirement ID as a Java class name.
func ClassNameOf(id string) string {
	var b strings.Builder
	for _, part := range strings.Split(id, "-") {
		if part == "" {
			continue
		}
		b.WriteString(strings.ToUpper(part[:1]))
		b.WriteString(strings.ToLower(part[1:]))
	}
	return b.String()
}

func (Dialect) Verify(ref string) string  { return "@Verifies(" + simple(ref) + ".class)" }
func (Dialect) Satisfy(ref string) string { return "@Satisfies(" + simple(ref) + ".class)" }
func (Dialect) Waive(rule string) string  { return `@Waive(rule = "` + rule + `", reason = …)` }

// Transition spells the annotation naming the state an event leads to.
//
// The JVM frontend recognises no events yet, so nothing produces this line
// today. It is written out rather than left to panic because the interface is
// the contract: a dialect that answers "not applicable" by crashing would turn
// a rule that merely does not apply into a broken run.
func (Dialect) Transition(event, state string) string {
	return "@Transition(event = " + simple(event) + `.class, to = "` + state + `")`
}

func (Dialect) Term(name string) string { return "@" + name }

// Status names an enum constant, which Java spells in screaming snake case.
//
// The rules hand over the name as the model spells it — OutOfScope — and a
// plain upper casing turns that into OUTOFSCOPE, which does not exist and
// cannot be pasted. This is the small kind of wrong that makes a reader stop
// trusting the rest of the message.
func (Dialect) Status(name string) string { return "Status." + screamingSnake(name) }

// screamingSnake turns OutOfScope into OUT_OF_SCOPE.
func screamingSnake(name string) string {
	var b strings.Builder
	for i, r := range name {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}

func simple(qualified string) string {
	if i := strings.LastIndexByte(qualified, '.'); i >= 0 {
		return qualified[i+1:]
	}
	return qualified
}
