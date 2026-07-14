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

// Site is the loaded content tree: all posts, newest first.
type Site struct {
	Posts []*Post
}

// LoadSite walks root expecting <category>/<slug>.md files. It parses every
// post and reports all broken files at once, each error prefixed with its
// file path.
func LoadSite(root string) (*Site, error) {
	site := &Site{}
	var errs []error
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		p, err := loadPost(path, rel)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", path, err))
			return nil
		}
		site.Posts = append(site.Posts, p)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	sort.Slice(site.Posts, func(i, j int) bool {
		return site.Posts[i].PublishDate.After(site.Posts[j].PublishDate)
	})
	return site, nil
}

func loadPost(path, rel string) (*Post, error) {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) != 2 {
		return nil, errors.New("expected content/<category>/<slug>.md layout")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParsePost(src, parts[0], strings.TrimSuffix(parts[1], ".md"))
}
