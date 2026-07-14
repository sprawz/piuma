package content

import (
	"strings"
	"testing"
	"time"
)

const samplePost = `---
publishDate: 2026-07-07T16:00:00Z
title: 'Writing hexdump from scratch'
excerpt: 'A hexdump rewrite.'
category: code
tags:
  - programming
  - C
metadata:
  description: 'Meta description.'
---

# Writing hexdump from scratch

Body text.
`

func TestParsePost(t *testing.T) {
	p, err := ParsePost([]byte(samplePost), "writing-hexdump-from-scratch")
	if err != nil {
		t.Fatalf("ParsePost: %v", err)
	}
	if p.Title != "Writing hexdump from scratch" {
		t.Errorf("Title = %q", p.Title)
	}
	if p.Excerpt != "A hexdump rewrite." {
		t.Errorf("Excerpt = %q", p.Excerpt)
	}
	if p.Description != "Meta description." {
		t.Errorf("Description = %q", p.Description)
	}
	want := time.Date(2026, 7, 7, 16, 0, 0, 0, time.UTC)
	if !p.PublishDate.Equal(want) {
		t.Errorf("PublishDate = %v, want %v", p.PublishDate, want)
	}
	if len(p.Tags) != 2 || p.Tags[0] != "programming" || p.Tags[1] != "C" {
		t.Errorf("Tags = %v", p.Tags)
	}
	if p.URL() != "/words/writing-hexdump-from-scratch" {
		t.Errorf("URL = %q", p.URL())
	}
	if !strings.Contains(string(p.Body), "Body text.") {
		t.Errorf("Body missing text: %q", p.Body)
	}
}

func TestParsePostRequiredFields(t *testing.T) {
	for name, src := range map[string]string{
		"missing title": "---\npublishDate: 2026-07-07T16:00:00Z\n---\nbody\n",
		"missing date":  "---\ntitle: x\n---\nbody\n",
		"bad date":      "---\ntitle: x\npublishDate: yesterday\n---\nbody\n",
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := ParsePost([]byte(src), "slug"); err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestTagSlug(t *testing.T) {
	for tag, want := range map[string]string{
		"Google Tag Manager": "google-tag-manager",
		"C":                  "c",
		"learn-in-public":    "learn-in-public",
		"GA4 / GTM":          "ga4-gtm",
		"  spaced  ":         "spaced",
	} {
		if got := TagSlug(tag); got != want {
			t.Errorf("TagSlug(%q) = %q, want %q", tag, got, want)
		}
	}
}

func TestParsePage(t *testing.T) {
	src := []byte("---\ntitle: 'About'\ndescription: 'Who I am.'\n---\n# About\n")
	p, err := ParsePage(src, "about")
	if err != nil {
		t.Fatalf("ParsePage: %v", err)
	}
	if p.Title != "About" || p.Description != "Who I am." {
		t.Errorf("Page = %+v", p)
	}
	if p.URL() != "/about" {
		t.Errorf("URL = %q", p.URL())
	}
	if home, _ := ParsePage(src, "index"); home.URL() != "/" {
		t.Errorf("index URL = %q", home.URL())
	}
}

func TestParsePageRequiresTitle(t *testing.T) {
	if _, err := ParsePage([]byte("---\ndescription: x\n---\nbody\n"), "about"); err == nil {
		t.Error("expected error for missing title")
	}
}
