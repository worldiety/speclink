package main

import (
	"strings"
	"testing"

	"github.com/worldiety/speclink/internal/ir"
)

func TestCoverProfileIsParsed(t *testing.T) {
	blocks := readCoverProfile(strings.NewReader(`mode: set
example.com/m/app/sales/uc_submit_quote.go:11.63,13.35 2 1
example.com/m/app/sales/uc_submit_quote.go:14.2,14.20 1 0
malformed
example.com/m/app/sales/model.go:5.1,5.2 1 3
`))

	if len(blocks) != 3 {
		t.Fatalf("got %d blocks, want 3 (the malformed line is skipped)", len(blocks))
	}
	if blocks[0].statements != 2 || !blocks[0].covered {
		t.Errorf("first block wrong: %+v", blocks[0])
	}
	if blocks[1].covered {
		t.Error("a block with count 0 was read as covered")
	}
}

// A block counts when it lies wholly inside the declaration. One that straddles
// the end belongs to whatever follows, and splitting the difference would
// produce a figure that is precise about something untrue.
func TestCoverageIsAttributedByExtent(t *testing.T) {
	c := ir.Construct{
		Name:    "example.com/m/app/sales.SubmitQuote",
		Package: "example.com/m/app/sales",
		Pos:     ir.Position{File: "/abs/path/uc_submit_quote.go", Line: 10},
		EndLine: 20,
	}
	blocks := []coverBlock{
		{file: "example.com/m/app/sales/uc_submit_quote.go", start: 11, end: 13, statements: 2, covered: true},
		{file: "example.com/m/app/sales/uc_submit_quote.go", start: 14, end: 15, statements: 3},
		// Outside the declaration.
		{file: "example.com/m/app/sales/uc_submit_quote.go", start: 21, end: 25, statements: 9, covered: true},
		// Straddles the end.
		{file: "example.com/m/app/sales/uc_submit_quote.go", start: 19, end: 24, statements: 4, covered: true},
		// Another file of the same package.
		{file: "example.com/m/app/sales/model.go", start: 11, end: 12, statements: 7, covered: true},
	}

	statements, covered := coverageOf(c, blocks)
	if statements != 5 || covered != 2 {
		t.Errorf("got %d of %d, want 2 of 5", covered, statements)
	}
}

// A declaration with no recorded extent is not measured, and must not come out
// as fully covered or fully uncovered.
func TestUnmeasuredDeclarationYieldsNothing(t *testing.T) {
	c := ir.Construct{Name: "x", Package: "p", Pos: ir.Position{File: "/a/b.go", Line: 1}}
	if statements, covered := coverageOf(c, []coverBlock{{file: "p/b.go", start: 1, end: 2, statements: 3}}); statements != 0 || covered != 0 {
		t.Errorf("got %d of %d, want nothing", covered, statements)
	}
}
