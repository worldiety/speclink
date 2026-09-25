// Package syntax reads Rust source as far as speclink needs to, and no further.
//
// It is not a Rust parser in the sense a compiler has one. It reads the items
// of a file — modules, uses, types, functions, constants, macro calls — and
// skips everything inside a function body or an expression as a balanced group
// of tokens. What it cannot represent without guessing it keeps as an opaque
// item with a position, so that a later stage can say "this was not read"
// rather than treat it as absent.
//
// That is enough because speclink does not have to understand the code it
// reads. cargo build has already established that every path in it resolves;
// what is left is to find the declarations and read their arguments.
//
// The lexer, on the other hand, has to be complete. Every later decision
// counts brackets, and a raw string holding a brace, a nested block comment or
// a lifetime taken for an unterminated character literal would shift every
// bracket after it. So it knows every token form the language has, even the
// ones nothing here ever looks at.
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
	// Ident is an identifier or keyword. Raw identifiers carry Raw.
	Ident
	// Lifetime is 'a, including the quote.
	Lifetime
	// Literal is a string, character or number literal; see LitKind.
	Literal
	// Punct is punctuation. The only multi character forms are ::, ->, =>,
	// .., ..= and ...; everything else is one character, so that >> closes
	// two generic lists without the parser having to split it.
	Punct
	// DocOuter is /// or /** */, documenting the item that follows.
	DocOuter
	// DocInner is //! or /*! */, documenting the enclosing item.
	DocInner
)

func (k Kind) String() string {
	switch k {
	case EOF:
		return "end of file"
	case Ident:
		return "identifier"
	case Lifetime:
		return "lifetime"
	case Literal:
		return "literal"
	case Punct:
		return "punctuation"
	case DocOuter, DocInner:
		return "doc comment"
	}
	return "token"
}

// LitKind distinguishes literals.
type LitKind uint8

const (
	LitNone LitKind = iota
	LitStr          // "…", r#"…"#, b"…", br"…", c"…", cr"…"
	LitChar         // 'x', b'x'
	LitNum          // 1, 0x1f, 1.5e3, 7u8
)

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
	Lit  LitKind
	// Text is the source text. For a raw identifier it is the name without
	// r#; for a doc comment it is the comment's content without its markers.
	Text string
	Raw  bool
	Pos  Pos
	// End is the byte offset just past the token.
	End int
}

// Is reports whether t is the punctuation or keyword s.
func (t Token) Is(s string) bool {
	return (t.Kind == Punct || t.Kind == Ident && !t.Raw) && t.Text == s
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
	toks []Token
	errs []Error
}

// Lex splits src into tokens. Ordinary comments and whitespace are dropped,
// doc comments are kept. It never stops early: a problem is reported and the
// rest of the file is still tokenised, which keeps the positions of everything
// after it meaningful.
func Lex(src string) ([]Token, []Error) {
	l := &lexer{src: src, line: 1, col: 1}
	l.shebang()
	for {
		l.skipSpace()
		if l.off >= len(l.src) {
			break
		}
		l.token()
	}
	l.toks = append(l.toks, Token{Kind: EOF, Pos: l.pos(), End: l.off})
	return l.toks, l.errs
}

func (l *lexer) pos() Pos { return Pos{Offset: l.off, Line: l.line, Col: l.col} }

func (l *lexer) peek(n int) byte {
	if l.off+n < len(l.src) {
		return l.src[l.off+n]
	}
	return 0
}

// advance moves n bytes forward, keeping line and column.
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

func (l *lexer) errorf(p Pos, format string, args ...any) {
	l.errs = append(l.errs, Error{Pos: p, Msg: fmt.Sprintf(format, args...)})
}

// shebang skips a #! line at the very start, unless it is the start of an
// inner attribute, which a shebang never is.
func (l *lexer) shebang() {
	if !strings.HasPrefix(l.src, "#!") {
		return
	}
	rest := strings.TrimLeft(l.src[2:], " \t")
	if strings.HasPrefix(rest, "[") {
		return
	}
	i := strings.IndexByte(l.src, '\n')
	if i < 0 {
		i = len(l.src)
	}
	l.advance(i)
}

