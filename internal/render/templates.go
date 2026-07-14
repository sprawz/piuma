package render

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"os"

	"github.com/paolobietolini/piuma/internal/content"
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

// funcs is available in all templates.
var funcs = template.FuncMap{"tagslug": content.TagSlug}

// Templates renders the site's pages. Layout scheme: layout.html is the
// page skeleton; post.html, page.html and index.html fill its "content"
// (and optionally "title" and "meta") blocks.
type Templates struct {
	post  *template.Template
	page  *template.Template
	index *template.Template
}

// LoadTemplates reads layout.html, post.html, page.html and index.html
// from dir. If dir is empty or missing, embedded defaults are used.
func LoadTemplates(dir string) (*Templates, error) {
	fsys, err := templateFS(dir)
	if err != nil {
		return nil, err
	}
	t := &Templates{}
	for name, dst := range map[string]**template.Template{
		"post.html":  &t.post,
		"page.html":  &t.page,
		"index.html": &t.index,
	} {
		*dst, err = template.New("").Funcs(funcs).ParseFS(fsys, "layout.html", name)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func templateFS(dir string) (fs.FS, error) {
	if dir != "" {
		if _, err := os.Stat(dir); err == nil {
			return os.DirFS(dir), nil
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return fs.Sub(defaultTemplates, "templates")
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
