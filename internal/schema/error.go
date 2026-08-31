package schema

import "fmt"

// Error says where in a shape the reader gave up.
//
// The position is carried because a shape is one long line and "unknown shape"
// on its own leaves a reader comparing two hundred characters by eye.
type Error struct {
	Shape string
	Pos   int
	Msg   string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s at offset %d of %q", e.Msg, e.Pos, e.Shape)
}

func errAt(shape string, pos int, msg string) error {
	return &Error{Shape: shape, Pos: pos, Msg: msg}
}
