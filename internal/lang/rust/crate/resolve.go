package crate

import (
	"strings"

	"github.com/worldiety/speclink/internal/lang/rust/cargo"
	"github.com/worldiety/speclink/internal/lang/rust/syntax"
)

// World is every crate a path may lead into: the crates of the workspace.
//
// Anything outside it — std, a registry dependency — is not read, and a path
// into it resolves to an error that says so. That is not a failure of the
// resolver. speclink only ever needs to resolve paths to declarations the
// project made itself, and a requirement declared in serde is not one.
type World struct {
	libs map[string]*Crate // by package directory
	// Define reports the names a macro call in item position declares. The
	// resolver cannot expand a macro, and it does not need to for the few
	// it has been told about: requirement!(R_X = …) declares R_X, and that
	// is known without looking inside.
	Define func(m *Module, it *syntax.Item) []string
}

// NewWorld indexes the library crates among cs.
func NewWorld(cs []*Crate) *World {
	w := &World{libs: map[string]*Crate{}}
	for _, c := range cs {
		if c.Kind == cargo.Lib && c.Package != nil {
			w.libs[c.Package.Dir] = c
		}
	}
	return w
}

// DefKind is what a path resolved to.
type DefKind uint8

const (
	DefModule DefKind = iota + 1
	// DefItem is an item of a module: a type, fn, const, static, trait or
	// a name a known macro declares.
	DefItem
	// DefMember is below an item: an enum variant, a method, an associated
	// const or type.
	DefMember
)

// Def is a resolved path.
type Def struct {
	Kind DefKind
	// Symbol is the canonical name, crate::module::Item, spelled from the
	// place of definition rather than from wherever it was imported
	// through. Two paths to one item have one symbol.
	Symbol string
	// Module is the module that holds the definition, or the module itself.
	Module *Module
	// Item is the item, or for a member the item it belongs to.
	Item *syntax.Item
	// Member is the member's own item where it is one (a method, an
	// associated const), nil for an enum variant.
	Member *syntax.Item
}

// Unresolved is a path that could not be followed, with the reason.
type Unresolved struct {
	Path   string
	Reason string
}

func (u *Unresolved) Error() string { return u.Path + ": " + u.Reason }

// maxDepth bounds the chains of re-exports and globs followed. A use cycle
// does not compile, so this is only reached by a tree the resolver misread.
const maxDepth = 16

// Resolve follows path as written in module from.
func (w *World) Resolve(from *Module, path []string) (Def, error) {
	return w.resolve(from, path, 0)
}

// ResolveText splits a path written as a::b::C and resolves it.
func (w *World) ResolveText(from *Module, path string) (Def, error) {
	var segs []string
	if strings.HasPrefix(path, "::") {
		segs = append(segs, "::")
		path = path[2:]
	}
	for _, s := range strings.Split(path, "::") {
		segs = append(segs, strings.TrimSpace(s))
	}
	return w.Resolve(from, segs)
}

func fail(path []string, reason string) error {
	return &Unresolved{Path: strings.Join(path, "::"), Reason: reason}
}

func moduleDef(m *Module) Def { return Def{Kind: DefModule, Symbol: m.Symbol(), Module: m} }

func (w *World) resolve(from *Module, path []string, depth int) (Def, error) {
	if depth > maxDepth {
		return Def{}, fail(path, "re-exports lead in a circle")
	}
	if len(path) == 0 || from == nil {
		return Def{}, fail(path, "empty path")
	}

	var cur Def
	rest := path[1:]
	switch head := path[0]; head {
	case "::":
		if len(path) < 2 {
			return Def{}, fail(path, "empty path")
		}
		ext := w.extern(from.Crate, path[1])
		if ext == nil {
			return Def{}, fail(path, "crate "+path[1]+" is not part of the workspace")
		}
		cur, rest = moduleDef(ext.Root), path[2:]
	case "crate":
		cur = moduleDef(from.Crate.Root)
	case "self":
		cur = moduleDef(from)
	case "super":
		if from.Parent == nil {
			return Def{}, fail(path, "super above the crate root")
		}
		cur = moduleDef(from.Parent)
	case "Self":
		return Def{}, fail(path, "Self depends on the impl it is written in")
	default:
		if d, ok := w.member(from, head, depth); ok {
			cur = d
		} else if ext := w.extern(from.Crate, head); ext != nil {
			cur = moduleDef(ext.Root)
		} else if head == "std" || head == "core" || head == "alloc" {
			return Def{}, fail(path, head+" is outside the workspace")
		} else {
			return Def{}, fail(path, head+" is not in scope in "+from.Symbol())
		}
	}

	for _, seg := range rest {
		switch cur.Kind {
		case DefModule:
			if seg == "super" {
				if cur.Module.Parent == nil {
					return Def{}, fail(path, "super above the crate root")
				}
				cur = moduleDef(cur.Module.Parent)
				continue
			}
			d, ok := w.member(cur.Module, seg, depth)
			if !ok {
				return Def{}, fail(path, cur.Symbol+" has no item "+seg)
			}
			cur = d
		case DefItem:
			d, ok := w.assoc(cur, seg, depth)
			if !ok {
				return Def{}, fail(path, cur.Symbol+" has no member "+seg)
			}
			cur = d
		default:
			return Def{}, fail(path, cur.Symbol+" has nothing below it")
		}
	}
	return cur, nil
}

