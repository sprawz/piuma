package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/paolobietolini/mdparser/internal/content"
)

func samplePost() *content.Post {
	return &content.Post{
		Slug:        "hello",
		Category:    "code",
		Title:       "Hello World",
		Excerpt:     "An excerpt.",
		Description: "A description.",
		PublishDate: time.Date(2026, 7, 7, 16, 0, 0, 0, time.UTC),
	}
}

func TestEmbeddedPostTemplate(t *testing.T) {
	tpl, err := LoadTemplates("")
	if err != nil {
		t.Fatalf("LoadTemplates: %v", err)
	}
	var b strings.Builder
	err = tpl.Post(&b, PostData{Post: samplePost(), Content: "<p>hi <em>there</em></p>"})
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	out := b.String()
	for _, want := range []string{"Hello World", "<em>there</em>", "A description."} {
		if !strings.Contains(out, want) {
			t.Errorf("post output missing %q:\n%s", want, out)
		}
	}
}

func TestEmbeddedIndexTemplate(t *testing.T) {
	tpl, err := LoadTemplates("")
	if err != nil {
		t.Fatalf("LoadTemplates: %v", err)
	}
	var b strings.Builder
	if err := tpl.Index(&b, IndexData{Posts: []*content.Post{samplePost()}}); err != nil {
		t.Fatalf("Index: %v", err)
	}
	out := b.String()
	for _, want := range []string{"Hello World", `href="/code/hello"`} {
		if !strings.Contains(out, want) {
			t.Errorf("index output missing %q:\n%s", want, out)
		}
	}
}

func TestLoadTemplatesFromDir(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"layout.html": `<html><body>{{block "content" .}}{{end}}</body></html>`,
		"post.html":   `{{define "content"}}CUSTOM {{.Post.Title}}{{end}}`,
		"index.html":  `{{define "content"}}CUSTOM INDEX{{end}}`,
	}
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	tpl, err := LoadTemplates(dir)
	if err != nil {
		t.Fatalf("LoadTemplates: %v", err)
	}
	var b strings.Builder
	if err := tpl.Post(&b, PostData{Post: samplePost()}); err != nil {
		t.Fatalf("Post: %v", err)
	}
	if !strings.Contains(b.String(), "CUSTOM Hello World") {
		t.Errorf("custom template not used:\n%s", b.String())
	}
}
