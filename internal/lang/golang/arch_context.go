package golang

import (
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

const (
	// RuleCtxNoUIImport fires when a domain package reaches into the user
	// interface.
	RuleCtxNoUIImport = "K6-CTX-NO-UI-IMPORT"
	// RuleCtxUIPackage fires when the ui directory does not declare ui<ctx>.
	RuleCtxUIPackage = "K6-CTX-UI-PKG"
	// RuleCtxUseCases fires when the use cases of a context are not bundled.
	RuleCtxUseCases = "K6-CTX-USECASES"
)

// CheckBoundedContexts verifies the layout and the dependency direction of the
// bounded contexts.
func CheckBoundedContexts(pkgs []*Package, cfg config.Config, root string, out *diag.Set) {
	for _, p := range pkgs {
		rel := p.relDir(root)
		ctx, inContext := cfg.InContextRoot(rel)
		if !inContext || cfg.Excluded(rel) {
			continue
		}

		switch role := contextRole(rel, cfg); role {
		case roleUI:
			// Only the ui directory itself carries the naming rule. The
			// presentation layer of a context is free to be more than one
			// package — an editor for one widget, a shared table renderer —
			// and those are ordinary Go packages named after what they do.
			// Demanding uiplatform of every one of them would force a dozen
			// identically named packages that could then only be imported
			// through aliases, which is the very thing the rule prevents.
			if isUIDir(rel) {
				p.checkUIPackageName(ctx, out)
			}
		case roleDomain:
			p.checkNoUIImport(rel, out)
			p.checkUseCaseBundle(ctx, out)
		case roleWiring:
			// The wiring layer connects views to use cases and therefore has
			// to see both. Excluding it is not a concession: it is the one
			// place where the two sides are supposed to meet.
		}
	}
}

type ctxRole int

const (
	roleDomain ctxRole = iota + 1
	roleUI
	roleWiring
)

// contextRole classifies a package inside a bounded context by its path.
func contextRole(rel string, cfg config.Config) ctxRole {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, part := range parts {
		switch part {
		case "ui":
			return roleUI
		case "cfg":
			return roleWiring
		}
	}
	return roleDomain
}

// isUIDir reports whether this directory is the ui directory of a context,
// rather than a package nested below it.
func isUIDir(rel string) bool {
	return filepath.Base(filepath.ToSlash(rel)) == "ui"
}

// checkUIPackageName enforces that a directory named ui declares ui<ctx>.
//
// Directory and package name carry the same fact twice, which is harmless
// exactly as long as it is checked. Unchecked, the two drift apart and imports
// start needing aliases nobody can predict.
func (p *Package) checkUIPackageName(ctx string, out *diag.Set) {
	want := "ui" + strings.ToLower(ctx)
	if p.pkg.Name == want {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 40),
		Pos:  p.filePos(),
		Rule: RuleCtxUIPackage,
		What: "the ui directory of " + ctx + " declares package " + p.pkg.Name + ", expected " + want + ".",
		Why:  "The directory is always called ui, so the package name is what tells a reader and an import which context's interface this is.",
		How:  "Rename the package to " + want + ".",
	})
}

// checkNoUIImport enforces the dependency direction of the architecture.
//
// A bounded context defines what the system does; the user interface is one way
// of reaching it, and there may be others. Once the domain imports the
// presentation layer that ordering is lost, the context can no longer be tested
// or reused without a renderer, and the two evolve as one lump.
func (p *Package) checkNoUIImport(rel string, out *diag.Set) {
	for _, f := range p.pkg.Syntax {
		for _, imp := range f.Imports {
			path, err := unquote(imp.Path.Value)
			if err != nil || !isUIPackage(path) {
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 41),
				Pos:  p.pos(imp.Pos()),
				Rule: RuleCtxNoUIImport,
				What: rel + " is a domain package but imports " + path + ".",
				Why:  "The domain defines what the system does; the interface is one way of reaching it. Importing it inverts that and makes the context untestable without a renderer.",
				How:  "Move the view code into the ui package of this context, or into its cfg package if it is wiring.",
			})
		}
	}
}