func (l *lexer) skipSpace() {
	for l.off < len(l.src) {
		c := l.src[l.off]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			l.advance(1)
		case c == '/' && l.peek(1) == '/':
			if l.lineDoc() {
				return
			}
			i := strings.IndexByte(l.src[l.off:], '\n')
			if i < 0 {
				i = len(l.src) - l.off
			}
			l.advance(i)
		case c == '/' && l.peek(1) == '*':
			if l.blockDoc() {
				return
			}
			l.blockComment()
		default:
			r, size := utf8.DecodeRuneInString(l.src[l.off:])
			if r != utf8.RuneError && unicode.IsSpace(r) {
				l.advance(size)
				continue
			}
			return
		}
	}
}

// lineDoc reports whether a // comment at the cursor is a doc comment, in
// which case the token loop reads it. //// is an ordinary comment.
func (l *lexer) lineDoc() bool {
	rest := l.src[l.off:]
	return strings.HasPrefix(rest, "//!") ||
		strings.HasPrefix(rest, "///") && !strings.HasPrefix(rest, "////")
}

// blockDoc is the same for /* */. /** and /*! are doc comments, /*** and the
// empty /**/ are not.
func (l *lexer) blockDoc() bool {
	rest := l.src[l.off:]
	if strings.HasPrefix(rest, "/*!") {
		return true
	}
	return strings.HasPrefix(rest, "/**") && !strings.HasPrefix(rest, "/***") && !strings.HasPrefix(rest, "/**/")
}

// blockComment skips a block comment. They nest, which is the whole reason this
// is a loop with a depth rather than a search for */.
func (l *lexer) blockComment() string {
	start := l.pos()
	l.advance(2)
	depth := 1
	from := l.off
	for l.off < len(l.src) {
		switch {
		case l.src[l.off] == '/' && l.peek(1) == '*':
			depth++
			l.advance(2)
		case l.src[l.off] == '*' && l.peek(1) == '/':
			depth--
			if depth == 0 {
				body := l.src[from:l.off]
				l.advance(2)
				return body
			}
			l.advance(2)
		default:
			l.advance(1)
		}
	}
	l.errorf(start, "unterminated block comment")
	return l.src[from:]
}

func (l *lexer) emit(t Token) { l.toks = append(l.toks, t) }

func (l *lexer) token() {
	start := l.pos()
	c := l.src[l.off]
	rest := l.src[l.off:]

	switch {
	case strings.HasPrefix(rest, "///") || strings.HasPrefix(rest, "//!"):
		kind := DocOuter
		if rest[2] == '!' {
			kind = DocInner
		}
		i := strings.IndexByte(rest, '\n')
		if i < 0 {
			i = len(rest)
		}
		text := strings.TrimSuffix(rest[3:i], "\r")
		l.advance(i)
		l.emit(Token{Kind: kind, Text: text, Pos: start, End: l.off})
		return
	case strings.HasPrefix(rest, "/**") || strings.HasPrefix(rest, "/*!"):
		kind := DocOuter
		if rest[2] == '!' {
			kind = DocInner
		}
		body := l.blockComment()
		if len(body) > 0 {
			body = body[1:] // the second * or the !
		}
		l.emit(Token{Kind: kind, Text: body, Pos: start, End: l.off})
		return
	}

	// Prefixed literals come before identifiers, because r, b, br, c and cr
	// are also the start of perfectly good names.
	if l.prefixedLiteral(start) {
		return
	}

	switch {
	case c == 'r' && l.peek(1) == '#' && isIdentStart(l.runeAt(2)):
		l.advance(2)
		name := l.identRun()
		l.emit(Token{Kind: Ident, Text: name, Raw: true, Pos: start, End: l.off})
	case isIdentStart(l.runeAt(0)):
		name := l.identRun()
		l.emit(Token{Kind: Ident, Text: name, Pos: start, End: l.off})
	case c >= '0' && c <= '9':
		l.number(start)
	case c == '"':
		l.quoted(start, '"', LitStr)
	case c == '\'':
		l.quote(start)
	default:
		for _, p := range []string{"..=", "...", "::", "->", "=>", ".."} {
			if strings.HasPrefix(rest, p) {
				l.advance(len(p))
				l.emit(Token{Kind: Punct, Text: p, Pos: start, End: l.off})
				return
			}
		}
		r, size := utf8.DecodeRuneInString(rest)
		if r == utf8.RuneError || r > unicode.MaxASCII {
			l.errorf(start, "unexpected character %q", r)
		}
		l.advance(size)
		l.emit(Token{Kind: Punct, Text: rest[:size], Pos: start, End: l.off})
	}
}

