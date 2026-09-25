package syntax

import (
	"strconv"
	"strings"
)

// Parse reads the items of one file.
//
// It returns a File even when there are errors, because a file with one
// unreadable item still has every other item in it, and the positions of what
// could be read are what the reader of a finding needs.
func Parse(path, src string) *File {
	toks, errs := Lex(src)
	f := &File{Path: path, Src: src, Toks: toks, Errors: errs}
	p := &parser{f: f, toks: toks}
	f.Items, f.Attrs, f.Doc = p.items(false)
	return f
}

type parser struct {
	f    *File
	toks []Token
	i    int
}

func (p *parser) tok() Token { return p.toks[p.i] }

func (p *parser) peek(n int) Token {
	if p.i+n < len(p.toks) {
		return p.toks[p.i+n]
	}
	return p.toks[len(p.toks)-1]
}

func (p *parser) at(s string) bool { return p.tok().Is(s) }
func (p *parser) eof() bool        { return p.tok().Kind == EOF }

func (p *parser) next() Token {
	t := p.toks[p.i]
	if t.Kind != EOF {
		p.i++
	}
	return t
}

func (p *parser) errorf(pos Pos, msg string) {
	p.f.Errors = append(p.f.Errors, Error{Pos: pos, Msg: msg})
}

func (p *parser) expect(s string) bool {
	if p.at(s) {
		p.next()
		return true
	}
	p.errorf(p.tok().Pos, "expected "+strconv.Quote(s)+", found "+describe(p.tok()))
	return false
}

func describe(t Token) string {
	if t.Kind == EOF {
		return "end of file"
	}
	return strconv.Quote(t.Text)
}

func isOpen(t Token) bool {
	return t.Kind == Punct && (t.Text == "(" || t.Text == "[" || t.Text == "{")
}
func isClose(t Token) bool {
	return t.Kind == Punct && (t.Text == ")" || t.Text == "]" || t.Text == "}")
}

// items reads items until the end of the file, or until the closing brace of
// the block the cursor is inside when closed is set.
func (p *parser) items(closed bool) (items []*Item, inner []Attr, doc string) {
	var docs []string
	for {
		t := p.tok()
		switch {
		case t.Kind == EOF:
			if closed {
				p.errorf(t.Pos, "missing }")
			}
			return items, inner, strings.Join(docs, "\n")
		case closed && t.Is("}"):
			p.next()
			return items, inner, strings.Join(docs, "\n")
		case t.Kind == DocInner:
			docs = append(docs, docLine(t.Text))
			p.next()
		case t.Is("#") && p.peek(1).Is("!") && p.peek(2).Is("["):
			inner = append(inner, p.attr())
		case t.Is(";"):
			p.next()
		case isClose(t):
			p.errorf(t.Pos, "unexpected "+strconv.Quote(t.Text))
			p.next()
		default:
			items = append(items, p.item())
		}
	}
}

// docLine drops the one space conventionally written after ///.
func docLine(s string) string { return strings.TrimPrefix(s, " ") }

// attr reads #[…] or #![…].
func (p *parser) attr() Attr {
	a := Attr{Pos: p.tok().Pos}
	p.next() // #
	if p.at("!") {
		a.Inner = true
		p.next()
	}
	if !p.at("[") {
		p.errorf(p.tok().Pos, "expected [ after #")
		return a
	}
	body := p.group()
	toks := p.f.Tokens(body)

	// The path runs up to the first delimiter or =.
	n := 0
	for n < len(toks) && (toks[n].Kind == Ident || toks[n].Is("::")) {
		n++
	}
	a.Path = Render(toks[:n])
	rest := toks[n:]
	switch {
	case len(rest) > 0 && rest[0].Is("="):
		a.Eq = true
		a.Args = rest[1:]
	case len(rest) >= 2 && isOpen(rest[0]) && isClose(rest[len(rest)-1]):
		a.Args = rest[1 : len(rest)-1]
	default:
		a.Args = rest
	}
	return a
}

