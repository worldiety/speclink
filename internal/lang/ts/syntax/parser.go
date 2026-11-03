package syntax

import "strings"

// Parse reads a whole TypeScript module.
func Parse(path, src string) *File { return ParseRegion(path, src, 0, len(src)) }

// ParseRegion reads the part of src between from and to as a module, with
// positions relative to the whole of src. It is how the script block of a
// single file component is read without losing the line numbers a reader sees.
func ParseRegion(path, src string, from, to int) *File {
	start := Pos{Offset: from, Line: 1 + strings.Count(src[:from], "\n"), Col: 1}
	if i := strings.LastIndexByte(src[:from], '\n'); i >= 0 {
		start.Col = from - i
	} else {
		start.Col = from + 1
	}
	toks, errs := Lex(src[from:to], start)
	f := &File{Path: path, Src: src, Toks: toks, Errors: errs}
	p := &parser{f: f, t: toks}
	p.module()
	return f
}

type parser struct {
	f *File
	t []Token
	i int
}

func (p *parser) at(i int) Token {
	if i < len(p.t) {
		return p.t[i]
	}
	return p.t[len(p.t)-1]
}

func (p *parser) errorf(at Token, msg string) {
	p.f.Errors = append(p.f.Errors, Error{Pos: at.Pos, Msg: msg})
}

func (p *parser) module() {
	for p.at(p.i).Kind != EOF {
		start := p.i
		if it := p.statement(); it != nil {
			if it.Range.Empty() {
				it.Range = Range{start, p.i}
			}
			p.f.Items = append(p.f.Items, it)
		}
		if p.i <= start {
			p.i = start + 1
		}
	}
}

func isOpener(t Token) bool {
	return t.Kind == TemplateHead || t.Kind == Punct && (t.Text == "(" || t.Text == "[" || t.Text == "{")
}

func isCloser(t Token) bool {
	return t.Kind == TemplateTail || t.Kind == Punct && (t.Text == ")" || t.Text == "]" || t.Text == "}")
}

// skipGroup returns the index just past the group opened at i.
func (p *parser) skipGroup(i int) int {
	depth := 0
	for j := i; ; j++ {
		t := p.at(j)
		if t.Kind == EOF {
			return j
		}
		if isOpener(t) {
			depth++
		} else if isCloser(t) {
			depth--
			if depth == 0 {
				return j + 1
			}
		}
	}
}

// skipAngles returns the index just past a type argument or parameter list
// opened at i.
func (p *parser) skipAngles(i int) int {
	depth := 0
	for j := i; ; {
		t := p.at(j)
		switch {
		case t.Kind == EOF:
			return j
		case isOpener(t):
			j = p.skipGroup(j)
			continue
		case isCloser(t) || t.Is(";"):
			return j
		case t.Is("<"):
			depth++
		case t.Is(">"):
			depth--
			if depth == 0 {
				return j + 1
			}
		}
		j++
	}
}

// endsStatement reports whether a token can be the last of a statement, which
// is the condition under which a line break ends one.
func endsStatement(t Token) bool {
	switch t.Kind {
	case Ident, PrivateName, String, Template, TemplateTail, Number, Regex:
		return true
	case Punct:
		switch t.Text {
		case ")", "]", "}", "++", "--":
			return true
		}
	}
	return false
}

// continues reports whether a token at the start of a line carries the
// previous line on rather than starting a statement of its own.
func continues(t Token) bool {
	switch t.Kind {
	case Punct:
		switch t.Text {
		case "!", "~", "++", "--", "{", "@", ";", "}":
			return false
		}
		return true
	case Ident:
		switch t.Text {
		case "as", "satisfies", "instanceof", "in":
			return true
		}
	}
	return false
}

// stmtEnd returns the index just past the statement starting at i: past its
// semicolon, or at the line break, closing brace or end of file that ends it.
func (p *parser) stmtEnd(i int) int {
	for j := i; ; {
		t := p.at(j)
		switch {
		case t.Kind == EOF:
			return j
		case j > i && t.NL && endsStatement(p.at(j-1)) && !continues(t):
			return j
		case t.Is(";"):
			return j + 1
		case isCloser(t):
			return j
		case isOpener(t):
			j = p.skipGroup(j)
			continue
		}
		j++
	}
}

