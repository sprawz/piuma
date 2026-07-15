package build

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

func fixtureSite(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "content/hello.md"),
		"---\npublishDate: 2026-07-07T16:00:00Z\ntitle: 'Hello'\nexcerpt: 'Hi.'\ntags:\n  - Google Tag Manager\n---\n# Hello\n\nSome *body*.\n")
	writeFile(t, filepath.Join(root, "static/robots.txt"), "User-agent: *\n")
	return root
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func TestBuildWithoutHomepage(t *testing.T) {
	root := fixtureSite(t)
	out := filepath.Join(root, "public")
	if err := Build(DefaultConfig(root)); err != nil {
		t.Fatalf("Build: %v", err)
	}

	postHTML := readFile(t, filepath.Join(out, "blog/hello/index.html"))
	for _, want := range []string{"Hello", "<em>body</em>", `href="/blog/tags/google-tag-manager/"`} {
		if !strings.Contains(postHTML, want) {
			t.Errorf("post page missing %q", want)
		}
	}
	// No pages/index.md: / and /blog/ both list posts.
	for _, p := range []string{"index.html", "blog/index.html"} {
		if html := readFile(t, filepath.Join(out, p)); !strings.Contains(html, `href="/blog/hello"`) {
			t.Errorf("%s missing post link:\n%s", p, html)
		}
	}
	tagHTML := readFile(t, filepath.Join(out, "blog/tags/google-tag-manager/index.html"))
	for _, want := range []string{"Google Tag Manager", `href="/blog/hello"`} {
		if !strings.Contains(tagHTML, want) {
			t.Errorf("tag page missing %q", want)
		}
	}
	if got := readFile(t, filepath.Join(out, "robots.txt")); got != "User-agent: *\n" {
		t.Errorf("static file not copied verbatim: %q", got)
	}
}

func TestBuildWithHomepage(t *testing.T) {
	root := fixtureSite(t)
	writeFile(t, filepath.Join(root, "pages/index.md"),
		"---\ntitle: 'Paolo Bietolini'\n---\nBuilding analytics infrastructure.\n")
	writeFile(t, filepath.Join(root, "pages/about.md"),
		"---\ntitle: 'About'\n---\nWho I am.\n")
	out := filepath.Join(root, "public")
	if err := Build(DefaultConfig(root)); err != nil {
		t.Fatalf("Build: %v", err)
	}

	home := readFile(t, filepath.Join(out, "index.html"))
	if !strings.Contains(home, "Building analytics infrastructure.") {
		t.Errorf("homepage not rendered from pages/index.md:\n%s", home)
	}
	if strings.Contains(home, `href="/blog/hello"`) {
		t.Errorf("homepage should not be the post listing")
	}
	if about := readFile(t, filepath.Join(out, "about/index.html")); !strings.Contains(about, "Who I am.") {
		t.Errorf("about page wrong:\n%s", about)
	}
	if listing := readFile(t, filepath.Join(out, "blog/index.html")); !strings.Contains(listing, `href="/blog/hello"`) {
		t.Errorf("/blog/ listing missing post")
	}
}

func TestBuildKeepsOldOutputOnRenderError(t *testing.T) {
	root := fixtureSite(t)
	if err := Build(DefaultConfig(root)); err != nil {
		t.Fatal(err)
	}
	good := filepath.Join(root, "public/blog/hello/index.html")
	if _, err := os.Stat(good); err != nil {
		t.Fatal(err)
	}
	// Break a template: the build must fail without touching public/.
	writeFile(t, filepath.Join(root, "templates/post.html"),
		`{{define "content"}}{{.Post.NoSuchField}}{{end}}`)
	if err := Build(DefaultConfig(root)); err == nil {
		t.Fatal("expected render error")
	}
	if _, err := os.Stat(good); err != nil {
		t.Errorf("previous output destroyed by failed build: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "public.staging")); !os.IsNotExist(err) {
		t.Errorf("staging directory left behind: %v", err)
	}
}

func TestBuildRefusesOutOverSources(t *testing.T) {
	root := fixtureSite(t)
	for _, out := range []string{root, filepath.Join(root, "content")} {
		cfg := DefaultConfig(root)
		cfg.OutDir = out
		if err := Build(cfg); err == nil || !strings.Contains(err.Error(), "source") {
			t.Errorf("out=%s: expected source-guard error, got %v", out, err)
		}
	}
	if _, err := os.Stat(filepath.Join(root, "content/hello.md")); err != nil {
		t.Fatalf("sources damaged: %v", err)
	}
}

func TestBuildSkipsEmptyTagSlug(t *testing.T) {
	root := fixtureSite(t)
	writeFile(t, filepath.Join(root, "content/odd.md"),
		"---\npublishDate: 2026-01-01T00:00:00Z\ntitle: 'Odd'\ntags:\n  - '!!!'\n---\nbody\n")
	if err := Build(DefaultConfig(root)); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "public/blog/tags/index.html")); !os.IsNotExist(err) {
		t.Errorf("empty tag slug produced phantom listing: %v", err)
	}
}

func TestBuildWipesStaleOutput(t *testing.T) {
	root := fixtureSite(t)
	stale := filepath.Join(root, "public/old/index.html")
	writeFile(t, stale, "stale")
	if err := Build(DefaultConfig(root)); err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Errorf("stale output survived: %v", err)
	}
}

func TestBuildReportsContentErrors(t *testing.T) {
	root := fixtureSite(t)
	writeFile(t, filepath.Join(root, "content/broken.md"), "no frontmatter")
	err := Build(DefaultConfig(root))
	if err == nil || !strings.Contains(err.Error(), "broken.md") {
		t.Fatalf("expected error naming broken.md, got %v", err)
	}
}
