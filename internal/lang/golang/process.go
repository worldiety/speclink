package golang

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/worldiety/speclink/internal/diag"
	"github.com/worldiety/speclink/internal/ir"
)

// ReadProcesses extracts the process declarations of a package's *.process.go
// files.
//
// Satisfies is recorded as qualified Go identifiers rather than requirement
// IDs, exactly as DerivedFrom is, so that the resolution happens once every
// declaration has been collected and the input order does not matter.
func (p *Package) ReadProcesses(out *diag.Set) []*ir.Process {
	var procs []*ir.Process
	for _, f := range p.processFiles {
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, s := range gd.Specs {
				vs, ok := s.(*ast.ValueSpec)
				if !ok || len(vs.Names) != 1 || len(vs.Values) != 1 {
					continue
				}
				lit, ok := vs.Values[0].(*ast.CompositeLit)
				if !ok || !p.isSpecType(lit, "Process") {
					continue
				}
				procs = append(procs, p.readProcess(vs, lit, out))
			}
		}
	}
	return procs
}

func (p *Package) readProcess(vs *ast.ValueSpec, lit *ast.CompositeLit, out *diag.Set) *ir.Process {
	proc := &ir.Process{
		GoIdent: p.PkgPath() + "." + vs.Names[0].Name,
		Pos:     p.pos(vs.Pos()),
	}

	for _, el := range lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}
		switch key.Name {
		case "ID":
			proc.ID, _ = p.stringArg(kv.Value)
		case "Title":
			proc.Title, _ = p.stringArg(kv.Value)
		case "Purpose":
			proc.Purpose, _ = p.stringArg(kv.Value)
		case "Satisfies":
			proc.Satisfies = p.identList(kv.Value)
		case "Nodes":
			proc.Nodes = p.readNodes(kv.Value, out)
		case "Edges":
			proc.Edges = p.readEdges(kv.Value, out)
		case "Drawn":
			if sel, ok := kv.Value.(*ast.SelectorExpr); ok {
				proc.Drawn, _ = ir.ViewOf(sel.Sel.Name)
			}
		}
	}
	return proc
}

// nodeKinds maps the constructor name to what it builds.
var nodeKinds = map[string]ir.NodeKind{
	"Start":     ir.NodeStart,
	"StartedBy": ir.NodeStart,
	"Send":      ir.NodeSend,
	"End":       ir.NodeEnd,
	"Do":        ir.NodeActivity,
	"Emit":      ir.NodeEmit,
	"On":        ir.NodeCatch,
	"Fork":      ir.NodeFork,
	"Join":      ir.NodeJoin,
	"Choice":    ir.NodeChoice,
	"Merge":     ir.NodeMerge,
}

func (p *Package) readNodes(expr ast.Expr, out *diag.Set) []ir.ProcessNode {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}

	var nodes []ir.ProcessNode
	for _, el := range lit.Elts {
		call, ok := el.(*ast.CallExpr)
		if !ok {
			continue
		}
		// A node may be wrapped in method calls that decorate it —
		// .Note("…") — so the constructor is found by peeling those off
		// first. Their arguments are collected on the way, because the
		// constructor call underneath has no idea they were made.
		call, decor := peelDecorations(p, call)
		name := p.specFuncName(call.Fun)
		kind, known := nodeKinds[name]
		if !known {
			// Anything else in this list is not part of the vocabulary. Saying
			// so beats ignoring it: a node that silently vanishes takes its
			// edges' endpoints with it and is reported as a dangling edge
			// somewhere else, which names the wrong line.
			out.Add(diag.Finding{
				Code: diag.Code(diag.PhaseBinding, 4),
				Pos:  p.pos(call.Pos()),
				What: "not a process node.",
				Why:  "The node list is a closed vocabulary; anything else here would be read as nothing and reported later as a broken edge.",
				How:  "Use one of Start, End, Do, Emit, On, Fork, Join, Choice or Merge.",
			})
			continue
		}

		node := ir.ProcessNode{Kind: kind, Pos: p.pos(call.Pos()), Note: decor.note}

		// StartedBy names the participant first, so every other argument
		// shifts along by one.
		shift := 0
		if name == "StartedBy" {
			shift = 1
			node.Actor = p.participantID(argAt(call, 0))
		}
		node.ID, _ = p.stringArg(argAt(call, shift))
		if kind == ir.NodeStart || kind == ir.NodeEnd {
			node.Label, _ = p.stringArg(argAt(call, shift+1))
		}
		if kind.References() {
			if t, ok := p.typeArg(call, 0); ok {
				node.Ref = typeName(t)
				node.RefPackage = typePackage(t)
			}
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func (p *Package) readEdges(expr ast.Expr, _ *diag.Set) []ir.ProcessEdge {
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return nil
	}

	var edges []ir.ProcessEdge
	for _, el := range lit.Elts {
		el, ok := el.(*ast.CompositeLit)
		if !ok {
			continue
		}
		edge := ir.ProcessEdge{Pos: p.pos(el.Pos())}
		for _, f := range el.Elts {
			kv, ok := f.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok {
				continue
			}
			switch key.Name {
			case "From":
				edge.From, _ = p.stringArg(kv.Value)
			case "To":
				edge.To, _ = p.stringArg(kv.Value)
			case "When":
				edge.When, _ = p.stringArg(kv.Value)
			}
		}
		edges = append(edges, edge)
	}
	return edges
}

// typePackage returns the import path a named type lives in.
//
// It is recorded beside the name because the two answer different questions. A
// reference that resolves to nothing is a mistake; a reference into a package
// the scope left out is not measurable, and reporting the second as the first
// would fail a project for a setting it chose on purpose.
func typePackage(t types.Type) string {
	named, ok := t.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return ""
	}
	return named.Obj().Pkg().Path()
}

// decorations are the optional extras a node carries, collected while the
// method calls wrapping its constructor are peeled off.
type decorations struct {
	note string
}

// peelDecorations unwraps X.Note("…") down to the constructor call.
func peelDecorations(p *Package, call *ast.CallExpr) (*ast.CallExpr, decorations) {
	var d decorations
	for {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return call, d
		}
		inner, ok := sel.X.(*ast.CallExpr)
		if !ok {
			return call, d
		}
		if sel.Sel.Name == "Note" {
			d.note, _ = p.stringArg(argAt(call, 0))
		}
		call = inner
	}
}

// participantID returns the qualified identifier of the actor a beginning
// names.
//
// The identifier and not the ID, because the participant may be declared in a
// package this one does not read. It is resolved in a second pass once the
// topology is in hand, the same shape the requirements and the themes go
// through and for the same reason: the order of files must not decide what a
// drawing says.
func (p *Package) participantID(e ast.Expr) string {
	id, ok := e.(*ast.Ident)
	if !ok {
		if sel, isSel := e.(*ast.SelectorExpr); isSel {
			id = sel.Sel
		} else {
			return ""
		}
	}
	obj := p.pkg.TypesInfo.Uses[id]
	if obj == nil {
		obj = p.pkg.TypesInfo.Defs[id]
	}
	if obj == nil {
		return ""
	}
	if obj.Pkg() == nil {
		return ""
	}
	return obj.Pkg().Path() + "." + obj.Name()
}
