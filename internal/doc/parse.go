package doc

import (
	"os"
	"strings"
)

// LoadMarkdown reads a prose file into blocks.
//
// The check that a chapter is usable and the rendering of that chapter go
// through here together, so a file that passes the check cannot turn out to
// produce something else when the document is written.
func LoadMarkdown(path string, base int) ([]Block, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseMarkdown(string(b), base), nil
}

// FirstHeading returns the heading a set of blocks opens with, empty when it
// opens with something else.
func FirstHeading(blocks []Block) string {
	if len(blocks) == 0 {
		return ""
	}
	h, ok := blocks[0].(*Heading)
	if !ok {
		return ""
	}
	return h.Text
}

// ParseMarkdown reads prose into document blocks.
//
// # Why this exists at all
//
// Every other chapter of the specification is derived: it says what the model
// says, and there is exactly one place each sentence is written. Prose that
// explains the system to a person cannot be derived from anything, and a
// document without it describes a module to somebody who already knows what
// the module is for.
//
// # Why it is not a Markdown implementation
//
// It recognises what the prose of a specification actually uses: headings,
// paragraphs, both kinds of list, tables, fenced listings and block quotes.
// It does not implement CommonMark, and deliberately so — the block model of
// this package is small on purpose, and a construct with nowhere to go would
// have to be dropped silently or approximated, which is how a document starts
// lying about its own contents. Anything unrecognised stays a paragraph and
// keeps its characters.
//
// base is the heading level a top level "#" becomes, so a chapter can be
// slotted under the generated ones without the prose having to know where it
// was placed.
func ParseMarkdown(src string, base int) []Block {
	p := &mdParser{lines: splitLines(src), base: base}
	return p.blocks()
}

