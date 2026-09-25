// Package cargo reads what speclink needs from Cargo manifests: which packages
// a workspace has, which crates each of them builds, where those crates start,
// and which other local packages they depend on.
//
// It does not run cargo. Asking cargo metadata would be more exact and would
// make speclink depend on a toolchain being installed where it runs, which is
// the one thing a static reader of a checked-in tree should not need. The
// rules cargo applies to find targets are few and documented, and they are
// written out below.
package cargo

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileName is the manifest's name.
const FileName = "Cargo.toml"

// Manifest is one Cargo.toml.
type Manifest struct {
	// Path is the absolute path of the manifest, Dir its directory.
	Path string
	Dir  string
	// Package is nil for a virtual manifest, which only declares a
	// workspace.
	Package *Package
	// Workspace is set when the manifest declares one.
	Workspace *WorkspaceTable
	// Lib is the library target, nil when the package has none.
	Lib *Target
	// Bins are the binary targets, explicit and discovered.
	Bins []Target
	// Deps are the dependencies by the name the code uses for them.
	Deps map[string]Dependency
}

// Package is the [package] table.
type Package struct {
	Name    string
	Edition string
	// AutoBins is false when the package switched off target discovery.
	AutoBins bool
}

// WorkspaceTable is the [workspace] table.
type WorkspaceTable struct {
	Members []string
	Exclude []string
	// Deps are [workspace.dependencies], which a member inherits with
	// workspace = true.
	Deps map[string]Dependency
}

// TargetKind distinguishes the crates a package builds.
type TargetKind uint8

const (
	Lib TargetKind = iota + 1
	Bin
)

func (k TargetKind) String() string {
	if k == Lib {
		return "lib"
	}
	return "bin"
}

// Target is one crate a package builds.
type Target struct {
	Kind TargetKind
	// Name is the crate name as code refers to it: a hyphen in the package
	// name is an underscore here.
	Name string
	// Path is the absolute path of the crate root.
	Path string
}

// Dependency is one entry of [dependencies] or [dev-dependencies].
type Dependency struct {
	// Name is how the code refers to it.
	Name string
	// Package is the package it names, which differs from Name when the
	// dependency was renamed.
	Package string
	// Path is the absolute directory of a local dependency, empty for one
	// from a registry or a repository.
	Path string
	Dev  bool
	// Inherited is set for workspace = true, which Load resolves.
	Inherited bool
}

// CrateName turns a package name into the name code uses for its crate.
func CrateName(pkg string) string { return strings.ReplaceAll(pkg, "-", "_") }

