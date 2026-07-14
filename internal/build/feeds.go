package build

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/paolobietolini/piuma/internal/content"
)

// writeFeeds emits sitemap.xml, atom.xml and llms.txt at the output
// root. base is the site's absolute URL, already normalized (no
// trailing slash). robots.txt is deliberately not generated: it is
// hand-written content that belongs in static/.
func writeFeeds(outDir, base string, site *content.Site) error {
	for name, write := range map[string]func(*os.File) error{
		"sitemap.xml": func(f *os.File) error { return writeSitemap(f, base, site) },
		"atom.xml":    func(f *os.File) error { return writeAtom(f, base, site) },
		"llms.txt":    func(f *os.File) error { return writeLLMs(f, base, site) },
	} {
		if err := writePage(filepath.Join(outDir, name), write); err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
	}
	return nil
}

type sitemapURL struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type sitemap struct {
	XMLName xml.Name     `xml:"urlset"`
	NS      string       `xml:"xmlns,attr"`
	URLs    []sitemapURL `xml:"url"`
}

// writeSitemap lists the homepage, standalone pages, the blog index and
// every post. Tag listings are excluded: index noise.
func writeSitemap(f *os.File, base string, site *content.Site) error {
	sm := sitemap{NS: "http://www.sitemaps.org/schemas/sitemap/0.9"}
	sm.URLs = append(sm.URLs, sitemapURL{Loc: base + "/"})
	for _, p := range site.Pages {
		if p.Slug != "index" {
			sm.URLs = append(sm.URLs, sitemapURL{Loc: base + p.URL()})
		}
	}
	sm.URLs = append(sm.URLs, sitemapURL{Loc: base + "/payload/"})
	for _, p := range site.Posts {
		sm.URLs = append(sm.URLs, sitemapURL{
			Loc:     base + p.URL(),
			LastMod: p.PublishDate.Format("2006-01-02"),
		})
	}
	return writeXML(f, sm)
}

type atomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
}

type atomEntry struct {
	Title   string   `xml:"title"`
	Link    atomLink `xml:"link"`
	ID      string   `xml:"id"`
	Updated string   `xml:"updated"`
	Summary string   `xml:"summary,omitempty"`
}

type atomFeed struct {
	XMLName xml.Name    `xml:"feed"`
	NS      string      `xml:"xmlns,attr"`
	Title   string      `xml:"title"`
	Link    atomLink    `xml:"link"`
	ID      string      `xml:"id"`
	Updated string      `xml:"updated"`
	Entries []atomEntry `xml:"entry"`
}

// writeAtom emits an Atom 1.0 feed of all posts, newest first (the
// site loader already sorts them). Entries carry the excerpt, not the
// full body.
func writeAtom(f *os.File, base string, site *content.Site) error {
	feed := atomFeed{
		NS:      "http://www.w3.org/2005/Atom",
		Title:   siteTitle(site),
		Link:    atomLink{Href: base + "/payload/", Rel: "alternate"},
		ID:      base + "/",
		Updated: time.Now().UTC().Format(time.RFC3339),
	}
	if len(site.Posts) > 0 {
		feed.Updated = site.Posts[0].PublishDate.UTC().Format(time.RFC3339)
	}
	for _, p := range site.Posts {
		feed.Entries = append(feed.Entries, atomEntry{
			Title:   p.Title,
			Link:    atomLink{Href: base + p.URL()},
			ID:      base + p.URL(),
			Updated: p.PublishDate.UTC().Format(time.RFC3339),
			Summary: p.Excerpt,
		})
	}
	return writeXML(f, feed)
}

// writeLLMs emits the llms.txt convention: the site title, then one
// markdown link line per post.
func writeLLMs(f *os.File, base string, site *content.Site) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n\n", siteTitle(site))
	for _, p := range site.Posts {
		fmt.Fprintf(&b, "- [%s](%s%s)", p.Title, base, p.URL())
		if p.Excerpt != "" {
			fmt.Fprintf(&b, ": %s", p.Excerpt)
		}
		b.WriteString("\n")
	}
	_, err := f.WriteString(b.String())
	return err
}

// siteTitle is the homepage title when one exists, "Blog" otherwise.
func siteTitle(site *content.Site) string {
	if home := site.Home(); home != nil {
		return home.Title
	}
	return "Blog"
}

func writeXML(f *os.File, v any) error {
	if _, err := f.WriteString(xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(f)
	enc.Indent("", "  ")
	return enc.Encode(v)
}
