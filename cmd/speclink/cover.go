package main

import (
	"bufio"
	"io"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/worldiety/speclink/internal/ir"
)

// coverBlock is one entry of a Go coverage profile.
type coverBlock struct {
	file       string
	start, end int
	statements int
	covered    bool
}

// readCoverProfile parses the profile `go test -coverprofile` writes.
//
// The format is one block per line after the mode:
//
//	example.com/m/app/sales/uc_submit_quote.go:11.63,13.35 2 1
//
// file:startLine.startCol,endLine.endCol statements count. Only the lines and
// the statement count matter here — a column would let this attribute a block
// to a declaration more finely than a declaration is worth attributing.
//
// A malformed line is skipped rather than fatal. The profile is produced by the
// go tool and read here as a courtesy; refusing to record any evidence at all
// because one line of it was odd would be the wrong trade.
func readCoverProfile(r io.Reader) []coverBlock {
	var blocks []coverBlock

	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}

		// Split off the two trailing numbers first: a file name may contain a
		// space, and the fields at the end never do.
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		stmts, err1 := strconv.Atoi(fields[len(fields)-2])
		count, err2 := strconv.Atoi(fields[len(fields)-1])
		if err1 != nil || err2 != nil {
			continue
		}

		spec := strings.Join(fields[:len(fields)-2], " ")
		colon := strings.LastIndexByte(spec, ':')
		if colon < 0 {
			continue
		}
		from, to, ok := strings.Cut(spec[colon+1:], ",")
		if !ok {
			continue
		}
		startLine, err1 := strconv.Atoi(firstField(from, '.'))
		endLine, err2 := strconv.Atoi(firstField(to, '.'))
		if err1 != nil || err2 != nil {
			continue
		}

		blocks = append(blocks, coverBlock{
			file:       spec[:colon],
			start:      startLine,
			end:        endLine,
			statements: stmts,
			covered:    count > 0,
		})
	}
	return blocks
}

func firstField(s string, sep byte) string {
	if i := strings.IndexByte(s, sep); i >= 0 {
		return s[:i]
	}
	return s
}

// coverageOf attributes the blocks of a profile to a declaration.
//
// A block counts when it lies wholly inside the declaration. Overlapping ones
// are left out rather than apportioned: a block that straddles the end of a
// declaration belongs to whatever follows it, and splitting the difference
// would produce a figure that is precise about something untrue.
func coverageOf(c ir.Construct, blocks []coverBlock) (statements, covered int) {
	if c.EndLine == 0 || c.Pos.File == "" {
		return 0, 0
	}
	// The profile names a file by import path and base name; a construct
	// carries an absolute path and its package. Joining the two is what makes
	// them comparable without guessing at a repository layout.
	want := path.Join(c.Package, filepath.Base(c.Pos.File))

	for _, b := range blocks {
		if b.file != want || b.start < c.Pos.Line || b.end > c.EndLine {
			continue
		}
		statements += b.statements
		if b.covered {
			covered += b.statements
		}
	}
	return statements, covered
}

// loadCoverProfile reads a profile from a path, empty when none was given.
func loadCoverProfile(path string) ([]coverBlock, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return readCoverProfile(f), nil
}
