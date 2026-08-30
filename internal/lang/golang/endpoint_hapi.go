package golang

import (
	"go/ast"
	"go/types"

	"github.com/worldiety/speclink/internal/ir"
)

// Recognising the routes of a fluent builder.
//
// # Why this is a second recogniser rather than a second spelling
//
// The standard library's router is mounted by a call that carries everything at
// once: a pattern and a handler, side by side, in one argument list. hapi
// carries the same facts across a chain — the method in the name of the
// function that starts it, the path in a struct literal handed to it, the work
// in closures passed to methods further along — and no single call in that
// chain holds more than a third of the answer. A recogniser that looked for one
// call would have to pick which third to believe.
//
// # What this buys that the standard library's router cannot
//
// The wire types. `hapi.Post[In]` and `hapi.ToJSON[In, Out]` state what crosses
// the boundary in type arguments the compiler has already resolved, so Request
// and Response are read here rather than guessed — the first dialect where
// speclink fills them in at all. On a bare mux they stay empty, because there
// the only way to obtain them would be to assume that the use case's own
// parameters are the wire shape, and a presentation layer that maps a request
// body onto a command makes that assumption silently wrong. The fixture for
// this recogniser maps SubmitQuoteBody onto SubmitQuoteCmd for exactly that
// reason: to pin that the two are not the same thing.

// hapiVerbs are the functions that begin a route, and the method each states.
//
// Endpoint is the general form and names no method of its own: it takes one
// from the operation, and reading it from there rather than defaulting to GET
// is the difference between what the route answers on and what the framework
// falls back to. A default is the framework's business; a catalogue that
// printed it as though the code had said it would be inventing a promise.
var hapiVerbs = map[string]string{
	"Post":     "POST",
	"Get":      "GET",
	"Put":      "PUT",
	"Delete":   "DELETE",
	"Endpoint": "",
}

// hapiSites finds every route mounted through the fluent builder.
func (p *Package) hapiSites(fw Framework) []site {
	if fw.Rest == "" {
		return nil
	}

	var out []site
	// A chain is several calls and only the outermost is a registration. The
	// inner ones are recorded so that the walk, which meets the outermost
	// first, does not report the same route once per link.
	inChain := map[ast.Node]bool{}

	for _, file := range p.pkg.Syntax {
		if p.isGeneratedByUs(file) {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok || inChain[call] {
				return true
			}
			links := chainOf(call)
			root := links[len(links)-1]
			verb, ok := p.hapiVerb(root, fw)
			if !ok {
				return true
			}
			for _, l := range links {
				inChain[l] = true
			}
			out = append(out, p.hapiSite(links, root, verb))
			return true
		})
	}
	return out
}

// chainOf returns the calls of a fluent chain, outermost first.
//
// The last one is the call that started it. Walking down rather than up because
// an expression knows what it was called on and not what will be called on it,
// and the outermost call is the only one the syntax hands over whole.
func chainOf(call *ast.CallExpr) []*ast.CallExpr {
	links := []*ast.CallExpr{call}
	for {
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return links
		}
		inner, ok := sel.X.(*ast.CallExpr)
		if !ok {
			return links
		}
		call = inner
		links = append(links, call)
	}
}

// hapiVerb reports whether a call begins a route, and which method it states.
func (p *Package) hapiVerb(call *ast.CallExpr, fw Framework) (string, bool) {
	fn, _ := p.calleeOf(call)
	if fn == nil || fn.Pkg() == nil || fn.Pkg().Path() != fw.Rest {
		return "", false
	}
	method, ok := hapiVerbs[fn.Name()]
	if !ok {
		return "", false
	}
	// Two arguments, the API and the operation. Anything else is a function
	// that shares a name with a verb and is not one.
	if len(call.Args) != 2 {
		return "", false
	}
	return method, true
}

// hapiSite reads one chain into a registration.
func (p *Package) hapiSite(links []*ast.CallExpr, root *ast.CallExpr, verb string) site {
	s := site{
		method:       verb,
		handler:      root,
		pos:          root.Pos(),
		pattern:      root.Args[1],
		shapesStated: true,
	}

	// The operation states the path, and the method too where the general form
	// was used. A field that is not a constant leaves the path empty, which is
	// the finding the pattern rules already make: an address that only exists
	// at run time is one no catalogue can name.
	if path, ok := p.hapiField(root.Args[1], "Path"); ok {
		s.path = path
	}
	if s.method == "" {
		if method, ok := p.hapiField(root.Args[1], "Method"); ok {
			s.method = method
		}
	}

	// What crosses is read from the options, not from the verb's own type
	// argument.
	//
	// That type is the assembled request model: the bearer subject, the
	// headers, the query and the body all mapped into one struct. It is what
	// the handler receives and it is emphatically not what the caller sends,
	// so promising it would promise a shape no client can produce — a field
	// carrying an auth.Subject among them. The body is stated separately, by
	// the option that reads it, and that is the one a caller depends on.

	// Every argument of every link is traced, and the links are traced apart
	// rather than as one expression. Request and Response are siblings in
	// meaning: authentication on one and the work on the other are both what
	// the route does, and the rule that stops at the shallowest evidence would
	// otherwise take whichever half the syntax happened to put outermost and
	// never look at the other.
	for _, l := range links {
		s.trace = append(s.trace, l.Args...)

		body := p.hapiBody(l, root)
		if body == nil {
			continue
		}
		switch linkName(l) {
		case "Request":
			if s.request == "" {
				s.request = typeName(body)
				s.requestShape = p.wireShape(body)
			}
		case "Response":
			if s.response == "" {
				s.response = typeName(body)
				s.responseShape = p.wireShape(body)
			}
		}
	}
	return s
}