// group consumes a delimited group and returns the range inside it. Kinds of
// bracket are not matched against each other: a file where they do not match
// does not compile, and cargo has said so before this runs.
func (p *parser) group() Range {
	open := p.next()
	from := p.i
	depth := 1
	for {
		t := p.tok()
		if t.Kind == EOF {
			p.errorf(open.Pos, "unclosed "+strconv.Quote(open.Text))
			return Range{from, p.i}
		}
		if isOpen(t) {
			depth++
		} else if isClose(t) {
			depth--
			if depth == 0 {
				r := Range{from, p.i}
				p.next()
				return r
			}
		}
		p.next()
	}
}

// scan advances to the first token at nesting depth zero for which stop is
// true, or to a closing bracket that ends the enclosing group, and returns
// what it passed over.
//
// With angles set, < and > count as brackets, which is right in a type and
// wrong in an expression: Vec<(A, B)> holds a comma that is not a separator,
// while a < b holds an angle that is not a bracket.
func (p *parser) scan(stop func(Token) bool, angles bool) Range {
	from := p.i
	depth, angle := 0, 0
	for {
		t := p.tok()
		if t.Kind == EOF {
			return Range{from, p.i}
		}
		if depth == 0 && angle == 0 && stop(t) {
			return Range{from, p.i}
		}
		switch {
		case isOpen(t):
			depth++
		case isClose(t):
			if depth == 0 {
				return Range{from, p.i}
			}
			depth--
		case angles && depth == 0 && t.Is("<"):
			angle++
		case angles && depth == 0 && t.Is(">") && angle > 0:
			angle--
		}
		p.next()
	}
}

func stopAt(words ...string) func(Token) bool {
	return func(t Token) bool {
		for _, w := range words {
			if t.Is(w) {
				return true
			}
		}
		return false
	}
}

// generics reads <…> if the cursor is at one.
func (p *parser) generics() Range {
	if !p.at("<") {
		return Range{}
	}
	open := p.next()
	from := p.i
	angle, brace := 1, 0
	for !p.eof() {
		t := p.tok()
		switch {
		case t.Is("{"):
			brace++
		case t.Is("}"):
			brace--
		case brace == 0 && t.Is("<"):
			angle++
		case brace == 0 && t.Is(">"):
			angle--
			if angle == 0 {
				r := Range{from, p.i}
				p.next()
				return r
			}
		}
		p.next()
	}
	p.errorf(open.Pos, "unclosed <")
	return Range{from, p.i}
}

func (p *parser) where() Range {
	if !p.at("where") {
		return Range{}
	}
	p.next()
	return p.scan(stopAt("{", ";", "="), true)
}

func (p *parser) name() string {
	if t := p.tok(); t.Kind == Ident {
		p.next()
		return t.Text
	}
	p.errorf(p.tok().Pos, "expected a name, found "+describe(p.tok()))
	return ""
}

// outer reads the attributes and documentation in front of an item or field.
func (p *parser) outer() (attrs []Attr, doc string) {
	var docs []string
	for {
		switch t := p.tok(); {
		case t.Kind == DocOuter || t.Kind == DocInner:
			docs = append(docs, docLine(t.Text))
			p.next()
		case t.Is("#") && (p.peek(1).Is("[") || p.peek(1).Is("!") && p.peek(2).Is("[")):
			a := p.attr()
			if a.Path == "doc" && a.Eq {
				if v, ok := a.Value(); ok {
					docs = append(docs, docLine(v))
				}
			}
			attrs = append(attrs, a)
		default:
			return attrs, strings.Join(docs, "\n")
		}
	}
}

// vis reads a visibility.
func (p *parser) vis() string {
	if !p.at("pub") {
		return ""
	}
	p.next()
	if p.at("(") {
		inner := p.peek(1)
		if inner.Is("crate") || inner.Is("super") || inner.Is("self") || inner.Is("in") {
			r := p.group()
			return "pub(" + p.f.Text(r) + ")"
		}
	}
	return "pub"
}

