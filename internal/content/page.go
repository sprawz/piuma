package content

import "fmt"

// Page is a standalone markdown page (about, now, the homepage).
// Unlike posts, pages are timeless: no publish date, no tags.
type Page struct {
	Slug        string // "index" is the homepage
	Title       string
	Description string
	Body        []byte
}

// URL is the page's site-absolute path.
func (p *Page) URL() string {
	if p.Slug == "index" {
		return "/"
	}
	return "/" + p.Slug
}

// ParsePage builds a Page from a raw markdown document.
func ParsePage(src []byte, slug string) (*Page, error) {
	fm, body, err := parseDoc(src)
	if err != nil {
		return nil, err
	}
	p := &Page{
		Slug:  slug,
		Title: str(fm, "title"),
		Body:  body,
	}
	if p.Title == "" {
		return nil, fmt.Errorf("missing required field %q", "title")
	}
	if meta, ok := fm["metadata"].(map[string]any); ok {
		p.Description = str(meta, "description")
	}
	if p.Description == "" {
		p.Description = str(fm, "description")
	}
	return p, nil
}
