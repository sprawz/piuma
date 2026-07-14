package content

import (
	"fmt"
	"time"
)

// Post is one markdown article. Category and slug come from the file's
// location on disk (content/<category>/<slug>.md), the rest from frontmatter.
type Post struct {
	Slug        string
	Category    string
	Title       string
	Excerpt     string
	Description string // meta description, falls back to Excerpt
	Tags        []string
	PublishDate time.Time
	Body        []byte // markdown body, without frontmatter
}

// URL is the post's site-absolute path.
func (p *Post) URL() string {
	return "/" + p.Category + "/" + p.Slug
}

// ParsePost builds a Post from a raw markdown document.
func ParsePost(src []byte, category, slug string) (*Post, error) {
	fmSrc, body, err := Split(src)
	if err != nil {
		return nil, err
	}
	fm, err := Parse(fmSrc)
	if err != nil {
		return nil, err
	}
	p := &Post{
		Slug:     slug,
		Category: category,
		Title:    str(fm, "title"),
		Excerpt:  str(fm, "excerpt"),
		Body:     body,
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

func str(m map[string]any, key string) string {
	s, _ := m[key].(string)
	return s
}
