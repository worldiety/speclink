// Package ts is the TypeScript frontend: a Vue single page application in the
// ports and adapters style.
//
// It reads source, not compiled output. A TypeScript build leaves nothing
// behind that is easier to read than what went into it — the JavaScript a
// bundler writes has lost its types, its imports and its file boundaries,
// which are exactly what the rules ask about. vue-tsc has already checked that
// every name resolves; what is left is to find the declarations and follow the
// imports, and a declaration level reader does both.
package ts

import (
	"encoding/json"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/worldiety/speclink/internal/config"
	"github.com/worldiety/speclink/internal/ir"
	"github.com/worldiety/speclink/internal/lang/ts/syntax"
)

// Module is one source file of the project.
type Module struct {
	// ID identifies the module in qualified names: its path relative to the
	// root without the .ts or .js extension. A component keeps .vue, because
	// Foo.ts and Foo.vue may sit side by side and are different modules.
	ID string
	// Path is the root relative path with extension, slash separated.
	Path string
	// Abs is the absolute path, which is what positions carry.
	Abs string
	// Dir is the root relative directory, the unit the scope is written in.
	Dir string
	Src string
	Vue bool
	// Scripts are the parsed scripts: one for a .ts file, one per script
	// block of a component.
	Scripts []*syntax.File
	// Templates are the template regions of a component.
	Templates []syntax.Block
	// Imports are the import declarations, resolved.
	Imports []*Ref

	binds map[string]binding
}

// Ref is one import, resolved against the project.
type Ref struct {
	Import *syntax.Import
	File   *syntax.File
	// Target is the module the import names, nil when it is a package or a
	// file that is not a module.
	Target *Module
	// Local is the root relative path of a project file that is not a module,
	// a stylesheet or a JSON catalogue.
	Local string
	// Package is the package name of a bare specifier.
	Package string
	Pos     ir.Position
}

// Project is every module of a project, with its imports resolved.
type Project struct {
	Root    string
	Layout  config.Config
	Modules []*Module
	byPath  map[string]*Module
	files   map[string]bool
	aliases []alias
}

// alias is one entry of compilerOptions.paths.
type alias struct {
	pattern string
	targets []string
}

// skipDirs are never descended into. They hold what was installed or built,
// not what was written.
var skipDirs = map[string]bool{
	"node_modules": true, "dist": true, "coverage": true, "build": true, "out": true,
}

