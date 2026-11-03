package ts

import (
	"path"

	"github.com/worldiety/speclink/internal/lang/ts/syntax"
)

// VocabularyFile is the name of the module that declares speclink's
// vocabulary.
//
// There is no package to install. A project writes the forty lines itself —
// requirement, bind, satisfies, waive and the rest — and speclink recognises a
// call by the file it was imported from, exactly as the JVM frontend
// recognises an annotation by its qualified name. One binary serves projects
// that are not all on the same version of anything, so it must not be part of
// what they depend on.
const VocabularyFile = "speclink.ts"

// isVocabulary reports whether a module is the project's speclink vocabulary.
func isVocabulary(m *Module) bool { return m != nil && path.Base(m.Path) == VocabularyFile }

// binding is what a local name of a module refers to through an import.
type binding struct {
	ref *Ref
	// name is the exported name, "default" for a default import and "*" for
	// a namespace.
	name string
}

// bindings maps every imported local name of a module to its import.
func (m *Module) bindings() map[string]binding {
	if m.binds != nil {
		return m.binds
	}
	out := map[string]binding{}
	m.binds = out
	for _, ref := range m.Imports {
		imp := ref.Import
		if imp.ReExport {
			continue
		}
		if imp.Default != "" {
			out[imp.Default] = binding{ref, "default"}
		}
		if imp.Namespace != "" {
			out[imp.Namespace] = binding{ref, "*"}
		}
		for _, n := range imp.Named {
			out[n.Local] = binding{ref, n.Name}
		}
	}
	return out
}

// vocabularyCall returns the speclink function a callee names, "" when it
// names something else.
//
// Decided by where the name was imported from, never by its spelling: a
// project function that happens to be called bind is not the annotation
// language, and an aliased import of the real one still is.
func (m *Module) vocabularyCall(callee []string) string {
	if len(callee) == 0 {
		return ""
	}
	b, ok := m.bindings()[callee[0]]
	if !ok || !isVocabulary(b.ref.Target) {
		return ""
	}
	switch {
	case b.name == "*" && len(callee) == 2:
		return callee[1]
	case b.name != "*" && b.name != "default" && len(callee) == 1:
		return b.name
	}
	return ""
}

// vocabularyName is vocabularyCall for a dotted name that is not called, an
// enum member such as Kind.Functional. It returns the enum and the member.
func (m *Module) vocabularyName(p []string) (string, string) {
	if len(p) < 2 {
		return "", ""
	}
	b, ok := m.bindings()[p[0]]
	if !ok || !isVocabulary(b.ref.Target) {
		return "", ""
	}
	switch {
	case b.name == "*" && len(p) == 3:
		return p[1], p[2]
	case b.name != "*" && len(p) == 2:
		return b.name, p[1]
	}
	return "", ""
}

// declared reports whether the module declares a name at its top level.
func (m *Module) declared(name string) bool {
	for _, it := range m.Items() {
		if it.Name == name && it.Import == nil {
			return true
		}
		for _, v := range it.Vars {
			if v.Name == name {
				return true
			}
		}
	}
	return false
}

// defaultExport returns the name of the default export, "" when there is none
// or it is anonymous.
func (m *Module) defaultExport() string {
	for _, it := range m.Items() {
		if !it.Default {
			continue
		}
		if it.Name != "" {
			return it.Name
		}
		toks := it.File.Tokens(it.Body)
		if len(toks) == 1 && toks[0].Kind == syntax.Ident {
			return toks[0].Text
		}
	}
	return ""
}

// symbol qualifies a name as written in a module: the module that declares
// it, a dot, and the name it is declared under. It is what a requirement's
// Symbol and a construct's Name are spelled in, so that a reference and a
// declaration meet on one string.
func (p *Project) symbol(m *Module, name []string) (string, bool) {
	if len(name) == 0 {
		return "", false
	}
	if len(name) == 1 && m.declared(name[0]) {
		return m.ID + "." + name[0], true
	}
	b, ok := m.bindings()[name[0]]
	if !ok || b.ref.Target == nil {
		return "", false
	}
	target := b.ref.Target
	switch {
	case b.name == "*" && len(name) == 2:
		return p.symbol(target, name[1:])
	case b.name == "default" && len(name) == 1:
		if d := target.defaultExport(); d != "" {
			return p.symbol(target, []string{d})
		}
		return "", false
	case len(name) == 1:
		// A re-export would lead further, and is not followed: a requirement
		// or a use case is named where it is declared.
		return target.ID + "." + b.name, target.declared(b.name)
	}
	return "", false
}
