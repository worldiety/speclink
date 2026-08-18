package golang

import (
	"go/ast"

	"github.com/worldiety/speclink/internal/diag"
)

// RuleNoGenericCRUD forbids the generic CRUD constructs of the framework.
//
// It is the only rule that forbids constructs instead of reporting
// contradictions, and it follows from P9 (analysability before brevity):
//
//   - These factories produce specification facts at run time — six
//     permissions derived from a prefix, a repository, three routes. A static
//     analysis can only see them by reimplementing framework internals, and if
//     the framework changes its naming rule the tool lies silently.
//   - They yield modules that can only be adapted as a whole. A hand written
//     use case can be changed one at a time.
//
// Generic factories exist to save humans typing. Once an LLM writes the code
// that benefit is worthless, while the opacity remains.
//
// The ban is empirically free: the reference project uses these constructs zero
// times, against 160 hand written permission.Declare calls. It forbids nothing
// that is in use; it closes a door before anyone walks through it.
const RuleNoGenericCRUD = "K4-NO-GENERIC-CRUD"

// forbidden lists the generic CRUD entry points, with what each creates at run
// time. The message names it, so the reader sees the cost rather than just the
// prohibition.
var forbidden = []struct {
	pkg, name, creates string
}{
	{nagoEntCfg, "Enable", "six permissions, a repository and three routes"},
	{nagoEnt, "DeclarePermissions", "six permission IDs from a prefix"},
	{nagoEnt, "NewUseCases", "six use cases"},
}

// CheckGenericCRUD reports every use of a generic CRUD factory.
func (p *Package) CheckGenericCRUD(out *diag.Set) {
	for _, f := range p.pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			for _, bad := range forbidden {
				if !p.callsInto(call.Fun, bad.pkg, bad.name) {
					continue
				}
				out.Add(diag.Finding{
					Code: diag.Code(diag.PhaseSemantic, 10),
					Pos:  p.pos(call.Pos()),
					Rule: RuleNoGenericCRUD,
					What: "generic CRUD factories are not permitted.",
					Why: shortPkg(bad.pkg) + "." + bad.name + " creates " + bad.creates +
						" at run time. Those facts are invisible to static analysis, and the module can only be adapted as a whole (P9: analysability before brevity).",
					How: "Write the use cases out one by one: a named func type with auth.Subject as the first parameter, one permission.Declare[UC](…) each, and subject.Audit(…) as the first statement of every Decide.",
				})
				return true
			}
			return true
		})
	}
	p.checkForbiddenImports(out)
}

// checkForbiddenImports catches the generic CRUD UI, which is used through its
// types rather than through a call.
func (p *Package) checkForbiddenImports(out *diag.Set) {
	for _, f := range p.pkg.Syntax {
		for _, imp := range f.Imports {
			path, err := unquote(imp.Path.Value)
			if err != nil || path != nagoUIEnt {
				continue
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseSemantic, 11),
				Pos:  p.pos(imp.Pos()),
				Rule: RuleNoGenericCRUD,
				What: "the generic CRUD user interface is not permitted.",
				Why:  shortPkg(nagoUIEnt) + " renders views whose routes, permissions and read models are assembled at run time, so no view can be traced back to a requirement.",
				How:  "Write the views explicitly and bind them with spec.Satisfies(…). Their look and feel may follow " + shortPkg(nagoUIEnt) + " closely, and may be lifted into a template; only the generic wiring is out.",
			})
		}
	}
}

// shortPkg renders an import path as the package name a reader would type.
func shortPkg(path string) string {
	// The generic CRUD user interface lives at application/ent/ui but is
	// imported as uient everywhere, so the last path segment is not the name a
	// reader would recognise.
	if path == nagoUIEnt {
		return "uient"
	}
	if i := lastIndexByte(path, '/'); i >= 0 {
		name := path[i+1:]
		// nago names its wiring packages cfg<domain>, e.g. application/ent/cfg
		// is imported as cfgent.
		if name == "cfg" {
			rest := path[:i]
			if j := lastIndexByte(rest, '/'); j >= 0 {
				return "cfg" + rest[j+1:]
			}
		}
		return name
	}
	return path
}

// unquote strips the quotes of an import path literal without pulling in
// strconv for a single call.
func unquote(lit string) (string, error) {
	if len(lit) < 2 || lit[0] != '"' || lit[len(lit)-1] != '"' {
		return "", errBadImport
	}
	return lit[1 : len(lit)-1], nil
}

var errBadImport = errString("malformed import path literal")

type errString string

func (e errString) Error() string { return string(e) }
