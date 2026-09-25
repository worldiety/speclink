package crate

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/worldiety/speclink/internal/lang/rust/cargo"
	"github.com/worldiety/speclink/internal/lang/rust/syntax"
)

func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func loadAll(t *testing.T, root string) (*cargo.Workspace, []*Crate) {
	t.Helper()
	w, err := cargo.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	var cs []*Crate
	for _, p := range w.Packages {
		for _, tg := range p.Targets() {
			cs = append(cs, Load(p, tg, nil))
		}
	}
	return w, cs
}

func modules(c *Crate) []string {
	var out []string
	for _, m := range c.Modules {
		rel, _ := filepath.Rel(c.Package.Dir, m.File)
		s := m.Symbol() + " " + filepath.ToSlash(rel)
		if m.Inline {
			s += " inline"
		}
		if len(m.Cfg) > 0 {
			s += " [" + strings.Join(m.Cfg, " | ") + "]"
		}
		out = append(out, s)
	}
	return out
}

func TestModuleTree(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"Cargo.toml": "[package]\nname = \"shop\"\n",
		"src/lib.rs": `
pub mod sales;
mod legacy;
#[cfg(feature = "pg")]
mod store;
mod inline { mod deep; #[path = "elsewhere.rs"] mod moved; }
#[path = "odd/place.rs"]
mod odd;
`,
		"src/sales.rs":            "mod quote;\n#[path = \"sales/quote.spec.rs\"]\nmod quote_spec;\n",
		"src/sales/quote.rs":      "",
		"src/sales/quote.spec.rs": "",
		"src/legacy/mod.rs":       "mod old;",
		"src/legacy/old.rs":       "",
		"src/store.rs":            "#![cfg(unix)]\nmod pg;",
		"src/store/pg.rs":         "",
		"src/inline/deep.rs":      "",
		"src/inline/elsewhere.rs": "",
		"src/odd/place.rs":        "mod near;",
		"src/odd/near.rs":         "",
	})
	_, cs := loadAll(t, root)
	c := cs[0]
	if len(c.Problems) > 0 {
		t.Fatalf("problems: %+v", c.Problems)
	}
	want := []string{
		"shop src/lib.rs",
		"shop::sales src/sales.rs",
		"shop::sales::quote src/sales/quote.rs",
		"shop::sales::quote_spec src/sales/quote.spec.rs",
		"shop::legacy src/legacy/mod.rs",
		"shop::legacy::old src/legacy/old.rs",
		`shop::store src/store.rs [feature = "pg" | unix]`,
		`shop::store::pg src/store/pg.rs [feature = "pg" | unix]`,
		"shop::inline src/lib.rs inline",
		"shop::inline::deep src/inline/deep.rs",
		"shop::inline::moved src/inline/elsewhere.rs",
		"shop::odd src/odd/place.rs",
		"shop::odd::near src/odd/near.rs",
	}
	if got := modules(c); !reflect.DeepEqual(got, want) {
		t.Errorf("modules:\n got  %q\n want %q", got, want)
	}
}

func TestModuleTreeProblems(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"Cargo.toml": "[package]\nname = \"x\"\n",
		"src/lib.rs": `
mod missing;
mod both;
mod a;
#[path = "a.rs"] mod again;
#[cfg_attr(unix, path = "u.rs")] mod plat;
`,
		"src/both.rs":     "",
		"src/both/mod.rs": "",
		"src/a.rs":        "",
		"src/plat.rs":     "",
	})
	_, cs := loadAll(t, root)
	var msgs []string
	for _, p := range cs[0].Problems {
		msgs = append(msgs, p.Msg)
	}
	joined := strings.Join(msgs, "\n")
	for _, want := range []string{"mod missing has no file", "mod both is both", "which is already x::a", "through cfg_attr"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing problem %q in:\n%s", want, joined)
		}
	}
}

