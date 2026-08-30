package ir

// EntryPoint is a program this module builds.
//
// A module routinely produces several — a service, a command line tool, a
// migration — and which of them a statement is about changes the answer to
// almost every question in this document. Until now the fact that there were
// several was visible only to the rule that insists they live in one place.
//
// Everything here is read from the source, and some of it is inferred rather
// than declared. What a binary is called and where it lives are facts. Which
// contexts it assembles is read from its imports and is a fact about the
// import graph, which is very nearly the same thing. How it is invoked is a
// guess, and is marked as one.
type EntryPoint struct {
	// Name is the binary, which for Go is the directory it lives in.
	Name string
	// Package is the import path.
	Package string
	// Dir is the path relative to the module root.
	Dir string
	// Doc is the package comment, which by Go convention describes the
	// command. Empty when nobody wrote one.
	Doc string
	// Contexts are the bounded contexts this program assembles.
	Contexts []string
	// Adapters are the concrete technologies it chooses. This is the list that
	// makes a deployment reviewable: it is the only place the choice is made,
	// and the only place it can be read off.
	Adapters []string
	// Verbs are subcommands the program appears to accept.
	//
	// Inferred from comparisons against the argument vector, so it is a guess
	// and the document says so. A program that dispatches through a table this
	// does not recognise will show none, which must never be reported as a
	// program that takes no subcommand.
	Verbs []string
	// Flags are the command line flags it appears to register, same caveat.
	Flags []string
	// Guessed records that Verbs and Flags come from inference rather than a
	// declaration, so a reader is told which half of this entry is solid.
	Guessed bool
	Pos     Position
}
