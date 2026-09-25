package syntax

import "strings"

// File is one parsed source file.
type File struct {
	Path string
	Src  string
	Toks []Token
	// Attrs are the inner attributes, #![…], of the file.
	Attrs []Attr
	// Doc is the inner documentation, //!, of the file.
	Doc    string
	Items  []*Item
	Errors []Error
}

// Range is a half open range of token indices into File.Toks. The zero value
// is empty.
type Range struct{ From, To int }

// Empty reports whether the range holds no token.
func (r Range) Empty() bool { return r.To <= r.From }

// Tokens returns the tokens of r.
func (f *File) Tokens(r Range) []Token {
	if r.Empty() {
		return nil
	}
	return f.Toks[r.From:r.To]
}

// Text renders r normalised: tokens joined with a single space where the
// language needs one and none where it does not, comments dropped. It is what
// a type is compared and displayed as, so that `Vec < u8 >` and `Vec<u8>` are
// the same type.
func (f *File) Text(r Range) string { return Render(f.Tokens(r)) }

// Source returns the exact source bytes of r, for fingerprints: a change to a
// comment inside a body is a change to the body a reviewer approved.
func (f *File) Source(r Range) string {
	if r.Empty() {
		return ""
	}
	return f.Src[f.Toks[r.From].Pos.Offset:f.Toks[r.To-1].End]
}

// Pos returns the position of the first token of r.
func (f *File) Pos(r Range) Pos {
	if r.From < len(f.Toks) {
		return f.Toks[r.From].Pos
	}
	return Pos{}
}

// Render joins tokens the way Text does.
func Render(toks []Token) string {
	var b strings.Builder
	for i, t := range toks {
		if t.Kind == EOF || t.Kind == DocOuter || t.Kind == DocInner {
			continue
		}
		if i > 0 && b.Len() > 0 && space(toks[i-1], t) {
			b.WriteByte(' ')
		}
		b.WriteString(tokenText(t))
	}
	return b.String()
}

func tokenText(t Token) string {
	if t.Raw {
		return "r#" + t.Text
	}
	return t.Text
}

func word(t Token) bool { return t.Kind == Ident || t.Kind == Literal || t.Kind == Lifetime }

func space(prev, t Token) bool {
	switch {
	case word(prev) && word(t):
		return true
	case prev.Is(","), prev.Is(";"):
		return true
	case t.Is("->") || prev.Is("->"), t.Is("=>") || prev.Is("=>"):
		return true
	case t.Is("=") || prev.Is("="), t.Is("+") || prev.Is("+"):
		return true
	case prev.Is(":") && !t.Is(":"):
		return true
	case prev.Is("dyn"), prev.Is("impl"), prev.Is("mut") && !t.Is(")"):
		return word(t) || t.Is("(") || t.Is("&") || t.Is("*") || t.Is("[")
	}
	return false
}

// Normalize renders a fragment of Rust source the way Text renders a range,
// so that a condition written in a configuration file compares equal to the
// same condition written in the code, whatever the spacing.
func Normalize(src string) string {
	toks, _ := Lex(src)
	return Render(toks)
}

// Attr is one attribute, #[…] or #![…].
type Attr struct {
	// Path is the attribute's name as written: cfg, path, serde, derive,
	// test, tokio::test.
	Path string
	// Args are the tokens inside the delimiters after the path, or after the
	// = of a key-value attribute.
	Args []Token
	// Eq is set for the key-value form, #[path = "x"].
	Eq    bool
	Inner bool
	Pos   Pos
}

// Text renders the arguments normalised.
func (a Attr) Text() string { return Render(a.Args) }

// Value returns the string of a key-value attribute, #[path = "x"].
func (a Attr) Value() (string, bool) {
	if !a.Eq || len(a.Args) != 1 {
		return "", false
	}
	return StringValue(a.Args[0])
}

// AttrsNamed returns the attributes of the given path.
func AttrsNamed(attrs []Attr, path string) []Attr {
	var out []Attr
	for _, a := range attrs {
		if a.Path == path {
			out = append(out, a)
		}
	}
	return out
}