// trimSemi drops a trailing semicolon from a range.
func (p *parser) trimSemi(r Range) Range {
	if !r.Empty() && p.at(r.To-1).Is(";") {
		r.To--
	}
	return r
}

func (p *parser) statement() *Item {
	t := p.at(p.i)
	switch {
	case t.Is(";"):
		p.i++
		return nil
	case t.Is("@"):
		p.skipDecorators()
		return p.statement()
	case t.Is("import"):
		next := p.at(p.i + 1)
		if next.Is("(") || next.Is(".") {
			return p.expression(false)
		}
		if next.Kind == Ident && !next.Is("type") && p.at(p.i+2).Is("=") {
			return p.other()
		}
		return p.importDecl()
	case t.Is("export"):
		return p.exportDecl()
	}
	if it := p.declaration(false, false); it != nil {
		return it
	}
	switch {
	case (t.Is("namespace") || t.Is("module")) && (p.at(p.i+1).Kind == Ident || p.at(p.i+1).Kind == String) && !p.at(p.i+1).NL:
		return p.other()
	case t.Is("if") || t.Is("for") || t.Is("while") || t.Is("do") || t.Is("try") || t.Is("switch") || t.Is("{"):
		return p.other()
	}
	return p.expression(false)
}

func (p *parser) skipDecorators() {
	for p.at(p.i).Is("@") {
		p.i++
		for p.at(p.i).Kind == Ident {
			p.i++
			if !p.at(p.i).Is(".") {
				break
			}
			p.i++
		}
		if p.at(p.i).Is("(") {
			p.i = p.skipGroup(p.i)
		}
	}
}

// other consumes a statement nothing reads, keeping only its extent.
func (p *parser) other() *Item {
	start := p.i
	end := p.stmtEnd(p.i)
	p.i = end
	return &Item{Kind: ItemOther, Range: Range{start, end}, Pos: p.at(start).Pos}
}

func (p *parser) expression(exported bool) *Item {
	start := p.i
	end := p.stmtEnd(p.i)
	p.i = end
	return &Item{
		Kind:     ItemExpr,
		Exported: exported,
		Range:    Range{start, end},
		Body:     p.trimSemi(Range{start, end}),
		Pos:      p.at(start).Pos,
	}
}

func (p *parser) importDecl() *Item {
	start := p.i
	p.i++
	imp := &Import{}
	if p.at(p.i).Is("type") && !p.at(p.i+1).Is(",") && !p.at(p.i+1).Is("from") {
		imp.TypeOnly = true
		p.i++
	}
	if t := p.at(p.i); t.Kind == String {
		imp.Spec, imp.SpecPos = t.Text, t.Pos
		p.i = p.stmtEnd(p.i)
		return &Item{Kind: ItemImport, Import: imp, Range: Range{start, p.i}, Pos: p.at(start).Pos}
	}
	if t := p.at(p.i); t.Kind == Ident && !t.Is("from") {
		imp.Default = t.Text
		p.i++
		if p.at(p.i).Is(",") {
			p.i++
		}
	}
	if p.at(p.i).Is("*") && p.at(p.i+1).Is("as") {
		imp.Namespace = p.at(p.i + 2).Text
		p.i += 3
	}
	if p.at(p.i).Is("{") {
		imp.Named = p.namedList()
	}
	if p.at(p.i).Is("from") && p.at(p.i+1).Kind == String {
		imp.Spec, imp.SpecPos = p.at(p.i+1).Text, p.at(p.i+1).Pos
	} else {
		p.errorf(p.at(p.i), "import without a module specifier")
	}
	p.i = p.stmtEnd(p.i)
	return &Item{Kind: ItemImport, Import: imp, Range: Range{start, p.i}, Pos: p.at(start).Pos}
}

// namedList reads { a, type b, c as d } starting at the brace.
func (p *parser) namedList() []Named {
	end := p.skipGroup(p.i)
	var out []Named
	j := p.i + 1
	for j < end-1 {
		var n Named
		if p.at(j).Is("type") && (p.at(j+1).Kind == Ident || p.at(j+1).Kind == String) && !p.at(j+1).Is("as") {
			n.TypeOnly = true
			j++
		}
		if t := p.at(j); t.Kind == Ident || t.Kind == String {
			n.Name, n.Local = t.Text, t.Text
			j++
		}
		if p.at(j).Is("as") {
			n.Local = p.at(j + 1).Text
			j += 2
		}
		if n.Name != "" {
			out = append(out, n)
		}
		for j < end-1 && !p.at(j).Is(",") {
			j++
		}
		j++
	}
	p.i = end
	return out
}

