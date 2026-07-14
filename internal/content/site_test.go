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

func page(title string) string {
	return "---\ntitle: '" + title + "'\n---\nbody\n"
}

func TestLoad(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "content/older.md"), post("2025-01-01T00:00:00Z", "Older"))
	writeFile(t, filepath.Join(root, "content/sub/newer.md"), post("2026-01-01T00:00:00Z", "Newer"))
	writeFile(t, filepath.Join(root, "content/notes.txt"), "not markdown")
	writeFile(t, filepath.Join(root, "pages/index.md"), page("Home"))
	writeFile(t, filepath.Join(root, "pages/about.md"), page("About"))

	site, err := Load(filepath.Join(root, "content"), filepath.Join(root, "pages"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(site.Posts) != 2 {
		t.Fatalf("got %d posts, want 2", len(site.Posts))
	}
	if site.Posts[0].Title != "Newer" || site.Posts[1].Title != "Older" {
		t.Errorf("posts not sorted newest first: %q, %q", site.Posts[0].Title, site.Posts[1].Title)
	}
	if got := site.Posts[1].URL(); got != "/words/older" {
		t.Errorf("URL = %q, want /words/older", got)
	}
	if len(site.Pages) != 2 {
		t.Fatalf("got %d pages, want 2", len(site.Pages))
	}
	if home := site.Home(); home == nil || home.Title != "Home" {
		t.Errorf("Home() = %+v", site.Home())
	}
}

func TestLoadWithoutPagesDir(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "content/a.md"), post("2026-01-01T00:00:00Z", "A"))
	site, err := Load(filepath.Join(root, "content"), filepath.Join(root, "pages"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(site.Posts) != 1 || site.Home() != nil {
		t.Errorf("Posts=%d Home=%v", len(site.Posts), site.Home())
	}
}

func TestLoadCollectsAllErrors(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "content/bad1.md"), "no frontmatter")
	writeFile(t, filepath.Join(root, "content/bad2.md"), "---\ntitle: x\n---\nno date\n")
	writeFile(t, filepath.Join(root, "content/tags.md"), post("2026-01-01T00:00:00Z", "Reserved"))
	writeFile(t, filepath.Join(root, "content/dup.md"), post("2026-01-01T00:00:00Z", "Dup"))
	writeFile(t, filepath.Join(root, "content/sub/dup.md"), post("2026-01-01T00:00:00Z", "Dup2"))
	writeFile(t, filepath.Join(root, "pages/words.md"), page("Reserved"))
	writeFile(t, filepath.Join(root, "pages/untitled.md"), "---\ndescription: x\n---\nbody\n")

	_, err := Load(filepath.Join(root, "content"), filepath.Join(root, "pages"))
	if err == nil {
		t.Fatal("expected errors")
	}
	for _, want := range []string{"bad1.md", "bad2.md", "reserved post slug", "duplicate slug", "reserved page slug", "untitled.md"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %q: %v", want, err)
		}
	}
}
