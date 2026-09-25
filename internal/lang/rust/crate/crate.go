// Package crate builds the module tree of a Rust crate from its source files,
// and resolves paths in it as far as that can be done without a compiler.
//
// The tree is exact where Rust's rules are: mod name; finds name.rs or
// name/mod.rs, #[path] overrides that, inline modules nest directories. It is
// incomplete where they depend on evaluation — a mod declared by a macro, a
// #[cfg_attr] that sets a path — and says so rather than guessing.
//
// It also records, for every module, the cfg conditions on the way down from
// the crate root. A module under #[cfg(feature = "x")] is only compiled when
// the feature is on, and speclink reads it whether it is or not; the chain is
// what lets a later rule say which of the files it read the compiler may never
// have seen.
package crate

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/worldiety/speclink/internal/lang/rust/cargo"
	"github.com/worldiety/speclink/internal/lang/rust/syntax"
)

// Crate is one compiled crate: a library or a binary of a package.
type Crate struct {
	// Name is how paths spell it: the lib name, or the binary's name.
	Name    string
	Kind    cargo.TargetKind
	Package *cargo.Manifest
	Root    *Module
	// Modules lists every module in the order the tree was walked.
	Modules []*Module
	// Files maps an absolute path to the modules whose items it holds. A
	// file normally belongs to one; #[path] can make it two.
	Files map[string][]*Module
	// Problems are what the walk could not settle.
	Problems []Problem
}

// Module is one module, inline or in a file of its own.
type Module struct {
	Crate *Crate
	// Path is the module's path below the crate root, nil for the root.
	Path []string
	// File is the file holding the module's items; for an inline module it
	// is the file of the enclosing one.
	File   string
	Syntax *syntax.File
	Items  []*syntax.Item
	// Attrs are the attributes of the declaration and, for a file module,
	// the inner attributes of the file.
	Attrs []syntax.Attr
	// Decl is the mod item that declared it, nil for the root.
	Decl   *syntax.Item
	Parent *Module
	Inline bool
	// Children are the submodules in declaration order.
	Children []*Module
	// Cfg are the cfg conditions from the crate root down to and including
	// this module, outermost first, each rendered normalised.
	Cfg []string

	// childDir is where mod name; inside this module looks for name.rs.
	// pathBase is what #[path] inside it is relative to. They differ for a
	// file that is not a mod.rs: src/a.rs keeps its children in src/a/, but
	// a #[path] written in it is relative to src/.
	childDir string
	pathBase string
}

// Name is the module's own name, "crate" for the root.
func (m *Module) Name() string {
	if len(m.Path) == 0 {
		return "crate"
	}
	return m.Path[len(m.Path)-1]
}

// Symbol is the module's qualified name, crate_name::a::b.
func (m *Module) Symbol() string {
	return strings.Join(append([]string{m.Crate.Name}, m.Path...), "::")
}

