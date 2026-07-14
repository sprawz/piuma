package content

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Reserved slugs: /words is the blog root, /words/tags the tag listings.
const (
	reservedPageSlug = "words"
	reservedPostSlug = "tags"
)

// Site is the loaded content tree: posts newest first, pages by slug.
type Site struct {
	Posts []*Post
	Pages []*Page
}

// Home returns the homepage (pages/index.md) or nil when there is none.
func (s *Site) Home() *Page {
	for _, p := range s.Pages {
		if p.Slug == "index" {
			return p
		}
	}
	return nil
}

// Load reads posts from contentDir and standalone pages from pagesDir
// (which may be absent). It parses everything and reports all broken
// files at once, each error prefixed with its file path.
func Load(contentDir, pagesDir string) (*Site, error) {
	site := &Site{}
	var errs []error

	seen := map[string]string{} // slug → file that claimed it
	errs = append(errs, walkMarkdown(contentDir, func(path, slug string) error {
		if slug == reservedPostSlug {
			return fmt.Errorf("%q is a reserved post slug", slug)
		}
		if prev, dup := seen[slug]; dup {
			return fmt.Errorf("duplicate slug %q (also %s)", slug, prev)
		}
		seen[slug] = path
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		p, err := ParsePost(src, slug)
		if err != nil {
			return err
		}
		site.Posts = append(site.Posts, p)
		return nil
	})...)

	seenPage := map[string]string{}
	errs = append(errs, walkMarkdown(pagesDir, func(path, slug string) error {
		if slug == reservedPageSlug {
			return fmt.Errorf("%q is a reserved page slug", slug)
		}
		if prev, dup := seenPage[slug]; dup {
			return fmt.Errorf("duplicate page slug %q (also %s)", slug, prev)
		}
		seenPage[slug] = path
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		p, err := ParsePage(src, slug)
		if err != nil {
			return err
		}
		site.Pages = append(site.Pages, p)
		return nil
	})...)

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	sort.Slice(site.Posts, func(i, j int) bool {
		return site.Posts[i].PublishDate.After(site.Posts[j].PublishDate)
	})
	return site, nil
}

// walkMarkdown calls fn for every .md file under root with the file's
// slug (base name without extension). Directories carry no meaning.
// A missing root is fine — it just yields nothing. Each failing file
// becomes one returned error, prefixed with its path.
func walkMarkdown(root string, fn func(path, slug string) error) []error {
	var errs []error
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		slug := strings.TrimSuffix(filepath.Base(path), ".md")
		if err := fn(path, slug); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
		}
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		errs = append(errs, err)
	}
	return errs
}
