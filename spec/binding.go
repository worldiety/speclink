package spec

import (
	"reflect"
	"runtime"
)

// Binding is the opaque result of a binding term. It carries no information and
// exists so that bindings can be written as package level `var _ = …`.
type Binding struct{}

// The five functions below are the entire side effect surface of the language.
// Each records its term in the runtime registry and captures its own source
// position via runtime.Caller.
//
// Capturing the position is not incidental. It carries two things:
//
//  1. It makes the cross check (speclink selfreport) position based, so the
//     static and the runtime view of the model can be compared as sets.
//  2. It works around a loss that cannot otherwise be avoided: ForVar receives
//     a pointer, and the variable name is gone at runtime. Keyed by position
//     the entry is still unambiguous.
//
// Comparison must be over sets keyed by position, never over sequences: Go
// initialises package level variables in dependency order, which is
// deterministic but not source order.

// For binds to a named type — use case func type, event struct, aggregate.
// The common case; the target is checked by the Go compiler.
func For[T any](as ...Assertion) Binding {
	return record(target{kind: targetType, typ: reflect.TypeFor[T]()}, as)
}

// ForFunc binds to a function. The argument is the function value itself.
func ForFunc(fn any, as ...Assertion) Binding {
	return record(target{kind: targetFunc, value: fn}, as)
}

// ForVar binds to a variable or constant. The argument is its address.
//
// The variable name is not recoverable at runtime; the recorded position
// identifies the entry instead.
func ForVar(ptr any, as ...Assertion) Binding {
	return record(target{kind: targetVar, value: ptr}, as)
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
	targetFunc
	targetVar
	targetField
	targetPackage
)

func (k targetKind) String() string {
	switch k {
	case targetType:
		return "type"
	case targetFunc:
		return "func"
	case targetVar:
		return "var"
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
	value any
	field string
}

// record appends the term to the registry. skip is fixed at 2: record itself
// and the exported binding function that called it.
func record(t target, as []Assertion) Binding {
	_, file, line, _ := runtime.Caller(2)
	appendEntry(entryOf(t, as, file, line))
	return Binding{}
}

func (t target) name() string {
	switch t.kind {
	case targetType:
		return typeName(t.typ)
	case targetFunc:
		if t.value == nil {
			return ""
		}
		pc := reflect.ValueOf(t.value).Pointer()
		if fn := runtime.FuncForPC(pc); fn != nil {
			return fn.Name()
		}
		return ""
	case targetVar:
		if t.value == nil {
			return ""
		}
		return typeName(reflect.TypeOf(t.value))
	case targetField:
		return typeName(t.typ) + "." + t.field
	}
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
