package source

import (
	"strings"
	"unicode"
)

// Headings returns the anchor slugs of all ATX headings in a Markdown document,
// in document order.
//
// speclink defines what an anchor is; converting some legacy convention into
// this format is project work, not the tool's job.
//
// Fenced code blocks are skipped so that comment lines starting with # inside a
// code sample are not mistaken for headings.
func Headings(md string) []string {
	var (
		out    []string
		inCode bool
	)
	for _, line := range strings.Split(md, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inCode = !inCode
			continue
		}
		if inCode || !strings.HasPrefix(trimmed, "#") {
			continue
		}
		text := strings.TrimLeft(trimmed, "#")
		// An ATX heading requires whitespace after the hashes. Without it the
		// line is ordinary text, e.g. a "#tag".
		if text == trimmed || (text != "" && !isSpace(text[0])) {
			continue
		}
		if slug := Slug(text); slug != "" {
			out = append(out, slug)
		}
	}
	return out
}

// Slug converts a heading into its anchor.
//
// The rule is the common Markdown one: lower case, spaces become "-",
// punctuation is dropped. "## 8.1 Angebot (Kopf)" yields "81-angebot-kopf".
//
// Letters and digits are kept for any script, so headings in German, French or
// Greek slug correctly rather than collapsing to empty.
func Slug(heading string) string {
	var b strings.Builder
	lastDash := true // suppresses a leading dash

	for _, r := range strings.TrimSpace(heading) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(unicode.ToLower(r))
			lastDash = false
		case r == ' ' || r == '\t' || r == '-' || r == '_':
			if !lastDash && b.Len() > 0 {
				b.WriteByte('-')
				lastDash = true
			}
		default:
			// punctuation is dropped without joining the neighbours
		}
	}
	return strings.TrimRight(b.String(), "-")
}

func isSpace(b byte) bool { return b == ' ' || b == '\t' }