func (p *parser) item() *Item {
	startIdx := p.i
	attrs, doc := p.outer()
	it := &Item{Attrs: attrs, Doc: doc, Start: p.toks[startIdx].Pos.Offset, Pos: p.tok().Pos}
	defer func() { it.All = Range{startIdx, p.i} }()

	if isClose(p.tok()) || p.eof() {
		p.errorf(it.Pos, "attributes without an item")
		it.Kind = ItemOpaque
		return it
	}
	it.Vis = p.vis()
	it.Pos = p.tok().Pos
	p.qualifiers(it)

	t := p.tok()
	switch {
	case t.Is("fn"):
		p.fn(it)
	case t.Is("mod"):
		p.mod(it)
	case t.Is("use"):
		p.use(it)
	case t.Is("struct"):
		it.Kind = ItemStruct
		p.structure(it)
	case t.Is("union") && p.peek(1).Kind == Ident:
		it.Kind = ItemUnion
		p.structure(it)
	case t.Is("enum"):
		p.enum(it)
	case t.Is("trait") || t.Is("auto") && p.peek(1).Is("trait"):
		p.trait(it)
	case t.Is("impl"):
		p.impl(it)
	case t.Is("const"):
		it.Kind = ItemConst
		p.constant(it)
	case t.Is("static"):
		it.Kind = ItemStatic
		p.constant(it)
	case t.Is("type"):
		p.alias(it)
	case t.Is("extern") && p.peek(1).Is("crate"):
		p.externCrate(it)
	case t.Is("macro_rules") && p.peek(1).Is("!"):
		p.macroRules(it)
	case p.macroAhead():
		p.macroCall(it)
	default:
		p.opaque(it)
	}
	return it
}

// qualifiers reads the words that may stand in front of fn, impl and trait.
func (p *parser) qualifiers(it *Item) {
	for {
		t, n := p.tok(), p.peek(1)
		switch {
		case t.Is("default") && n.Kind == Ident:
			p.next()
		case t.Is("const") && (n.Is("fn") || n.Is("unsafe") || n.Is("async") || n.Is("extern")):
			it.Const = true
			p.next()
		case t.Is("async") && (n.Is("fn") || n.Is("unsafe") || n.Is("extern")):
			it.Async = true
			p.next()
		case t.Is("unsafe") && (n.Is("fn") || n.Is("impl") || n.Is("trait") || n.Is("extern") || n.Is("auto") || n.Is("mod")):
			it.Unsafe = true
			p.next()
		case t.Is("safe") && n.Kind == Ident:
			p.next()
		case t.Is("extern") && n.Is("fn"):
			p.next()
		case t.Is("extern") && n.Kind == Literal && (p.peek(2).Is("fn") || p.peek(2).Is("unsafe")):
			p.next()
			p.next()
		default:
			return
		}
	}
}

func (p *parser) fn(it *Item) {
	it.Kind = ItemFn
	p.next()
	it.Name = p.name()
	it.Generics = p.generics()
	if p.at("(") {
		it.Params = p.group()
	} else {
		p.errorf(p.tok().Pos, "expected ( after the name of fn "+it.Name)
	}
	if p.at("->") {
		p.next()
		it.Ret = p.scan(stopAt("{", ";", "where"), true)
	}
	it.Where = p.where()
	switch {
	case p.at("{"):
		from := p.i
		p.group()
		it.Body = Range{from, p.i}
	case p.at(";"):
		p.next()
	default:
		p.errorf(p.tok().Pos, "expected the body of fn "+it.Name)
	}
}

func (p *parser) mod(it *Item) {
	it.Kind = ItemMod
	p.next()
	it.Name = p.name()
	switch {
	case p.at(";"):
		p.next()
		it.External = true
	case p.at("{"):
		p.next()
		var doc string
		it.Items, it.Inner, doc = p.items(true)
		if doc != "" {
			it.Doc = strings.TrimPrefix(it.Doc+"\n"+doc, "\n")
		}
	default:
		p.errorf(p.tok().Pos, "expected ; or { after mod "+it.Name)
	}
}

func (p *parser) use(it *Item) {
	it.Kind = ItemUse
	p.next()
	p.useTree(nil, &it.Uses)
	p.expect(";")
}