// ItemKind is the sort of an item.
type ItemKind uint8

const (
	// ItemOpaque is anything this does not read. It carries its position
	// and its tokens and nothing else, and a later stage decides whether
	// its presence matters.
	ItemOpaque ItemKind = iota
	ItemMod
	ItemUse
	ItemExternCrate
	ItemStruct
	ItemUnion
	ItemEnum
	ItemTrait
	ItemImpl
	ItemFn
	ItemConst
	ItemStatic
	ItemType
	// ItemMacroCall is a macro invoked in item position, name!(…).
	ItemMacroCall
	// ItemMacroRules is a macro_rules! definition.
	ItemMacroRules
)

var itemKindNames = [...]string{
	ItemOpaque: "unread item", ItemMod: "module", ItemUse: "use", ItemExternCrate: "extern crate",
	ItemStruct: "struct", ItemUnion: "union", ItemEnum: "enum", ItemTrait: "trait", ItemImpl: "impl",
	ItemFn: "fn", ItemConst: "const", ItemStatic: "static", ItemType: "type alias",
	ItemMacroCall: "macro call", ItemMacroRules: "macro_rules",
}

func (k ItemKind) String() string {
	if int(k) < len(itemKindNames) {
		return itemKindNames[k]
	}
	return "item"
}

// Item is one item of a module, trait or impl.
//
// It is one struct rather than one type per kind because nearly every
// consumer switches on the kind anyway, and most fields are ranges that are
// simply empty where they do not apply.
type Item struct {
	Kind ItemKind
	// Name is the declared name; empty for use, impl and an unnamed macro
	// call. _ for const _.
	Name string
	// Vis is the visibility as written, normalised: "", "pub", "pub(crate)",
	// "pub(super)", "pub(in a::b)".
	Vis   string
	Attrs []Attr
	Doc   string
	// Pos is where the item proper starts, after its attributes; Start is
	// the first token of its attributes and documentation.
	Pos   Pos
	Start int
	// All covers the whole item including its attributes.
	All Range

	// Inner holds the inner attributes of an inline module, a trait or an
	// impl.
	Inner []Attr
	// Items holds the items of an inline module, a trait or an impl.
	Items []*Item
	// External is set for mod name;, whose items are in another file.
	External bool

	Uses []Use

	Generics Range
	Where    Range
	Fields   []Field
	// Tuple is set for a tuple struct.
	Tuple    bool
	Variants []Variant

	Async, Const, Unsafe bool
	// Params is inside the parentheses of a function, Ret after its arrow.
	Params Range
	Ret    Range
	// Body includes its braces. Empty for a function without one.
	Body Range

	// Type is the declared type of a const or static, or the aliased type.
	Type Range
	// Value is the expression of a const or static.
	Value Range
	Mut   bool

	// Trait is the implemented trait of an impl, empty for an inherent one.
	Trait    Range
	SelfType Range
	Negative bool

	// Macro is the invoked path, satisfies or speclink::satisfies.
	Macro string
	// Delim is the opening delimiter of the macro's arguments.
	Delim string
	// Args are inside the delimiters.
	Args Range
}

// Use is one imported path of a use declaration, flattened: use a::{b, c as
// d} is two.
type Use struct {
	Path []string
	// Alias is the name it is imported as. For a glob it is empty.
	Alias string
	Glob  bool
	Pos   Pos
}

// Name is the name the use binds in scope.
func (u Use) Name() string {
	if u.Alias != "" {
		return u.Alias
	}
	if len(u.Path) == 0 {
		return ""
	}
	return u.Path[len(u.Path)-1]
}

// Field is one field of a struct, union or variant. Tuple fields are named
// by their index.
type Field struct {
	Name  string
	Vis   string
	Attrs []Attr
	Doc   string
	Type  Range
	Pos   Pos
}

// Variant is one variant of an enum.
type Variant struct {
	Name   string
	Attrs  []Attr
	Doc    string
	Fields []Field
	Tuple  bool
	// Unit is set for a variant with no fields at all.
	Unit         bool
	Discriminant Range
	Pos          Pos
}
