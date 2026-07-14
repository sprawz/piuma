// Package content models the markdown sources of the site: frontmatter,
// posts, and the content tree on disk.
package content

import (
	"fmt"
	"strings"
)

const fmDelimiter = "---"

// Split separates a document into its frontmatter source and markdown body.
// The document must start with a `---` line and contain a closing one.
func Split(src []byte) (fm, body []byte, err error) {
	s := string(src)
	rest, ok := strings.CutPrefix(s, fmDelimiter+"\n")
	if !ok {
		return nil, nil, fmt.Errorf("missing frontmatter opening %q", fmDelimiter)
	}
	fmSrc, bodySrc, ok := strings.Cut(rest, "\n"+fmDelimiter+"\n")
	if !ok {
		return nil, nil, fmt.Errorf("missing frontmatter closing %q", fmDelimiter)
	}
	return []byte(fmSrc + "\n"), []byte(bodySrc), nil
}

// Parse reads the YAML subset used by the blog's frontmatter: string
// scalars (optionally quoted), lists of scalars, and nested maps by
// indentation. Values are string, []string, or map[string]any.
func Parse(src []byte) (map[string]any, error) {
	lines, err := scanLines(src)
	if err != nil {
		return nil, err
	}
	m, rest, err := parseMap(lines, 0)
	if err != nil {
		return nil, err
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("line %d: unexpected indentation", rest[0].num)
	}
	return m, nil
}

type line struct {
	indent int
	text   string // content with indentation stripped
	num    int    // 1-based line number in the frontmatter block
}

func scanLines(src []byte) ([]line, error) {
	var lines []line
	for i, raw := range strings.Split(string(src), "\n") {
		trimmed := strings.TrimLeft(raw, " ")
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "\t") {
			return nil, fmt.Errorf("line %d: tabs are not allowed for indentation", i+1)
		}
		lines = append(lines, line{indent: len(raw) - len(trimmed), text: trimmed, num: i + 1})
	}
	return lines, nil
}

func parseMap(lines []line, indent int) (map[string]any, []line, error) {
	m := map[string]any{}
	for len(lines) > 0 && lines[0].indent == indent {
		ln := lines[0]
		lines = lines[1:]
		key, val, ok := strings.Cut(ln.text, ":")
		if !ok || key == "" || strings.HasPrefix(ln.text, "- ") {
			return nil, nil, fmt.Errorf("line %d: expected \"key: value\"", ln.num)
		}
		val = strings.TrimSpace(val)
		if val != "" {
			full, rest, err := continueQuoted(val, lines, ln.num)
			if err != nil {
				return nil, nil, err
			}
			m[key], lines = unquote(full), rest
			continue
		}
		if len(lines) == 0 || lines[0].indent <= indent {
			return nil, nil, fmt.Errorf("line %d: key %q has no value", ln.num, key)
		}
		child, rest, err := parseBlock(lines, lines[0].indent)
		if err != nil {
			return nil, nil, err
		}
		m[key], lines = child, rest
	}
	return m, lines, nil
}

// parseBlock parses an indented value block: either a list or a nested map.
func parseBlock(lines []line, indent int) (any, []line, error) {
	if !strings.HasPrefix(lines[0].text, "- ") {
		return parseMap(lines, indent)
	}
	var items []string
	for len(lines) > 0 && lines[0].indent == indent && strings.HasPrefix(lines[0].text, "- ") {
		items = append(items, unquote(strings.TrimPrefix(lines[0].text, "- ")))
		lines = lines[1:]
	}
	return items, lines, nil
}

// continueQuoted consumes follow-up lines of a quoted scalar that opens on
// one line and closes on a later one. Line breaks fold to spaces, as YAML
// does for flow scalars.
func continueQuoted(val string, lines []line, num int) (string, []line, error) {
	q := val[0]
	if (q != '\'' && q != '"') || quoteClosed(val, q) {
		return val, lines, nil
	}
	for len(lines) > 0 {
		val += " " + lines[0].text
		lines = lines[1:]
		if quoteClosed(val, q) {
			return val, lines, nil
		}
	}
	return "", nil, fmt.Errorf("line %d: unterminated quoted value", num)
}

func quoteClosed(s string, q byte) bool {
	return len(s) >= 2 && s[len(s)-1] == q
}

func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[len(s)-1] == s[0] {
		return s[1 : len(s)-1]
	}
	return s
}
