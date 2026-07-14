package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/paolobietolini/piuma/internal/content"
)

func samplePost() *content.Post {
	return &content.Post{
		Slug:        "hello",
		Title:       "Hello World",
		Excerpt:     "An excerpt.",
		Description: "A description.",
		Tags:        []string{"Google Tag Manager"},
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
	for _, want := range []string{
		"Hello World", "<em>there</em>", "A description.",
		`href="/payload/tags/google-tag-manager/"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("post output missing %q:\n%s", want, out)
		}
	}
}

func TestEmbeddedPageTemplate(t *testing.T) {
	tpl, err := LoadTemplates("")
	if err != nil {
		t.Fatalf("LoadTemplates: %v", err)
	}
	pg := &content.Page{Slug: "index", Title: "Paolo Bietolini", Description: "Bio."}
	var b strings.Builder
	if err := tpl.Page(&b, PageData{Page: pg, Content: "<p>presenting</p>"}); err != nil {
		t.Fatalf("Page: %v", err)
	}
	out := b.String()
	for _, want := range []string{"Paolo Bietolini", "<p>presenting</p>", "Bio."} {
		if !strings.Contains(out, want) {
			t.Errorf("page output missing %q:\n%s", want, out)
		}
	}
}

func TestEmbeddedIndexTemplate(t *testing.T) {
	tpl, err := LoadTemplates("")
	if err != nil {
		t.Fatalf("LoadTemplates: %v", err)
	}
	var b strings.Builder
	if err := tpl.Index(&b, IndexData{Heading: "gtm", Posts: []*content.Post{samplePost()}}); err != nil {
		t.Fatalf("Index: %v", err)
	}
	out := b.String()
	for _, want := range []string{"Hello World", `href="/payload/hello"`, "<h1>gtm</h1>"} {
		if !strings.Contains(out, want) {
			t.Errorf("index output missing %q:\n%s", want, out)
		}
	}
}

func TestLoadTemplatesPartialOverride(t *testing.T) {
	// A site overriding only index.html must get embedded defaults for
	// layout, post and page — no need to copy every template.
	dir := t.TempDir()
	custom := `{{define "content"}}CUSTOM INDEX{{end}}`
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}
	tpl, err := LoadTemplates(dir)
	if err != nil {
		t.Fatalf("LoadTemplates: %v", err)
	}
	var b strings.Builder
	if err := tpl.Index(&b, IndexData{}); err != nil {
		t.Fatalf("Index: %v", err)
	}
	if !strings.Contains(b.String(), "CUSTOM INDEX") {
		t.Errorf("custom index not used:\n%s", b.String())
	}
	if !strings.Contains(b.String(), "<!doctype html>") {
		t.Errorf("embedded layout not used around custom index:\n%s", b.String())
	}
	b.Reset()
	if err := tpl.Post(&b, PostData{Post: samplePost()}); err != nil {
		t.Fatalf("Post: %v", err)
	}
	if !strings.Contains(b.String(), "Hello World") {
		t.Errorf("embedded post template broken:\n%s", b.String())
	}
}

func TestLoadTemplatesFromDir(t *testing.T) {
	dir := t.TempDir()
	files := map[string]string{
		"layout.html": `<html><body>{{block "content" .}}{{end}}</body></html>`,
		"post.html":   `{{define "content"}}CUSTOM {{.Post.Title}}{{end}}`,
		"page.html":   `{{define "content"}}CUSTOM PAGE{{end}}`,
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