func (p *parser) exportDecl() *Item {
	start := p.i
	p.i++
	t := p.at(p.i)
	switch {
	case t.Is("default"):
		p.i++
		if it := p.declaration(true, true); it != nil {
			it.Range.From = start
			it.Pos = p.at(start).Pos
			return it
		}
		it := p.expression(true)
		it.Default = true
		it.Range.From = start
		it.Pos = p.at(start).Pos
		return it
	case t.Is("*"):
		imp := &Import{ReExport: true}
		p.i++
		if p.at(p.i).Is("as") {
			imp.Namespace = p.at(p.i + 1).Text
			p.i += 2
		}
		if p.at(p.i).Is("from") {
			imp.Spec, imp.SpecPos = p.at(p.i+1).Text, p.at(p.i+1).Pos
		}
		p.i = p.stmtEnd(p.i)
		return &Item{Kind: ItemImport, Import: imp, Range: Range{start, p.i}, Pos: p.at(start).Pos}
	case t.Is("{") || t.Is("type") && p.at(p.i+1).Is("{"):
		typeOnly := t.Is("type")
		if typeOnly {
			p.i++
		}
		names := p.namedList()
		if p.at(p.i).Is("from") && p.at(p.i+1).Kind == String {
			imp := &Import{ReExport: true, TypeOnly: typeOnly, Named: names,
				Spec: p.at(p.i + 1).Text, SpecPos: p.at(p.i + 1).Pos}
			p.i = p.stmtEnd(p.i)
			return &Item{Kind: ItemImport, Import: imp, Range: Range{start, p.i}, Pos: p.at(start).Pos}
		}
		p.i = p.stmtEnd(p.i)
		return &Item{Kind: ItemExportList, Exports: names, Range: Range{start, p.i}, Pos: p.at(start).Pos}
	case t.Is("=") || t.Is("as") || t.Is("import"):
		return p.other()
	}
	if it := p.declaration(true, false); it != nil {
		it.Range.From = start
		it.Pos = p.at(start).Pos
		return it
	}
	return p.other()
}

// declaration reads a declaration at p.i, or returns nil without consuming
// anything when there is none.
func (p *parser) declaration(exported, def bool) *Item {
	start := p.i
	declare := false
	if p.at(p.i).Is("declare") && !p.at(p.i+1).NL && p.at(p.i+1).Kind == Ident {
		declare = true
		p.i++
	}
	if p.at(p.i).Is("@") {
		p.skipDecorators()
	}
	t := p.at(p.i)
	next := p.at(p.i + 1)

	var it *Item
	switch {
	case t.Is("class"):
		it = p.class()
	case t.Is("abstract") && next.Is("class") && !next.NL:
		p.i++
		it = p.class()
		it.Abstract = true
	case t.Is("interface") && next.Kind == Ident && !next.NL:
		it = p.interfaceDecl()
	case t.Is("enum") && next.Kind == Ident:
		it = p.enumDecl()
	case t.Is("const") && next.Is("enum"):
		p.i++
		it = p.enumDecl()
	case t.Is("type") && next.Kind == Ident && !next.NL && (p.at(p.i+2).Is("=") || p.at(p.i+2).Is("<")):
		it = p.typeAlias()
	case t.Is("function"):
		it = p.function()
	case t.Is("async") && next.Is("function") && !next.NL:
		p.i++
		it = p.function()
	case (t.Is("const") || t.Is("let") || t.Is("var")) && !def &&
		(next.Kind == Ident || next.Is("{") || next.Is("[")):
		it = p.variable()
	default:
		p.i = start
		return nil
	}
	it.Exported = exported
	it.Default = def
	it.Declare = declare
	it.Range.From = start
	it.Pos = p.at(start).Pos
	return it
}

