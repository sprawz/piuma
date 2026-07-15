package render

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/sprawz/piuma/internal/content"
)

//go:embed templates/*.html
var defaultTemplates embed.FS

// HTML marks pre-rendered markup as safe to inject into templates.
type HTML = template.HTML

// PostData is what the post page template receives.
type PostData struct {
	Post    *content.Post
	Content HTML // rendered markdown body
}

// PageData is what a standalone page template receives.
type PageData struct {
	Page    *content.Page
	Content HTML
}

// IndexData is what a listing template receives: the full blog index
// (empty Heading) or a tag listing (Heading is the tag name).
type IndexData struct {
	Heading string
	Posts   []*content.Post
}

// funcs is available in all templates. blogroot is the blog's absolute
// path ("/blog"), so templates never spell the segment out themselves.
var funcs = template.FuncMap{
	"tagslug":  content.TagSlug,
	"blogroot": func() string { return "/" + content.BlogRoot },
}

// Templates renders the site's pages. Layout scheme: layout.html is the
// page skeleton; post.html, page.html and index.html fill its "content"
// (and optionally "title" and "meta") blocks.
type Templates struct {
	post  *template.Template
	page  *template.Template
	index *template.Template
}

// LoadTemplates reads layout.html, post.html, page.html and index.html.
// Each file is looked up in dir first and falls back to the embedded
// default, so a site only keeps the templates it actually customizes.
func LoadTemplates(dir string) (*Templates, error) {
	layout, err := readTemplate(dir, "layout.html")
	if err != nil {
		return nil, err
	}
	t := &Templates{}
	for name, dst := range map[string]**template.Template{
		"post.html":  &t.post,
		"page.html":  &t.page,
		"index.html": &t.index,
	} {
		src, err := readTemplate(dir, name)
		if err != nil {
			return nil, err
		}
		tpl, err := template.New("layout.html").Funcs(funcs).Parse(string(layout))
		if err != nil {
			return nil, fmt.Errorf("layout.html: %w", err)
		}
		if _, err := tpl.New(name).Parse(string(src)); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
		*dst = tpl
	}
	return t, nil
}

// readTemplate returns dir/name when it exists, the embedded default
// otherwise.
func readTemplate(dir, name string) ([]byte, error) {
	if dir != "" {
		src, err := os.ReadFile(filepath.Join(dir, name))
		if err == nil {
			return src, nil
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return defaultTemplates.ReadFile("templates/" + name)
}

// Post renders a single post page.
func (t *Templates) Post(w io.Writer, data PostData) error {
	return t.post.ExecuteTemplate(w, "layout.html", data)
}

// Page renders a standalone page.
func (t *Templates) Page(w io.Writer, data PageData) error {
	return t.page.ExecuteTemplate(w, "layout.html", data)
}

// Index renders a post listing.
func (t *Templates) Index(w io.Writer, data IndexData) error {
	return t.index.ExecuteTemplate(w, "layout.html", data)
}
