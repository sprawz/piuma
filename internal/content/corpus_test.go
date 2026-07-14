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
	root := filepath.Join(home, "code/paolobietolini.com/src/content/blog")
	if _, err := os.Stat(root); err != nil {
		t.Skipf("real blog not available: %v", err)
	}
	n := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(path) != ".md" {
			return err
		}
		n++
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		fm, _, err := Split(src)
		if err != nil {
			t.Errorf("%s: %v", path, err)
			return nil
		}
		if _, err := Parse(fm); err != nil {
			t.Errorf("%s: %v", path, err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("parsed %d posts", n)
}
