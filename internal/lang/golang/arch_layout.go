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

// Rule IDs of the layout linters. They appear in diagnostics and in spec.Waive
// calls, so they are public surface and must stay stable.
const (
	// RuleMainExists fires when a module has no entry point at all.
	RuleMainExists = "K8-MAIN-EXISTS"
	// RuleMainLocation fires for a main package outside the command root.
	RuleMainLocation = "K8-MAIN-LOCATION"
	// RuleInfraDomainFree fires when infrastructure carries domain knowledge.
	RuleInfraDomainFree = "K7-INFRA-DOMAIN-FREE"
)

// CheckMainPackages verifies that the module has an entry point and that every
// entry point lives where entry points belong.
//
// Both halves matter. A module without a main package is a library, and saying
// so out loud is cheap; a main package scattered somewhere in the tree is the
// classic way a "small helper program" grows into a second application nobody
// reviews.
func CheckMainPackages(pkgs []*Package, cfg config.Config, root string, out *diag.Set) {
	found := 0

	for _, p := range pkgs {
		if p.pkg.Name != "main" {
			continue
		}
		// Counted before the scope is consulted. Where the entry point lives is
		// a question about the package and belongs to the scope; whether the
		// module has one at all is a question about the module, and a scope
		// that happens to exclude cmd/ must not answer it with "no".
		found++

		rel := p.relDir(root)
		if !cfg.InScope(rel) {
			continue
		}
		if cfg.UnderCmdRoot(rel) {
			continue
		}
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 30),
			Pos:  p.filePos(),
			Rule: RuleMainLocation,
			What: "main package " + rel + " does not live under " + cfg.CmdRoot + "/.",
			Why:  "Entry points belong in one place. Scattered ones are how a helper program grows into a second application nobody reviews.",
			How:  "Move it to " + cfg.CmdRoot + "/" + filepath.Base(rel) + "/, or exclude the path in " + config.FileName + " if it is an example.",
		})
	}

	if found == 0 && len(pkgs) > 0 {
		out.Add(diag.Finding{
			Code: diag.Code(diag.PhaseSemantic, 31),
			Pos:  ir.Position{File: root, Line: 1, Col: 1},
			Rule: RuleMainExists,
			What: "the module has no main package.",
			Why:  "Without an entry point nothing assembles the bounded contexts, and no wiring can be checked against the specification.",
			How:  "Add " + cfg.CmdRoot + "/<name>/main.go, or waive the rule if this module is a library.",
		})
	}
}

// CheckInfrastructure verifies that infrastructure packages carry no domain
// knowledge.
//
// The rule is stated in reverse on purpose. Deciding what counts as an
// infrastructure helper is a judgement call speclink cannot make; deciding
// whether a package under pkg/ knows about the domain is exact. Two markers
// suffice, and both are unambiguous:
//
//  1. it imports a bounded context
//  2. it declares a use case, i.e. a function taking an auth subject
//
// Either one means the dependency points the wrong way: infrastructure is what
// the domain builds on, not the other way round.
func CheckInfrastructure(pkgs []*Package, cfg config.Config, root string, out *diag.Set) {
	contexts := contextPaths(pkgs, cfg, root)

	for _, p := range pkgs {
		rel := p.relDir(root)
		if !cfg.InInfraRoot(rel) || !cfg.InScope(rel) {
			continue
		}

		for _, imp := range p.imports() {
			ctx, ok := contexts[imp]
			if !ok {
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 32),
				Pos:  p.filePos(),
				Rule: RuleInfraDomainFree,
				What: rel + " is infrastructure but imports the bounded context " + ctx + ".",
				Why:  "Infrastructure is what the domain builds on. An import in this direction inverts the dependency and ties a general helper to one domain.",
				How:  "Move the domain specific part into " + ctx + ", or pass what is needed in as a parameter or interface.",
			})
			break
		}

		if name, pos, ok := p.firstUseCaseSignature(); ok {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 33),
				Pos:  pos,
				Rule: RuleInfraDomainFree,
				What: rel + " is infrastructure but declares the use case " + name + ".",
				Why:  "A function taking auth.Subject as its first parameter is a use case, and use cases belong to a bounded context, not to a general purpose package.",
				How:  "Move " + name + " into the bounded context it serves.",
			})
		}
	}
}

// contextPaths maps the import path of every bounded context package to its
// context name.
func contextPaths(pkgs []*Package, cfg config.Config, root string) map[string]string {
	out := map[string]string{}
	for _, p := range pkgs {
		rel := p.relDir(root)
		if ctx, ok := cfg.InContextRoot(rel); ok {
			out[p.PkgPath()] = ctx
		}
	}
	return out
}

// relDir returns the package directory relative to the project root.
func (p *Package) relDir(root string) string {
	if len(p.pkg.GoFiles) == 0 && len(p.pkg.CompiledGoFiles) == 0 {
		return ""
	}
	files := p.pkg.GoFiles
	if len(files) == 0 {
		files = p.pkg.CompiledGoFiles
	}
	rel, err := filepath.Rel(root, filepath.Dir(files[0]))
	if err != nil {
		return ""
	}
	return filepath.ToSlash(rel)
}

// filePos returns a position pointing at the package, for findings that concern
// the package as a whole rather than one statement.
func (p *Package) filePos() ir.Position {
	for _, f := range p.pkg.Syntax {
		return p.pos(f.Package)
	}
	return ir.Position{File: p.PkgPath(), Line: 1, Col: 1}
}

// imports returns the import paths of the package.
func (p *Package) imports() []string {
	out := make([]string, 0, len(p.pkg.Imports))
	for path := range p.pkg.Imports {
		out = append(out, path)
	}
	return out
}

// isUIPackage reports whether an import path denotes a user interface package:
// either the framework presentation layer, or a package whose name starts with
// "ui".
func isUIPackage(path string) bool {
	if strings.Contains(path, "/presentation/") || strings.HasSuffix(path, "/presentation") {
		return true
	}
	base := path
	if i := strings.LastIndexByte(base, '/'); i >= 0 {
		base = base[i+1:]
	}
	return strings.HasPrefix(base, "ui") && base != "uuid"
}

// firstUseCaseSignature returns the first declared function type in the package
// whose first parameter is an auth subject.
func (p *Package) firstUseCaseSignature() (string, ir.Position, bool) {
	for _, f := range p.pkg.Syntax {
		if p.isGeneratedByUs(f) {
			continue
		}
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
				if obj == nil {
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
				return ts.Name.Name, p.pos(ts.Pos()), true
			}
		}
	}
	return "", ir.Position{}, false
}

// CheckArchitecture runs every architecture rule this frontend has.
//
// They are collected behind one call because they are one thing: the invariants
// of a nago project, which another frontend has no counterpart for and would
// replace wholesale rather than pick from. A caller naming them individually
// would be a caller that knows which of them exist, and that knowledge belongs
// on this side of the boundary.
//
// The set is not decoration. The recognisers only work because the architecture
// holds: K4-NO-GENERIC-CRUD bans factories that produce specification facts at
// run time precisely so that a static analysis can see them at all. Enforcing
// the architecture is what makes inferring the model possible.
func CheckArchitecture(pkgs []*Package, cfg config.Config, root string, style Style, waived ir.Waivers, out *diag.Set) {
	CheckUseCases(pkgs, cfg, root, style, waived, out)
	CheckBoundedContexts(pkgs, cfg, root, out)
	CheckInfrastructure(pkgs, cfg, root, out)
	CheckMainPackages(pkgs, cfg, root, out)
}