func (l *lexer) runeAt(n int) rune {
	if l.off+n >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.off+n:])
	return r
}

func isIdentStart(r rune) bool { return r == '_' || unicode.IsLetter(r) }
func isIdentPart(r rune) bool  { return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r) }

func (l *lexer) identRun() string {
	from := l.off
	for l.off < len(l.src) {
		r, size := utf8.DecodeRuneInString(l.src[l.off:])
		if !isIdentPart(r) {
			break
		}
		l.advance(size)
	}
	return l.src[from:l.off]
}

// prefixedLiteral reads b"…", b'…', br"…", r"…", r#"…"#, c"…" and cr"…".
func (l *lexer) prefixedLiteral(start Pos) bool {
	rest := l.src[l.off:]
	for _, p := range []string{"br", "cr", "r"} {
		if strings.HasPrefix(rest, p) {
			after := rest[len(p):]
			if strings.HasPrefix(after, "\"") || strings.HasPrefix(after, "#") && strings.HasPrefix(strings.TrimLeft(after, "#"), "\"") {
				l.advance(len(p))
				l.rawString(start)
				return true
			}
		}
	}
	switch {
	case strings.HasPrefix(rest, "b\"") || strings.HasPrefix(rest, "c\""):
		l.advance(1)
		l.quoted(start, '"', LitStr)
		return true
	case strings.HasPrefix(rest, "b'"):
		l.advance(1)
		l.quoted(start, '\'', LitChar)
		return true
	}
	return false
}

// rawString reads the part of a raw string after its prefix: hashes, a quote,
// anything, a quote and the same number of hashes. Nothing inside is an
// escape, which is what makes it the one place a stray brace is common.
func (l *lexer) rawString(start Pos) {
	hashes := 0
	for l.peek(0) == '#' {
		hashes++
		l.advance(1)
	}
	l.advance(1) // the opening quote
	closing := "\"" + strings.Repeat("#", hashes)
	i := strings.Index(l.src[l.off:], closing)
	if i < 0 {
		l.errorf(start, "unterminated raw string")
		l.advance(len(l.src) - l.off)
	} else {
		l.advance(i + len(closing))
	}
	l.emit(Token{Kind: Literal, Lit: LitStr, Text: l.src[start.Offset:l.off], Pos: start, End: l.off})
}

// quoted reads a literal delimited by q, honouring backslash escapes.
func (l *lexer) quoted(start Pos, q byte, kind LitKind) {
	l.advance(1)
	for {
		if l.off >= len(l.src) {
			l.errorf(start, "unterminated %s literal", map[LitKind]string{LitStr: "string", LitChar: "character"}[kind])
			break
		}
		c := l.src[l.off]
		if c == '\\' {
			l.advance(2)
			continue
		}
		l.advance(1)
		if c == q {
			break
		}
	}
	l.suffix()
	l.emit(Token{Kind: Literal, Lit: kind, Text: l.src[start.Offset:l.off], Pos: start, End: l.off})
}

