package golang

import (
	"go/ast"
	"go/types"
	"sort"
	"strconv"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// Schema reading: the persisted shape of a type, in the form it takes on the
// wire rather than the form it takes in the source.
//
// The distinction is the whole point. A field declared as a named type with
// underlying string is a string on the wire, so renaming string to RelationID
// is invisible to a reader of stored data and must stay invisible to the check.
// The shape is therefore recorded as the underlying structure, and a rename
// that preserves it produces an identical fingerprint — no comparison rule
// needed, because there is nothing to compare.

// ReadSchema returns the persisted shape of every type of this package whose
// form outlives the code.
//
// Two things qualify. An event is self evident: it implements Evolve and
// carries a discriminator, so the struct is the wire format by definition.
// A persistence model is not, because nothing in its declaration says it is
// stored — that is decided where a repository is built over it, possibly in
// another package, so the names are collected beforehand and passed in.
//
// Only the wire relevant facts are collected: the discriminator, and per field
// its Go name, the name it carries in JSON and its underlying shape. Everything
// else about the type may change freely.
func (p *Package) ReadSchema(models map[string]bool) []ir.SchemaType {
	var out []ir.SchemaType

	for _, f := range p.pkg.Syntax {
		if p.isGeneratedByUs(f) {
			continue
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			for _, s := range gd.Specs {
				ts, ok := s.(*ast.TypeSpec)
				if !ok {
					continue
				}
				obj := p.pkg.TypesInfo.Defs[ts.Name]
				if obj == nil {
					continue
				}
				named, ok := obj.Type().(*types.Named)
				if !ok {
					continue
				}
				st, ok := named.Underlying().(*types.Struct)
				if !ok {
					continue
				}

				name := p.PkgPath() + "." + ts.Name.Name
				event := p.hasMethods(named, "Evolve", "Discriminator")
				if !event && !models[name] {
					continue
				}

				entry := ir.SchemaType{
					Name:    name,
					Package: p.PkgPath(),
					Fields:  p.readFields(st),
					Pos:     p.pos(ts.Pos()),
				}
				// Only an event is decoded by a tag. A persistence model is
				// found by its key, so it has none and none is expected.
				if event {
					entry.Discriminator = p.discriminatorOf(ts.Name.Name)
				}
				out = append(out, entry)
			}
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// discriminatorOf finds the constant the Discriminator method returns.
//
// Only a plain literal return is recognised. A computed discriminator would be
// a fact the tool cannot see, and a persisted identity that cannot be read
// statically is one that cannot be protected — so leaving it empty is the
// honest answer, and the comparison reports it.
func (p *Package) discriminatorOf(typeName string) string {
	for _, f := range p.pkg.Syntax {
		for _, decl := range f.Decls {
			fd, ok := decl.(*ast.FuncDecl)
			if !ok || fd.Name.Name != "Discriminator" || fd.Recv == nil || fd.Body == nil {
				continue
			}
			if receiverName(fd) != typeName {
				continue
			}
			if lit, ok := singleReturnString(fd.Body); ok {
				return lit
			}
		}
	}
	return ""
}

// receiverName returns the base type name of a method receiver, with any
// pointer stripped.
func receiverName(fd *ast.FuncDecl) string {
	if fd.Recv == nil || len(fd.Recv.List) == 0 {
		return ""
	}
	expr := fd.Recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

// singleReturnString reports the string literal of a body that does nothing but
// return one.
func singleReturnString(body *ast.BlockStmt) (string, bool) {
	if len(body.List) != 1 {
		return "", false
	}
	ret, ok := body.List[0].(*ast.ReturnStmt)
	if !ok || len(ret.Results) != 1 {
		return "", false
	}
	expr := ret.Results[0]
	// A conversion such as evs.Discriminator("x") wraps the literal.
	if call, ok := expr.(*ast.CallExpr); ok && len(call.Args) == 1 {
		expr = call.Args[0]
	}
	lit, ok := expr.(*ast.BasicLit)
	if !ok {
		return "", false
	}
	s, err := unquote(lit.Value)
	if err != nil {
		return "", false
	}
	return s, true
}

// readFields collects the exported fields of a struct in declaration order.
//
// Unexported fields never reach the wire, so they are not part of the promise
// and may change freely.
func (p *Package) readFields(st *types.Struct) []ir.SchemaField {
	var out []ir.SchemaField
	for i := 0; i < st.NumFields(); i++ {
		f := st.Field(i)
		if !f.Exported() {
			continue
		}
		wire, skipped := wireName(f, st.Tag(i))
		if skipped {
			continue
		}
		out = append(out, ir.SchemaField{
			Name:      f.Name(),
			Wire:      wire,
			Shape:     shapeOf(f.Type(), map[*types.Named]bool{}),
			Type:      declaredName(f.Type()),
			Doc:       p.docs.of(f.Pos()),
			OmitEmpty: hasTagOption(st.Tag(i), "omitempty"),
			Pos:       p.pos(f.Pos()),
		})
	}
	return out
}

// wireName resolves the JSON name of a field, and reports whether the field is
// excluded from serialisation altogether.
func wireName(f *types.Var, tag string) (name string, skipped bool) {
	value := reflectTag(tag, "json")
	if value == "" {
		return f.Name(), false
	}
	parts := strings.Split(value, ",")
	if parts[0] == "-" && len(parts) == 1 {
		return "", true
	}
	if parts[0] == "" {
		return f.Name(), false
	}
	return parts[0], false
}

// declaredName is the qualified name of a named type, empty for anything else.
//
// A pointer and a slice are unwrapped, because a rule about the values of a
// type holds however many of them travel together and whether or not one may
// be absent.
func declaredName(t types.Type) string {
	switch u := t.(type) {
	case *types.Pointer:
		return declaredName(u.Elem())
	case *types.Slice:
		return declaredName(u.Elem())
	case *types.Array:
		return declaredName(u.Elem())
	case *types.Named:
		return typeName(u)
	case *types.Alias:
		return declaredName(types.Unalias(u))
	}
	return ""
}

// hasTagOption reports whether the json tag carries an option such as
// omitempty. Read separately from the name so that the name resolution stays
// the one thing a rule compares.
func hasTagOption(tag, option string) bool {
	parts := strings.Split(reflectTag(tag, "json"), ",")
	for _, p := range parts[min(1, len(parts)):] {
		if strings.TrimSpace(p) == option {
			return true
		}
	}
	return false
}

// reflectTag reads one key out of a struct tag without pulling in reflect.
func reflectTag(tag, key string) string {
	for tag != "" {
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := tag[:i]
		tag = tag[i+1:]

		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		quoted := tag[:i+1]
		tag = tag[i+1:]

		if name == key {
			value, err := strconv.Unquote(quoted)
			if err != nil {
				return ""
			}
			return value
		}
	}
	return ""
}

// basicShape renders a basic type, collapsing every integer width into one
// token.
//
// A change of width is not a change of shape: on the wire an integer is a JSON
// number either way, and the promise made about such a field is that it holds a
// whole number, not that it holds sixty-four bits of one. Recording the class
// rather than the width means int64 and int32 produce the same fingerprint and
// no comparison rule is needed at all — the same trick as expanding named
// types, and for the same reason: a difference that cannot be observed in
// stored data should not be representable in the record.
//
// Floating point is deliberately not collapsed. Narrowing a float64 to a
// float32 loses precision silently, which is exactly the kind of change that
// deserves to be stopped.
func basicShape(b *types.Basic) string {
	if b.Info()&types.IsInteger != 0 {
		return "int"
	}
	return b.Name()
}

// shapeOf renders the underlying structure of a type as a stable fingerprint.
//
// Named types are expanded to what they are made of, because that is what a
// stored message contains. seen guards against a type that reaches itself,
// which is legal through a pointer or a slice.
func shapeOf(t types.Type, seen map[*types.Named]bool) string {
	switch u := t.(type) {
	case *types.Alias:
		return shapeOf(types.Unalias(u), seen)

	case *types.Named:
		if seen[u] {
			return "..."
		}
		seen[u] = true
		defer delete(seen, u)
		return shapeOf(u.Underlying(), seen)

	case *types.Basic:
		return basicShape(u)

	case *types.Pointer:
		// A pointer is absence on the wire, not a different shape.
		return shapeOf(u.Elem(), seen)

	case *types.Slice:
		return "[]" + shapeOf(u.Elem(), seen)

	case *types.Array:
		return "[]" + shapeOf(u.Elem(), seen)

	case *types.Map:
		return "map[" + shapeOf(u.Key(), seen) + "]" + shapeOf(u.Elem(), seen)

	case *types.Struct:
		var parts []string
		for i := 0; i < u.NumFields(); i++ {
			f := u.Field(i)
			if !f.Exported() {
				continue
			}
			wire, skipped := wireName(f, u.Tag(i))
			if skipped {
				continue
			}
			parts = append(parts, wire+":"+shapeOf(f.Type(), seen))
		}
		return "{" + strings.Join(parts, ",") + "}"

	case *types.Interface:
		// An interface carries no shape of its own; whatever is stored through
		// it is decided elsewhere and cannot be promised here.
		return "any"
	}
	return "unknown"
}
