package golang

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/worldiety/speclink/internal/ir"
)

// The architectural roles of a project without a framework.
//
// Four rather than nago's eight, and the four that are missing are missing for
// a reason rather than out of thrift. Command, event and projection are the
// vocabulary of event sourcing, and an architecture that stores current state
// has nothing for those words to name. Query is absent because nothing tells it
// from a use case: nago separates them by whether a commit sequence comes back,
// and here every use case returns (Out, error). Inventing a distinction the
// code does not carry would be guessing.
var (
	// A named func type whose first parameter is a subject.
	ConstructBareUseCase = ir.NewConstructKind("use case", "a use case",
		ir.NeedsRequirement(), ir.PerformsWork())

	// A consistency boundary with an identity, or a type marked as storage.
	ConstructBareAggregate = ir.NewConstructKind("aggregate", "an aggregate",
		ir.IsDomainModel(), ir.EmbodiesStorageDecision())

	// A named type over data.Repository, or an interface marked as storage.
	ConstructBareRepository = ir.NewConstructKind("repository", "a repository",
		ir.EmbodiesStorageDecision())

	// Declared by permission.Declare and bound to a use case through its type
	// parameter.
	ConstructBarePermission = ir.NewConstructKind("permission", "a permission")
)

// InferBare recognises the roles of a project that has no framework.
//
// It reads the same shapes as the nago recogniser and fewer of them, with one
// addition: a type carrying spec.Persistence is storage. That term exists
// because this architecture cannot say it any other way — nago hands out a
// Repository type and thereby states it, a hand written interface states
// nothing — and it is the only place here where a fact is annotated rather than
// inferred.
func (p *Package) InferBare(f Framework, marked map[string]bool) []ir.Construct {
	var out []ir.Construct

	for _, file := range p.pkg.Syntax {
		if p.isGeneratedByUs(file) {
			continue
		}
		for _, decl := range file.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, s := range gd.Specs {
				ts, ok := s.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if c, ok := p.inferBareType(ts, f, marked); ok {
					out = append(out, c)
				}
			}
		}
	}
	return append(out, p.inferPermissionsIn(f.Permission, ConstructBarePermission)...)
}

// inferBareType classifies one declared type.
//
// The order is from the most specific evidence to the least. A marked interface
// is a repository before anything else can claim it, because the mark is a
// statement somebody made rather than a shape that happened to match.
func (p *Package) inferBareType(ts *ast.TypeSpec, f Framework, marked map[string]bool) (ir.Construct, bool) {
	obj, ok := p.pkg.TypesInfo.Defs[ts.Name]
	if !ok || obj == nil {
		return ir.Construct{}, false
	}
	named, ok := obj.Type().(*types.Named)
	if !ok {
		return ir.Construct{}, false
	}

	base := ir.Construct{
		Name:    p.PkgPath() + "." + ts.Name.Name,
		Package: p.PkgPath(),
		Pos:     p.pos(ts.Name.Pos()),
	}

	if marked[base.Name] {
		if types.IsInterface(named.Underlying()) {
			base.Kind = ConstructBareRepository
			base.Evidence = "is marked as storage with spec.Persistence, and is an interface, so it is a port"
		} else {
			base.Kind = ConstructBareAggregate
			base.Evidence = "is marked as storage with spec.Persistence"
			base.Fields = p.fieldsOf(named)
		}
		return base, true
	}

	if f.Data != "" {
		if inst, ok := p.repositoryInstanceIn(ts, f.Data); ok {
			base.Kind = ConstructBareRepository
			base.Evidence = "stands for " + inst
			return base, true
		}
	}

	if p.hasMethods(named, "Identity") {
		base.Kind = ConstructBareAggregate
		base.Evidence = "has an Identity method, which is what makes it an aggregate root"
		base.Fields = p.fieldsOf(named)
		return base, true
	}

	sig, ok := named.Underlying().(*types.Signature)
	if !ok || !f.isSubject(firstParam(sig)) {
		return ir.Construct{}, false
	}
	base.Kind = ConstructBareUseCase
	base.Evidence = "is a named func type taking a subject as its first parameter"
	return base, true
}

// fieldsOf returns the exported fields of a struct type, empty for anything
// else. A domain model states what the system believes about the thing it
// describes, and that is what the field level rules work from.
func (p *Package) fieldsOf(named *types.Named) []ir.SchemaField {
	st, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil
	}
	return p.readFields(st)
}

// firstParam returns the type of a signature's first parameter, nil when it has
// none.
func firstParam(sig *types.Signature) types.Type {
	if sig == nil || sig.Params().Len() == 0 {
		return nil
	}
	return sig.Params().At(0).Type()
}

// PersistenceMarks returns the qualified names of the types a package declared
// as storage.
//
// Read from the bindings rather than from the types, because that is where the
// statement is. It is collected across the whole project before anything is
// inferred: an annotation file names its neighbour, and the neighbour may sit
// in a package the recogniser reaches first.
func PersistenceMarks(bindings []ir.Binding) map[string]bool {
	out := map[string]bool{}
	for _, b := range bindings {
		if b.Target.Kind != ir.TargetType {
			continue
		}
		for _, a := range b.Assertions {
			if a.Kind == ir.AssertPersistence {
				out[b.Target.Name] = true
			}
		}
	}
	return out
}

// StoredForms maps a domain type to the shape it is written down as.
//
// Read from the bindings for the same reason PersistenceMarks is: the fact
// lives in the annotation, not in the types. Two structs in two packages look
// identical whether one maps to the other or neither knows the other exists.
func StoredForms(bindings []ir.Binding) map[string]string {
	out := map[string]string{}
	for _, b := range bindings {
		if b.Target.Kind != ir.TargetType {
			continue
		}
		for _, a := range b.Assertions {
			if a.Kind == ir.AssertStoredAs && a.DomainType != "" {
				out[a.DomainType] = b.Target.Name
			}
		}
	}
	return out
}

// RepositoryElements returns the domain types this package's repositories store.
//
// A repository is declared as `type R data.Repository[E, ID]`, and E is what
// ends up on disk. This is the whole of what a project without a framework
// offers: there is no constructor naming a persistence model, so the element
// type of the port is the only structural statement that anything is stored at
// all.
//
// The element type is read from the syntax rather than the underlying type,
// for the reason repositoryInstance already documents: a defined interface type
// erases the generic it came from.
func (p *Package) RepositoryElements(f Framework) []string {
	if f.Data == "" {
		return nil
	}

	var out []string
	for _, file := range p.pkg.Syntax {
		if p.isGeneratedByUs(file) {
			continue
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, s := range gen.Specs {
				ts, ok := s.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if _, isRepo := p.repositoryInstanceIn(ts, f.Data); !isRepo {
					continue
				}
				if name := p.elementOf(ts); name != "" {
					out = append(out, name)
				}
			}
		}
	}
	return out
}

// elementOf returns the qualified name of the first type argument of a
// repository declaration, which is the aggregate it stores.
func (p *Package) elementOf(ts *ast.TypeSpec) string {
	var first ast.Expr
	switch t := ts.Type.(type) {
	case *ast.IndexListExpr:
		if len(t.Indices) == 0 {
			return ""
		}
		first = t.Indices[0]
	case *ast.IndexExpr:
		first = t.Index
	default:
		return ""
	}
	if tv, ok := p.pkg.TypesInfo.Types[first]; ok {
		return typeName(tv.Type)
	}
	return ""
}