// quote decides between a lifetime and a character literal, the one genuinely
// ambiguous case in the lexical grammar. 'a' is a character; 'a is a lifetime;
// 'ab' is neither and reads as a lifetime followed by a stray quote, which is
// also what the compiler says about it.
func (l *lexer) quote(start Pos) {
	r := l.runeAt(1)
	if isIdentStart(r) {
		_, size := utf8.DecodeRuneInString(l.src[l.off+1:])
		if l.peek(1+size) != '\'' {
			l.advance(1)
			name := l.identRun()
			l.emit(Token{Kind: Lifetime, Text: "'" + name, Pos: start, End: l.off})
			return
		}
	}
	l.quoted(start, '\'', LitChar)
}

// number reads an integer or float literal with its suffix. A dot is only
// part of it when a digit follows, so that 1..2 is a range and 1.max(2) a call.
func (l *lexer) number(start Pos) {
	hex := strings.HasPrefix(l.src[l.off:], "0x") || strings.HasPrefix(l.src[l.off:], "0X")
	l.digits(hex)
	if l.peek(0) == '.' && l.peek(1) >= '0' && l.peek(1) <= '9' {
		l.advance(1)
		l.digits(false)
	} else if l.peek(0) == '.' && l.peek(1) != '.' && !isIdentStart(l.runeAt(1)) {
		// 1. is a float with no fractional part.
		l.advance(1)
	}
	l.emit(Token{Kind: Literal, Lit: LitNum, Text: l.src[start.Offset:l.off], Pos: start, End: l.off})
}

func (l *lexer) digits(hex bool) {
	for l.off < len(l.src) {
		c := l.src[l.off]
		if !hex && (c == 'e' || c == 'E') && (l.peek(1) == '+' || l.peek(1) == '-') {
			l.advance(2)
			continue
		}
		if c == '_' || c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
			l.advance(1)
			continue
		}
		break
	}
}

// suffix reads a literal suffix such as the u8 of b'a'u8, which is legal
// syntax even where it is not a legal program.
func (l *lexer) suffix() {
	if isIdentStart(l.runeAt(0)) {
		l.identRun()
	}
}

// StringValue returns the value of a string literal token, with escapes
// applied. ok is false for anything that is not a string literal, and for a
// byte or C string, whose value is not text.
func StringValue(t Token) (string, bool) {
	if t.Kind != Literal || t.Lit != LitStr {
		return "", false
	}
	s := t.Text
	if strings.HasPrefix(s, "b") || strings.HasPrefix(s, "c") {
		return "", false
	}
	if strings.HasPrefix(s, "r") {
		s = s[1:]
		h := 0
		for h < len(s) && s[h] == '#' {
			h++
		}
		if len(s) < 2*h+2 {
			return "", false
		}
		return s[h+1 : len(s)-h-1], true
	}
	end := strings.LastIndexByte(s, '"')
	if end <= 0 {
		return "", false
	}
	return unescape(s[1:end])
}

func unescape(s string) (string, bool) {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' {
			b.WriteByte(c)
			continue
		}
		i++
		if i >= len(s) {
			return "", false
		}
		switch s[i] {
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case 'r':
			b.WriteByte('\r')
		case '0':
			b.WriteByte(0)
		case '\\', '"', '\'':
			b.WriteByte(s[i])
		case 'x':
			if i+3 > len(s) {
				return "", false
			}
			var v byte
			if _, err := fmt.Sscanf(s[i+1:i+3], "%02x", &v); err != nil {
				return "", false
			}
			b.WriteByte(v)
			i += 2
		case 'u':
			end := strings.IndexByte(s[i:], '}')
			if i+1 >= len(s) || s[i+1] != '{' || end < 0 {
				return "", false
			}
			var v rune
			if _, err := fmt.Sscanf(strings.ReplaceAll(s[i+2:i+end], "_", ""), "%x", &v); err != nil {
				return "", false
			}
			b.WriteRune(v)
			i += end
		case '\n':
			// A line continuation swallows the newline and the indentation
			// of the next line.
			for i+1 < len(s) && (s[i+1] == ' ' || s[i+1] == '\t' || s[i+1] == '\n' || s[i+1] == '\r') {
				i++
			}
		default:
			return "", false
		}
	}
	return b.String(), true
}
