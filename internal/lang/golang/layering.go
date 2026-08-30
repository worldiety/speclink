package golang

import (
	"go/ast"
	"go/types"
	"strings"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/diag"
)

// The layering rules of an architecture that names its own layers.
//
// nago has one presentation layer called ui and keeps its storage in the
// context root, so the only direction worth checking there is that the domain
// does not import the interface. An architecture with a presentation layer, an
// adapter layer and hand written ports has three directions, and the one that
// matters most is the one nago does not have at all: the domain must not know
// how it is stored.
const (
	// RuleCtxNoPresentationImport keeps the domain from importing the layers
	// that present it.
	RuleCtxNoPresentationImport = "K6-CTX-NO-PRESENTATION-IMPORT"
	// RuleAdapterWiredInCmd keeps everything but the entry point from knowing
	// which adapter is in use.
	RuleAdapterWiredInCmd = "K6-ADAPTER-WIRED-IN-CMD"
	// RulePresentationNoBundle keeps a handler from depending on a whole
	// context.
	RulePresentationNoBundle = "K6-PRESENTATION-NO-BUNDLE"
	// RuleCtxPresentationPkg names the presentation packages after their layer
	// and context.
	RuleCtxPresentationPkg = "K6-CTX-PRESENTATION-PKG"
)

// presentationDirs are the layers that show a context to somebody.
var presentationDirs = []string{"rest", "cli"}

// CheckLayering enforces the dependency directions of a layered context.
func CheckLayering(pkgs []*Package, cfg config.Config, root string, out *diag.Set) {
	for _, p := range pkgs {
		rel := p.relDir(root)
		ctx, inContext := cfg.InContextRoot(rel)
		if !inContext || !cfg.InScope(rel) {
			continue
		}

		switch layerOf(rel, ctx, cfg) {
		case layerDomain:
			p.checkNoPresentationImport(rel, out)
			p.checkNoAdapterImport(rel, "a domain package", out)
		case layerPresentation:
			p.checkPresentationPkgName(rel, ctx, out)
			p.checkNoAdapterImport(rel, "presentation", out)
			p.checkNoBundleParameter(rel, out)
		}
	}

	// Only cmd may import an adapter, and cmd is outside the context root, so
	// the remaining direction is checked over everything else.
	for _, p := range pkgs {
		rel := p.relDir(root)
		if !cfg.InScope(rel) || cfg.UnderCmdRoot(rel) || isAdapterDir(rel) {
			continue
		}
		if _, inContext := cfg.InContextRoot(rel); inContext {
			continue // already covered above, with a better description
		}
		p.checkNoAdapterImport(rel, rel, out)
	}
}

type layer int

const (
	layerDomain layer = iota + 1
	layerPresentation
	layerAdapter
)

// layerOf reads the layer out of the path below a context.
func layerOf(rel, ctx string, cfg config.Config) layer {
	parts := strings.Split(rel, "/")
	for i, part := range parts {
		if part != ctx || i+1 >= len(parts) {
			continue
		}
		switch next := parts[i+1]; {
		case next == "adapter":
			return layerAdapter
		case contains(presentationDirs, next):
			return layerPresentation
		}
	}
	return layerDomain
}

func isAdapterDir(rel string) bool {
	for _, part := range strings.Split(rel, "/") {
		if part == "adapter" {
			return true
		}
	}
	return false
}

// checkNoPresentationImport keeps the domain from importing what presents it.
//
// The same reasoning as the ui rule it generalises: a context defines what the
// system does, and a presentation is one way of reaching it. Once the domain
// imports one, that ordering is lost and the context can no longer be tested or
// reused without it.
func (p *Package) checkNoPresentationImport(rel string, out *diag.Set) {
	for _, f := range p.pkg.Syntax {
		for _, imp := range f.Imports {
			path, err := unquote(imp.Path.Value)
			if err != nil || !isPresentationPackage(path) {
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 45),
				Pos:  p.pos(imp.Pos()),
				Rule: RuleCtxNoPresentationImport,
				What: rel + " is a domain package but imports " + path + ".",
				Why:  "The domain defines what the system does; a presentation is one way of reaching it, and there is more than one. Importing it inverts that and makes the context untestable without a transport.",
				How:  "Move the code into the presentation package, or pass what is needed in as a parameter.",
			})
		}
	}
}