// extern returns the crate a name refers to from inside c: a local
// dependency, or for a binary its own package's library.
func (w *World) extern(c *Crate, name string) *Crate {
	pkg := c.Package
	if pkg == nil {
		return nil
	}
	if c.Kind == cargo.Bin && pkg.Lib != nil && pkg.Lib.Name == name {
		return w.libs[pkg.Dir]
	}
	if d, ok := pkg.Deps[name]; ok && d.Path != "" {
		return w.libs[d.Path]
	}
	return nil
}

// member finds name among what module m declares or imports.
//
// Visibility is not checked. The compiler has already refused every path
// that reaches something private, so the only effect checking it here could
// have is to disagree with the compiler.
func (w *World) member(m *Module, name string, depth int) (Def, bool) {
	if name == "_" || name == "" {
		return Def{}, false
	}
	if c := m.Child(name); c != nil && c.Syntax != nil {
		return moduleDef(c), true
	}
	for _, it := range m.Items {
		if declares(it, name) {
			return Def{Kind: DefItem, Symbol: m.Symbol() + "::" + name, Module: m, Item: it}, true
		}
		if it.Kind == syntax.ItemMacroCall && w.Define != nil {
			for _, n := range w.Define(m, it) {
				if n == name {
					return Def{Kind: DefItem, Symbol: m.Symbol() + "::" + name, Module: m, Item: it}, true
				}
			}
		}
	}
	for _, it := range m.Items {
		if it.Kind != syntax.ItemUse && it.Kind != syntax.ItemExternCrate {
			continue
		}
		for _, u := range it.Uses {
			if u.Glob || u.Name() != name {
				continue
			}
			if d, err := w.resolve(m, u.Path, depth+1); err == nil {
				return d, true
			}
		}
	}
	for _, it := range m.Items {
		if it.Kind != syntax.ItemUse {
			continue
		}
		for _, u := range it.Uses {
			if !u.Glob {
				continue
			}
			target, err := w.resolve(m, u.Path, depth+1)
			if err != nil {
				continue
			}
			switch target.Kind {
			case DefModule:
				if d, ok := w.member(target.Module, name, depth+1); ok {
					return d, true
				}
			case DefItem:
				if target.Item.Kind == syntax.ItemEnum {
					if d, ok := w.assoc(target, name, depth+1); ok {
						return d, true
					}
				}
			}
		}
	}
	return Def{}, false
}

func declares(it *syntax.Item, name string) bool {
	if it.Name != name {
		return false
	}
	switch it.Kind {
	case syntax.ItemStruct, syntax.ItemUnion, syntax.ItemEnum, syntax.ItemTrait, syntax.ItemFn,
		syntax.ItemConst, syntax.ItemStatic, syntax.ItemType, syntax.ItemMacroRules:
		return true
	}
	return false
}

// assoc finds a member of an item: a variant of an enum, an item of a trait,
// or an item of an impl of the type anywhere in its crate.
func (w *World) assoc(d Def, name string, depth int) (Def, bool) {
	member := func(sub *syntax.Item) Def {
		return Def{Kind: DefMember, Symbol: d.Symbol + "::" + name, Module: d.Module, Item: d.Item, Member: sub}
	}
	switch d.Item.Kind {
	case syntax.ItemEnum:
		for _, v := range d.Item.Variants {
			if v.Name == name {
				return member(nil), true
			}
		}
	case syntax.ItemTrait:
		for _, sub := range d.Item.Items {
			if sub.Name == name {
				return member(sub), true
			}
		}
	}
	if depth > maxDepth {
		return Def{}, false
	}
	for _, m := range d.Module.Crate.Modules {
		for _, it := range m.Items {
			if it.Kind != syntax.ItemImpl {
				continue
			}
			var hit *syntax.Item
			for _, sub := range it.Items {
				if sub.Name == name {
					hit = sub
				}
			}
			if hit == nil || !w.implFor(m, it, d, depth) {
				continue
			}
			return member(hit), true
		}
	}
	return Def{}, false
}

// implFor reports whether an impl in module m is for the item d. The self
// type is resolved where the impl is written; a self type the resolver cannot
// follow is matched by its last name, which is right far more often than a
// refusal would be useful.
func (w *World) implFor(m *Module, impl *syntax.Item, d Def, depth int) bool {
	path := typePath(m.Syntax.Tokens(impl.SelfType))
	if len(path) == 0 {
		return false
	}
	if got, err := w.resolve(m, path, depth+1); err == nil {
		return got.Symbol == d.Symbol
	}
	return path[len(path)-1] == d.Item.Name
}

// typePath reads the path at the start of a type, up to its generic
// arguments: crate::a::Quote<T> is crate, a, Quote.
func typePath(toks []syntax.Token) []string {
	var out []string
	for _, t := range toks {
		switch {
		case t.Kind == syntax.Ident && !t.Is("dyn") && !t.Is("mut"):
			out = append(out, t.Text)
		case t.Is("::"):
			if len(out) == 0 {
				out = append(out, "::")
			}
		case t.Is("&"):
		default:
			return out
		}
	}
	return out
}
