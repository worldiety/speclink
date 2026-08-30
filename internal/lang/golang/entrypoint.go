package golang

import (
	"go/ast"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// EntryPoints reports the programs this module builds.
//
// Detection is the package name, which is what the language actually
// guarantees: a main package is a program and nothing else is. The rest is
// read off the imports, and the parts that cannot be read off anything are
// marked as inferred rather than presented alongside the facts.
//
// Scope is deliberately the whole module. Which programs exist is a question
// about the module, and answering it from a narrowed load would report that a
// service does not exist because nobody asked about its directory — the exact
// class of defect that the measured and loaded sets were separated to prevent.
func (m *Model) EntryPoints() []ir.EntryPoint {
	// Without the module path the imports cannot be told apart from the
	// standard library's, so the contexts and adapters come out empty. That is
	// a poorer answer, not a wrong one, and losing the list of programs
	// entirely because go.mod could not be read is the worse trade.
	module, _ := ModulePath(m.All)

	var out []ir.EntryPoint
	for _, p := range m.All {
		if p.pkg.Name != "main" || p.isTest {
			continue
		}
		out = append(out, m.entryPoint(p, module))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Dir < out[j].Dir })
	return out
}

func (m *Model) entryPoint(p *Package, module string) ir.EntryPoint {
	dir := p.relDir(m.Root)
	e := ir.EntryPoint{
		Name:    path.Base(p.PkgPath()),
		Package: p.PkgPath(),
		Dir:     dir,
		Doc:     packageDoc(p),
		Guessed: true,
	}

	ctxRoot := m.Layout.ContextRoot
	if ctxRoot == "" {
		ctxRoot = "app"
	}
	seenCtx, seenAdapter := map[string]bool{}, map[string]bool{}
	for _, imp := range p.imports() {
		rel, ok := withinModule(imp, module)
		if !ok {
			continue
		}
		if !strings.HasPrefix(rel, ctxRoot+"/") {
			continue
		}
		rest := strings.TrimPrefix(rel, ctxRoot+"/")
		name, _, _ := strings.Cut(rest, "/")
		if name != "" && !seenCtx[name] {
			seenCtx[name] = true
			e.Contexts = append(e.Contexts, name)
		}
		if i := strings.Index(rest, "adapter/"); i >= 0 && !seenAdapter[rel] {
			seenAdapter[rel] = true
			e.Adapters = append(e.Adapters, rel)
		}
	}
	sort.Strings(e.Contexts)
	sort.Strings(e.Adapters)

	e.Verbs, e.Flags = invocation(p)
	if len(p.pkg.Syntax) > 0 {
		e.Pos = p.pos(p.pkg.Syntax[0].Package)
	}
	return e
}

// packageDoc is the comment above the package clause.
//
// Go's convention is that a command's doc begins "Command <name>", so this is
// where a project already writes what a binary is for. Reading it costs
// nothing and asking authors to write it a second time somewhere structured
// would be asking them to state a fact twice.
func packageDoc(p *Package) string {
	for _, f := range p.pkg.Syntax {
		if f.Doc == nil {
			continue
		}
		text := strings.TrimSpace(f.Doc.Text())
		if text == "" {
			continue
		}
		// The first paragraph only. The rest is usually the design notes of
		// whoever wrote it, which belong in the file and not in a table.
		if i := strings.Index(text, "\n\n"); i >= 0 {
			text = text[:i]
		}
		return strings.Join(strings.Fields(text), " ")
	}
	return ""
}

// invocation guesses how the program is called.
//
// Two shapes, both common and both shallow: a string compared against an
// element of the argument vector, and a flag registered by name. Neither is a
// declaration, and a program that dispatches through a map or a third party
// router will yield nothing here — which is why the result is labelled a guess
// wherever it is printed. A silent empty list presented as fact would be worse
// than no chapter.
func invocation(p *Package) (verbs, flags []string) {
	seenV, seenF := map[string]bool{}, map[string]bool{}
	for _, f := range p.pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			switch t := n.(type) {
			case *ast.BinaryExpr:
				// os.Args[1] == "serve", either way round.
				if lit, ok := argComparison(t); ok && !seenV[lit] {
					seenV[lit] = true
					verbs = append(verbs, lit)
				}
			case *ast.CallExpr:
				if name, ok := flagRegistration(t); ok && !seenF[name] {
					seenF[name] = true
					flags = append(flags, name)
				}
			}
			return true
		})
	}
	sort.Strings(verbs)
	sort.Strings(flags)
	return verbs, flags
}

// argComparison recognises a literal tested against the argument vector.
func argComparison(b *ast.BinaryExpr) (string, bool) {
	lit, other := stringLit(b.X), b.Y
	if lit == "" {
		lit, other = stringLit(b.Y), b.X
	}
	if lit == "" || !indexesArgs(other) {
		return "", false
	}
	return lit, true
}

// indexesArgs reports whether the expression reads os.Args at some index.
func indexesArgs(e ast.Expr) bool {
	idx, ok := e.(*ast.IndexExpr)
	if !ok {
		return false
	}
	sel, ok := idx.X.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Args" {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "os"
}

// flagRegistration recognises flag.String("name", …) and its siblings, on the
// package and on a FlagSet alike.
func flagRegistration(c *ast.CallExpr) (string, bool) {
	sel, ok := c.Fun.(*ast.SelectorExpr)
	if !ok || len(c.Args) < 2 {
		return "", false
	}
	switch sel.Sel.Name {
	case "String", "Bool", "Int", "Duration", "Float64", "Int64", "Uint", "Var",
		"StringVar", "BoolVar", "IntVar", "DurationVar", "Float64Var", "Int64Var", "UintVar":
	default:
		return "", false
	}
	// The name is the first string literal argument, second for the Var forms
	// whose first argument is the destination pointer.
	for _, a := range c.Args[:min(3, len(c.Args))] {
		if s := stringLit(a); s != "" {
			return s, true
		}
	}
	return "", false
}

func stringLit(e ast.Expr) string {
	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind.String() != "STRING" {
		return ""
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return ""
	}
	return s
}

// withinModule strips the module prefix from an import path.
func withinModule(imp, module string) (string, bool) {
	if module == "" || imp == module {
		return "", false
	}
	if !strings.HasPrefix(imp, module+"/") {
		return "", false
	}
	return strings.TrimPrefix(imp, module+"/"), true
}
