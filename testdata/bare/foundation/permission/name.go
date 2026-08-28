package permission

import "reflect"

// nameOf renders the use case a permission guards, for the listing only.
//
// Reflection is used here and nowhere else, and only on a type parameter that
// the compiler has already resolved. Nothing decides anything on it.
func nameOf[T any]() string { return reflect.TypeFor[T]().String() }
