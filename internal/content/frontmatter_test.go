package content

import (
	"reflect"
	"testing"
)

func TestSplit(t *testing.T) {
	src := []byte("---\ntitle: hello\n---\n\n# Body\n")
	fm, body, err := Split(src)
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if got, want := string(fm), "title: hello\n"; got != want {
		t.Errorf("frontmatter = %q, want %q", got, want)
	}
	if got, want := string(body), "\n# Body\n"; got != want {
		t.Errorf("body = %q, want %q", got, want)
	}
}

func TestSplitErrors(t *testing.T) {
	for name, src := range map[string]string{
		"no frontmatter": "# Just a doc\n",
		"unterminated":   "---\ntitle: x\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := Split([]byte(src)); err == nil {
				t.Errorf("Split(%q): expected error", src)
			}
		})
	}
}

func TestSplitCRLF(t *testing.T) {
	fm, body, err := Split([]byte("---\r\ntitle: hello\r\n---\r\nBody\r\n"))
	if err != nil {
		t.Fatalf("Split: %v", err)
	}
	if got := string(fm); got != "title: hello\n" {
		t.Errorf("frontmatter = %q", got)
	}
	if got := string(body); got != "Body\n" {
		t.Errorf("body = %q", got)
	}
}

func TestParse(t *testing.T) {
	src := []byte(`publishDate: 2026-07-07T16:00:00Z
title: 'Writing hexdump from scratch'
excerpt: "Say we want to inspect a file."
category: code
tags:
  - programming
  - C
metadata:
  canonical: https://paolobietolini.com/code/writing-hexdump-from-scratch
  description: 'A description.'
`)
	got, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := map[string]any{
		"publishDate": "2026-07-07T16:00:00Z",
		"title":       "Writing hexdump from scratch",
		"excerpt":     "Say we want to inspect a file.",
		"category":    "code",
		"tags":        []string{"programming", "C"},
		"metadata": map[string]any{
			"canonical":   "https://paolobietolini.com/code/writing-hexdump-from-scratch",
			"description": "A description.",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestParseMultilineQuotedScalar(t *testing.T) {
	src := []byte("excerpt: 'First line: intro.\nSecond line.'\ncategory: code\n")
	got, err := Parse(src)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := map[string]any{
		"excerpt":  "First line: intro. Second line.",
		"category": "code",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %#v, want %#v", got, want)
	}
}

func TestParseEscapedSingleQuote(t *testing.T) {
	got, err := Parse([]byte("title: 'L''obiettivo'\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if got["title"] != "L'obiettivo" {
		t.Errorf("title = %q, want %q", got["title"], "L'obiettivo")
	}
}

func TestParseSkipsBlankLines(t *testing.T) {
	got, err := Parse([]byte("title: a\n\ncategory: b\n"))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := map[string]any{"title": "a", "category": "b"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Parse = %#v, want %#v", got, want)
	}
}

func TestParseErrors(t *testing.T) {
	for name, src := range map[string]string{
		"no colon":           "just text\n",
		"tab indent":         "tags:\n\t- a\n",
		"mixed list and map": "tags:\n  - a\n  key: v\n",
		"empty nested block": "metadata:\n",
		"scalar then indent": "title: x\n  stray: y\n",
		"unterminated quote": "excerpt: 'never closed\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(src)); err == nil {
				t.Errorf("Parse(%q): expected error", src)
			}
		})
	}
}
