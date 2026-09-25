package cargo

import (
	"fmt"
	"strconv"
	"strings"
)

// parseTOML reads the part of TOML a Cargo manifest uses.
//
// That is most of it — tables, arrays of tables, dotted and quoted keys, all
// four string forms, arrays and inline tables spanning lines — and not the
// typed scalars: numbers, dates and times come back as their source text.
// Nothing speclink reads from a manifest is a number, and a parser that
// converted them would be a second place for a version string like 1.0 to
// turn into a float.
//
// It is written here rather than imported because the whole of speclink
// depends on one library outside the standard one, and a manifest is not worth
// a second.
func parseTOML(src string) (map[string]any, error) {
	p := &tomlParser{s: src, line: 1}
	root := map[string]any{}
	cur := root
	for {
		p.skipBlank()
		if p.eof() {
			return root, nil
		}
		if p.s[p.i] == '[' {
			t, err := p.header(root)
			if err != nil {
				return nil, err
			}
			cur = t
			continue
		}
		keys, err := p.key()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if !p.consume('=') {
			return nil, p.errorf("expected = after key %s", strings.Join(keys, "."))
		}
		v, err := p.value()
		if err != nil {
			return nil, err
		}
		if err := set(cur, keys, v); err != nil {
			return nil, p.errorf("%v", err)
		}
		p.skipSpace()
		p.skipComment()
		if !p.eof() && p.s[p.i] != '\n' && p.s[p.i] != '\r' {
			return nil, p.errorf("unexpected %q after value", p.s[p.i])
		}
	}
}

type tomlParser struct {
	s    string
	i    int
	line int
}

func (p *tomlParser) eof() bool { return p.i >= len(p.s) }

func (p *tomlParser) errorf(format string, args ...any) error {
	return fmt.Errorf("line %d: %s", p.line, fmt.Sprintf(format, args...))
}

func (p *tomlParser) consume(c byte) bool {
	if !p.eof() && p.s[p.i] == c {
		p.i++
		return true
	}
	return false
}

func (p *tomlParser) skipSpace() {
	for !p.eof() && (p.s[p.i] == ' ' || p.s[p.i] == '\t') {
		p.i++
	}
}

func (p *tomlParser) skipComment() {
	if !p.eof() && p.s[p.i] == '#' {
		for !p.eof() && p.s[p.i] != '\n' {
			p.i++
		}
	}
}

// skipBlank skips whitespace, newlines and comments.
func (p *tomlParser) skipBlank() {
	for !p.eof() {
		switch p.s[p.i] {
		case ' ', '\t', '\r':
			p.i++
		case '\n':
			p.line++
			p.i++
		case '#':
			p.skipComment()
		default:
			return
		}
	}
}

func (p *tomlParser) header(root map[string]any) (map[string]any, error) {
	array := strings.HasPrefix(p.s[p.i:], "[[")
	if array {
		p.i += 2
	} else {
		p.i++
	}
	keys, err := p.key()
	if err != nil {
		return nil, err
	}
	p.skipSpace()
	closing := "]"
	if array {
		closing = "]]"
	}
	if !strings.HasPrefix(p.s[p.i:], closing) {
		return nil, p.errorf("expected %s", closing)
	}
	p.i += len(closing)

	parent := root
	for _, k := range keys[:len(keys)-1] {
		next, err := descend(parent, k)
		if err != nil {
			return nil, p.errorf("%v", err)
		}
		parent = next
	}
	last := keys[len(keys)-1]
	if array {
		t := map[string]any{}
		list, _ := parent[last].([]map[string]any)
		parent[last] = append(list, t)
		return t, nil
	}
	t, err := descend(parent, last)
	if err != nil {
		return nil, p.errorf("%v", err)
	}
	return t, nil
}

// descend returns the table under k, creating it. Under an array of tables it
// is the last one, which is what a header like [bin.x] after [[bin]] means.
func descend(m map[string]any, k string) (map[string]any, error) {
	switch v := m[k].(type) {
	case nil:
		t := map[string]any{}
		m[k] = t
		return t, nil
	case map[string]any:
		return v, nil
	case []map[string]any:
		if len(v) > 0 {
			return v[len(v)-1], nil
		}
	}
	return nil, fmt.Errorf("%s is not a table", k)
}

func set(m map[string]any, keys []string, v any) error {
	for _, k := range keys[:len(keys)-1] {
		next, err := descend(m, k)
		if err != nil {
			return err
		}
		m = next
	}
	last := keys[len(keys)-1]
	if _, dup := m[last]; dup {
		return fmt.Errorf("key %s defined twice", last)
	}
	m[last] = v
	return nil
}

func (p *tomlParser) key() ([]string, error) {
	var keys []string
	for {
		p.skipSpace()
		if p.eof() {
			return nil, p.errorf("expected a key")
		}
		switch c := p.s[p.i]; {
		case c == '"':
			s, err := p.basic()
			if err != nil {
				return nil, err
			}
			keys = append(keys, s)
		case c == '\'':
			s, err := p.literal()
			if err != nil {
				return nil, err
			}
			keys = append(keys, s)
		default:
			from := p.i
			for !p.eof() && bare(p.s[p.i]) {
				p.i++
			}
			if from == p.i {
				return nil, p.errorf("expected a key, found %q", c)
			}
			keys = append(keys, p.s[from:p.i])
		}
		p.skipSpace()
		if !p.consume('.') {
			return keys, nil
		}
	}
}

