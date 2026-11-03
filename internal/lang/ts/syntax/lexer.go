// Package syntax reads TypeScript as far as speclink needs to, and no further.
//
// It is not a TypeScript parser in the sense a compiler has one. It reads the
// statements at the top of a module — imports, exports, classes, interfaces,
// enums, type aliases, functions, variables — and inside a class the head of
// every member. Everything else, a function body, an initializer, a type, is
// kept as a balanced range of tokens that a rule may scan but nothing here
// interprets.
//
// That is enough because speclink does not have to understand the code it
// reads. vue-tsc has already established that every name in it resolves; what
// is left is to find the declarations, their imports and the few calls the
// rules ask about.
//
// The lexer, on the other hand, has to be complete. Every later decision counts
// brackets, and a template literal holding a brace, a regular expression
// holding a quote or a comment holding a parenthesis would shift every bracket
// after it. So it knows every token form the language has, even the ones
// nothing here ever looks at.
package syntax

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Kind is the class of a token.
type Kind uint8

const (
	EOF Kind = iota
	// Ident is an identifier or keyword.
	Ident
	// PrivateName is #name, a private class member.
	PrivateName
	// String is a single or double quoted literal. Text is its value.
	String
	// Template is a template literal without substitutions. Text is its raw
	// content.
	Template
	// TemplateHead opens a template literal with substitutions: `…${
	TemplateHead
	// TemplateMiddle continues one: }…${
	TemplateMiddle
	// TemplateTail closes one: }…`
	TemplateTail
	// Number is a numeric literal.
	Number
	// Regex is a regular expression literal.
	Regex
	// Punct is punctuation. The multi character forms are the ones whose
	// splitting would change what a rule sees — =>, ..., ?., and the
	// comparison and logical operators. < and > are always single, so that a
	// generic list closes without the parser having to split >>.
	Punct
)

func (k Kind) String() string {
	switch k {
	case EOF:
		return "end of file"
	case Ident:
		return "identifier"
	case PrivateName:
		return "private name"
	case String:
		return "string"
	case Template, TemplateHead, TemplateMiddle, TemplateTail:
		return "template literal"
	case Number:
		return "number"
	case Regex:
		return "regular expression"
	case Punct:
		return "punctuation"
	}
	return "token"
}

// Pos is a source position. Line and Col are 1 based; Col counts bytes.
type Pos struct {
	Offset int
	Line   int
	Col    int
}

func (p Pos) String() string { return fmt.Sprintf("%d:%d", p.Line, p.Col) }

// Token is one lexical token.
type Token struct {
	Kind Kind
	Text string
	Pos  Pos
	// End is the byte offset just past the token.
	End int
	// NL records that a line break precedes the token, which is what
	// automatic semicolon insertion is decided on.
	NL bool
}

// Is reports whether t is the punctuation or identifier s.
func (t Token) Is(s string) bool {
	return (t.Kind == Punct || t.Kind == Ident) && t.Text == s
}

// Error is a lexical or syntactic problem, positioned.
type Error struct {
	Pos Pos
	Msg string
}

func (e Error) Error() string { return e.Pos.String() + ": " + e.Msg }

type lexer struct {
	src  string
	off  int
	line int
	col  int
	nl   bool
	toks []Token
	errs []Error
	// braces holds, for every template substitution still open, the brace
	// depth at which it was opened, so that the } closing it resumes the
	// template rather than ending a block.
	braces []int
	depth  int
}

// Lex splits src into tokens. start is the position of src within the file it
// was taken from, which is how the script block of a single file component
// keeps the line numbers of the file rather than of the block.
func Lex(src string, start Pos) ([]Token, []Error) {
	if start.Line == 0 {
		start = Pos{Line: 1, Col: 1}
	}
	l := &lexer{src: src, line: start.Line, col: start.Col}
	base := start.Offset
	l.run()
	if base != 0 {
		for i := range l.toks {
			l.toks[i].Pos.Offset += base
			l.toks[i].End += base
		}
		for i := range l.errs {
			l.errs[i].Pos.Offset += base
		}
	}
	return l.toks, l.errs
}

