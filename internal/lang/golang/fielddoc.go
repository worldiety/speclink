package golang

import (
	"go/ast"
	"go/token"
	"strings"
)

// fieldDocs maps the position of a struct field onto the comment above it.
//
// # Why a comment and not a declaration
//
// What a field means is prose, and prose about a field belongs beside the
// field. Every alternative puts it somewhere else: a tag turns documentation
// into a string literal nobody formats, a separate schema file is the same
// fact in two places, and a spec.Help binding would be a second thing to keep
// in step with the first. The doc comment is already where a Go programmer
// writes it and already what every editor shows on hover.
//
// # Why the index is module wide
//
// A message payload is declared in the package that owns the type, and the
// channel carrying it is declared somewhere else entirely. The package reading
// the channel therefore has the type and not the syntax that produced it, so
// the comments are collected once across everything the run loaded and looked
// up by position.
type fieldDocs struct {
	byPos map[token.Pos]string
}

// collectFieldDocs indexes every struct field comment in the loaded packages.
//
// Fields of a package the run did not load carry no comment, which is a gap
// that reports itself: the schema simply says nothing about that field rather
// than saying something invented.
func collectFieldDocs(pkgs []*Package) *fieldDocs {
	d := &fieldDocs{byPos: map[token.Pos]string{}}
	for _, p := range pkgs {
		for _, f := range p.pkg.Syntax {
			ast.Inspect(f, func(n ast.Node) bool {
				st, ok := n.(*ast.StructType)
				if !ok || st.Fields == nil {
					return true
				}
				for _, field := range st.Fields.List {
					text := docText(field)
					if text == "" {
						continue
					}
					for _, name := range field.Names {
						d.byPos[name.Pos()] = text
					}
					if len(field.Names) == 0 {
						// An embedded field has no name of its own; its
						// position is the type.
						d.byPos[field.Type.Pos()] = text
					}
				}
				return true
			})
		}
	}
	return d
}

// docText is the comment belonging to one field, preferring the block above it
// over the one trailing it on the same line.
//
// Both are read because both are used: a sentence goes above, a word goes
// beside. Taking only one of them would leave half the fields of a real struct
// undocumented while their authors could see the words right there.
func docText(f *ast.Field) string {
	if f.Doc != nil {
		return commentText(f.Doc)
	}
	if f.Comment != nil {
		return commentText(f.Comment)
	}
	return ""
}

// commentText folds a comment group into one line.
//
// The line breaks of a comment are a fact about the width of an editor, and
// carrying them into a JSON string or a table cell would put that fact in
// front of a reader who cannot act on it.
func commentText(g *ast.CommentGroup) string {
	var parts []string
	for _, line := range strings.Split(strings.TrimSpace(g.Text()), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts = append(parts, line)
	}
	return strings.Join(parts, " ")
}

// of returns the comment for a field, empty where none was found.
func (d *fieldDocs) of(pos token.Pos) string {
	if d == nil {
		return ""
	}
	return d.byPos[pos]
}
