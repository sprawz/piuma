package render

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"os"

	"github.com/paolobietolini/mdparser/internal/content"
)

//go:embed templates/*.html
var defaultTemplates embed.FS

// HTML marks pre-rendered markup as safe to inject into templates.
type HTML = template.HTML

// PostData is what the post page template receives.
type PostData struct {
	Post    *content.Post
	Content template.HTML // rendered markdown body
}

// IndexData is what the index page template receives.
type IndexData struct {
	Posts []*content.Post
}

// Templates renders the site's pages. Layout scheme: layout.html is the
// page skeleton; post.html and index.html fill its "content" (and
// optionally "title" and "meta") blocks.
type Templates struct {
	post  *template.Template
	index *template.Template
}

// LoadTemplates reads layout.html, post.html and index.html from dir.
// If dir is empty or missing, embedded defaults are used.
func LoadTemplates(dir string) (*Templates, error) {
	fsys, err := templateFS(dir)
	if err != nil {
		return nil, err
	}
	post, err := template.ParseFS(fsys, "layout.html", "post.html")
	if err != nil {
		return nil, err
	}
	index, err := template.ParseFS(fsys, "layout.html", "index.html")
	if err != nil {
		return nil, err
	}
	return &Templates{post: post, index: index}, nil
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

// Index renders the post listing page.
func (t *Templates) Index(w io.Writer, data IndexData) error {
	return t.index.ExecuteTemplate(w, "layout.html", data)
}