func (p *parser) class() *Item {
	it := &Item{Kind: ItemClass}
	p.i++ // class
	if t := p.at(p.i); t.Kind == Ident && !t.Is("extends") && !t.Is("implements") {
		it.Name = t.Text
		p.i++
	}
	if p.at(p.i).Is("<") {
		p.i = p.skipAngles(p.i)
	}
	for {
		switch {
		case p.at(p.i).Is("extends"):
			p.i++
			it.Extends = p.heritage()
			continue
		case p.at(p.i).Is("implements"):
			p.i++
			it.Implements = p.heritage()
			continue
		}
		break
	}
	if !p.at(p.i).Is("{") {
		p.errorf(p.at(p.i), "expected the body of class "+it.Name)
		p.i = p.stmtEnd(p.i)
		return it
	}
	end := p.skipGroup(p.i)
	it.Body = Range{p.i + 1, end - 1}
	it.Members = p.members(p.i+1, end-1)
	p.i = end
	it.Range.To = end
	return it
}

// heritage reads a comma separated list of dotted names, skipping their type
// arguments and anything else up to the next clause or the body.
func (p *parser) heritage() []string {
	var out []string
	for {
		t := p.at(p.i)
		switch {
		case t.Kind == EOF || t.Is("{") || t.Is("implements") || t.Is("extends"):
			return out
		case t.Is(","):
			p.i++
		case t.Kind == Ident:
			name := t.Text
			p.i++
			for p.at(p.i).Is(".") && p.at(p.i+1).Kind == Ident {
				name += "." + p.at(p.i+1).Text
				p.i += 2
			}
			out = append(out, name)
		case t.Is("<"):
			p.i = p.skipAngles(p.i)
		case isOpener(t):
			p.i = p.skipGroup(p.i)
		default:
			p.i++
		}
	}
}

var memberModifiers = map[string]bool{
	"public": true, "private": true, "protected": true, "static": true,
	"readonly": true, "abstract": true, "async": true, "override": true,
	"declare": true, "accessor": true, "get": true, "set": true,
}

func nameLike(t Token) bool {
	return t.Kind == Ident || t.Kind == String || t.Kind == Number || t.Kind == PrivateName || t.Is("[")
}

func (p *parser) members(from, to int) []*Member {
	var out []*Member
	j := from
	for j < to {
		t := p.at(j)
		if t.Is(";") || t.Is(",") {
			j++
			continue
		}
		start := j
		for p.at(j).Is("@") {
			j++
			for p.at(j).Kind == Ident {
				j++
				if !p.at(j).Is(".") {
					break
				}
				j++
			}
			if p.at(j).Is("(") {
				j = p.skipGroup(j)
			}
		}
		m := &Member{Pos: p.at(j).Pos}
		kind := MemberMethod
		for {
			t := p.at(j)
			if t.Kind != Ident || !memberModifiers[t.Text] || !nameLike(p.at(j+1)) && !p.at(j+1).Is("*") && !p.at(j+1).Is("{") {
				break
			}
			if p.at(j + 1).NL {
				break
			}
			switch t.Text {
			case "public", "private", "protected":
				m.Access = t.Text
			case "static":
				m.Static = true
			case "readonly":
				m.Readonly = true
			case "abstract":
				m.Abstract = true
			case "async":
				m.Async = true
			case "get":
				kind = MemberGetter
			case "set":
				kind = MemberSetter
			}
			j++
			if t.Text == "static" && p.at(j).Is("{") {
				break
			}
		}
		if m.Static && p.at(j).Is("{") {
			end := p.skipGroup(j)
			m.Kind = MemberStatic
			m.Body = Range{j + 1, end - 1}
			m.Range = Range{start, end}
			out = append(out, m)
			j = end
			continue
		}
		if p.at(j).Is("*") {
			j++
		}

		index := false
		switch t := p.at(j); {
		case t.Is("["):
			index = p.at(j+1).Kind == Ident && p.at(j+2).Is(":")
			m.Name = p.f.Source(Range{j, p.skipGroup(j)})
			j = p.skipGroup(j)
		case nameLike(t):
			m.Name = t.Text
			j++
		default:
			p.errorf(t, "expected a class member")
			j = p.memberEnd(j+1, to)
			continue
		}
		if p.at(j).Is("?") || p.at(j).Is("!") {
			j++
		}
		if p.at(j).Is("<") {
			j = p.skipAngles(j)
		}

		if p.at(j).Is("(") && !index {
			m.Kind = kind
			if m.Name == "constructor" {
				m.Kind = MemberConstructor
			}
			m.Params, j = p.params(j)
			if p.at(j).Is(":") {
				j = p.skipType(j + 1)
			}
			if p.at(j).Is("{") {
				end := p.skipGroup(j)
				m.Body = Range{j + 1, end - 1}
				j = end
			} else if p.at(j).Is(";") {
				j++
			}
			m.Range = Range{start, j}
			out = append(out, m)
			continue
		}

		m.Kind = MemberProperty
		if index {
			m.Kind = MemberIndex
		}
		end := p.memberEnd(j, to)
		if p.at(j).Is(":") {
			j = p.skipType(j + 1)
		}
		if p.at(j).Is("=") && j < end {
			m.Init = p.trimSemi(Range{j + 1, end})
		}
		m.Range = Range{start, end}
		out = append(out, m)
		j = end
	}
	return out
}