// useTree flattens one use tree into its imported paths. A leading :: is
// kept as the first segment, because it means an external crate and nothing
// else.
func (p *parser) useTree(prefix []string, out *[]Use) {
	pos := p.tok().Pos
	path := append([]string(nil), prefix...)
	if p.at("::") {
		p.next()
		if len(path) == 0 {
			path = append(path, "::")
		}
	}
	for {
		switch t := p.tok(); {
		case t.Is("*"):
			p.next()
			*out = append(*out, Use{Path: path, Glob: true, Pos: pos})
			return
		case t.Is("{"):
			p.next()
			for !p.at("}") && !p.eof() {
				p.useTree(path, out)
				if !p.at(",") {
					break
				}
				p.next()
			}
			p.expect("}")
			return
		case t.Kind == Ident:
			p.next()
			seg := t.Text
			if p.at("::") {
				p.next()
				path = append(path, seg)
				continue
			}
			alias := ""
			if p.at("as") {
				p.next()
				alias = p.name()
			}
			// use a::{self} imports a itself.
			if seg != "self" || len(path) == 0 {
				path = append(path, seg)
			}
			*out = append(*out, Use{Path: path, Alias: alias, Pos: pos})
			return
		default:
			p.errorf(t.Pos, "unreadable use tree at "+describe(t))
			return
		}
	}
}

func (p *parser) structure(it *Item) {
	p.next()
	it.Name = p.name()
	it.Generics = p.generics()
	it.Where = p.where()
	switch {
	case p.at("{"):
		it.Fields = p.namedFields()
	case p.at("("):
		it.Tuple = true
		it.Fields = p.tupleFields()
		it.Where = p.where()
		p.expect(";")
	case p.at(";"):
		p.next()
	default:
		p.errorf(p.tok().Pos, "expected the fields of "+it.Name)
	}
}

func (p *parser) namedFields() []Field {
	p.next() // {
	var out []Field
	for !p.at("}") && !p.eof() {
		before := p.i
		attrs, doc := p.outer()
		if p.at("}") {
			break
		}
		f := Field{Attrs: attrs, Doc: doc, Vis: p.vis(), Pos: p.tok().Pos}
		f.Name = p.name()
		p.expect(":")
		f.Type = p.scan(stopAt(","), true)
		out = append(out, f)
		if p.at(",") {
			p.next()
		}
		if p.i == before {
			p.next()
		}
	}
	p.expect("}")
	return out
}

func (p *parser) tupleFields() []Field {
	p.next() // (
	var out []Field
	for !p.at(")") && !p.eof() {
		before := p.i
		attrs, doc := p.outer()
		if p.at(")") {
			break
		}
		f := Field{Attrs: attrs, Doc: doc, Vis: p.vis(), Pos: p.tok().Pos, Name: strconv.Itoa(len(out))}
		f.Type = p.scan(stopAt(","), true)
		out = append(out, f)
		if p.at(",") {
			p.next()
		}
		if p.i == before {
			p.next()
		}
	}
	p.expect(")")
	return out
}

func (p *parser) enum(it *Item) {
	it.Kind = ItemEnum
	p.next()
	it.Name = p.name()
	it.Generics = p.generics()
	it.Where = p.where()
	if !p.expect("{") {
		return
	}
	for !p.at("}") && !p.eof() {
		before := p.i
		attrs, doc := p.outer()
		if p.at("}") {
			break
		}
		p.vis()
		v := Variant{Attrs: attrs, Doc: doc, Pos: p.tok().Pos}
		v.Name = p.name()
		switch {
		case p.at("{"):
			v.Fields = p.namedFields()
		case p.at("("):
			v.Tuple = true
			v.Fields = p.tupleFields()
		default:
			v.Unit = true
		}
		if p.at("=") {
			p.next()
			v.Discriminant = p.scan(stopAt(","), false)
		}
		it.Variants = append(it.Variants, v)
		if p.at(",") {
			p.next()
		}
		if p.i == before {
			p.next()
		}
	}
	p.expect("}")
}

func (p *parser) trait(it *Item) {
	it.Kind = ItemTrait
	if p.at("auto") {
		p.next()
	}
	p.next() // trait
	it.Name = p.name()
	it.Generics = p.generics()
	if p.at(":") {
		p.next()
		p.scan(stopAt("{", "where", ";", "="), true)
	}
	if p.at("=") {
		// A trait alias has no items to read.
		p.next()
		p.scan(stopAt(";"), true)
		p.expect(";")
		return
	}
	it.Where = p.where()
	if p.expect("{") {
		it.Items, it.Inner, _ = p.items(true)
	}
}

