package jvm

import (
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/jvm/classfile"
)

// Infer recognises the architectural roles of a Spring project.
//
// This is the capability that makes forward coverage possible, and the reason
// it is worth having at all: without it there is no set of constructs to hold
// accountable, so "every construct names a requirement" is not a weaker claim
// but no claim. An annotation that was forgotten leaves nothing behind to
// notice — which is the failure every marker based traceability tool has, and
// the one thing speclink was built not to have.
//
// It also answers a question the Go frontend could not. There, recognition
// meant reasoning about method sets, type aliases and signatures, and the
// suspicion was that inference only worked because nago's shapes happened to be
// legible. Spring declares its architecture in annotations, and annotations are
// exactly what a declaration level reader sees. Inference is not a property of
// nago; it is a property of a framework that says what it is.
func (r *Reader) Infer() []ir.Construct {
	var out []ir.Construct

	for _, c := range r.classes {
		if c.IsSynthetic() {
			continue
		}
		switch {
		case hasAny(c.Annotations, springWeb+".RestController", springStereotype+".Controller"):
			out = append(out, r.endpoints(c)...)
		case has(c.Annotations, springStereotype+".Service"):
			out = append(out, r.operations(c)...)
		}
		if k, ok := r.typeRole(c); ok {
			out = append(out, r.typeConstruct(c, k))
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// typeRole classifies a whole type.
func (r *Reader) typeRole(c *Class) (ir.ConstructKind, bool) {
	switch {
	case hasAny(c.Annotations, jakartaPersist+".Entity", javaxPersist+".Entity"):
		return ConstructEntity, true
	case has(c.Annotations, springStereotype+".Repository"):
		return ConstructRepository, true
	case r.extendsSpringData(c):
		return ConstructRepository, true
	}
	return ir.ConstructKind{}, false
}

// extendsSpringData reports whether an interface is a Spring Data repository.
//
// The check is on the directly named interfaces, not on the transitive
// hierarchy. A class file names its supertypes outright, so following the chain
// would work — but it would leave the project and walk into Spring itself,
// where the interfaces this recognises are declared, and a project that has not
// been compiled against Spring would then be classified by whether its jars
// happened to be around. Matching what the source actually wrote is both
// cheaper and the same answer in every case anybody writes.
func (r *Reader) extendsSpringData(c *Class) bool {
	for _, name := range c.Interfaces {
		if name == springData+".Repository" || name == springData+".CrudRepository" {
			return true
		}
		// Spring Data's own subinterfaces — JpaRepository, PagingAndSorting —
		// all live under the same package and all mean the same thing here.
		if strings.HasPrefix(name, springData+".") && strings.HasSuffix(name, "Repository") {
			return true
		}
		if strings.HasPrefix(name, "org.springframework.data.") && strings.HasSuffix(name, "Repository") {
			return true
		}
	}
	return false
}

func (r *Reader) typeConstruct(c *Class, kind ir.ConstructKind) ir.Construct {
	file, line := r.pos.Of(c, c.Simple())
	construct := ir.Construct{
		Kind:     kind,
		Name:     c.Name,
		Package:  c.Package(),
		Evidence: evidenceFor(kind),
		Pos:      ir.Position{File: file, Line: line, Col: 1},
	}
	if kind.IsDomainModel() {
		construct.Fields = r.fields(c)
	}
	return construct
}

// endpoints returns the request mapped methods of a controller.
func (r *Reader) endpoints(c *Class) []ir.Construct {
	var out []ir.Construct
	names := methodNames(c)

	for _, m := range c.Methods {
		if !declared(m) || !hasAny(m.Annotations, mappingAnnotations...) {
			continue
		}
		out = append(out, r.methodConstruct(c, m, names, ConstructEndpoint,
			"answers a request on "+strings.TrimPrefix(mappingOf(m), springWeb+".")))
	}
	return out
}

// operations returns the public methods of an application service.
//
// Every one of them, which is demanding on purpose. A public method of an
// application service is one operation somebody asked for, and that is what a
// use case is — the same reading that makes a named func type a use case in the
// other frontend. A service with twelve public methods is twelve things the
// specification has to account for, and if that is noisy the noise is the
// finding.
func (r *Reader) operations(c *Class) []ir.Construct {
	var out []ir.Construct
	names := methodNames(c)

	for _, m := range c.Methods {
		if !declared(m) || m.Access&classfile.AccPublic == 0 || isObjectMethod(m) {
			continue
		}
		out = append(out, r.methodConstruct(c, m, names, ConstructService,
			"is a public method of a class annotated as a service"))
	}
	return out
}

func (r *Reader) methodConstruct(c *Class, m classfile.Member, names map[string]int, kind ir.ConstructKind, evidence string) ir.Construct {
	file, line := r.pos.Of(c, m.Name)
	if m.Line > 0 {
		// A method does have a line in the class file, and it beats a text
		// search: the search finds the first mention of the name, the table
		// records where the code is.
		line = m.Line
	}
	return ir.Construct{
		Kind:     kind,
		Name:     c.Name + "." + uniqueName(m, names),
		Package:  c.Package(),
		Evidence: evidence,
		Pos:      ir.Position{File: file, Line: line, Col: 1},
	}
}

// uniqueName returns the method name, made unique when the class overloads it.
//
// Java allows two methods of one name and Go does not, so this is a problem the
// other frontend never had. Collapsing them would be the wrong kind of wrong:
// annotating one overload would mark both as bound, and a construct that nobody
// wrote a requirement for would report as covered. The descriptor is appended
// only where there is an ambiguity, so ordinary names stay readable.
func uniqueName(m classfile.Member, names map[string]int) string {
	if names[m.Name] < 2 {
		return m.Name
	}
	return m.Name + m.Descriptor
}

func methodNames(c *Class) map[string]int {
	out := map[string]int{}
	for _, m := range c.Methods {
		if declared(m) {
			out[m.Name]++
		}
	}
	return out
}

// fields returns the declared instance fields of a type.
func (r *Reader) fields(c *Class) []ir.SchemaField {
	var out []ir.SchemaField
	for _, f := range c.Fields {
		if f.IsSynthetic() || f.Access&classfile.AccStatic != 0 {
			continue
		}
		file, line := r.pos.Of(c, f.Name)
		out = append(out, ir.SchemaField{
			Name: f.Name,
			Wire: f.Name,
			Pos:  ir.Position{File: file, Line: line, Col: 1},
		})
	}
	return out
}

// declared filters out what a compiler produced rather than a person wrote:
// constructors, static initialisers, bridges and the rest.
func declared(m classfile.Member) bool {
	return !m.IsSynthetic() && !strings.HasPrefix(m.Name, "<") && !strings.Contains(m.Name, "$")
}

// isObjectMethod reports whether a method only overrides Object.
//
// They are on every class and mean nothing about the architecture. Matching by
// name and descriptor rather than by name alone, so a service that genuinely
// declares an operation called toString(String) is not lost with them.
func isObjectMethod(m classfile.Member) bool {
	switch m.Name + m.Descriptor {
	case "toString()Ljava/lang/String;", "hashCode()I", "equals(Ljava/lang/Object;)Z":
		return true
	}
	return false
}

func mappingOf(m classfile.Member) string {
	for _, want := range mappingAnnotations {
		if has(m.Annotations, want) {
			return want
		}
	}
	return "a request mapping"
}

func evidenceFor(kind ir.ConstructKind) string {
	switch kind {
	case ConstructEntity:
		return "is annotated as a persistent entity"
	case ConstructRepository:
		return "is annotated as a repository, or extends one from Spring Data"
	}
	return "is recognised by the framework annotations it carries"
}

func has(in []classfile.Annotation, typ string) bool {
	_, ok := find(in, typ)
	return ok
}

func hasAny(in []classfile.Annotation, types ...string) bool {
	for _, t := range types {
		if has(in, t) {
			return true
		}
	}
	return false
}

// Constructs implements lang.ConstructInferrer.
func (m *Model) Constructs(out *diag.Set) []ir.Construct { return m.r.Infer() }