// memberEnd is stmtEnd bounded by the class body.
func (p *parser) memberEnd(j, to int) int {
	end := p.stmtEnd(j)
	if end > to {
		return to
	}
	if end == j {
		return j + 1
	}
	return end
}

// params reads a parameter list starting at its parenthesis.
func (p *parser) params(i int) ([]Param, int) {
	end := p.skipGroup(i)
	var out []Param
	j := i + 1
	for j < end-1 {
		// One parameter runs to the next comma at this depth.
		segEnd := j
		for segEnd < end-1 && !p.at(segEnd).Is(",") {
			if isOpener(p.at(segEnd)) {
				segEnd = p.skipGroup(segEnd)
				continue
			}
			if p.at(segEnd).Is("<") {
				segEnd = p.skipAngles(segEnd)
				continue
			}
			segEnd++
		}
		var prm Param
		k := j
		for p.at(k).Is("@") {
			k++
			for p.at(k).Kind == Ident || p.at(k).Is(".") {
				k++
			}
			if p.at(k).Is("(") {
				k = p.skipGroup(k)
			}
		}
	modifiers:
		for k < segEnd {
			t := p.at(k)
			if t.Kind != Ident || !nameLike(p.at(k+1)) && !p.at(k+1).Is("{") && !p.at(k+1).Is("...") {
				break
			}
			switch t.Text {
			case "public", "private", "protected":
				prm.Access = t.Text
			case "readonly":
				prm.Readonly = true
			case "override":
			default:
				break modifiers
			}
			k++
		}
		if p.at(k).Is("...") {
			k++
		}
		prm.Pos = p.at(k).Pos
		if p.at(k).Kind == Ident {
			prm.Name = p.at(k).Text
			k++
		} else if isOpener(p.at(k)) {
			k = p.skipGroup(k)
		}
		if p.at(k).Is("?") {
			k++
		}
		if p.at(k).Is(":") {
			tEnd := p.skipType(k + 1)
			if tEnd > segEnd {
				tEnd = segEnd
			}
			prm.Type = Range{k + 1, tEnd}
		}
		out = append(out, prm)
		j = segEnd + 1
	}
	return out, end
}

var typePrefixes = map[string]bool{
	"keyof": true, "typeof": true, "readonly": true, "unique": true,
	"infer": true, "asserts": true, "new": true, "abstract": true,
}