// Child returns the submodule of that name.
func (m *Module) Child(name string) *Module {
	for _, c := range m.Children {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// Problem is something the walk could not settle, positioned.
type Problem struct {
	File string
	Pos  syntax.Pos
	Msg  string
}

// ParseFunc reads and parses one file. It is a parameter so that a caller
// holding many crates can parse a shared file once.
type ParseFunc func(path string) (*syntax.File, error)

// ReadFile is the ParseFunc that reads from disk.
func ReadFile(path string) (*syntax.File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return syntax.Parse(path, string(data)), nil
}

// Load walks the module tree of one target.
func Load(pkg *cargo.Manifest, t cargo.Target, parse ParseFunc) *Crate {
	if parse == nil {
		parse = ReadFile
	}
	c := &Crate{Name: t.Name, Kind: t.Kind, Package: pkg, Files: map[string][]*Module{}}
	f, err := parse(t.Path)
	if err != nil {
		c.Problems = append(c.Problems, Problem{File: t.Path, Msg: "cannot read the crate root: " + err.Error()})
		return c
	}
	dir := filepath.Dir(t.Path)
	c.Root = &Module{Crate: c, File: t.Path, Syntax: f, Items: f.Items, Attrs: f.Attrs, childDir: dir, pathBase: dir}
	c.Root.Cfg = cfgOf(nil, f.Attrs)
	c.add(c.Root)
	c.walk(c.Root, parse)
	return c
}

func (c *Crate) add(m *Module) {
	c.Modules = append(c.Modules, m)
	c.Files[m.File] = append(c.Files[m.File], m)
}

func (c *Crate) problem(file string, pos syntax.Pos, msg string) {
	c.Problems = append(c.Problems, Problem{File: file, Pos: pos, Msg: msg})
}

// walk visits the mod declarations of m.
func (c *Crate) walk(m *Module, parse ParseFunc) {
	for _, it := range m.Items {
		if it.Kind != syntax.ItemMod || it.Name == "" {
			continue
		}
		child := &Module{
			Crate:  c,
			Path:   append(append([]string(nil), m.Path...), it.Name),
			Decl:   it,
			Parent: m,
			Attrs:  it.Attrs,
		}
		m.Children = append(m.Children, child)
		for _, a := range syntax.AttrsNamed(it.Attrs, "cfg_attr") {
			if strings.Contains(a.Text(), "path") {
				c.problem(m.File, a.Pos, "mod "+it.Name+" sets its path through cfg_attr, which is not evaluated; the file read may not be the one compiled")
			}
		}
		pathAttr, hasPath := pathOf(it.Attrs)

		if !it.External {
			child.Inline = true
			child.File = m.File
			child.Syntax = m.Syntax
			child.Items = it.Items
			child.Attrs = append(append([]syntax.Attr(nil), it.Attrs...), it.Inner...)
			child.childDir = filepath.Join(m.childDir, it.Name)
			if hasPath {
				child.childDir = filepath.Join(m.childDir, pathAttr)
			}
			child.pathBase = child.childDir
			child.Cfg = cfgOf(m.Cfg, child.Attrs)
			c.add(child)
			c.walk(child, parse)
			continue
		}

		var file string
		modRS := false
		if hasPath {
			file = filepath.Join(m.pathBase, pathAttr)
			// A file reached through #[path] keeps its children beside it,
			// the way a mod.rs does.
			modRS = true
		} else {
			flat := filepath.Join(m.childDir, it.Name+".rs")
			nested := filepath.Join(m.childDir, it.Name, "mod.rs")
			switch fe, ne := isFile(flat), isFile(nested); {
			case fe && ne:
				c.problem(m.File, it.Pos, "mod "+it.Name+" is both "+rel(m, flat)+" and "+rel(m, nested)+"; cargo refuses this")
				file = flat
			case fe:
				file = flat
			case ne:
				file, modRS = nested, true
			default:
				c.problem(m.File, it.Pos, "mod "+it.Name+" has no file; looked for "+rel(m, flat)+" and "+rel(m, nested))
				child.Cfg = cfgOf(m.Cfg, it.Attrs)
				continue
			}
		}

		f, err := parse(file)
		if err != nil {
			c.problem(m.File, it.Pos, "mod "+it.Name+": "+err.Error())
			child.Cfg = cfgOf(m.Cfg, it.Attrs)
			continue
		}
		child.File = file
		child.Syntax = f
		child.Items = f.Items
		child.Attrs = append(append([]syntax.Attr(nil), it.Attrs...), f.Attrs...)
		child.pathBase = filepath.Dir(file)
		if modRS {
			child.childDir = filepath.Dir(file)
		} else {
			child.childDir = filepath.Join(filepath.Dir(file), strings.TrimSuffix(filepath.Base(file), ".rs"))
		}
		child.Cfg = cfgOf(m.Cfg, child.Attrs)

		if prior := c.Files[file]; len(prior) > 0 {
			c.problem(m.File, it.Pos, "mod "+it.Name+" includes "+rel(m, file)+", which is already "+prior[0].Symbol())
		}
		c.add(child)
		if len(child.Path) > 64 {
			c.problem(file, it.Pos, "module nesting deeper than 64; a #[path] probably includes a file in itself")
			continue
		}
		c.walk(child, parse)
	}
}

func pathOf(attrs []syntax.Attr) (string, bool) {
	for _, a := range syntax.AttrsNamed(attrs, "path") {
		if v, ok := a.Value(); ok {
			return v, true
		}
	}
	return "", false
}

// cfgOf extends the parent's chain with the cfg attributes of a module.
func cfgOf(parent []string, attrs []syntax.Attr) []string {
	out := append([]string(nil), parent...)
	for _, a := range syntax.AttrsNamed(attrs, "cfg") {
		out = append(out, a.Text())
	}
	return out
}

func isFile(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// rel renders a path relative to the package, which is how a finding names it.
func rel(m *Module, path string) string {
	if m.Crate.Package != nil {
		if r, err := filepath.Rel(m.Crate.Package.Dir, path); err == nil {
			return filepath.ToSlash(r)
		}
	}
	return path
}
