package jvm

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// SourceRoots are where the toolchains keep source, for recovering line numbers.
var SourceRoots = []string{
	"src/main/java", "src/main/kotlin",
	"src/test/java", "src/test/kotlin",
	"src", // Maven-less projects and the fixture.
}

// positions recovers declaration lines from source text.
//
// It exists because the class file format has no answer. LineNumberTable is an
// attribute of Code (JVMS §4.7.12) and a field has no code, so a field's
// declaration line is not missing from the reader — it does not exist. Class
// level lines are equally absent; only methods have any.
//
// That matters more here than it would elsewhere. This tool's diagnostics are
// meant to be acted on directly, and K1-FIELD-UNBOUND pointing at a file with
// no line would be the difference between a fix and a search.
//
// So this searches the source text, and it is a search and not a parser. It
// looks for the word, in a line that is not a comment, and takes the first
// match. It knows nothing about Java or Kotlin grammar and makes no attempt to:
// a parser here would be a second, worse implementation of the thing that was
// deliberately not built, and it would have to be right about generics, nested
// types, annotations and string literals to earn its keep.
//
// The rule it follows instead is that a wrong line is worse than no line. A
// reader sent to line 40 when the declaration is at line 12 will conclude the
// tool is confused; a reader given only the file will open it and look. So
// every heuristic here fails to zero rather than to a guess.
type positions struct {
	root        string
	sourceRoots []string

	mu    sync.Mutex
	cache map[string][]string // source path -> lines
}

func newPositions(root string, sourceRoots []string) *positions {
	if len(sourceRoots) == 0 {
		sourceRoots = SourceRoots
	}
	return &positions{root: root, sourceRoots: sourceRoots, cache: map[string][]string{}}
}

// Of returns the file and line a declaration was written at.
//
// name is the simple name being looked for: a class, a field or a method. An
// empty name asks only for the file, which is what a class level finding gets
// when nothing matches.
func (p *positions) Of(c *Class, name string) (file string, line int) {
	for _, candidate := range c.SourcePath(p.sourceRoots) {
		lines, ok := p.read(candidate)
		if !ok {
			continue
		}
		if name == "" {
			return candidate, 0
		}
		if n := findDeclaration(lines, name); n > 0 {
			return candidate, n
		}
		// The file is right even when the word is not found in it, which
		// happens for a generated member or a name that only appears in a
		// comment. Returning the file alone is still better than the class
		// file's own path.
		return candidate, 0
	}
	return c.File, 0
}

func (p *positions) read(rel string) ([]string, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if lines, ok := p.cache[rel]; ok {
		return lines, lines != nil
	}
	data, err := os.ReadFile(filepath.Join(p.root, rel))
	if err != nil {
		p.cache[rel] = nil
		return nil, false
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r\n", "\n"), "\n")
	p.cache[rel] = lines
	return lines, true
}

// findDeclaration returns the one based line the name is declared on, or 0.
//
// A declaration is taken to be the first occurrence of the name as a whole word
// on a line that is not a comment. That is wrong for a name used before it is
// declared — a field referenced in a method above its own declaration — and
// right for the overwhelming majority of real code, where types and fields are
// declared before or above their uses.
//
// Being wrong here costs a line number on a finding whose file is still
// correct. Being clever here would cost a grammar.
func findDeclaration(lines []string, name string) int {
	inBlockComment := false

	for i, raw := range lines {
		line := strings.TrimSpace(raw)

		if inBlockComment {
			if idx := strings.Index(line, "*/"); idx >= 0 {
				inBlockComment = false
				line = line[idx+2:]
			} else {
				continue
			}
		}
		// A line opening a block comment that does not close it takes the rest
		// of the line with it. Javadoc above a declaration is the single most
		// common place for the name to appear before it is declared.
		if idx := strings.Index(line, "/*"); idx >= 0 {
			if !strings.Contains(line[idx:], "*/") {
				inBlockComment = true
			}
			line = line[:idx]
		}
		if idx := strings.Index(line, "//"); idx >= 0 {
			line = line[:idx]
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		if containsWord(line, name) {
			return i + 1
		}
	}
	return 0
}

// containsWord reports whether name appears as a whole identifier.
//
// Without the boundary check, looking for "id" would match "quoteId", "valid"
// and "identity", and the first of those is on almost every line of a data
// class.
func containsWord(line, name string) bool {
	for i := 0; ; {
		idx := strings.Index(line[i:], name)
		if idx < 0 {
			return false
		}
		start := i + idx
		end := start + len(name)

		before := start == 0 || !isIdentRune(line[start-1])
		after := end == len(line) || !isIdentRune(line[end])
		if before && after {
			return true
		}
		i = start + 1
		if i >= len(line) {
			return false
		}
	}
}

func isIdentRune(b byte) bool {
	return b == '_' || b == '$' ||
		(b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}
