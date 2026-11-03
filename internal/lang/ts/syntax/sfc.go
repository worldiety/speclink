package syntax

import (
	"regexp"
	"strings"
)

// Block is one top level block of a Vue single file component.
type Block struct {
	// Tag is script, style or template.
	Tag string
	// Attrs is the attribute text of the opening tag, verbatim.
	Attrs string
	// From and To delimit the content, without the tags.
	From, To int
}

// Setup reports whether a script block is <script setup>.
func (b Block) Setup() bool {
	for _, f := range strings.Fields(b.Attrs) {
		if f == "setup" || strings.HasPrefix(f, "setup=") {
			return true
		}
	}
	return false
}

var sfcBlock = regexp.MustCompile(`(?s)<(script|style)\b([^>]*)>(.*?)</(?:script|style)\s*>`)

// SFCBlocks splits a single file component into its script and style blocks,
// and returns everything else as template blocks.
//
// The template is not parsed as HTML. What a rule asks of it — which
// translation keys it uses — is answered by its text, and a markup parser here
// would be a second, larger thing to get right for no better answer.
func SFCBlocks(src string) []Block {
	var out []Block
	last := 0
	for _, m := range sfcBlock.FindAllStringSubmatchIndex(src, -1) {
		if m[0] > last {
			out = append(out, Block{Tag: "template", From: last, To: m[0]})
		}
		out = append(out, Block{
			Tag:   src[m[2]:m[3]],
			Attrs: strings.TrimSpace(src[m[4]:m[5]]),
			From:  m[6],
			To:    m[7],
		})
		last = m[1]
	}
	if last < len(src) {
		out = append(out, Block{Tag: "template", From: last, To: len(src)})
	}
	return out
}

// LineOf returns the 1 based line of a byte offset.
func LineOf(src string, off int) int {
	if off > len(src) {
		off = len(src)
	}
	return 1 + strings.Count(src[:off], "\n")
}