// Load reads every module below root.
//
// Everything is loaded whatever the scope says, for the reason every frontend
// gives: a rule that follows an import must see where it leads, and a verdict
// that changed with the configured scope would be worse than no verdict.
func Load(root string, layout config.Config) (*Project, []error) {
	p := &Project{Root: root, Layout: layout, byPath: map[string]*Module{}, files: map[string]bool{}}
	var errs []error

	err := filepath.WalkDir(root, func(abs string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() {
			if abs != root && (strings.HasPrefix(name, ".") || skipDirs[name]) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(root, abs)
		rel = filepath.ToSlash(rel)
		p.files[rel] = true
		if !isModuleFile(name) {
			return nil
		}
		data, readErr := os.ReadFile(abs)
		if readErr != nil {
			errs = append(errs, readErr)
			return nil
		}
		p.add(rel, abs, string(data))
		return nil
	})
	if err != nil {
		errs = append(errs, err)
	}
	sort.Slice(p.Modules, func(i, j int) bool { return p.Modules[i].Path < p.Modules[j].Path })

	p.aliases = readAliases(root)
	for _, m := range p.Modules {
		p.resolveImports(m)
	}
	return p, errs
}

func isModuleFile(name string) bool {
	switch {
	case strings.HasSuffix(name, ".d.ts"):
		return false
	case strings.HasSuffix(name, ".ts"), strings.HasSuffix(name, ".vue"),
		strings.HasSuffix(name, ".js"), strings.HasSuffix(name, ".mjs"):
		return true
	}
	return false
}

func (p *Project) add(rel, abs, src string) {
	m := &Module{
		ID:   moduleID(rel),
		Path: rel,
		Abs:  abs,
		Dir:  path.Dir(rel),
		Src:  src,
		Vue:  strings.HasSuffix(rel, ".vue"),
	}
	if m.Vue {
		for _, b := range syntax.SFCBlocks(src) {
			switch b.Tag {
			case "script":
				m.Scripts = append(m.Scripts, syntax.ParseRegion(abs, src, b.From, b.To))
			case "template":
				m.Templates = append(m.Templates, b)
			}
		}
	} else {
		m.Scripts = []*syntax.File{syntax.Parse(abs, src)}
	}
	p.Modules = append(p.Modules, m)
	p.byPath[rel] = m
}

func moduleID(rel string) string {
	for _, ext := range []string{".ts", ".mjs", ".js"} {
		if strings.HasSuffix(rel, ext) {
			return strings.TrimSuffix(rel, ext)
		}
	}
	return rel
}

// Module returns the module at a root relative path.
func (p *Project) Module(rel string) *Module { return p.byPath[rel] }

// Pos turns a token position of a module into a reported position.
func (m *Module) Pos(at syntax.Pos) ir.Position {
	return ir.Position{File: m.Abs, Line: at.Line, Col: at.Col}
}

// Items returns the top level statements of every script of the module.
func (m *Module) Items() []itemOf {
	var out []itemOf
	for _, f := range m.Scripts {
		for _, it := range f.Items {
			out = append(out, itemOf{f, it})
		}
	}
	return out
}

type itemOf struct {
	File *syntax.File
	*syntax.Item
}

func (p *Project) resolveImports(m *Module) {
	for _, f := range m.Scripts {
		for _, it := range f.Items {
			if it.Import == nil || it.Import.Spec == "" {
				continue
			}
			ref := &Ref{Import: it.Import, File: f, Pos: m.Pos(it.Import.SpecPos)}
			p.resolve(m, ref)
			m.Imports = append(m.Imports, ref)
		}
	}
}

// resolve finds what a specifier names: a module, another file of the
// project, or a package.
func (p *Project) resolve(m *Module, ref *Ref) {
	spec := ref.Import.Spec
	var candidates []string
	switch {
	case strings.HasPrefix(spec, "./") || strings.HasPrefix(spec, "../") || spec == "." || spec == "..":
		candidates = []string{path.Join(m.Dir, spec)}
	case strings.HasPrefix(spec, "/"):
		candidates = []string{strings.TrimPrefix(spec, "/")}
	default:
		candidates = p.aliased(spec)
		if len(candidates) == 0 {
			ref.Package = packageName(spec)
			return
		}
	}
	for _, c := range candidates {
		if target, local := p.find(path.Clean(c)); target != nil || local != "" {
			ref.Target, ref.Local = target, local
			return
		}
	}
	// Named a project path and found nothing there. Reported as a local file
	// that does not exist rather than mistaken for a package.
	ref.Local = path.Clean(candidates[0])
}

// find tries the extensions a bundler would.
func (p *Project) find(c string) (*Module, string) {
	tries := []string{c}
	for _, ext := range []string{".js", ".mjs"} {
		if strings.HasSuffix(c, ext) {
			tries = append(tries, strings.TrimSuffix(c, ext)+".ts")
		}
	}
	tries = append(tries, c+".ts", c+".vue", c+".js", c+"/index.ts", c+"/index.js")
	for _, t := range tries {
		if m := p.byPath[t]; m != nil {
			return m, ""
		}
	}
	if p.files[c] {
		return nil, c
	}
	return nil, ""
}

func packageName(spec string) string {
	parts := strings.Split(spec, "/")
	if strings.HasPrefix(spec, "@") && len(parts) > 1 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

func (p *Project) aliased(spec string) []string {
	var out []string
	for _, a := range p.aliases {
		prefix, wildcard := strings.CutSuffix(a.pattern, "*")
		switch {
		case wildcard && strings.HasPrefix(spec, prefix):
			rest := strings.TrimPrefix(spec, prefix)
			for _, t := range a.targets {
				out = append(out, strings.Replace(t, "*", rest, 1))
			}
		case !wildcard && spec == a.pattern:
			out = append(out, a.targets...)
		}
	}
	return out
}

// readAliases collects compilerOptions.paths from every tsconfig in the root.
//
// A Vue project splits its configuration over several files that reference
// one another, and following the references would be reimplementing the
// compiler's project model to learn one table. The aliases are the same in all
// of them in every project anybody writes, so they are simply collected.
func readAliases(root string) []alias {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []alias
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "tsconfig") || !strings.HasSuffix(name, ".json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			continue
		}
		var cfg struct {
			CompilerOptions struct {
				BaseURL string              `json:"baseUrl"`
				Paths   map[string][]string `json:"paths"`
			} `json:"compilerOptions"`
		}
		if json.Unmarshal(StripJSONC(data), &cfg) != nil {
			continue
		}
		base := cfg.CompilerOptions.BaseURL
		if base == "" {
			base = "."
		}
		keys := make([]string, 0, len(cfg.CompilerOptions.Paths))
		for k := range cfg.CompilerOptions.Paths {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			if seen[k] {
				continue
			}
			seen[k] = true
			var targets []string
			for _, t := range cfg.CompilerOptions.Paths[k] {
				targets = append(targets, path.Clean(path.Join(base, t)))
			}
			out = append(out, alias{pattern: k, targets: targets})
		}
	}
	// Longest pattern first, so that a specific alias wins over a general one.
	sort.SliceStable(out, func(i, j int) bool { return len(out[i].pattern) > len(out[j].pattern) })
	return out
}

// StripJSONC removes comments and trailing commas, which tsconfig files allow
// and encoding/json does not.
func StripJSONC(data []byte) []byte {
	var out []byte
	inString := false
	for i := 0; i < len(data); i++ {
		c := data[i]
		if inString {
			out = append(out, c)
			if c == '\\' && i+1 < len(data) {
				i++
				out = append(out, data[i])
			} else if c == '"' {
				inString = false
			}
			continue
		}
		switch {
		case c == '"':
			inString = true
			out = append(out, c)
		case c == '/' && i+1 < len(data) && data[i+1] == '/':
			for i < len(data) && data[i] != '\n' {
				i++
			}
			out = append(out, '\n')
		case c == '/' && i+1 < len(data) && data[i+1] == '*':
			i += 2
			for i+1 < len(data) && !(data[i] == '*' && data[i+1] == '/') {
				i++
			}
			i++
		case c == ',':
			j := i + 1
			for j < len(data) && strings.ContainsRune(" \t\r\n", rune(data[j])) {
				j++
			}
			if j < len(data) && (data[j] == '}' || data[j] == ']') {
				continue
			}
			out = append(out, c)
		default:
			out = append(out, c)
		}
	}
	return out
}
