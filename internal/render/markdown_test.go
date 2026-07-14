package render

import (
	"strings"
	"testing"
)

func TestMarkdown(t *testing.T) {
	got, err := Markdown([]byte("# Title\n\nSome *text*.\n"))
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}
	html := string(got)
	for _, want := range []string{"<h1", "Title", "<em>text</em>"} {
		if !strings.Contains(html, want) {
			t.Errorf("output missing %q:\n%s", want, html)
		}
	}
}

func TestMarkdownKeepsRawHTML(t *testing.T) {
	// Posts embed iframes and other raw HTML; it must pass through.
	got, err := Markdown([]byte("<iframe src=\"https://example.com\"></iframe>\n"))
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}
	if !strings.Contains(string(got), "<iframe") {
		t.Errorf("raw HTML stripped:\n%s", got)
	}
}

func TestMarkdownTables(t *testing.T) {
	got, err := Markdown([]byte("| a | b |\n|---|---|\n| 1 | 2 |\n"))
	if err != nil {
		t.Fatalf("Markdown: %v", err)
	}
	if !strings.Contains(string(got), "<table>") {
		t.Errorf("GFM table not rendered:\n%s", got)
	}
}
