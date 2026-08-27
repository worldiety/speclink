package golang

import (
	"go/ast"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// ReadVerifications collects the spec.Verified calls of a test package.
//
// This is the one reader that does not look at an annotation file, and the one
// assertion whose subject is the file it stands in. Everywhere else a binding
// is a sentence about something else; here a test says something about itself.
//
// It also reads a claim rather than a fact. The call is written down
// statically, so a missing one can be reported — but its presence proves only
// that somebody typed it, not that it ever ran. The other half of the answer
// comes from the line the call writes when it executes, which is recorded
// separately. What is read here is what should have happened; what is recorded
// is what did.
//
// Only calls inside a function body count, and the enclosing function is the
// target. A call at package level would have no test to belong to, and there is
// nowhere sensible to attribute it.
func (p *Package) ReadVerifications(out *diag.Set) []ir.Binding {
	if !p.isTest {
		return nil
	}

	var bindings []ir.Binding
	for _, f := range p.testFiles {
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if b, found := p.readVerified(fn, out); found {
				bindings = append(bindings, b)
			}
		}
	}
	return bindings
}

// readVerified gathers every spec.Verified call in one function.
//
// The calls are folded into a single binding rather than one each, because they
// state one thing about one test: which requirements it demonstrates. Two calls
// on two branches are two statements of the same kind about the same subject,
// and splitting them would make a test that verifies a requirement twice look
// like two tests.
func (p *Package) readVerified(fn *ast.FuncDecl, out *diag.Set) (ir.Binding, bool) {
	binding := ir.Binding{
		Target: ir.Target{Kind: ir.TargetFunc, Name: p.PkgPath() + "." + fn.Name.Name},
		Pos:    p.pos(fn.Name.Pos()),
	}

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok || p.specFuncName(call.Fun) != "Verified" {
			return true
		}

		a := ir.Assertion{Kind: ir.AssertVerified, Pos: p.pos(call.Lparen)}
		// The first argument is the logger the line is written through. It
		// carries no specification meaning and is skipped.
		for _, arg := range call.Args[min(1, len(call.Args)):] {
			if name := p.objectName(arg); name != "" {
				a.Requirements = append(a.Requirements, name)
			}
		}
		if len(a.Requirements) == 0 {
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseBinding, 10),
				Pos:  a.Pos,
				What: "spec.Verified names no requirement.",
				Why:  "The call exists to say which requirement this test demonstrates. Without one it writes a line that records nothing and satisfies no obligation, while looking like a test that verifies something.",
				How:  "Pass the requirements, for example spec.Verified(t, quote.RQuoteSubmit).",
			})
			return true
		}
		binding.Assertions = append(binding.Assertions, a)
		return true
	})

	return binding, len(binding.Assertions) > 0
}

// CheckVerifiedOutsideTests rejects a spec.Verified call in production code.
//
// It means nothing there. Nothing runs it under a test runner, so no line is
// ever attributed to a test, and the claim it makes could never be backed by
// anything. Left unreported it would look like verification that simply never
// gets recorded, which is the most expensive kind of silence in this tool.
func (p *Package) CheckVerifiedOutsideTests(out *diag.Set) {
	if p.isTest {
		return
	}
	for _, f := range p.pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || p.specFuncName(call.Fun) != "Verified" {
				return true
			}
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseBinding, 11),
				Pos:  p.pos(call.Lparen),
				What: "spec.Verified outside a test.",
				Why:  "Verification is evidence that a test ran and passed. Outside a test nothing runs it under a test runner, so the line it writes is attributed to nothing and the requirement stays unverified — while the call suggests otherwise.",
				How:  "Move the call into a *_test.go file, or delete it.",
			})
			return true
		})
	}
}
