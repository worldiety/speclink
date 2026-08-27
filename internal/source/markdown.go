package source

import (
	"strings"
)

// informativeMarker excludes a section from the coverage obligation.
//
// It is an HTML comment because that is the only annotation Markdown has that
// every renderer ignores. The source documents belong to the departments who
// write them; a marker that showed up in the rendered text would be a tool
// leaking into somebody else's document.
const informativeMarker = "<!-- speclink:informative -->"

// preambleID addresses the text before the first heading.
//
// The preamble is a segment like any other and carries the same obligation.
// Exempting it would create the one place in a document where a requirement can
// be written and then legitimately produce nothing, which is precisely the
// defect this package exists to find. Documents that genuinely open with prose
// mark it informative, one line, once.
const preambleID = "_preamble"

// segmentMarkdown splits a Markdown document at its headings.
//
// Segments are disjoint: a section ends at the next heading of any level, so
// nested content belongs to the innermost heading above it and to nothing else.
// Nesting the other way — a level two section containing its level three
// children — would count the same words several times, and a requirement citing
// the parent would silently satisfy the coverage obligation of every child.
func segmentMarkdown(doc string, md string) ([]Segment, []error) {
	type raw struct {
		id    string
		title string
		line  int
		body  []string
	}

	var (
		segs   []raw
		cur    = raw{id: preambleID, title: "(preamble)", line: 1}
		inCode bool
	)

	lines := strings.Split(normalizeNewlines(md), "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Fenced blocks are skipped so a comment inside a code sample is not
		// mistaken for a heading. The fence toggles even inside a heading-less
		// preamble, which is why this runs before the heading test.
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inCode = !inCode
			cur.body = append(cur.body, line)
			continue
		}

		if inCode || !isATXHeading(trimmed) {
			cur.body = append(cur.body, line)
			continue
		}

		segs = append(segs, cur)
		title := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
		cur = raw{id: Slug(title), title: title, line: i + 1}
	}
	segs = append(segs, cur)

	var (
		out  []Segment
		errs []error
		seen = map[string]int{}
	)
	for _, s := range segs {
		body := strings.Join(s.body, "\n")

		// A preamble of nothing but blank lines is not a segment. Every
		// document that opens directly with a heading would otherwise carry an
		// empty obligation that can only ever be discharged by marking it
		// informative.
		if s.id == preambleID && strings.TrimSpace(body) == "" {
			continue
		}

		if s.id == "" {
			errs = append(errs, &SegmentError{
				Doc:  doc,
				Line: s.line,
				Msg:  "heading " + quote(s.title) + " has no addressable slug",
				Why:  "The slug keeps letters and digits of any script. A heading made only of punctuation cannot be cited by a requirement.",
				How:  "Give the heading at least one letter or digit.",
			})
			continue
		}

		// Two headings slugging the same make every citation of that slug
		// ambiguous. Numbering the duplicate away, as some renderers do, would
		// mean the address of a section depends on how many sections above it
		// happen to share a name — a reference would break when an unrelated
		// heading is added.
		if prev, dup := seen[s.id]; dup {
			errs = append(errs, &SegmentError{
				Doc:  doc,
				Line: s.line,
				Msg:  "two headings share the slug " + quote(s.id),
				Why:  "A slug is the address of a section. Two sections with one address make every citation of it ambiguous. First occurrence: line " + itoa(prev) + ".",
				How:  "Rename one of the headings.",
			})
			continue
		}
		seen[s.id] = s.line

		out = append(out, Segment{
			Doc:         doc,
			ID:          s.id,
			Kind:        KindMarkdown,
			Title:       s.title,
			Fingerprint: fingerprint([]byte(canonical(body))),
			Informative: strings.Contains(body, informativeMarker),
			Pos:         Pos{File: doc, Line: s.line},
		})
	}
	return out, errs
}

// isATXHeading reports whether the line opens a heading.
//
// An ATX heading requires whitespace after the hashes; without it the line is
// ordinary text such as a "#tag".
func isATXHeading(trimmed string) bool {
	if !strings.HasPrefix(trimmed, "#") {
		return false
	}
	rest := strings.TrimLeft(trimmed, "#")
	if rest == trimmed {
		return false
	}
	return rest == "" || rest[0] == ' ' || rest[0] == '\t'
}

// canonical reduces a section body to what a change to it would have to mean.
//
// Trailing whitespace and the number of blank lines around a paragraph are
// invisible to the reader and are routinely rewritten by editors and
// formatters. Reporting them as drift would make the rule fire on
// reformatting, and a rule that fires on reformatting is one that gets
// waived by habit.
//
// Everything else is kept verbatim. Inner whitespace, capitalisation and
// punctuation all carry meaning in a requirement document, and a hash that
// forgives them would be quietly deciding that some rewrites do not count.
func canonical(body string) string {
	var (
		out   []string
		blank int
	)
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimRight(line, " \t")
		if line == "" {
			blank++
			continue
		}
		if blank > 0 && len(out) > 0 {
			out = append(out, "")
		}
		blank = 0
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func normalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}