// checkUseCaseBundle enforces that the use cases of a context are collected in
// a UseCases struct built by NewUseCases.
//
// The bundle is what lets a caller depend on the capabilities of a context
// rather than on its internals, and it is the single place where the shared
// dependencies of a context are threaded through.
func (p *Package) checkUseCaseBundle(ctx string, out *diag.Set) {
	useCases := p.useCaseTypes()
	if len(useCases) == 0 {
		return
	}

	bundle, bundlePos, hasBundle := p.lookupStruct("UseCases")
	if !hasBundle {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 42),
			Pos:  useCases[0].pos,
			Rule: RuleCtxUseCases,
			What: "context " + ctx + " declares use cases but no UseCases struct.",
			Why:  "The bundle is what a caller depends on instead of the internals, and the one place the shared dependencies of a context are threaded through.",
			How:  "Add `type UseCases struct { … }` and `func NewUseCases(…) UseCases` in usecases.go.",
		})
		return
	}

	if _, ok := p.lookupFunc("NewUseCases"); !ok {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 43),
			Pos:  bundlePos,
			Rule: RuleCtxUseCases,
			What: "context " + ctx + " has a UseCases struct but no NewUseCases constructor.",
			Why:  "Without a constructor every caller wires the context by hand, and the shared dependencies are threaded through in as many ways as there are callers.",
			How:  "Add `func NewUseCases(…) UseCases` in usecases.go that calls every New… constructor.",
		})
	}

	fields := map[string]bool{}
	for i := 0; i < bundle.NumFields(); i++ {
		if named, ok := bundle.Field(i).Type().(*types.Named); ok {
			fields[named.Obj().Name()] = true
		}
	}
	for _, uc := range useCases {
		if fields[uc.name] {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 44),
			Pos:  uc.pos,
			Rule: RuleCtxUseCases,
			What: "use case " + uc.name + " is missing from the UseCases struct of " + ctx + ".",
			Why:  "A use case outside the bundle is reachable only by knowing the internals of the context, which defeats the purpose of having a bundle.",
			How:  "Add a field of type " + uc.name + " to UseCases and set it in NewUseCases.",
		})
	}
}

// useCase is a use case type declared in the package.
type useCase struct {
	name string
	pos  ir.Position
	spec *ast.TypeSpec
	sig  *types.Signature
	file string
}

// useCaseTypes returns every named func type of the package whose first
// parameter is an auth subject.
func (p *Package) useCaseTypes() []useCase {
	var out []useCase
	for _, f := range p.pkg.Syntax {
		if p.isGeneratedByUs(f) {
			continue
		}
		file := p.pkg.Fset.Position(f.Pos()).Filename
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.TYPE {
				continue
			}
			for _, s := range gd.Specs {
				ts, ok := s.(*ast.TypeSpec)
				if !ok {
					continue
				}
				obj := p.pkg.TypesInfo.Defs[ts.Name]
				if obj == nil || !obj.Exported() {
					continue
				}
				named, ok := obj.Type().(*types.Named)
				if !ok {
					continue
				}
				sig, ok := named.Underlying().(*types.Signature)
				if !ok || !p.firstParamIsSubject(sig) {
					continue
				}
				out = append(out, useCase{
					name: ts.Name.Name,
					pos:  p.pos(ts.Pos()),
					spec: ts,
					sig:  sig,
					file: file,
				})
			}
		}
	}
	return out
}

// lookupStruct resolves a package level struct type by name.
func (p *Package) lookupStruct(name string) (*types.Struct, ir.Position, bool) {
	obj := p.pkg.Types.Scope().Lookup(name)
	if obj == nil {
		return nil, ir.Position{}, false
	}
	st, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		return nil, ir.Position{}, false
	}
	return st, p.pos(obj.Pos()), true
}

// lookupFunc resolves a package level function by name.
func (p *Package) lookupFunc(name string) (*types.Func, bool) {
	obj := p.pkg.Types.Scope().Lookup(name)
	fn, ok := obj.(*types.Func)
	return fn, ok
}

// DomainPackages returns the import paths of the packages that are the domain
// side of a bounded context.
//
// It exists so that checks outside this package can ask the same question the
// architecture rules ask, instead of reaching into infrastructure. Which
// directories are contexts is the one thing speclink cannot infer, and asking
// a shared helper keeps the answer in one place: an infrastructure type that
// happens to carry an identity is not an aggregate anybody decided about.
func DomainPackages(pkgs []*Package, cfg config.Config, root string) map[string]bool {
	out := map[string]bool{}
	for _, p := range pkgs {
		rel := p.relDir(root)
		if _, inContext := cfg.InContextRoot(rel); !inContext {
			continue
		}
		if cfg.Excluded(rel) || contextRole(rel, cfg) != roleDomain {
			continue
		}
		out[p.PkgPath()] = true
	}
	return out
}
