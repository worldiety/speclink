// Package golang is the Go frontend: it binds annotations to the constructs of
// the host language and lowers them into the language neutral ir.
//
// It is the only package allowed to know go/ast and go/types. Everything it
// produces is ir, so rules, diagnostics and backends stay language agnostic and
// a second language frontend can be added without rewriting them.
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

	"github.com/worldiety/speclink/internal/config"
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
	// TopologySuffix marks a declaration of the world outside: boundary.topology.go.
	TopologySuffix = ".topology.go"
	// ProcessSuffix marks a process declaration: quote-to-invoice.process.go.
	//
	// It is its own suffix rather than a requirement file because a process
	// imports the contexts it names, and a requirement must import nothing:
	// the requirement tree is the leaf of the dependency graph, and a cycle
	// through it would make the whole thing unloadable.
	ProcessSuffix = ".process.go"
	// TestSuffix marks a test file. Go requires it to be the final suffix, so
	// a test file can never also carry one of the two above.
	TestSuffix = "_test.go"
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
	// processFiles are the *.process.go files of this package.
	processFiles []*ast.File
	// topologyFiles are the *.topology.go files of this package.
	topologyFiles []*ast.File
	// sourceNames holds the base names of ordinary Go files, used to detect
	// orphaned annotation files.
	sourceNames map[string]bool
	// testFiles are the *_test.go files, empty unless the load asked for them.
	testFiles []*ast.File
	// isTest marks a test variant of a package rather than the package itself.
	isTest bool
}

// PkgPath returns the import path of the package.
//
// It is not an identity. go/packages gives the in-package test variant the same
// PkgPath as the package it tests, so two entries in one load can answer this
// identically — measured on the reference project the moment a single _test.go
// file exists. Anything keying on a package must therefore ask IsTest as well,
// or ask ID.
func (p *Package) PkgPath() string { return p.pkg.PkgPath }

// ID is the unique identity of a loaded package, which PkgPath is not.
func (p *Package) ID() string { return p.pkg.ID }

// IsTest reports whether this is a test variant rather than a package proper.
//
// Every rule that existed before test loading was introduced must skip these.
// They are the same source seen twice, so letting them through would double
// every construct, every schema and every finding derived from them — and the
// generated <pkg>.test binary is a main package outside cmd/, which K8-MAIN
// -LOCATION would report in every package that has a test.
func (p *Package) IsTest() bool { return p.isTest }

// Load loads the given patterns from dir with full type information.
//
// It returns an error only for failures of the load itself. Type errors in the
// loaded packages are reported through TypeErrors, because they belong to phase
// V2 which is the Go compilation and not a speclink phase.
func Load(dir string, patterns ...string) ([]*Package, error) {
	return load(dir, false, patterns...)
}

// LoadWithTests additionally loads the test variants of the matched packages.
//
// It is separate because it is not free. Asking for tests roughly doubles the
// load: go/packages returns the package, the package recompiled with its
// in-package tests, the external test package and a generated main, and every
// one of them is parsed and type checked. On the reference project the load
// step went from 0.35s to 0.71s the moment one test file existed.
//
// Only verify needs it, because only K14 asks a question about tests. freeze,
// inventory and impact would pay the same price for nothing.
func LoadWithTests(dir string, patterns ...string) ([]*Package, error) {
	return load(dir, true, patterns...)
}

func load(dir string, tests bool, patterns ...string) ([]*Package, error) {
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}
	cfg := &packages.Config{Mode: loadMode, Dir: dir, Tests: tests}
	loaded, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("load packages: %w", err)
	}

	out := make([]*Package, 0, len(loaded))
	for _, lp := range loaded {
		p := &Package{pkg: lp, sourceNames: map[string]bool{}, isTest: isTestVariant(lp)}
		for _, f := range lp.Syntax {
			name := filepath.Base(lp.Fset.Position(f.Pos()).Filename)
			switch {
			case strings.HasSuffix(name, TestSuffix):
				p.testFiles = append(p.testFiles, f)
			case strings.HasSuffix(name, AnnotationSuffix):
				p.annotationFiles = append(p.annotationFiles, f)
			case strings.HasSuffix(name, TopologySuffix):
				p.topologyFiles = append(p.topologyFiles, f)
			case strings.HasSuffix(name, ProcessSuffix):
				p.processFiles = append(p.processFiles, f)
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

// isTestVariant reports whether a loaded package is a test build rather than
// the package itself.
//
// The decision is made on ID, never on PkgPath. go/packages gives the
// in-package test variant the same PkgPath as its subject and distinguishes
// them only in the ID, which carries the bracketed form
// "example.com/erp/app/sales [example.com/erp/app/sales.test]". A check on
// PkgPath would therefore separate nothing at all, silently, and every rule
// would see the package twice.
func isTestVariant(p *packages.Package) bool {
	return strings.Contains(p.ID, ".test]") || strings.HasSuffix(p.ID, ".test")
}

// Tests returns the loaded packages that are test variants.
func Tests(pkgs []*Package) []*Package { return filterTests(pkgs, true) }

// NonTests returns the loaded packages that are not test variants.
//
// Every rule written before test loading existed takes this, so that asking
// for tests cannot change what any of them see.
func NonTests(pkgs []*Package) []*Package { return filterTests(pkgs, false) }

func filterTests(pkgs []*Package, want bool) []*Package {
	out := make([]*Package, 0, len(pkgs))
	for _, p := range pkgs {
		if p.isTest == want {
			out = append(out, p)
		}
	}
	return out
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

// InScope keeps the packages the configuration says are measured.
//
// The filter runs once, on the loaded set, before any rule sees it. That is the
// only place it can run: a rule that skipped out-of-scope packages itself would
// leave them in the set every other rule works from, and the evolution rules in
// particular decide what counts as *removed* from exactly that set. A promised
// type in a package nobody looked at must not read as deleted.
//
// A package holding requirement declarations is always kept. The scope says
// which code is measured; the requirement tree is not code, and a scope that
// happened to exclude it would leave the run with nothing to measure against
// and every requirement reading as uncovered.
func InScope(pkgs []*Package, cfg config.Config, root string) []*Package {
	out := make([]*Package, 0, len(pkgs))
	for _, p := range pkgs {
		if len(p.requirementFiles) > 0 || cfg.InScope(p.relDir(root)) {
			out = append(out, p)
		}
	}
	return out
}

// OutOfScope returns the packages the configuration excludes, so a run can say
// how much it did not look at.
func OutOfScope(pkgs []*Package, cfg config.Config, root string) []*Package {
	out := make([]*Package, 0, len(pkgs))
	for _, p := range pkgs {
		if len(p.requirementFiles) == 0 && !cfg.InScope(p.relDir(root)) {
			out = append(out, p)
		}
	}
	return out
}