func (p *parser) impl(it *Item) {
	it.Kind = ItemImpl
	p.next()
	it.Generics = p.generics()
	if p.at("const") {
		p.next()
	}
	if p.at("!") {
		it.Negative = true
		p.next()
	}

	// The header is either a type or a trait, for, and a type. The for is
	// found at nesting depth zero, and one followed by < is a higher ranked
	// bound rather than the separator.
	from := p.i
	forAt := -1
	depth, angle := 0, 0
	for !p.eof() {
		t := p.tok()
		if depth == 0 && angle == 0 && (t.Is("{") || t.Is("where")) {
			break
		}
		switch {
		case isOpen(t):
			depth++
		case isClose(t):
			depth--
		case depth == 0 && t.Is("<"):
			angle++
		case depth == 0 && t.Is(">") && angle > 0:
			angle--
		case depth == 0 && angle == 0 && t.Is("for") && !p.peek(1).Is("<") && forAt < 0:
			forAt = p.i
		}
		p.next()
	}
	if forAt >= 0 {
		it.Trait = Range{from, forAt}
		it.SelfType = Range{forAt + 1, p.i}
	} else {
		it.SelfType = Range{from, p.i}
	}
	it.Where = p.where()
	if p.expect("{") {
		it.Items, it.Inner, _ = p.items(true)
	}
}

func (p *parser) constant(it *Item) {
	p.next()
	if it.Kind == ItemStatic && p.at("mut") {
		it.Mut = true
		p.next()
	}
	it.Name = p.name()
	it.Generics = p.generics()
	if p.at(":") {
		p.next()
		it.Type = p.scan(stopAt("=", ";"), true)
	}
	if p.at("=") {
		p.next()
		it.Value = p.scan(stopAt(";"), false)
	}
	p.expect(";")
}

func (p *parser) alias(it *Item) {
	it.Kind = ItemType
	p.next()
	it.Name = p.name()
	it.Generics = p.generics()
	if p.at(":") {
		p.next()
		p.scan(stopAt("=", ";", "where"), true)
	}
	it.Where = p.where()
	if p.at("=") {
		p.next()
		it.Type = p.scan(stopAt(";", "where"), true)
		if w := p.where(); !w.Empty() {
			it.Where = w
		}
	}
	p.expect(";")
}

func (p *parser) externCrate(it *Item) {
	it.Kind = ItemExternCrate
	p.next()
	p.next()
	it.Name = p.name()
	u := Use{Path: []string{"::", it.Name}, Pos: it.Pos}
	if p.at("as") {
		p.next()
		u.Alias = p.name()
	}
	it.Uses = []Use{u}
	p.expect(";")
}

func (p *parser) macroRules(it *Item) {
	it.Kind = ItemMacroRules
	it.Macro = "macro_rules"
	p.next()
	p.next()
	it.Name = p.name()
	p.macroArgs(it)
}

// macroAhead reports whether the cursor is at path!( or path![ or path!{.
func (p *parser) macroAhead() bool {
	n := 0
	if p.peek(n).Is("::") {
		n++
	}
	for {
		if p.peek(n).Kind != Ident {
			return false
		}
		n++
		if p.peek(n).Is("::") {
			n++
			continue
		}
		return p.peek(n).Is("!") && isOpen(p.peek(n+1))
	}
}

func (p *parser) macroCall(it *Item) {
	it.Kind = ItemMacroCall
	from := p.i
	for !p.at("!") {
		p.next()
	}
	it.Macro = p.f.Text(Range{from, p.i})
	p.next() // !
	p.macroArgs(it)
}

func (p *parser) macroArgs(it *Item) {
	if !isOpen(p.tok()) {
		p.errorf(p.tok().Pos, "expected the arguments of "+it.Macro+"!")
		return
	}
	it.Delim = p.tok().Text
	it.Args = p.group()
	if it.Delim != "{" {
		p.expect(";")
	}
}

// opaque consumes what this does not read: up to a semicolon, or through a
// braced block, at depth zero. It always consumes at least one token.
func (p *parser) opaque(it *Item) {
	it.Kind = ItemOpaque
	start := p.i
	for !p.eof() {
		t := p.tok()
		if isClose(t) {
			break
		}
		if t.Is(";") {
			p.next()
			return
		}
		if isOpen(t) {
			brace := t.Is("{")
			p.group()
			if brace {
				if p.at(";") {
					p.next()
				}
				return
			}
			continue
		}
		p.next()
	}
	if p.i == start {
		p.next()
	}
}
