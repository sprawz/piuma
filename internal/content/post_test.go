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
	p, err := ParsePost([]byte(samplePost), "code", "writing-hexdump-from-scratch")
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
	if got := []string{"programming", "C"}; len(p.Tags) != 2 || p.Tags[0] != got[0] || p.Tags[1] != got[1] {
		t.Errorf("Tags = %v", p.Tags)
	}
	if p.URL() != "/code/writing-hexdump-from-scratch" {
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
			if _, err := ParsePost([]byte(src), "code", "slug"); err == nil {
				t.Error("expected error")
			}
		})
	}
}