type mdParser struct {
	lines []string
	i     int
	base  int
	out   []Block
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

func (p *mdParser) done() bool      { return p.i >= len(p.lines) }
func (p *mdParser) peek() string    { return p.lines[p.i] }
func (p *mdParser) next() string    { s := p.lines[p.i]; p.i++; return s }
func (p *mdParser) add(blk Block)   { p.out = append(p.out, blk) }
func (p *mdParser) blank() bool     { return strings.TrimSpace(p.peek()) == "" }
func (p *mdParser) trimmed() string { return strings.TrimSpace(p.peek()) }

func (p *mdParser) blocks() []Block {
	p.frontmatter()
	for !p.done() {
		switch {
		case p.blank():
			p.i++
		case p.fence() != "":
			p.listing()
		case headingLevel(p.trimmed()) > 0:
			p.heading()
		case strings.HasPrefix(p.trimmed(), ">"):
			p.quote()
		case p.tableAhead():
			p.table()
		case bulletOf(p.peek()) != "":
			p.list()
		case isRule(p.trimmed()):
			p.i++
		default:
			p.paragraph()
		}
	}
	return p.out
}

// frontmatter drops a leading YAML block.
//
// The prose files of a specification carry metadata for whatever assembled
// them before. It is not content, and rendering it would put a row of colons
// at the top of a chapter.
func (p *mdParser) frontmatter() {
	if p.done() || strings.TrimSpace(p.peek()) != "---" {
		return
	}
	for j := p.i + 1; j < len(p.lines); j++ {
		if strings.TrimSpace(p.lines[j]) == "---" {
			p.i = j + 1
			return
		}
	}
	// No closing marker: it was a thematic break after all, not frontmatter.
}

func (p *mdParser) heading() {
	line := p.trimmed()
	level := headingLevel(line)
	text := strings.TrimSpace(strings.TrimLeft(line, "#"))
	p.i++
	p.add(&Heading{Level: min(p.base+level-1, 6), Text: text})
}

func (p *mdParser) paragraph() {
	var b []string
	for !p.done() && !p.blank() && !p.interrupts() {
		b = append(b, strings.TrimSpace(p.next()))
	}
	if len(b) == 0 {
		return
	}
	p.add(&Para{Text: ParseInline(strings.Join(b, " "))})
}

// interrupts reports whether the current line ends a paragraph without a blank
// line before it, which every block construct except another paragraph does.
func (p *mdParser) interrupts() bool {
	t := p.trimmed()
	return headingLevel(t) > 0 || p.fence() != "" || strings.HasPrefix(t, ">") ||
		bulletOf(p.peek()) != "" || isRule(t) || strings.HasPrefix(t, "|")
}

func (p *mdParser) quote() {
	var b []string
	for !p.done() && strings.HasPrefix(p.trimmed(), ">") {
		b = append(b, strings.TrimSpace(strings.TrimPrefix(p.trimmed(), ">")))
		p.i++
	}
	p.add(&Note{Text: ParseInline(strings.TrimSpace(strings.Join(b, " ")))})
}

// fence returns the backtick or tilde run opening a listing, empty when the
// line does not open one.
func (p *mdParser) fence() string {
	t := p.trimmed()
	for _, c := range []string{"```", "~~~"} {
		if strings.HasPrefix(t, c) {
			return c
		}
	}
	return ""
}

func (p *mdParser) listing() {
	f := p.fence()
	lang := strings.TrimSpace(strings.TrimPrefix(p.trimmed(), f))
	p.i++
	var b []string
	for !p.done() && !strings.HasPrefix(p.trimmed(), f) {
		b = append(b, p.next())
	}
	if !p.done() {
		p.i++ // the closing fence
	}
	p.add(&Listing{Lang: lang, Text: strings.Join(b, "\n")})
}

func (p *mdParser) list() {
	ordered := orderedItem(p.peek())
	l := &List{Ordered: ordered}
	for !p.done() {
		marker := bulletOf(p.peek())
		if marker == "" {
			break
		}
		// An indented item hangs under the one before it, whichever marker it
		// uses: a bulleted list under a numbered step is the ordinary way to
		// write "this step, and these details".
		nested := indentOf(p.peek()) >= 2 && len(l.Items) > 0
		if !nested && orderedItem(p.peek()) != ordered {
			break
		}
		text := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(p.next()), marker))
		// Items are hard wrapped in prose written for review, so an indented
		// line that starts no item of its own continues this one. Without
		// this the tail of every wrapped item breaks out of the list and
		// lands in the text as a paragraph of its own, which also restarts
		// the numbering at one.
		for !p.done() && !p.blank() && bulletOf(p.peek()) == "" && indentOf(p.peek()) > 0 {
			text += " " + strings.TrimSpace(p.next())
		}
		if nested {
			// The model allows one level and no more, so a deeper one is
			// folded into this one rather than dropped.
			last := &l.Items[len(l.Items)-1]
			last.Sub = append(last.Sub, ParseInline(text))
			continue
		}
		l.Items = append(l.Items, Bullet{Text: ParseInline(text)})
	}
	if len(l.Items) > 0 {
		p.add(l)
	}
}

func (p *mdParser) tableAhead() bool {
	if !strings.HasPrefix(p.trimmed(), "|") {
		return false
	}
	return p.i+1 < len(p.lines) && isTableRule(strings.TrimSpace(p.lines[p.i+1]))
}

func (p *mdParser) table() {
	head := splitRow(p.trimmed())
	p.i++
	align := alignOf(splitRow(strings.TrimSpace(p.next())))
	t := &Table{Head: head, Align: align}
	for !p.done() && strings.HasPrefix(p.trimmed(), "|") {
		var row [][]Inline
		for _, c := range splitRow(strings.TrimSpace(p.next())) {
			row = append(row, ParseInline(c))
		}
		t.Rows = append(t.Rows, row)
	}
	p.add(t)
}

func splitRow(line string) []string {
	line = strings.TrimSuffix(strings.TrimPrefix(line, "|"), "|")
	out := strings.Split(line, "|")
	for i := range out {
		out[i] = strings.TrimSpace(out[i])
	}
	return out
}

