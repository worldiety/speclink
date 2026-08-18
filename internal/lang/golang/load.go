// Package golang is the Go frontend: it binds annotations to the constructs of
// the host language and lowers them into the language neutral ir.
//
// It is the only package allowed to know go/ast and go/types. Everything it
// produces is ir, so rules, diagnostics and backends stay language agnostic
// (see docs/plan.md §3.1).
//
// There is no extraction step, no synthetic file and no position mapping:
// annotation files are ordinary Go and part of the normal build, so the Go
// compiler has already checked arity, argument types, field names, enum values
// and every identifier reference before speclink looks at anything.
package golang

import (
	"fmt"
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"

	"github.com/worldiety/speclink/internal/ir"
)

// SpecPkgPath is the import path of the public directive catalogue. Calls
// resolving into this package are the annotation language; everything else in
// an annotation file is a whitelist violation.
const SpecPkgPath = "github.com/worldiety/speclink/spec"

const (
	// AnnotationSuffix marks a sidecar file that asserts facts about the
	// constructs of its neighbour: commands.go -> commands.annotation.go.
	AnnotationSuffix = ".annotation.go"
	// RequirementSuffix marks a requirement declaration file in the
	// requirement tree: R-QUOTE-SUBMIT.spec.go.
	RequirementSuffix = ".spec.go"
)

// loadMode is deliberately explicit. NeedTypes and NeedTypesInfo are what makes
// the whole design work: without type resolution an identifier in the AST is
// just a name, and neither generic type arguments nor cross package references
// could be resolved.
//
// NeedDeps is deliberately absent. It would apply this whole mode recursively
// and make go/packages parse and type check every transitive dependency from
// source; without it the dependencies arrive as export data, which is all that
// is needed here. Nothing reads the syntax or the type info of a dependency —
// only the keys of Imports, to decide whether a package imports a bounded
// context or the presentation layer.
//
// The distinction is invisible against a small dependency tree and decisive
// against a real one: the framework alone is several hundred packages, and
// parsing it on every run would be paid on every single invocation.
const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedSyntax |
	packages.NeedTypes |
	packages.NeedTypesInfo |
	packages.NeedTypesSizes |
	packages.NeedImports |
	packages.NeedModule

// Package is a loaded Go package together with the classification of its files.
type Package struct {
	pkg *packages.Package

	// annotationFiles are the *.annotation.go files of this package.
	annotationFiles []*ast.File
	// requirementFiles are the *.spec.go files of this package.
	requirementFiles []*ast.File
	// sourceNames holds the base names of ordinary Go files, used to detect
	// orphaned annotation files.
	sourceNames map[string]bool
}

// PkgPath returns the import path of the package.
func (p *Package) PkgPath() string { return p.pkg.PkgPath }

// Load loads the given patterns from dir with full type information.
//
// It returns an error only for failures of the load itself. Type errors in the
// loaded packages are reported through TypeErrors, because they belong to phase
// V2 which is the Go compilation and not a speclink phase.
func Load(dir string, patterns ...string) ([]*Package, error) {
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}
	cfg := &packages.Config{Mode: loadMode, Dir: dir, Tests: false}
	loaded, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	out := make([]*Package, 0, len(loaded))
	for _, lp := range loaded {
		p := &Package{pkg: lp, sourceNames: map[string]bool{}}
		for _, f := range lp.Syntax {
			name := filepath.Base(lp.Fset.Position(f.Pos()).Filename)
			switch {
			case strings.HasSuffix(name, AnnotationSuffix):
				p.annotationFiles = append(p.annotationFiles, f)
			case strings.HasSuffix(name, RequirementSuffix):
				p.requirementFiles = append(p.requirementFiles, f)
			default:
				p.sourceNames[name] = true
			}
		}
		out = append(out, p)
	}
	return out, nil
}

// TypeErrors collects the Go compiler errors of the loaded packages.
//
// Phase V2 is the Go compilation itself. When it fails there is no annotation
// feedback at all, and the loop runner has to prioritise accordingly: the build
// order is Go compiler, then speclink, then tests.
func TypeErrors(pkgs []*Package) []packages.Error {
	var errs []packages.Error
	for _, p := range pkgs {
		errs = append(errs, p.pkg.Errors...)
	}
	return errs
}

// pos converts a token.Pos into a language neutral position.
func (p *Package) pos(at token.Pos) ir.Position {
	pp := p.pkg.Fset.Position(at)
	return ir.Position{File: pp.Filename, Line: pp.Line, Col: pp.Column}
}

// specFuncName returns the name of the speclink/spec function that expr calls,
// or "" when expr does not call into the spec package.
//
// This is the single point where "is this the annotation language?" is decided,
// and it is decided over resolved types, never over the spelling of the
// selector. An alias import or a shadowed identifier cannot fool it.
func (p *Package) specFuncName(fun ast.Expr) string {
	id := calleeIdent(fun)
	if id == nil {
		return ""
	}
	obj := p.pkg.TypesInfo.Uses[id]
	if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != SpecPkgPath {
		return ""
	}
	return obj.Name()
}

// calleeIdent unwraps a call target down to its identifier, seeing through
// generic instantiation (IndexExpr, IndexListExpr) and selectors.
func calleeIdent(e ast.Expr) *ast.Ident {
	switch x := e.(type) {
	case *ast.IndexExpr:
		return calleeIdent(x.X)
	case *ast.IndexListExpr:
		return calleeIdent(x.X)
	case *ast.SelectorExpr:
		return x.Sel
	case *ast.Ident:
		return x
	case *ast.ParenExpr:
		return calleeIdent(x.X)
	}
	return nil
}