func bare(c byte) bool {
	return c == '_' || c == '-' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

func (p *tomlParser) value() (any, error) {
	p.skipSpace()
	if p.eof() {
		return nil, p.errorf("expected a value")
	}
	rest := p.s[p.i:]
	switch {
	case strings.HasPrefix(rest, `"""`):
		return p.multiline(`"""`, true)
	case strings.HasPrefix(rest, `'''`):
		return p.multiline(`'''`, false)
	case rest[0] == '"':
		return p.basic()
	case rest[0] == '\'':
		return p.literal()
	case rest[0] == '[':
		return p.array()
	case rest[0] == '{':
		return p.inline()
	case strings.HasPrefix(rest, "true"):
		p.i += 4
		return true, nil
	case strings.HasPrefix(rest, "false"):
		p.i += 5
		return false, nil
	}
	from := p.i
	for !p.eof() && !strings.ContainsRune(" \t\r\n,]}#", rune(p.s[p.i])) {
		p.i++
	}
	if from == p.i {
		return nil, p.errorf("expected a value, found %q", rest[0])
	}
	return p.s[from:p.i], nil
}

func (p *tomlParser) basic() (string, error) {
	p.i++
	var b strings.Builder
	for !p.eof() {
		c := p.s[p.i]
		switch c {
		case '"':
			p.i++
			return b.String(), nil
		case '\n':
			return "", p.errorf("newline in string")
		case '\\':
			if err := p.escape(&b); err != nil {
				return "", err
			}
			continue
		}
		b.WriteByte(c)
		p.i++
	}
	return "", p.errorf("unterminated string")
}

func (p *tomlParser) escape(b *strings.Builder) error {
	p.i++
	if p.eof() {
		return p.errorf("unterminated escape")
	}
	c := p.s[p.i]
	p.i++
	switch c {
	case 'n':
		b.WriteByte('\n')
	case 't':
		b.WriteByte('\t')
	case 'r':
		b.WriteByte('\r')
	case 'b':
		b.WriteByte('\b')
	case 'f':
		b.WriteByte('\f')
	case '"', '\\':
		b.WriteByte(c)
	case 'u', 'U':
		n := 4
		if c == 'U' {
			n = 8
		}
		if p.i+n > len(p.s) {
			return p.errorf("short unicode escape")
		}
		v, err := strconv.ParseUint(p.s[p.i:p.i+n], 16, 32)
		if err != nil {
			return p.errorf("bad unicode escape")
		}
		b.WriteRune(rune(v))
		p.i += n
	default:
		return p.errorf("unknown escape \\%c", c)
	}
	return nil
}

func (p *tomlParser) literal() (string, error) {
	p.i++
	end := strings.IndexAny(p.s[p.i:], "'\n")
	if end < 0 || p.s[p.i+end] != '\'' {
		return "", p.errorf("unterminated string")
	}
	s := p.s[p.i : p.i+end]
	p.i += end + 1
	return s, nil
}

func (p *tomlParser) multiline(delim string, escapes bool) (string, error) {
	p.i += 3
	// A newline right after the opening delimiter is not part of the value.
	if strings.HasPrefix(p.s[p.i:], "\r\n") {
		p.i += 2
		p.line++
	} else if p.consume('\n') {
		p.line++
	}
	var b strings.Builder
	for !p.eof() {
		if strings.HasPrefix(p.s[p.i:], delim) {
			p.i += 3
			return b.String(), nil
		}
		c := p.s[p.i]
		if c == '\n' {
			p.line++
		}
		if escapes && c == '\\' {
			// A backslash at the end of a line joins it with the next.
			rest := strings.TrimLeft(p.s[p.i+1:], " \t\r")
			if strings.HasPrefix(rest, "\n") {
				p.i = len(p.s) - len(strings.TrimLeft(rest, " \t\r\n"))
				continue
			}
			if err := p.escape(&b); err != nil {
				return "", err
			}
			continue
		}
		b.WriteByte(c)
		p.i++
	}
	return "", p.errorf("unterminated multi-line string")
}

func (p *tomlParser) array() ([]any, error) {
	p.i++
	var out []any
	for {
		p.skipBlank()
		if p.consume(']') {
			return out, nil
		}
		v, err := p.value()
		if err != nil {
			return nil, err
		}
		out = append(out, v)
		p.skipBlank()
		if p.consume(']') {
			return out, nil
		}
		if !p.consume(',') {
			return nil, p.errorf("expected , or ] in array")
		}
	}
}

func (p *tomlParser) inline() (map[string]any, error) {
	p.i++
	out := map[string]any{}
	for {
		p.skipBlank()
		if p.consume('}') {
			return out, nil
		}
		keys, err := p.key()
		if err != nil {
			return nil, err
		}
		p.skipSpace()
		if !p.consume('=') {
			return nil, p.errorf("expected = in inline table")
		}
		v, err := p.value()
		if err != nil {
			return nil, err
		}
		if err := set(out, keys, v); err != nil {
			return nil, p.errorf("%v", err)
		}
		p.skipBlank()
		if p.consume('}') {
			return out, nil
		}
		if !p.consume(',') {
			return nil, p.errorf("expected , or } in inline table")
		}
	}
}