// ReadManifest reads one manifest and works out its targets.
func ReadManifest(path string) (*Manifest, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	doc, err := parseTOML(string(data))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", abs, err)
	}
	m := &Manifest{Path: abs, Dir: filepath.Dir(abs), Deps: map[string]Dependency{}}

	if t, ok := doc["package"].(map[string]any); ok {
		m.Package = &Package{Name: str(t["name"]), Edition: str(t["edition"]), AutoBins: true}
		if v, ok := t["autobins"].(bool); ok {
			m.Package.AutoBins = v
		}
		if m.Package.Name == "" {
			return nil, fmt.Errorf("%s: [package] has no name", abs)
		}
	}
	if t, ok := doc["workspace"].(map[string]any); ok {
		m.Workspace = &WorkspaceTable{
			Members: strs(t["members"]),
			Exclude: strs(t["exclude"]),
			Deps:    deps(m.Dir, t["dependencies"], false),
		}
	}
	for k, v := range deps(m.Dir, doc["dependencies"], false) {
		m.Deps[k] = v
	}
	for k, v := range deps(m.Dir, doc["dev-dependencies"], true) {
		if _, normal := m.Deps[k]; !normal {
			m.Deps[k] = v
		}
	}
	if m.Package != nil {
		m.targets(doc)
	}
	return m, nil
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func strs(v any) []string {
	list, _ := v.([]any)
	var out []string
	for _, e := range list {
		if s, ok := e.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func deps(dir string, v any, dev bool) map[string]Dependency {
	table, _ := v.(map[string]any)
	out := map[string]Dependency{}
	for name, spec := range table {
		d := Dependency{Name: CrateName(name), Package: name, Dev: dev}
		if t, ok := spec.(map[string]any); ok {
			if p := str(t["package"]); p != "" {
				d.Package = p
			}
			if p := str(t["path"]); p != "" {
				d.Path = filepath.Clean(filepath.Join(dir, p))
			}
			if w, ok := t["workspace"].(bool); ok && w {
				d.Inherited = true
			}
		}
		out[d.Name] = d
	}
	return out
}

// targets applies cargo's target discovery.
//
// The library is src/lib.rs unless [lib] says otherwise. Binaries are the
// [[bin]] entries, and — unless autobins is off — src/main.rs as a binary
// named after the package, src/bin/*.rs, and src/bin/*/main.rs. Examples,
// tests and benches are crates too, but nothing speclink reads may live in
// them, so they are not listed.
func (m *Manifest) targets(doc map[string]any) {
	pkg := m.Package.Name
	if t, ok := doc["lib"].(map[string]any); ok {
		lib := Target{Kind: Lib, Name: CrateName(pkg), Path: filepath.Join(m.Dir, "src", "lib.rs")}
		if n := str(t["name"]); n != "" {
			lib.Name = CrateName(n)
		}
		if p := str(t["path"]); p != "" {
			lib.Path = filepath.Join(m.Dir, p)
		}
		m.Lib = &lib
	} else if exists(filepath.Join(m.Dir, "src", "lib.rs")) {
		m.Lib = &Target{Kind: Lib, Name: CrateName(pkg), Path: filepath.Join(m.Dir, "src", "lib.rs")}
	}

	used := map[string]bool{}
	bins, _ := doc["bin"].([]map[string]any)
	for _, t := range bins {
		name := str(t["name"])
		path := str(t["path"])
		switch {
		case path != "":
			path = filepath.Join(m.Dir, path)
		case name == pkg && exists(filepath.Join(m.Dir, "src", "main.rs")):
			path = filepath.Join(m.Dir, "src", "main.rs")
		case exists(filepath.Join(m.Dir, "src", "bin", name+".rs")):
			path = filepath.Join(m.Dir, "src", "bin", name+".rs")
		default:
			path = filepath.Join(m.Dir, "src", "bin", name, "main.rs")
		}
		if name == "" {
			name = strings.TrimSuffix(filepath.Base(path), ".rs")
		}
		m.Bins = append(m.Bins, Target{Kind: Bin, Name: CrateName(name), Path: path})
		used[path] = true
	}
	if !m.Package.AutoBins {
		return
	}

	add := func(name, path string) {
		if !used[path] && exists(path) {
			m.Bins = append(m.Bins, Target{Kind: Bin, Name: CrateName(name), Path: path})
			used[path] = true
		}
	}
	add(pkg, filepath.Join(m.Dir, "src", "main.rs"))
	entries, _ := os.ReadDir(filepath.Join(m.Dir, "src", "bin"))
	for _, e := range entries {
		switch {
		case !e.IsDir() && strings.HasSuffix(e.Name(), ".rs"):
			add(strings.TrimSuffix(e.Name(), ".rs"), filepath.Join(m.Dir, "src", "bin", e.Name()))
		case e.IsDir():
			add(e.Name(), filepath.Join(m.Dir, "src", "bin", e.Name(), "main.rs"))
		}
	}
}

func exists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// Targets returns the library first and then the binaries.
func (m *Manifest) Targets() []Target {
	var out []Target
	if m.Lib != nil {
		out = append(out, *m.Lib)
	}
	return append(out, m.Bins...)
}

// Workspace is the set of packages under one root.
type Workspace struct {
	Root     *Manifest
	Packages []*Manifest
}

// Load reads the workspace rooted at dir.
//
// A root with [workspace] lists its members, globs included; a root that is
// only a package is a workspace of one. A member's inherited dependencies are
// resolved against the root, so that a path dependency written once at the
// top is known to every member that uses it.
func Load(dir string) (*Workspace, error) {
	root, err := ReadManifest(filepath.Join(dir, FileName))
	if err != nil {
		return nil, err
	}
	w := &Workspace{Root: root}
	seen := map[string]bool{}
	addPkg := func(m *Manifest) {
		if m.Package != nil && !seen[m.Dir] {
			seen[m.Dir] = true
			w.Packages = append(w.Packages, m)
		}
	}
	addPkg(root)

	if root.Workspace != nil {
		excluded := map[string]bool{}
		for _, e := range root.Workspace.Exclude {
			matches, _ := filepath.Glob(filepath.Join(root.Dir, e))
			for _, m := range matches {
				excluded[filepath.Clean(m)] = true
			}
		}
		for _, pattern := range root.Workspace.Members {
			matches, err := filepath.Glob(filepath.Join(root.Dir, pattern))
			if err != nil {
				return nil, fmt.Errorf("workspace member %q: %w", pattern, err)
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("workspace member %q matches no directory", pattern)
			}
			sort.Strings(matches)
			for _, d := range matches {
				d = filepath.Clean(d)
				if excluded[d] || !exists(filepath.Join(d, FileName)) {
					continue
				}
				m, err := ReadManifest(filepath.Join(d, FileName))
				if err != nil {
					return nil, err
				}
				addPkg(m)
			}
		}
		for _, m := range w.Packages {
			for k, d := range m.Deps {
				if !d.Inherited {
					continue
				}
				if top, ok := root.Workspace.Deps[k]; ok {
					top.Dev = d.Dev
					m.Deps[k] = top
				}
			}
		}
	}
	if len(w.Packages) == 0 {
		return nil, fmt.Errorf("%s declares no package and no member", root.Path)
	}
	return w, nil
}

// PackageAt returns the package whose directory is dir, nil if none.
func (w *Workspace) PackageAt(dir string) *Manifest {
	dir = filepath.Clean(dir)
	for _, m := range w.Packages {
		if m.Dir == dir {
			return m
		}
	}
	return nil
}
