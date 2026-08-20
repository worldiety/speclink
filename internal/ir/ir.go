// Package ir holds the language neutral intermediate model.
//
// The ir is not an external interface. It has no serialisation format, no
// version and needs no round trip tests: one binary, one process, no boundary.
//
// Its only purpose is type isolation:
//
//	go/types.Type, token.Pos and ast.Node must never reach
//	internal/check, internal/diag or internal/backend.
//
// Without that boundary the rules and backends would depend directly on Go type
// information, and a second language frontend could not be added without
// rewriting them. The boundary costs almost nothing; it only decides where the
// data types live.
//
// A direct consequence: positions here are File/Line/Col, not token.Pos, which
// is meaningless without a FileSet and therefore language bound.
package ir

import "fmt"

// Position is a source location. Column is 1 based, 0 when unknown.
type Position struct {
	File string
	Line int
	Col  int
}

func (p Position) String() string {
	if p.File == "" {
		return "<unknown>"
	}
	if p.Col > 0 {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Col)
	}
	return fmt.Sprintf("%s:%d", p.File, p.Line)
}

// TargetKind is the sort of construct a binding attaches to.
//
// Func, Var and Const are not chosen by the author. spec.ForDecl names a
// declaration and the type checker decides which of the three it is, so the
// annotation cannot disagree with the code.
type TargetKind int

const (
	TargetType TargetKind = iota + 1
	TargetFunc
	TargetVar
	TargetConst
	TargetField
	TargetPackage
)

func (k TargetKind) String() string {
	switch k {
	case TargetType:
		return "type"
	case TargetFunc:
		return "func"
	case TargetVar:
		return "var"
	case TargetConst:
		return "constant"
	case TargetField:
		return "field"
	case TargetPackage:
		return "package"
	}
	return "unknown"
}

// Target names the construct a binding attaches to.
//
// Name is fully qualified, e.g. "example.com/m/sales.SubmitQuoteUC". Field is
// set for TargetField only.
type Target struct {
	Kind    TargetKind
	Package string
	Name    string
	Field   string
}

func (t Target) String() string {
	if t.Kind == TargetField {
		return t.Name + "." + t.Field
	}
	if t.Name == "" {
		return t.Package
	}
	return t.Name
}

// AssertionKind discriminates the payload of an [Assertion].
type AssertionKind int

const (
	AssertSatisfies AssertionKind = iota + 1
	AssertTransition
	AssertExternal
	AssertHelp
	AssertTerm
	AssertRationale
	AssertWaive
	AssertDraft
	AssertOptional
)

func (k AssertionKind) String() string {
	switch k {
	case AssertSatisfies:
		return "satisfies"
	case AssertTransition:
		return "transition"
	case AssertExternal:
		return "external"
	case AssertHelp:
		return "help"
	case AssertTerm:
		return "term"
	case AssertRationale:
		return "rationale"
	case AssertWaive:
		return "waive"
	case AssertDraft:
		return "draft"
	case AssertOptional:
		return "optional"
	}
	return "unknown"
}

// Assertion is one statement about the construct named by its binding.
// Only the fields relevant for Kind are populated.
type Assertion struct {
	Kind AssertionKind
	Pos  Position

	// Requirements holds the requirement IDs of an AssertSatisfies.
	Requirements []string
	// EventType is the fully qualified event type of an AssertTransition.
	EventType string
	// State is the target lifecycle state of an AssertTransition.
	State string
	// Text carries help, rationale or waiver reason.
	Text string
	// Term is the glossary ID of an AssertTerm.
	Term string
	// Rule is the suspended rule ID of an AssertWaive.
	Rule string
}

// Binding attaches a set of assertions to one construct.
type Binding struct {
	Target     Target
	Assertions []Assertion
	Pos        Position
}
