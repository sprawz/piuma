package content

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRealBlogCorpus parses every post of the real blog when present on
// this machine, guarding the frontmatter subset against real-world files.
func TestRealBlogCorpus(t *testing.T) {
	home, _ := os.UserHomeDir()
	root := firstExisting(
		filepath.Join(home, "code/paolobietolini.com/content"),
		filepath.Join(home, "code/paolobietolini.com/src/content/blog"),
	)
	if root == "" {
		t.Skip("real blog not available")
	}
	n := 0
	for _, e := range walkMarkdown(root, func(path, slug string) error {
		n++
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = ParsePost(src, slug)
		return err
	}) {
		t.Error(e)
	}
	t.Logf("parsed %d posts from %s", n, root)
}

func firstExisting(paths ...string) string {
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}
