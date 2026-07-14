package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, data string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
}

func post(date, title string) string {
	return "---\npublishDate: " + date + "\ntitle: '" + title + "'\n---\nbody\n"
}

func TestLoadSite(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "code/older.md"), post("2025-01-01T00:00:00Z", "Older"))
	writeFile(t, filepath.Join(root, "analytics/newer.md"), post("2026-01-01T00:00:00Z", "Newer"))
	writeFile(t, filepath.Join(root, "code/notes.txt"), "not markdown")

	site, err := LoadSite(root)
	if err != nil {
		t.Fatalf("LoadSite: %v", err)
	}
	if len(site.Posts) != 2 {
		t.Fatalf("got %d posts, want 2", len(site.Posts))
	}
	if site.Posts[0].Title != "Newer" || site.Posts[1].Title != "Older" {
		t.Errorf("posts not sorted newest first: %q, %q", site.Posts[0].Title, site.Posts[1].Title)
	}
	if got := site.Posts[1].URL(); got != "/code/older" {
		t.Errorf("URL = %q, want /code/older", got)
	}
}

func TestLoadSiteCollectsAllErrors(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "code/bad1.md"), "no frontmatter")
	writeFile(t, filepath.Join(root, "code/bad2.md"), "---\ntitle: x\n---\nno date\n")
	writeFile(t, filepath.Join(root, "stray.md"), post("2026-01-01T00:00:00Z", "Stray"))

	_, err := LoadSite(root)
	if err == nil {
		t.Fatal("expected errors")
	}
	for _, want := range []string{"bad1.md", "bad2.md", "stray.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %s: %v", want, err)
		}
	}
}