// checkNoAdapterImport keeps everything but the entry point from knowing which
// adapter is in use.
//
// This is the direction nago does not have. There a repository is built in the
// context root, so the domain imports its own storage; here the port is
// declared in the domain, implemented under adapter, and joined in cmd. A
// project that imports an adapter anywhere else has decided its storage in a
// place that cannot be swapped for a test.
func (p *Package) checkNoAdapterImport(rel, what string, out *diag.Set) {
	for _, f := range p.pkg.Syntax {
		if p.declaresSpecFile(f) {
			// A specification file names types; it does not build anything.
			// The reason for this rule is that an import decides which
			// implementation is in use somewhere that cannot be swapped for a
			// test, and a declaration that says "this is the shape that
			// crosses the boundary to the file store" decides nothing and
			// constructs nothing. Refusing it would leave the one place that
			// can state what a stored shape is unable to name it.
			continue
		}
		for _, imp := range f.Imports {
			path, err := unquote(imp.Path.Value)
			if err != nil || !isAdapterPackage(path) {
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 46),
				Pos:  p.pos(imp.Pos()),
				Rule: RuleAdapterWiredInCmd,
				What: rel + " is " + what + " but imports the adapter " + path + ".",
				Why:  "A port is declared where it is used and implemented under adapter, so that which implementation is in use is decided in one place. An import here decides it somewhere that cannot be swapped for a test.",
				How:  "Take the port as a parameter and let the main package under cmd/ pass the adapter in.",
			})
		}
	}
}

// checkNoBundleParameter keeps a handler from depending on a whole context.
//
// The bundle exists to take the weight out of the place a context is
// assembled — one constructor rather than a dozen. Handing it to a handler
// undoes that: the handler then depends on every use case the context has, and
// a test for one route has to build all of them.
func (p *Package) checkNoBundleParameter(rel string, out *diag.Set) {
	for _, f := range p.pkg.Syntax {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || fn.Type.Params == nil {
				continue
			}
			for _, param := range fn.Type.Params.List {
				name, ok := p.namedTypeOf(param.Type)
				if !ok || name != "UseCases" {
					continue
				}
				out.Add(diag.Finding{
					Code: diag.Code(diag.PhaseSemantic, 47),
					Pos:  p.pos(param.Pos()),
					Rule: RulePresentationNoBundle,
					What: fn.Name.Name + " in " + rel + " takes the whole UseCases bundle.",
					Why:  "The bundle exists to take the weight out of the place a context is assembled. A handler that takes it depends on every use case the context has, and a test for one route has to build all of them.",
					How:  "Take the use cases this handler actually calls, one parameter each.",
				})
				break
			}
		}
	}
}

// namedTypeOf returns the name of a named type expression.
func (p *Package) namedTypeOf(e ast.Expr) (string, bool) {
	tv, ok := p.pkg.TypesInfo.Types[e]
	if !ok {
		return "", false
	}
	named, ok := types.Unalias(tv.Type).(*types.Named)
	if !ok {
		return "", false
	}
	return named.Obj().Name(), true
}

// checkPresentationPkgName names a presentation package after its layer and
// context.
//
// The directory is always called rest or cli, so the package name is what tells
// a reader and an import which context's interface this is — and it is what
// keeps two contexts from colliding in the main package that wires them.
func (p *Package) checkPresentationPkgName(rel, ctx string, out *diag.Set) {
	layer := lastSegment(rel)
	want := layer + ctx
	if p.pkg.Name == want {
		return
	}
	out.Add(diag.Finding{
		Code: diag.Code(diag.PhaseSemantic, 48),
		Pos:  p.filePos(),
		Rule: RuleCtxPresentationPkg,
		What: "the " + layer + " directory of " + ctx + " declares package " + p.pkg.Name + ", expected " + want + ".",
		Why:  "The directory is always called " + layer + ", so the package name is what tells a reader which context this is — and what keeps two contexts from colliding in the main package that wires them.",
		How:  "Rename the package to " + want + ".",
	})
}

func isPresentationPackage(path string) bool {
	last := lastSegment(path)
	return contains(presentationDirs, last)
}

func isAdapterPackage(path string) bool {
	for _, part := range strings.Split(path, "/") {
		if part == "adapter" {
			return true
		}
	}
	return false
}

func lastSegment(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