func (l *lexer) pos() Pos { return Pos{Offset: l.off, Line: l.line, Col: l.col} }

func (l *lexer) peek(n int) byte {
	if l.off+n < len(l.src) {
		return l.src[l.off+n]
	}
	return 0
}

func (l *lexer) advance(n int) {
	for i := 0; i < n && l.off < len(l.src); i++ {
		if l.src[l.off] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.off++
	}
}

func (l *lexer) emit(k Kind, text string, p Pos) {
	l.toks = append(l.toks, Token{Kind: k, Text: text, Pos: p, End: l.off, NL: l.nl})
	l.nl = false
}

func (l *lexer) errorf(p Pos, format string, args ...any) {
	l.errs = append(l.errs, Error{Pos: p, Msg: fmt.Sprintf(format, args...)})
}

func (l *lexer) run() {
	for {
		l.skipSpace()
		if l.off >= len(l.src) {
			p := l.pos()
			l.toks = append(l.toks, Token{Kind: EOF, Pos: p, End: l.off, NL: l.nl})
			return
		}
		l.next()
	}
}

func (l *lexer) skipSpace() {
	for l.off < len(l.src) {
		c := l.src[l.off]
		switch {
		case c == '\n':
			l.nl = true
			l.advance(1)
		case c == ' ' || c == '\t' || c == '\r' || c == '\f' || c == '\v':
			l.advance(1)
		case c == '/' && l.peek(1) == '/':
			for l.off < len(l.src) && l.src[l.off] != '\n' {
				l.advance(1)
			}
		case c == '/' && l.peek(1) == '*':
			p := l.pos()
			end := strings.Index(l.src[l.off+2:], "*/")
			if end < 0 {
				l.errorf(p, "unterminated block comment")
				l.advance(len(l.src) - l.off)
				return
			}
			if strings.Contains(l.src[l.off:l.off+2+end], "\n") {
				l.nl = true
			}
			l.advance(end + 4)
		case c >= 0x80:
			r, size := utf8.DecodeRuneInString(l.src[l.off:])
			if r == '\u2028' || r == '\u2029' {
				l.nl = true
			}
			if !unicode.IsSpace(r) && r != '\uFEFF' {
				return
			}
			l.off += size
			l.col += size
		default:
			return
		}
	}
}

func (l *lexer) next() {
	p := l.pos()
	c := l.src[l.off]

	switch {
	case c == '"' || c == '\'':
		l.str(p, c)
	case c == '`':
		l.advance(1)
		l.template(p, true)
	case c == '#' && l.off+1 < len(l.src) && isIdentStart(l.runeAt(l.off+1)):
		l.advance(1)
		name := l.ident()
		l.emit(PrivateName, "#"+name, p)
	case isDigit(c) || c == '.' && isDigit(l.peek(1)):
		l.number(p)
	case isIdentStart(l.runeAt(l.off)) || c == '\\':
		l.emit(Ident, l.ident(), p)
	case c == '/' && l.regexAllowed():
		l.regex(p)
	case c == '}' && len(l.braces) > 0 && l.braces[len(l.braces)-1] == l.depth:
		// The } closing a template substitution resumes the template.
		l.braces = l.braces[:len(l.braces)-1]
		l.advance(1)
		l.template(p, false)
	default:
		l.punct(p)
	}
}

func (l *lexer) runeAt(off int) rune {
	r, _ := utf8.DecodeRuneInString(l.src[off:])
	return r
}

func (l *lexer) ident() string {
	var b strings.Builder
	for l.off < len(l.src) {
		if l.src[l.off] == '\\' && l.peek(1) == 'u' {
			// A unicode escape in an identifier. Rare enough to keep
			// verbatim; it only has to be one token.
			start := l.off
			l.advance(2)
			if l.peek(0) == '{' {
				for l.off < len(l.src) && l.src[l.off] != '}' {
					l.advance(1)
				}
				l.advance(1)
			} else {
				l.advance(4)
			}
			b.WriteString(l.src[start:l.off])
			continue
		}
		r, size := utf8.DecodeRuneInString(l.src[l.off:])
		if !isIdentPart(r) {
			break
		}
		b.WriteString(l.src[l.off : l.off+size])
		l.off += size
		l.col += size
	}
	return b.String()
}

