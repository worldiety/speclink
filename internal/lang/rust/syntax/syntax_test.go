package syntax

import (
	"reflect"
	"strings"
	"testing"
)

func kinds(toks []Token) []string {
	var out []string
	for _, t := range toks {
		if t.Kind == EOF {
			break
		}
		out = append(out, t.Kind.String()+":"+t.Text)
	}
	return out
}

func TestLex(t *testing.T) {
	tests := []struct {
		name string
		src  string
		want []string
	}{
		{"lifetime is not a char", "&'a str", []string{"punctuation:&", "lifetime:'a", "identifier:str"}},
		{"char literal", "'a' '\\'' '\\u{1F600}'", []string{"literal:'a'", `literal:'\''`, `literal:'\u{1F600}'`}},
		{"static lifetime", "'static", []string{"lifetime:'static"}},
		{"raw string holds braces", `r#"{ "} "#`, []string{`literal:r#"{ "} "#`}},
		{"raw string without hashes", `r"a\b"`, []string{`literal:r"a\b"`}},
		{"byte forms", `b"x" b'y' br"z"`, []string{`literal:b"x"`, "literal:b'y'", `literal:br"z"`}},
		{"raw identifier", "r#type", []string{"identifier:type"}},
		{"nested block comment", "a /* x /* y */ z */ b", []string{"identifier:a", "identifier:b"}},
		{"line comment", "a // b\nc", []string{"identifier:a", "identifier:c"}},
		{"four slashes are no doc", "//// no\na", []string{"identifier:a"}},
		{"outer doc", "/// hi\nfn", []string{"doc comment: hi", "identifier:fn"}},
		{"inner doc", "//! top", []string{"doc comment: top"}},
		{"block doc", "/** b */", []string{"doc comment: b "}},
		{"range is not a float", "1..2", []string{"literal:1", "punctuation:..", "literal:2"}},
		{"method on int", "1.max(2)", []string{"literal:1", "punctuation:.", "identifier:max", "punctuation:(", "literal:2", "punctuation:)"}},
		{"floats", "1.5e-3 2.0f32 0x1F_u8", []string{"literal:1.5e-3", "literal:2.0f32", "literal:0x1F_u8"}},
		{"multi punct", "a::b -> => ..= >>", []string{"identifier:a", "punctuation:::", "identifier:b", "punctuation:->", "punctuation:=>", "punctuation:..=", "punctuation:>", "punctuation:>"}},
		{"string with escaped quote", `"a\"b" c`, []string{`literal:"a\"b"`, "identifier:c"}},
		{"unicode ident", "größe", []string{"identifier:größe"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toks, errs := Lex(tt.src)
			if len(errs) > 0 {
				t.Fatalf("errors: %v", errs)
			}
			if got := kinds(toks); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got  %q\nwant %q", got, tt.want)
			}
		})
	}
}

func TestLexPositions(t *testing.T) {
	toks, _ := Lex("a\n  /* x\n */ bc")
	if p := toks[1].Pos; p.Line != 3 || p.Col != 5 || p.Offset != 13 {
		t.Errorf("bc at %+v", p)
	}
}

func TestLexReportsUnterminated(t *testing.T) {
	for _, src := range []string{`"abc`, "/* abc", `r#"abc"`} {
		if _, errs := Lex(src); len(errs) == 0 {
			t.Errorf("%q: no error", src)
		}
	}
}

func TestShebangIsSkippedButInnerAttributeIsNot(t *testing.T) {
	toks, _ := Lex("#!/usr/bin/env run\nfn")
	if toks[0].Text != "fn" {
		t.Errorf("shebang not skipped: %q", kinds(toks))
	}
	toks, _ = Lex("#![allow(x)]")
	if toks[0].Text != "#" {
		t.Errorf("inner attribute taken for a shebang: %q", kinds(toks))
	}
}

