// Package build orchestrates a full site build: load content, render
// pages, copy static files into the output directory.
package build

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sprawz/piuma/internal/content"
	"github.com/sprawz/piuma/internal/render"
)

// Config holds the directories of a site project.
type Config struct {
	ContentDir  string
	PagesDir    string
	TemplateDir string
	StaticDir   string
	OutDir      string
	// BaseURL is the site's absolute URL (https://example.com). When
	// set, sitemap.xml, atom.xml and llms.txt are generated at the
	// output root; when empty they are skipped.
	BaseURL string
}

// DefaultConfig returns the conventional layout rooted at dir:
// content/, pages/, templates/, static/, public/.
func DefaultConfig(dir string) Config {
	return Config{
		ContentDir:  filepath.Join(dir, "content"),
		PagesDir:    filepath.Join(dir, "pages"),
		TemplateDir: filepath.Join(dir, "templates"),
		StaticDir:   filepath.Join(dir, "static"),
		OutDir:      filepath.Join(dir, "public"),
	}
}

// Build renders the whole site into cfg.OutDir. The output directory is
// wiped first: it is disposable by contract.
//
// Output layout:
//
//	/                        homepage (pages/index.md) or the post listing
//	/<blog>/                 post listing (always)
//	/<blog>/<slug>/          posts
//	/<blog>/tags/<tag>/      tag listings
//	/<page>/                 standalone pages
//
// <blog> is content.BlogRoot.
//
// The site is rendered into a staging directory next to OutDir and
// swapped in only on success, so a failing build never destroys the
// previous output.
func Build(cfg Config) error {
	if err := guardOutDir(cfg); err != nil {
		return err
	}
	site, err := content.Load(cfg.ContentDir, cfg.PagesDir)
	if err != nil {
		return err
	}
	tpl, err := render.LoadTemplates(cfg.TemplateDir)
	if err != nil {
		return err
	}
	stage := cfg.OutDir + ".staging"
	if err := os.RemoveAll(stage); err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(stage) }()
	if err := renderSite(stage, cfg, tpl, site); err != nil {
		return err
	}
	if err := os.RemoveAll(cfg.OutDir); err != nil {
		return err
	}
	return os.Rename(stage, cfg.OutDir)
}

// Validate runs the entire build pipeline — load, templates, render —
// into a throwaway directory and discards the result. It is the check
// behind `piuma format`: if Validate passes, Build will too.
func Validate(cfg Config) (*content.Site, error) {
	site, err := content.Load(cfg.ContentDir, cfg.PagesDir)
	if err != nil {
		return nil, err
	}
	tpl, err := render.LoadTemplates(cfg.TemplateDir)
	if err != nil {
		return nil, err
	}
	dir, err := os.MkdirTemp("", "piuma-validate-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	// A dummy base exercises the feed writers even when the real build
	// will run without -base.
	cfg.BaseURL = "https://validate.invalid"
	if err := renderSite(dir, cfg, tpl, site); err != nil {
		return nil, err
	}
	return site, nil
}

// guardOutDir refuses output locations that would make the build delete
// its own sources (e.g. -out pointing at the site root).
func guardOutDir(cfg Config) error {
	out, err := filepath.Abs(cfg.OutDir)
	if err != nil {
		return err
	}
	for _, src := range []string{cfg.ContentDir, cfg.PagesDir, cfg.TemplateDir, cfg.StaticDir} {
		abs, err := filepath.Abs(src)
		if err != nil {
			return err
		}
		if abs == out || strings.HasPrefix(abs, out+string(filepath.Separator)) {
			return fmt.Errorf("output directory %s would overwrite source directory %s", cfg.OutDir, src)
		}
	}
	return nil
}

func renderSite(outDir string, cfg Config, tpl *render.Templates, site *content.Site) error {
	if err := copyStatic(cfg.StaticDir, outDir); err != nil {
		return err
	}
	for _, p := range site.Posts {
		if err := writePost(outDir, tpl, p); err != nil {
			return fmt.Errorf("%s: %w", p.URL(), err)
		}
	}
	for _, p := range site.Pages {
		if err := writeStandalone(outDir, tpl, p); err != nil {
			return fmt.Errorf("%s: %w", p.URL(), err)
		}
	}
	if err := writeListings(outDir, tpl, site); err != nil {
		return err
	}
	if base := strings.TrimRight(cfg.BaseURL, "/"); base != "" {
		return writeFeeds(outDir, base, site)
	}
	return nil
}

func writePost(outDir string, tpl *render.Templates, p *content.Post) error {
	body, err := render.Markdown(p.Body)
	if err != nil {
		return err
	}
	path := filepath.Join(outDir, content.BlogRoot, p.Slug, "index.html")
	return writePage(path, func(f *os.File) error {
		return tpl.Post(f, render.PostData{Post: p, Content: render.HTML(body)})
	})
}

func writeStandalone(outDir string, tpl *render.Templates, p *content.Page) error {
	body, err := render.Markdown(p.Body)
	if err != nil {
		return err
	}
	path := filepath.Join(outDir, p.Slug, "index.html")
	if p.Slug == "index" {
		path = filepath.Join(outDir, "index.html")
	}
	return writePage(path, func(f *os.File) error {
		return tpl.Page(f, render.PageData{Page: p, Content: render.HTML(body)})
	})
}

// writeListings renders the blog index at /<BlogRoot>/, one listing per
// tag under /<BlogRoot>/tags/, and — when no homepage page exists — the
// blog index again at / so zero-config sites keep working.
func writeListings(outDir string, tpl *render.Templates, site *content.Site) error {
	all := render.IndexData{Posts: site.Posts}
	if err := writeIndex(filepath.Join(outDir, content.BlogRoot), tpl, all); err != nil {
		return err
	}
	if site.Home() == nil {
		if err := writeIndex(outDir, tpl, all); err != nil {
			return err
		}
	}
	for slug, tag := range tagIndex(site.Posts) {
		data := render.IndexData{Heading: tag.name, Posts: tag.posts}
		if err := writeIndex(filepath.Join(outDir, content.BlogRoot, "tags", slug), tpl, data); err != nil {
			return err
		}
	}
	return nil
}

type tagGroup struct {
	name  string
	posts []*content.Post
}

// tagIndex groups posts by tag slug. The first spelling of a tag wins
// as the display name.
func tagIndex(posts []*content.Post) map[string]*tagGroup {
	tags := map[string]*tagGroup{}
	for _, p := range posts {
		for _, t := range p.Tags {
			slug := content.TagSlug(t)
			if slug == "" {
				continue // tag with no usable characters
			}
			if tags[slug] == nil {
				tags[slug] = &tagGroup{name: t}
			}
			tags[slug].posts = append(tags[slug].posts, p)
		}
	}
	return tags
}

func writeIndex(dir string, tpl *render.Templates, data render.IndexData) error {
	return writePage(filepath.Join(dir, "index.html"), func(f *os.File) error {
		return tpl.Index(f, data)
	})
}

func writePage(path string, render func(*os.File) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := render(f); err != nil {
		_ = f.Close() // the render error is the one worth reporting
		return err
	}
	return f.Close()
}

func copyStatic(staticDir, outDir string) error {
	if _, err := os.Stat(staticDir); os.IsNotExist(err) {
		return os.MkdirAll(outDir, 0o755)
	}
	return os.CopyFS(outDir, os.DirFS(staticDir))
}