func (l *lexer) str(p Pos, quote byte) {
	l.advance(1)
	var b strings.Builder
	for {
		if l.off >= len(l.src) || l.src[l.off] == '\n' {
			l.errorf(p, "unterminated string")
			break
		}
		c := l.src[l.off]
		if c == quote {
			l.advance(1)
			break
		}
		if c == '\\' {
			l.escape(&b)
			continue
		}
		b.WriteByte(c)
		l.advance(1)
	}
	l.emit(String, b.String(), p)
}

// escape reads one escape sequence and writes what it stands for.
func (l *lexer) escape(b *strings.Builder) {
	l.advance(1)
	if l.off >= len(l.src) {
		return
	}
	c := l.src[l.off]
	switch c {
	case 'n':
		b.WriteByte('\n')
	case 't':
		b.WriteByte('\t')
	case 'r':
		b.WriteByte('\r')
	case 'b':
		b.WriteByte('\b')
	case 'f':
		b.WriteByte('\f')
	case 'v':
		b.WriteByte('\v')
	case '0':
		b.WriteByte(0)
	case '\r':
		l.advance(1)
		if l.peek(0) == '\n' {
			l.advance(1)
		}
		return
	case '\n':
		// A line continuation stands for nothing.
	case 'u':
		l.advance(1)
		var hex string
		if l.peek(0) == '{' {
			end := strings.IndexByte(l.src[l.off:], '}')
			if end < 0 {
				return
			}
			hex = l.src[l.off+1 : l.off+end]
			l.advance(end + 1)
		} else {
			if l.off+4 > len(l.src) {
				return
			}
			hex = l.src[l.off : l.off+4]
			l.advance(4)
		}
		var r rune
		if _, err := fmt.Sscanf(hex, "%x", &r); err == nil {
			b.WriteRune(r)
		}
		return
	case 'x':
		l.advance(1)
		if l.off+2 > len(l.src) {
			return
		}
		var r rune
		if _, err := fmt.Sscanf(l.src[l.off:l.off+2], "%x", &r); err == nil {
			b.WriteRune(r)
		}
		l.advance(2)
		return
	default:
		r, size := utf8.DecodeRuneInString(l.src[l.off:])
		b.WriteRune(r)
		l.advance(size)
		return
	}
	l.advance(1)
}

// template reads template text up to its end or its next substitution. head
// says whether this is the opening part, which decides the token kind.
func (l *lexer) template(p Pos, head bool) {
	var b strings.Builder
	for {
		if l.off >= len(l.src) {
			l.errorf(p, "unterminated template literal")
			l.emit(Template, b.String(), p)
			return
		}
		c := l.src[l.off]
		switch {
		case c == '`':
			l.advance(1)
			if head {
				l.emit(Template, b.String(), p)
			} else {
				l.emit(TemplateTail, b.String(), p)
			}
			return
		case c == '$' && l.peek(1) == '{':
			l.advance(2)
			if head {
				l.emit(TemplateHead, b.String(), p)
			} else {
				l.emit(TemplateMiddle, b.String(), p)
			}
			l.braces = append(l.braces, l.depth)
			return
		case c == '\\':
			l.escape(&b)
		default:
			b.WriteByte(c)
			l.advance(1)
		}
	}
}

