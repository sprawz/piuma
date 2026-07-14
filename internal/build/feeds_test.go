package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func feedsFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "content/older.md"),
		"---\npublishDate: 2025-01-01T00:00:00Z\ntitle: 'Q&A <old>'\nexcerpt: 'Old & dusty.'\n---\nbody\n")
	writeFile(t, filepath.Join(root, "content/newer.md"),
		"---\npublishDate: 2026-02-02T00:00:00Z\ntitle: 'Newer'\ntags:\n  - go\n---\nbody\n")
	writeFile(t, filepath.Join(root, "pages/index.md"),
		"---\ntitle: 'My Site'\n---\nhello\n")
	writeFile(t, filepath.Join(root, "pages/about.md"),
		"---\ntitle: 'About'\n---\nwho\n")
	return root
}

func TestBuildWithBaseGeneratesFeeds(t *testing.T) {
	root := feedsFixture(t)
	cfg := DefaultConfig(root)
	cfg.BaseURL = "https://example.com/" // trailing slash must normalize
	if err := Build(cfg); err != nil {
		t.Fatalf("Build: %v", err)
	}
	out := cfg.OutDir

	sitemap := readFile(t, filepath.Join(out, "sitemap.xml"))
	for _, want := range []string{
		"<loc>https://example.com/</loc>",
		"<loc>https://example.com/about</loc>",
		"<loc>https://example.com/payload/</loc>",
		"<loc>https://example.com/payload/newer</loc>",
		"<lastmod>2025-01-01</lastmod>",
	} {
		if !strings.Contains(sitemap, want) {
			t.Errorf("sitemap missing %q:\n%s", want, sitemap)
		}
	}
	if strings.Contains(sitemap, "/payload/tags/") {
		t.Errorf("sitemap must not list tag pages:\n%s", sitemap)
	}
	if strings.Contains(sitemap, "example.com//") {
		t.Errorf("double slash in sitemap URLs:\n%s", sitemap)
	}

	atom := readFile(t, filepath.Join(out, "atom.xml"))
	for _, want := range []string{
		"Q&amp;A &lt;old&gt;",
		`href="https://example.com/payload/newer"`,
		"<updated>2026-02-02T00:00:00Z</updated>",
	} {
		if !strings.Contains(atom, want) {
			t.Errorf("atom missing %q:\n%s", want, atom)
		}
	}
	if strings.Index(atom, "Newer") > strings.Index(atom, "old") {
		t.Errorf("atom entries not newest first:\n%s", atom)
	}

	llms := readFile(t, filepath.Join(out, "llms.txt"))
	for _, want := range []string{
		"# My Site",
		"- [Newer](https://example.com/payload/newer)",
		"- [Q&A <old>](https://example.com/payload/older): Old & dusty.",
	} {
		if !strings.Contains(llms, want) {
			t.Errorf("llms.txt missing %q:\n%s", want, llms)
		}
	}
}

func TestBuildWithoutBaseSkipsFeeds(t *testing.T) {
	root := feedsFixture(t)
	if err := Build(DefaultConfig(root)); err != nil {
		t.Fatalf("Build: %v", err)
	}
	for _, f := range []string{"sitemap.xml", "atom.xml", "llms.txt"} {
		if _, err := os.Stat(filepath.Join(root, "public", f)); !os.IsNotExist(err) {
			t.Errorf("%s generated without -base: %v", f, err)
		}
	}
}
