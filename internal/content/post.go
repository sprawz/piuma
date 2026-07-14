package content

import (
	"fmt"
	"strings"
	"time"
)

// Post is one markdown article. The slug comes from the file name
// (content/<slug>.md), the rest from frontmatter.
type Post struct {
	Slug        string
	Title       string
	Excerpt     string
	Description string // meta description, falls back to Excerpt
	Tags        []string
	PublishDate time.Time
	Body        []byte // markdown body, without frontmatter
}

// URL is the post's site-absolute path: all posts live under /words.
func (p *Post) URL() string {
	return "/words/" + p.Slug
}

// TagSlug turns a tag name into its URL segment: lowercase, runs of
// non-alphanumerics collapse to one dash ("Google Tag Manager" →
// "google-tag-manager").
func TagSlug(tag string) string {
	var b strings.Builder
	dash := false
	for _, r := range strings.ToLower(tag) {
		switch {
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			b.WriteRune(r)
			dash = false
		case !dash && b.Len() > 0:
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// ParsePost builds a Post from a raw markdown document.
func ParsePost(src []byte, slug string) (*Post, error) {
	fm, body, err := parseDoc(src)
	if err != nil {
		return nil, err
	}
	p := &Post{
		Slug:    slug,
		Title:   str(fm, "title"),
		Excerpt: str(fm, "excerpt"),
		Body:    body,
	}
	if p.Title == "" {
		return nil, fmt.Errorf("missing required field %q", "title")
	}
	date := str(fm, "publishDate")
	if date == "" {
		return nil, fmt.Errorf("missing required field %q", "publishDate")
	}
	if p.PublishDate, err = time.Parse(time.RFC3339, date); err != nil {
		return nil, fmt.Errorf("publishDate: %v", err)
	}
	if tags, ok := fm["tags"].([]string); ok {
		p.Tags = tags
	}
	if meta, ok := fm["metadata"].(map[string]any); ok {
		p.Description = str(meta, "description")
	}
	if p.Description == "" {
		p.Description = p.Excerpt
	}
	return p, nil
}

// parseDoc splits a document and parses its frontmatter in one step.
func parseDoc(src []byte) (map[string]any, []byte, error) {
	fmSrc, body, err := Split(src)
	if err != nil {
		return nil, nil, err
	}
	fm, err := Parse(fmSrc)
	if err != nil {
		return nil, nil, err
	}
	return fm, body, nil
}

func str(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}
