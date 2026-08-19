package ir

// The persisted shape of a type, in the form it takes on the wire.
//
// This is the part of a declaration that outlives the code declaring it. A
// stored message is decoded by its discriminator and read field by field, so
// those are the facts that become a promise the moment the first message is
// written; everything else about the type may change freely.

// SchemaType is the wire shape of one persisted type.
type SchemaType struct {
	// Name is the fully qualified name, e.g. "example.com/m/sales.QuoteSubmitted".
	Name string
	// Package is the import path the type was found in.
	Package string
	// Discriminator is the serialisation tag a stored message is decoded by.
	// Empty when it could not be read statically, which is itself reportable.
	Discriminator string
	// Fields are the serialised fields in declaration order.
	Fields []SchemaField
	Pos    Position
}

// Field returns the field with the given Go name.
func (t SchemaType) Field(name string) (SchemaField, bool) {
	for _, f := range t.Fields {
		if f.Name == name {
			return f, true
		}
	}
	return SchemaField{}, false
}

// SchemaField is one serialised field.
type SchemaField struct {
	// Name is the Go field name. It is what a binding refers to, and it may
	// change without breaking anything as long as Wire stays put.
	Name string
	// Wire is the name the field carries in stored data.
	Wire string
	// Shape is the underlying structure of the field type, expanded through
	// named types: a field declared as RelationID with underlying string has
	// the shape "string", because that is what a stored message contains.
	Shape string
	// Optional reports whether the field may be absent from stored data. It is
	// set for fields added after the type was promised.
	Optional bool
}
