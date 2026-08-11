package spec

import (
	"reflect"
	"runtime"
)

// Binding is the opaque result of a binding term. It carries no information and
// exists so that bindings can be written as package level `var _ = …`.
type Binding struct{}

// The four functions below are the entire side effect surface of the language.
// Each records its term in the runtime registry and captures its own source
// position via runtime.Caller.
//
// Capturing the position is not incidental. It carries two things:
//
//  1. It makes the cross check (speclink selfreport) position based, so the
//     static and the runtime view of the model can be compared as sets.
//  2. It works around a loss that cannot otherwise be avoided: ForDecl receives
//     a value, and the identifier that named it is gone at runtime. Keyed by
//     position the entry is still unambiguous.
//
// Comparison must be over sets keyed by position, never over sequences: Go
// initialises package level variables in dependency order, which is
// deterministic but not source order.

// For binds to a named type — use case func type, event struct, aggregate.
// The common case; the target is checked by the Go compiler.
func For[T any](as ...Assertion) Binding {
	return record(target{kind: targetType, typ: reflect.TypeFor[T]()}, as)
}

// ForDecl binds to a declared function, variable or constant.
//
// The argument names the declaration. Which kind it is — func, var or const —
// follows from the type checker and is not asserted by the author, so the two
// cannot disagree.
//
// Generic rather than any, so the value is not boxed at initialisation time.
// The value itself is never read; it exists only so the Go compiler verifies
// that the declaration exists.
func ForDecl[T any](ref T, as ...Assertion) Binding {
	return record(target{kind: targetDecl}, as)
}

// ForField binds to a struct field of T.
//
// The field name is a string and thus the only reference in the whole language
// that the Go compiler does not check. speclink checks it against the type.
func ForField[T any](field string, as ...Assertion) Binding {
	return record(target{kind: targetField, typ: reflect.TypeFor[T](), field: field}, as)
}

// ForPackage binds to the package of the neighbouring source file.
func ForPackage(as ...Assertion) Binding {
	return record(target{kind: targetPackage}, as)
}

type targetKind int

const (
	targetType targetKind = iota + 1
	targetDecl
	targetField
	targetPackage
)

func (k targetKind) String() string {
	switch k {
	case targetType:
		return "type"
	case targetDecl:
		return "decl"
	case targetField:
		return "field"
	case targetPackage:
		return "package"
	}
	return "unknown"
}

type target struct {
	kind  targetKind
	typ   reflect.Type
	field string
}

// record appends the term to the registry. skip is fixed at 2: record itself
// and the exported binding function that called it.
//
// The runtime reports one kind "decl" for functions, variables and constants,
// because reflection cannot tell a constant from a variable of the same type.
// The static reader refines this into func, var and const from the type
// checker. The cross check therefore compares positions and assertions; the
// target kind is a static only enrichment and must not be treated as a
// mismatch.
func record(t target, as []Assertion) Binding {
	_, file, line, _ := runtime.Caller(2)
	appendEntry(entryOf(t, as, file, line))
	return Binding{}
}

func (t target) name() string {
	switch t.kind {
	case targetType:
		return typeName(t.typ)
	case targetField:
		return typeName(t.typ) + "." + t.field
	}
	// A declaration cannot name itself at runtime: the identifier is gone by
	// the time the value arrives. The recorded position identifies the entry.
	return ""
}

func typeName(t reflect.Type) string {
	if t == nil {
		return ""
	}
	if t.PkgPath() != "" && t.Name() != "" {
		return t.PkgPath() + "." + t.Name()
	}
	return t.String()
}