func TestStringValue(t *testing.T) {
	tests := map[string]string{
		`"a\nb"`:           "a\nb",
		`"\x41\u{42}"`:     "AB",
		`r#"a"b"#`:         `a"b`,
		`"a\` + "\n   b\"": "ab",
	}
	for src, want := range tests {
		toks, _ := Lex(src)
		got, ok := StringValue(toks[0])
		if !ok || got != want {
			t.Errorf("%s: got %q, %v", src, got, ok)
		}
	}
	toks, _ := Lex(`b"x"`)
	if _, ok := StringValue(toks[0]); ok {
		t.Error("a byte string has no text value")
	}
}

func parse(t *testing.T, src string) *File {
	t.Helper()
	f := Parse("x.rs", src)
	if len(f.Errors) > 0 {
		t.Fatalf("errors: %v", f.Errors)
	}
	return f
}

func TestParseItems(t *testing.T) {
	f := parse(t, `
//! The crate.
#![allow(dead_code)]

use std::collections::{HashMap, hash_map::{self, Entry as E}};
use super::*;
pub(crate) mod inner { pub fn f() {} }
mod outer;
extern crate alloc as a;

/// A quote.
#[derive(Debug, Clone)]
pub struct Quote<T: Clone = u8> {
    /// The number.
    pub number: String,
    items: HashMap<String, Vec<(T, u32)>>,
    #[serde(rename = "x")] pub(super) f: fn(u8, u8) -> u8,
}

pub struct Id(pub u64, String);
struct Unit;

pub enum State { Open, Closed { at: u64, by: String }, Moved(u8, u8) = 3, }

pub trait Repo: Send + Sync where Self: Sized {
    fn save(&self, q: &Quote) -> Result<(), Error>;
    type Item;
    const N: usize = 1;
}

impl<T> Repo for Store<T> where T: Clone {
    fn save(&self, q: &Quote) -> Result<(), Error> { let x = { 1 }; Ok(()) }
    type Item = u8;
}

impl Quote<u8> { pub async fn submit(self) -> Self { self } }
unsafe impl Send for Quote<u8> {}
impl<T> From<T> for Id where T: Into<u64> { fn from(t: T) -> Self { Id(t.into(), String::new()) } }

pub const R_X: Requirement = Requirement { id: "R-X", ..Requirement::EMPTY };
const _: () = { let _ = 1 < 2; };
static mut COUNT: u32 = 0;
pub type Res<T> = Result<T, Error>;
pub const unsafe fn g<'a>(x: &'a u8) -> &'a u8 where u8: Copy { x }

satisfies!(Quote => R_X);
speclink::rationale! { Quote => "because" }
macro_rules! m { ($x:expr) => { $x }; }
extern "C" { fn abs(x: i32) -> i32; }
`)

	if f.Doc != "The crate." || len(f.Attrs) != 1 || f.Attrs[0].Path != "allow" {
		t.Errorf("file doc/attrs: %q %+v", f.Doc, f.Attrs)
	}

	var got []string
	for _, it := range f.Items {
		got = append(got, it.Kind.String()+" "+it.Name)
	}
	want := []string{
		"use ", "use ", "module inner", "module outer", "extern crate alloc",
		"struct Quote", "struct Id", "struct Unit", "enum State", "trait Repo",
		"impl ", "impl ", "impl ", "impl ",
		"const R_X", "const _", "static COUNT", "type alias Res", "fn g",
		"macro call ", "macro call ", "macro_rules m", "unread item ",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("items:\n got  %q\n want %q", got, want)
	}
	it := func(i int) *Item { return f.Items[i] }

	var uses []string
	for _, u := range append(it(0).Uses, it(1).Uses...) {
		s := strings.Join(u.Path, "::")
		if u.Glob {
			s += "::*"
		}
		if u.Alias != "" {
			s += " as " + u.Alias
		}
		uses = append(uses, s)
	}
	wantUses := []string{"std::collections::HashMap", "std::collections::hash_map", "std::collections::hash_map::Entry as E", "super::*"}
	if !reflect.DeepEqual(uses, wantUses) {
		t.Errorf("uses: %q", uses)
	}

	if m := it(2); m.Vis != "pub(crate)" || m.External || len(m.Items) != 1 || m.Items[0].Name != "f" {
		t.Errorf("inline mod: %+v", m)
	}
	if !it(3).External {
		t.Error("mod outer; is external")
	}
	if u := it(4).Uses[0]; u.Alias != "a" {
		t.Errorf("extern crate alias: %+v", u)
	}

	q := it(5)
	if q.Doc != "A quote." || len(q.Attrs) != 1 || q.Attrs[0].Text() != "Debug, Clone" || f.Text(q.Generics) != "T: Clone = u8" {
		t.Errorf("struct header: doc %q attrs %+v generics %q", q.Doc, q.Attrs, f.Text(q.Generics))
	}
	var fields []string
	for _, fl := range q.Fields {
		fields = append(fields, fl.Vis+"|"+fl.Name+"|"+f.Text(fl.Type)+"|"+fl.Doc)
	}
	wantFields := []string{"pub|number|String|The number.", "|items|HashMap<String, Vec<(T, u32)>>|", "pub(super)|f|fn(u8, u8) -> u8|"}
	if !reflect.DeepEqual(fields, wantFields) {
		t.Errorf("fields:\n got  %q\n want %q", fields, wantFields)
	}
	if a := q.Fields[2].Attrs[0]; a.Path != "serde" || a.Text() != `rename = "x"` {
		t.Errorf("field attr: %+v", a)
	}

	if id := it(6); !id.Tuple || len(id.Fields) != 2 || id.Fields[1].Name != "1" || f.Text(id.Fields[1].Type) != "String" {
		t.Errorf("tuple struct: %+v", id.Fields)
	}

	st := it(8)
	if len(st.Variants) != 3 || !st.Variants[0].Unit || len(st.Variants[1].Fields) != 2 || !st.Variants[2].Tuple || f.Text(st.Variants[2].Discriminant) != "3" {
		t.Errorf("enum: %+v", st.Variants)
	}

	repo := it(9)
	if len(repo.Items) != 3 || repo.Items[0].Kind != ItemFn || !repo.Items[0].Body.Empty() || f.Text(repo.Items[0].Ret) != "Result<(), Error>" {
		t.Errorf("trait items: %+v", repo.Items)
	}

	impl := it(10)
	if f.Text(impl.Trait) != "Repo" || f.Text(impl.SelfType) != "Store<T>" || f.Text(impl.Where) != "T: Clone" || len(impl.Items) != 2 || impl.Items[0].Body.Empty() {
		t.Errorf("trait impl: trait %q self %q where %q", f.Text(impl.Trait), f.Text(impl.SelfType), f.Text(impl.Where))
	}
	if inh := it(11); !inh.Trait.Empty() || f.Text(inh.SelfType) != "Quote<u8>" || !inh.Items[0].Async {
		t.Errorf("inherent impl: %q", f.Text(inh.SelfType))
	}
	if u := it(12); !u.Unsafe || f.Text(u.Trait) != "Send" {
		t.Errorf("unsafe impl: %+v", u)
	}
	if from := it(13); f.Text(from.Trait) != "From<T>" || f.Text(from.SelfType) != "Id" {
		t.Errorf("generic trait impl: %q for %q", f.Text(from.Trait), f.Text(from.SelfType))
	}

	if c := it(14); f.Text(c.Type) != "Requirement" || !strings.Contains(f.Text(c.Value), `id: "R-X"`) {
		t.Errorf("const: %q = %q", f.Text(c.Type), f.Text(c.Value))
	}
	if s := it(16); !s.Mut {
		t.Error("static mut")
	}
	if ty := it(17); f.Text(ty.Type) != "Result<T, Error>" {
		t.Errorf("alias: %q", f.Text(ty.Type))
	}
	if g := it(18); !g.Const || !g.Unsafe || f.Text(g.Ret) != "&'a u8" || f.Text(g.Params) != "x: &'a u8" {
		t.Errorf("fn: ret %q params %q", f.Text(g.Ret), f.Text(g.Params))
	}

	if m := it(19); m.Macro != "satisfies" || m.Delim != "(" || f.Text(m.Args) != "Quote => R_X" {
		t.Errorf("macro call: %q %q %q", m.Macro, m.Delim, f.Text(m.Args))
	}
	if m := it(20); m.Macro != "speclink::rationale" || m.Delim != "{" {
		t.Errorf("braced macro call: %q", m.Macro)
	}
}

// Every item that is not read must still take its position with it, and the
// items after it must be read as if it were not there.
func TestUnreadItemsDoNotDisturbTheRest(t *testing.T) {
	f := Parse("x.rs", `
extern "C" { fn a(); }
global_asm!("nop");
struct After;
`)
	last := f.Items[len(f.Items)-1]
	if last.Kind != ItemStruct || last.Name != "After" || last.Pos.Line != 4 {
		t.Errorf("last item: %s %s at %v", last.Kind, last.Name, last.Pos)
	}
	if f.Items[0].Kind != ItemOpaque || f.Items[0].Pos.Line != 2 {
		t.Errorf("first item: %s at %v", f.Items[0].Kind, f.Items[0].Pos)
	}
}

func TestBodiesAreSkippedWhole(t *testing.T) {
	f := parse(t, `
fn f() {
    let s = r#"}"#;
    let c = '}';
    let l: &'static str = "{";
    /* } */
    struct Inner;
}
struct Next;
`)
	if len(f.Items) != 2 || f.Items[1].Name != "Next" {
		t.Fatalf("items: %d", len(f.Items))
	}
	body := f.Source(f.Items[0].Body)
	if !strings.HasPrefix(body, "{") || !strings.HasSuffix(body, "}") || !strings.Contains(body, "/* } */") {
		t.Errorf("body source: %q", body)
	}
}

func TestNormalize(t *testing.T) {
	for in, want := range map[string]string{
		`feature="postgres"`:         `feature = "postgres"`,
		`feature = "postgres"`:       `feature = "postgres"`,
		`all( unix , feature="x" )`:  `all(unix, feature = "x")`,
		"Vec < ( u8 , u8 ) >":        "Vec<(u8, u8)>",
		"&'a  mut  dyn Fn(u8)->bool": "&'a mut dyn Fn(u8) -> bool",
	} {
		if got := Normalize(in); got != want {
			t.Errorf("Normalize(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseErrorsArePositioned(t *testing.T) {
	f := Parse("x.rs", "struct A {\n  x: u8,\n")
	if len(f.Errors) == 0 {
		t.Fatal("no error for an unclosed struct")
	}
}