func TestResolve(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"Cargo.toml":      "[workspace]\nmembers = [\"spec\", \"app\"]\n",
		"spec/Cargo.toml": "[package]\nname = \"shop-spec\"\n",
		"spec/src/lib.rs": `
pub mod requirements;
pub use requirements::quote::R_QUOTE_SUBMIT as SUBMIT;
`,
		"spec/src/requirements/mod.rs": "pub mod quote;\npub mod dec;\npub use dec::*;\n",
		"spec/src/requirements/quote.rs": `
use super::dec::R_DEC_NUMBERING;
requirement!(R_QUOTE_SUBMIT = "R-QUOTE-SUBMIT" { derived_from: &[&R_DEC_NUMBERING] });
`,
		"spec/src/requirements/dec.rs": "pub const R_DEC_NUMBERING: Requirement = Requirement::EMPTY;\n",
		"app/Cargo.toml":               "[package]\nname = \"app\"\n[dependencies]\nshop-spec = { path = \"../spec\" }\n",
		"app/src/lib.rs": `
pub mod sales;
pub use sales::Quote as Offer;
`,
		"app/src/sales.rs": `
use shop_spec::requirements::{self as reqs, quote};
use crate::Offer;
pub struct Quote { pub number: String }
impl Quote { pub fn submit(&self) {} }
pub enum State { Open, Closed }
pub trait Repo { fn save(&self); }
pub mod inner { pub fn helper() {} }
`,
		"app/src/main.rs": "use app::sales::Quote;\nfn main() {}\n",
	})
	_, cs := loadAll(t, root)
	w := NewWorld(cs)
	w.Define = func(m *Module, it *syntax.Item) []string {
		if toks := m.Syntax.Tokens(it.Args); it.Macro == "requirement" && len(toks) > 0 {
			return []string{toks[0].Text}
		}
		return nil
	}

	var app, bin *Crate
	for _, c := range cs {
		if len(c.Problems) > 0 {
			t.Fatalf("%s: %+v", c.Name, c.Problems)
		}
		if c.Name == "app" && c.Kind == cargo.Lib {
			app = c
		}
		if c.Kind == cargo.Bin {
			bin = c
		}
	}
	sales := app.Root.Child("sales")

	tests := []struct {
		from *Module
		path string
		want string
	}{
		{sales, "Quote", "app::sales::Quote"},
		{sales, "self::Quote", "app::sales::Quote"},
		{sales, "crate::sales::Quote", "app::sales::Quote"},
		{sales, "Offer", "app::sales::Quote"},
		{sales, "super::Offer", "app::sales::Quote"},
		{sales, "Quote::submit", "app::sales::Quote::submit"},
		{sales, "State::Closed", "app::sales::State::Closed"},
		{sales, "Repo::save", "app::sales::Repo::save"},
		{sales, "inner::helper", "app::sales::inner::helper"},
		{sales, "quote::R_QUOTE_SUBMIT", "shop_spec::requirements::quote::R_QUOTE_SUBMIT"},
		{sales, "reqs::quote::R_QUOTE_SUBMIT", "shop_spec::requirements::quote::R_QUOTE_SUBMIT"},
		{sales, "shop_spec::SUBMIT", "shop_spec::requirements::quote::R_QUOTE_SUBMIT"},
		{sales, "::shop_spec::SUBMIT", "shop_spec::requirements::quote::R_QUOTE_SUBMIT"},
		{sales, "reqs::R_DEC_NUMBERING", "shop_spec::requirements::dec::R_DEC_NUMBERING"},
		{bin.Root, "Quote", "app::sales::Quote"},
		{bin.Root, "app::Offer", "app::sales::Quote"},
	}
	for _, tt := range tests {
		d, err := w.ResolveText(tt.from, tt.path)
		if err != nil {
			t.Errorf("%s from %s: %v", tt.path, tt.from.Symbol(), err)
			continue
		}
		if d.Symbol != tt.want {
			t.Errorf("%s from %s = %s, want %s", tt.path, tt.from.Symbol(), d.Symbol, tt.want)
		}
	}

	for _, path := range []string{"Missing", "Quote::nope", "std::fmt::Debug", "serde::Serialize", "super::super::X"} {
		if d, err := w.ResolveText(sales, path); err == nil {
			t.Errorf("%s resolved to %s", path, d.Symbol)
		}
	}
}