// skipType returns the index just past the type starting at j.
//
// A type is read only so that it can be stepped over: where a return type ends
// and a method body begins is the one question that needs it, because both can
// open with a brace.
func (p *parser) skipType(j int) int {
	conditional := false
	for {
		for p.at(j).Is("|") || p.at(j).Is("&") {
			j++
		}
		for p.at(j).Kind == Ident && typePrefixes[p.at(j).Text] && !p.at(j+1).NL &&
			(nameLike(p.at(j+1)) || isOpener(p.at(j+1)) || p.at(j+1).Is("<")) {
			j++
		}
		t := p.at(j)
		switch {
		case t.Is("<"):
			j = p.skipAngles(j)
			continue
		case isOpener(t):
			j = p.skipGroup(j)
		case t.Kind == Ident || t.Kind == String || t.Kind == Number || t.Kind == Template:
			j++
			for p.at(j).Is(".") && p.at(j+1).Kind == Ident {
				j += 2
			}
		case t.Is("-") && p.at(j+1).Kind == Number:
			j += 2
		default:
			return j
		}
		for {
			switch {
			case p.at(j).Is("<") && !p.at(j).NL:
				j = p.skipAngles(j)
				continue
			case p.at(j).Is("[") && !p.at(j).NL:
				j = p.skipGroup(j)
				continue
			}
			break
		}
		switch t := p.at(j); {
		case t.Is("=>") || t.Is("|") || t.Is("&") || t.Is("is"):
			j++
			continue
		case t.Is("extends"):
			conditional = true
			j++
			continue
		case conditional && (t.Is("?") || t.Is(":")):
			j++
			continue
		}
		return j
	}
}

func (p *parser) interfaceDecl() *Item {
	it := &Item{Kind: ItemInterface, Name: p.at(p.i + 1).Text}
	p.i += 2
	if p.at(p.i).Is("<") {
		p.i = p.skipAngles(p.i)
	}
	if p.at(p.i).Is("extends") {
		p.i++
		it.Extends = p.heritage()
	}
	if p.at(p.i).Is("{") {
		end := p.skipGroup(p.i)
		it.Body = Range{p.i + 1, end - 1}
		p.i = end
	}
	it.Range.To = p.i
	return it
}

func (p *parser) enumDecl() *Item {
	it := &Item{Kind: ItemEnum, Name: p.at(p.i + 1).Text}
	p.i += 2
	if p.at(p.i).Is("{") {
		end := p.skipGroup(p.i)
		it.Body = Range{p.i + 1, end - 1}
		p.i = end
	}
	it.Range.To = p.i
	return it
}

func (p *parser) typeAlias() *Item {
	it := &Item{Kind: ItemType, Name: p.at(p.i + 1).Text}
	start := p.i
	p.i = p.stmtEnd(p.i)
	it.Body = p.trimSemi(Range{start, p.i})
	it.Range.To = p.i
	return it
}

func (p *parser) function() *Item {
	it := &Item{Kind: ItemFunction}
	p.i++ // function
	if p.at(p.i).Is("*") {
		p.i++
	}
	if p.at(p.i).Kind == Ident {
		it.Name = p.at(p.i).Text
		p.i++
	}
	if p.at(p.i).Is("<") {
		p.i = p.skipAngles(p.i)
	}
	if p.at(p.i).Is("(") {
		var params []Param
		params, p.i = p.params(p.i)
		it.Members = []*Member{{Kind: MemberMethod, Name: it.Name, Params: params}}
	}
	if p.at(p.i).Is(":") {
		p.i = p.skipType(p.i + 1)
	}
	if p.at(p.i).Is("{") {
		end := p.skipGroup(p.i)
		it.Body = Range{p.i + 1, end - 1}
		p.i = end
	} else {
		p.i = p.stmtEnd(p.i)
	}
	it.Range.To = p.i
	return it
}

func (p *parser) variable() *Item {
	it := &Item{Kind: ItemVar, VarKind: p.at(p.i).Text}
	start := p.i
	end := p.stmtEnd(p.i)
	j := p.i + 1
	for j < end {
		v := Var{Pos: p.at(j).Pos}
		if p.at(j).Kind == Ident {
			v.Name = p.at(j).Text
			j++
		} else if isOpener(p.at(j)) {
			j = p.skipGroup(j)
		}
		if p.at(j).Is("!") {
			j++
		}
		if p.at(j).Is(":") {
			j = p.skipType(j + 1)
		}
		if p.at(j).Is("=") {
			k := j + 1
			for k < end && !p.at(k).Is(",") && !p.at(k).Is(";") {
				if isOpener(p.at(k)) {
					k = p.skipGroup(k)
					continue
				}
				k++
			}
			v.Init = Range{j + 1, k}
			j = k
		}
		it.Vars = append(it.Vars, v)
		for j < end && !p.at(j).Is(",") {
			j++
		}
		j++
	}
	if len(it.Vars) > 0 {
		it.Name = it.Vars[0].Name
	}
	p.i = end
	it.Range = Range{start, end}
	return it
}