func (l *lexer) number(p Pos) {
	start := l.off
	if l.src[l.off] == '0' && strings.ContainsRune("xXoObB", rune(l.peek(1))) {
		l.advance(2)
		for l.off < len(l.src) && (isHex(l.src[l.off]) || l.src[l.off] == '_') {
			l.advance(1)
		}
	} else {
		for l.off < len(l.src) && (isDigit(l.src[l.off]) || l.src[l.off] == '_') {
			l.advance(1)
		}
		if l.peek(0) == '.' && l.peek(1) != '.' {
			l.advance(1)
			for l.off < len(l.src) && (isDigit(l.src[l.off]) || l.src[l.off] == '_') {
				l.advance(1)
			}
		}
		if c := l.peek(0); c == 'e' || c == 'E' {
			if n := l.peek(1); isDigit(n) || (n == '+' || n == '-') && isDigit(l.peek(2)) {
				l.advance(2)
				for l.off < len(l.src) && isDigit(l.src[l.off]) {
					l.advance(1)
				}
			}
		}
	}
	if l.peek(0) == 'n' {
		l.advance(1)
	}
	l.emit(Number, l.src[start:l.off], p)
}

// regexAllowed decides whether a slash begins a regular expression or is a
// division. It is the one place the lexer needs to know about the grammar, and
// the classic answer is enough: after something that ends an operand it is a
// division, everywhere else a literal.
func (l *lexer) regexAllowed() bool {
	if l.peek(1) == '=' && len(l.toks) > 0 && endsOperand(l.toks[len(l.toks)-1]) {
		return false
	}
	if len(l.toks) == 0 {
		return true
	}
	return !endsOperand(l.toks[len(l.toks)-1])
}

// endsOperand reports whether a token can be the last of an operand, after
// which a slash divides.
func endsOperand(t Token) bool {
	switch t.Kind {
	case Number, String, Template, TemplateTail, Regex, PrivateName:
		return true
	case Ident:
		switch t.Text {
		case "return", "typeof", "case", "do", "else", "in", "instanceof", "new",
			"delete", "void", "throw", "yield", "await", "of":
			return false
		}
		return true
	case Punct:
		return t.Text == ")" || t.Text == "]" || t.Text == "}"
	}
	return false
}

func (l *lexer) regex(p Pos) {
	start := l.off
	l.advance(1)
	inClass := false
	for {
		if l.off >= len(l.src) || l.src[l.off] == '\n' {
			l.errorf(p, "unterminated regular expression")
			break
		}
		c := l.src[l.off]
		if c == '\\' {
			l.advance(2)
			continue
		}
		if c == '[' {
			inClass = true
		} else if c == ']' {
			inClass = false
		} else if c == '/' && !inClass {
			l.advance(1)
			break
		}
		l.advance(1)
	}
	for l.off < len(l.src) && isIdentPart(l.runeAt(l.off)) {
		l.advance(1)
	}
	l.emit(Regex, l.src[start:l.off], p)
}

// puncts are the multi character forms kept together, longest first.
var puncts = []string{
	"...", "===", "!==", "**=", "&&=", "||=", "??=",
	"=>", "==", "!=", "&&", "||", "??", "?.", "++", "--",
	"+=", "-=", "*=", "/=", "%=", "&=", "|=", "^=", "**",
}

func (l *lexer) punct(p Pos) {
	rest := l.src[l.off:]
	for _, s := range puncts {
		if !strings.HasPrefix(rest, s) {
			continue
		}
		if s == "?." && len(rest) > 2 && isDigit(rest[2]) {
			// a ? .5 : b is a conditional, not an optional chain.
			continue
		}
		l.advance(len(s))
		l.emit(Punct, s, p)
		return
	}
	c := l.src[l.off]
	switch c {
	case '{':
		l.depth++
	case '}':
		l.depth--
	}
	if c >= 0x80 {
		r, size := utf8.DecodeRuneInString(rest)
		l.errorf(p, "unexpected character %q", r)
		l.off += size
		l.col += size
		return
	}
	l.advance(1)
	l.emit(Punct, string(c), p)
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func isHex(c byte) bool {
	return isDigit(c) || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func isIdentStart(r rune) bool {
	return r == '_' || r == '$' || unicode.IsLetter(r)
}

func isIdentPart(r rune) bool {
	return isIdentStart(r) || unicode.IsDigit(r) || r == '\u200c' || r == '\u200d' ||
		unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Pc, r)
}
