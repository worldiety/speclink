package syntax

// File is one parsed module.
type File struct {
	Path   string
	Src    string
	Toks   []Token
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

// Source returns the exact source bytes of r.
func (f *File) Source(r Range) string {
	if r.Empty() {
		return ""
	}
	from := f.Toks[r.From].Pos.Offset
	to := f.Toks[r.To-1].End
	if from < 0 || to > len(f.Src) || from > to {
		return ""
	}
	return f.Src[from:to]
}

// EndLine returns the line of the last token of r.
func (f *File) EndLine(r Range) int {
	if r.Empty() {
		return 0
	}
	return f.Toks[r.To-1].Pos.Line
}

// ItemKind is the sort of a top level statement.
type ItemKind uint8

const (
	// ItemImport is an import declaration, or a re-export with a from clause.
	ItemImport ItemKind = iota + 1
	ItemClass
	ItemInterface
	ItemEnum
	ItemType
	ItemFunction
	// ItemVar is a const, let or var declaration.
	ItemVar
	// ItemExpr is an expression statement, or the expression of an export
	// default that is not a declaration.
	ItemExpr
	// ItemExportList is export { a, b as c } without a from clause.
	ItemExportList
	// ItemOther is a statement nothing here reads: a namespace, an import
	// alias, a control flow statement at the top level.
	ItemOther
)

func (k ItemKind) String() string {
	switch k {
	case ItemImport:
		return "import"
	case ItemClass:
		return "class"
	case ItemInterface:
		return "interface"
	case ItemEnum:
		return "enum"
	case ItemType:
		return "type alias"
	case ItemFunction:
		return "function"
	case ItemVar:
		return "variable"
	case ItemExpr:
		return "expression statement"
	case ItemExportList:
		return "export list"
	}
	return "statement"
}

// Item is one top level statement.
type Item struct {
	Kind ItemKind
	// Name is the declared name, empty for an anonymous default export and
	// for statements that declare nothing.
	Name     string
	Exported bool
	// Default marks export default.
	Default  bool
	Declare  bool
	Abstract bool
	// VarKind is const, let or var.
	VarKind string
	Vars    []Var
	// Range is the whole statement.
	Range Range
	// Body is the braces of a class, interface, enum or function, without
	// the braces themselves. For an expression statement it is the
	// expression.
	Body Range
	// Extends and Implements name the heritage of a class or interface, as
	// written: a dotted name without its type arguments.
	Extends    []string
	Implements []string
	Members    []*Member
	Import     *Import
	// Exports are the names of an export list.
	Exports []Named
	Pos     Pos
}

// Var is one declarator of a variable statement.
type Var struct {
	// Name is the bound identifier, empty for a destructuring pattern.
	Name string
	Init Range
	Pos  Pos
}

// Import is an import declaration or a re-export.
type Import struct {
	Spec      string
	SpecPos   Pos
	TypeOnly  bool
	Default   string
	Namespace string
	Named     []Named
	// ReExport marks export … from, which loads the module like an import
	// and binds nothing locally.
	ReExport bool
}

// Named is one name of an import or export list.
type Named struct {
	// Name is the exported name, Local the one bound here.
	Name, Local string
	TypeOnly    bool
}

// MemberKind is the sort of a class member.
type MemberKind uint8

const (
	MemberMethod MemberKind = iota + 1
	MemberProperty
	MemberConstructor
	MemberGetter
	MemberSetter
	// MemberIndex is an index signature.
	MemberIndex
	// MemberStatic is a static initialisation block.
	MemberStatic
)

// Member is one member of a class.
type Member struct {
	Kind MemberKind
	Name string
	// Access is public, private or protected as written, empty when none is.
	Access   string
	Static   bool
	Readonly bool
	Abstract bool
	Async    bool
	Params   []Param
	// Init is the initializer of a property.
	Init Range
	// Body is the braces of a method, without the braces themselves.
	Body  Range
	Range Range
	Pos   Pos
}

// Private reports whether the member is unreachable from outside the class.
func (m *Member) Private() bool {
	return m.Access == "private" || m.Access == "protected" || len(m.Name) > 0 && m.Name[0] == '#'
}

// Param is one parameter of a constructor or method.
type Param struct {
	Name string
	// Access is the modifier of a parameter property.
	Access   string
	Readonly bool
	Type     Range
	Pos      Pos
}

// GroupEnd returns the index just past the bracket group that opens at token
// i: the parenthesis of a call, the brace of a block.
func (f *File) GroupEnd(i int) int {
	p := &parser{f: f, t: f.Toks}
	return p.skipGroup(i)
}
