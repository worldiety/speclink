package syntax

// ValueKind is the sort of a literal value.
type ValueKind uint8

const (
	// ValOther is anything that is not a literal: an expression that would
	// have to be evaluated to be known.
	ValOther ValueKind = iota
	ValString
	ValNumber
	ValBool
	// ValIdent is a bare identifier; Path holds it.
	ValIdent
	// ValMember is a dotted name, Kind.Functional; Path holds its parts.
	ValMember
	ValArray
	ValObject
	// ValCall is a call of a dotted name; Path is the callee, Elems the
	// arguments.
	ValCall
)

// Value is a literal as written, read without evaluating anything.
//
// It is the closed subset speclink's declarations are written in. A value
// outside it is not an error here — it is ValOther with a position — so that
// the caller can say which declaration needed a literal and did not get one.
type Value struct {
	Kind   ValueKind
	Str    string
	Path   []string
	Elems  []Value
	Fields []Field
	Pos    Pos
	Range  Range
}

// Field is one property of an object literal.
type Field struct {
	Key   string
	Value Value
	Pos   Pos
}

// Get returns the field of an object literal by key.
func (v Value) Get(key string) (Value, bool) {
	for _, f := range v.Fields {
		if f.Key == key {
			return f.Value, true
		}
	}
	return Value{}, false
}

// Last returns the last part of a dotted name, the whole of an identifier.
func (v Value) Last() string {
	if len(v.Path) == 0 {
		return ""
	}
	return v.Path[len(v.Path)-1]
}

// Value reads the literal in r.
func (f *File) Value(r Range) Value {
	if r.Empty() {
		return Value{Kind: ValOther, Range: r}
	}
	vp := &valueParser{f: f, p: &parser{f: f, t: f.Toks}, end: r.To}
	v, next := vp.expr(r.From)
	// A trailing assertion changes the type and not the value.
	for next < r.To && (f.Toks[next].Is("as") || f.Toks[next].Is("satisfies")) {
		next = vp.p.skipType(next + 1)
	}
	if next < r.To {
		return Value{Kind: ValOther, Pos: f.Toks[r.From].Pos, Range: r}
	}
	return v
}

type valueParser struct {
	f   *File
	p   *parser
	end int
}

func (vp *valueParser) other(from, to int) (Value, int) {
	return Value{Kind: ValOther, Pos: vp.p.at(from).Pos, Range: Range{from, to}}, to
}

// expr reads a primary value and any string concatenation after it.
func (vp *valueParser) expr(i int) (Value, int) {
	v, j := vp.primary(i)
	for j < vp.end && vp.p.at(j).Is("+") {
		w, k := vp.primary(j + 1)
		if v.Kind != ValString || w.Kind != ValString {
			return vp.other(i, vp.skipToEnd(k))
		}
		v.Str += w.Str
		v.Range.To = k
		j = k
	}
	return v, j
}

// skipToEnd steps to the next comma or the end of the enclosing group.
func (vp *valueParser) skipToEnd(j int) int {
	for j < vp.end {
		t := vp.p.at(j)
		if t.Is(",") || isCloser(t) {
			return j
		}
		if isOpener(t) {
			j = vp.p.skipGroup(j)
			continue
		}
		j++
	}
	return j
}

func (vp *valueParser) primary(i int) (Value, int) {
	t := vp.p.at(i)
	switch {
	case t.Kind == String || t.Kind == Template:
		return Value{Kind: ValString, Str: t.Text, Pos: t.Pos, Range: Range{i, i + 1}}, i + 1
	case t.Kind == Number:
		return Value{Kind: ValNumber, Str: t.Text, Pos: t.Pos, Range: Range{i, i + 1}}, i + 1
	case t.Is("true") || t.Is("false"):
		return Value{Kind: ValBool, Str: t.Text, Pos: t.Pos, Range: Range{i, i + 1}}, i + 1
	case t.Is("["):
		return vp.array(i)
	case t.Is("{"):
		return vp.object(i)
	case t.Kind == Ident:
		path := []string{t.Text}
		j := i + 1
		for vp.p.at(j).Is(".") && vp.p.at(j+1).Kind == Ident {
			path = append(path, vp.p.at(j+1).Text)
			j += 2
		}
		if vp.p.at(j).Is("<") {
			// Type arguments of a call.
			if k := vp.p.skipAngles(j); vp.p.at(k).Is("(") {
				j = k
			}
		}
		if vp.p.at(j).Is("(") {
			end := vp.p.skipGroup(j)
			args := vp.list(j+1, end-1)
			return Value{Kind: ValCall, Path: path, Elems: args, Pos: t.Pos, Range: Range{i, end}}, end
		}
		kind := ValMember
		if len(path) == 1 {
			kind = ValIdent
		}
		return Value{Kind: kind, Path: path, Pos: t.Pos, Range: Range{i, j}}, j
	}
	end := vp.skipToEnd(i)
	if end <= i {
		end = i + 1
	}
	return vp.other(i, end)
}

// list reads comma separated values between from and to.
func (vp *valueParser) list(from, to int) []Value {
	saved := vp.end
	vp.end = to
	defer func() { vp.end = saved }()

	var out []Value
	j := from
	for j < to {
		if vp.p.at(j).Is(",") {
			j++
			continue
		}
		v, k := vp.expr(j)
		if k < to && !vp.p.at(k).Is(",") {
			v, k = vp.other(j, vp.skipToEnd(k))
		}
		out = append(out, v)
		if k <= j {
			k = j + 1
		}
		j = k
	}
	return out
}

func (vp *valueParser) array(i int) (Value, int) {
	end := vp.p.skipGroup(i)
	return Value{Kind: ValArray, Elems: vp.list(i+1, end-1), Pos: vp.p.at(i).Pos, Range: Range{i, end}}, end
}

func (vp *valueParser) object(i int) (Value, int) {
	end := vp.p.skipGroup(i)
	v := Value{Kind: ValObject, Pos: vp.p.at(i).Pos, Range: Range{i, end}}

	saved := vp.end
	vp.end = end - 1
	defer func() { vp.end = saved }()

	j := i + 1
	for j < end-1 {
		t := vp.p.at(j)
		if t.Is(",") {
			j++
			continue
		}
		if (t.Kind == Ident || t.Kind == String || t.Kind == Number) && vp.p.at(j+1).Is(":") {
			val, k := vp.expr(j + 2)
			if k < end-1 && !vp.p.at(k).Is(",") {
				val, k = vp.other(j+2, vp.skipToEnd(k))
			}
			v.Fields = append(v.Fields, Field{Key: t.Text, Value: val, Pos: t.Pos})
			j = k
			continue
		}
		if t.Kind == Ident && (vp.p.at(j+1).Is(",") || j+1 == end-1) {
			// Shorthand: { derivedFrom } stands for { derivedFrom: derivedFrom }.
			v.Fields = append(v.Fields, Field{Key: t.Text, Pos: t.Pos,
				Value: Value{Kind: ValIdent, Path: []string{t.Text}, Pos: t.Pos, Range: Range{j, j + 1}}})
			j++
			continue
		}
		// A spread, a method or a computed key: not a literal.
		k := vp.skipToEnd(j + 1)
		v.Fields = append(v.Fields, Field{Key: "", Pos: t.Pos, Value: Value{Kind: ValOther, Pos: t.Pos, Range: Range{j, k}}})
		if k <= j {
			k = j + 1
		}
		j = k
	}
	return v, end
}
