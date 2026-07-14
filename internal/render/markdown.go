// Package render turns loaded content into HTML: markdown conversion and
// page templates. Goldmark is confined to this file so replacing it later
// (a planned rewrite) touches nothing else.
package render

import (
	"bytes"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	// Posts contain raw HTML (iframes, embeds); trusted input, keep it.
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

// Markdown converts markdown source to HTML.
func Markdown(src []byte) ([]byte, error) {
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