func alignOf(cells []string) []Align {
	out := make([]Align, len(cells))
	for i, c := range cells {
		if strings.HasSuffix(c, ":") && !strings.HasPrefix(c, ":") {
			out[i] = Right
		}
	}
	return out
}

func isTableRule(s string) bool {
	if !strings.HasPrefix(s, "|") {
		return false
	}
	for _, c := range s {
		if c != '|' && c != '-' && c != ':' && c != ' ' {
			return false
		}
	}
	return strings.ContainsRune(s, '-')
}

// isRule reports a thematic break, which the block model has no node for and
// which carries no meaning a chapter needs.
func isRule(s string) bool {
	if len(s) < 3 {
		return false
	}
	for _, c := range s {
		if c != '-' && c != '*' && c != '_' && c != ' ' {
			return false
		}
	}
	return true
}

func headingLevel(s string) int {
	n := 0
	for n < len(s) && s[n] == '#' {
		n++
	}
	if n == 0 || n > 6 || n >= len(s) || s[n] != ' ' {
		return 0
	}
	return n
}

func indentOf(s string) int {
	n := 0
	for n < len(s) && (s[n] == ' ' || s[n] == '\t') {
		n++
	}
	return n
}

// bulletOf returns the list marker of a line, empty when it starts no item.
func bulletOf(line string) string {
	t := strings.TrimSpace(line)
	for _, m := range []string{"- ", "* ", "+ "} {
		if strings.HasPrefix(t, m) {
			return strings.TrimSpace(m)
		}
	}
	if d := orderedMarker(t); d != "" {
		return d
	}
	return ""
}

func orderedItem(line string) bool { return orderedMarker(strings.TrimSpace(line)) != "" }

// orderedMarker returns "12." for a line beginning with that, empty otherwise.
func orderedMarker(t string) string {
	n := 0
	for n < len(t) && t[n] >= '0' && t[n] <= '9' {
		n++
	}
	if n == 0 || n+1 >= len(t) || t[n] != '.' || t[n+1] != ' ' {
		return ""
	}
	return t[:n+1]
}

// ParseInline reads the span level markup of one line.
//
// Code spans are taken first and their contents left alone: a backtick span is
// the one place where an asterisk is an asterisk, and handling it later would
// turn a documented glob pattern into emphasis.
func ParseInline(s string) []Inline {
	var out []Inline
	var lit strings.Builder
	flush := func() {
		if lit.Len() > 0 {
			out = append(out, Text(lit.String()))
			lit.Reset()
		}
	}
	for i := 0; i < len(s); {
		switch {
		case s[i] == '`':
			if j := strings.IndexByte(s[i+1:], '`'); j >= 0 {
				flush()
				out = append(out, Code(s[i+1:i+1+j]))
				i += j + 2
				continue
			}
		case strings.HasPrefix(s[i:], "**"):
			if j := strings.Index(s[i+2:], "**"); j >= 0 {
				flush()
				out = append(out, Strong(s[i+2:i+2+j]))
				i += j + 4
				continue
			}
		case s[i] == '*':
			if j := strings.IndexByte(s[i+1:], '*'); j > 0 {
				flush()
				out = append(out, Emph(s[i+1:i+1+j]))
				i += j + 2
				continue
			}
		case s[i] == '[':
			if text, url, n, ok := mdLink(s[i:]); ok {
				flush()
				out = append(out, Link{Text: text, URL: url})
				i += n
				continue
			}
		}
		lit.WriteByte(s[i])
		i++
	}
	flush()
	return out
}

// mdLink reads "[text](url)" at the start of s.
func mdLink(s string) (text, url string, n int, ok bool) {
	close := strings.IndexByte(s, ']')
	if close < 0 || close+1 >= len(s) || s[close+1] != '(' {
		return "", "", 0, false
	}
	end := strings.IndexByte(s[close+2:], ')')
	if end < 0 {
		return "", "", 0, false
	}
	return s[1:close], s[close+2 : close+2+end], close + 2 + end + 1, true
}