// linkName returns the method a chain link calls, empty for the call that
// started the chain.
func linkName(link *ast.CallExpr) string {
	sel, ok := deparen(link.Fun).(*ast.SelectorExpr)
	if !ok {
		return ""
	}
	return sel.Sel.Name
}

// hapiBody reads the body an option maps, in whichever direction it maps it.
//
// Recognised by shape rather than by name: an option that carries two type
// arguments whose first is the request model is one that states a structured
// body, and that is true of the ones that read JSON in and of the ones that
// write JSON out alike. The options that map a header, a query parameter or a
// file carry one type argument and are correctly not bodies.
//
// A response written as raw bytes states nothing here, and that is the truthful
// answer rather than a gap: the route genuinely promises no shape.
func (p *Package) hapiBody(link *ast.CallExpr, root *ast.CallExpr) types.Type {
	request := ""
	if args := p.typeArgsOf(root); args != nil && args.Len() > 0 {
		request = typeName(args.At(0))
	}
	for _, arg := range link.Args {
		call, ok := arg.(*ast.CallExpr)
		if !ok {
			continue
		}
		args := p.typeArgsOf(call)
		if args == nil || args.Len() != 2 {
			continue
		}
		if typeName(args.At(0)) != request {
			continue
		}
		return args.At(1)
	}
	return nil
}

// wireShape reads a body the way a persisted type is read.
//
// The same expansion through named types, so that a change inside a nested
// struct reaches the string a rule compares. Fields are collected where the
// body is a struct, because the two directions of a boundary break
// asymmetrically and only a field by field view can tell a removal from an
// addition; a body that is a list or a string has none, and the whole shape is
// then what carries the promise.
func (p *Package) wireShape(t types.Type) *ir.WireShape {
	w := &ir.WireShape{Type: typeName(t), Shape: shapeOf(t, map[*types.Named]bool{})}
	if st, ok := t.Underlying().(*types.Struct); ok {
		w.Fields = p.readFields(st)
	}
	return w
}

// hapiField reads a constant string out of a field of a struct literal.
//
// Only from a literal written at the call. An operation built elsewhere and
// passed in is not read, and leaving the path empty then is deliberate: the
// pattern rules report an address this run could not determine, which is the
// honest outcome and the one that gets fixed.
func (p *Package) hapiField(e ast.Expr, name string) (string, bool) {
	lit, ok := deparen(e).(*ast.CompositeLit)
	if !ok {
		return "", false
	}
	for _, elt := range lit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != name {
			continue
		}
		return p.stringValue(kv.Value)
	}
	return "", false
}

// calleeOf returns the function a call resolves to, through any instantiation.
func (p *Package) calleeOf(call *ast.CallExpr) (*types.Func, *ast.Ident) {
	fun := deparen(call.Fun)
	switch x := fun.(type) {
	case *ast.IndexExpr:
		fun = deparen(x.X)
	case *ast.IndexListExpr:
		fun = deparen(x.X)
	}

	var id *ast.Ident
	switch x := fun.(type) {
	case *ast.SelectorExpr:
		id = x.Sel
	case *ast.Ident:
		id = x
	default:
		return nil, nil
	}
	fn, _ := p.pkg.TypesInfo.Uses[id].(*types.Func)
	return fn, id
}

// typeArgsOf returns the type arguments a call was instantiated with.
//
// Read from the type checker rather than from the syntax, so that a call which
// left them to inference is understood exactly as one that wrote them out. The
// two spellings mean the same thing to the compiler and must mean the same
// thing to a catalogue.
func (p *Package) typeArgsOf(call *ast.CallExpr) *types.TypeList {
	_, id := p.calleeOf(call)
	if id == nil {
		return nil
	}
	inst, ok := p.pkg.TypesInfo.Instances[id]
	if !ok {
		return nil
	}
	return inst.TypeArgs
}

// deparen removes redundant parentheses.
func deparen(e ast.Expr) ast.Expr {
	for {
		par, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = par.X
	}
}
