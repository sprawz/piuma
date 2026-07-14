# piuma

A minimal static site builder: markdown in, HTML out. Written in Go,
almost everything standard library.

Piuma (Italian for *feather*) exists because a personal blog does not
need a JavaScript framework. It replaces an Astro setup with one small
binary: no node_modules, no config files, no plugins. You write
markdown, it writes HTML.

## Goal and philosophy

- **KISS / YAGNI.** Small functions, small files, one job per package.
  Features are added when a real site needs them, not before.
- **Few dependencies.** The only third-party module is
  [goldmark](https://github.com/yuin/goldmark) (markdown → HTML), and
  it is deliberately isolated in one file so it can be replaced by a
  hand-written parser later. Everything else — frontmatter parsing,
  templates, dev server, file watching — is standard library.
- **Convention over configuration.** There is no config file. The
  directory layout is the configuration.

## Requirements

- Go 1.23 or newer to build (the module currently targets 1.26; the
  hard floor is 1.23 for `os.CopyFS`). No C toolchain, no other tools.
- Nothing at runtime: the result is one static binary, and the built
  site is plain files that any web server can serve.
- Works on Linux and anything else Go targets.

## Install

From a checkout:

```bash
git clone <this repo> && cd piuma
go install .
```

That places the `piuma` binary in `$(go env GOPATH)/bin` (usually
`~/go/bin`); make sure it is on your PATH:

```bash
export PATH="$PATH:$HOME/go/bin"   # add to your shell rc
```

Alternatives: `go build -o piuma .` for a binary in the current
directory, or `go run . <command>` to run without installing.

## Commands

```
usage: piuma <command> [flags]

  build    render the site into the output directory
  dev      serve the site locally, rebuilding on every change
  format   validate all content; exits 1 listing every problem
  help     show usage

Flags (all commands):
  -dir string    site root directory (default ".")

Flags (build):
  -out string    output directory (default <dir>/public); wiped on every build

Flags (dev):
  -addr string   listen address (default ":8080")
```

`piuma <command> -h` prints that command's flags. Typical loop:

```bash
cd ~/code/mysite
piuma dev              # write, save, refresh browser
piuma format           # sanity-check everything
piuma build            # produce public/ for deployment
rsync -av public/ server:/var/www/   # deploy however you like
```

## Site layout

A piuma site is a directory with this shape (only `content/` is
required):

```
mysite/
  content/            posts: one <slug>.md per article
  pages/              standalone pages: about.md, now.md, ...
    index.md          the homepage, if you want one
  templates/          optional template overrides (see below)
  static/             copied verbatim into the output root
  public/             build output — disposable, wiped every build
```

Subdirectories under `content/` are allowed for your own organization
but carry no meaning: only the file name matters.

## URLs

```
/                    homepage (pages/index.md) — or the post listing
                     when no homepage page exists
/<page>              each standalone page (pages/about.md → /about)
/payload/            the post listing, always
/payload/<slug>      each post (content/hello.md → /payload/hello)
/payload/tags/<tag>  all posts carrying that tag
```

The blog lives under its own path so a subdomain can serve it: point a
vhost at `public/payload/` and posts appear at `blog.example.com/<slug>`.

Tag URLs are slugified: lowercase, runs of non-alphanumerics become one
dash (`Google Tag Manager` → `google-tag-manager`). In the default post
template every tag is a link to its listing.

Reserved names: a page cannot be called `payload`, a post cannot be
called `tags`, and duplicate slugs are errors — `format` and `build`
refuse with a message naming the offending files.

## Content format

Posts are markdown with YAML-style frontmatter:

```markdown
---
publishDate: 2026-07-07T16:00:00Z
title: 'Writing hexdump from scratch'
excerpt: 'Shown in listings.'
tags:
  - programming
  - C
metadata:
  description: 'Overrides excerpt as the meta description.'
---

# Writing hexdump from scratch

Body in markdown. Raw HTML (iframes, embeds) passes through untouched.
```

`title` and `publishDate` (RFC 3339) are required; the rest optional.
Pages need only `title` — they have no date and no tags.

The frontmatter parser is hand-written and intentionally supports only
this subset of YAML: string scalars (quoted or bare, multi-line quoted
folds to spaces), flat lists, one level of nested map. No anchors, no
types, no surprises. Unknown keys are ignored, so legacy frontmatter
(e.g. an old `category:` key) is harmless.

Markdown is rendered by goldmark with GitHub-flavored extensions
(tables, strikethrough, autolinks) and raw HTML enabled — content is
trusted, it's your own site.

## Templates

Piuma ships embedded default templates, so a bare `content/` directory
already builds a working site. To customize, create `templates/` in the
site and add **only the files you want to change** — each template
falls back to the embedded default when missing:

```
templates/
  layout.html   the skeleton around every page: <head>, nav, footer
  index.html    the post listing (also used for tag pages)
  post.html     a single post
  page.html     a standalone page
```

They are Go `html/template` files. `layout.html` defines the page and
leaves holes; the others fill them:

```html
<!-- layout.html -->
<title>{{block "title" .}}My site{{end}}</title>
{{block "meta" .}}{{end}}
<main>{{block "content" .}}{{end}}</main>

<!-- index.html -->
{{define "content"}}
{{with .Heading}}<h1>{{.}}</h1>{{end}}   <!-- tag name on tag pages -->
{{range .Posts}}<a href="{{.URL}}">{{.Title}}</a>{{end}}
{{end}}
```

Data available: `post.html` gets `.Post` (Slug, Title, Excerpt,
Description, Tags, PublishDate) and `.Content` (rendered body);
`page.html` gets `.Page` and `.Content`; `index.html` gets `.Posts`
(newest first) and `.Heading`. The `tagslug` function is available
everywhere: `/payload/tags/{{tagslug .}}/`.

CSS, fonts, favicons and images belong in `static/` and are referenced
by absolute path (`/styles.css`, `/fonts/...`) so they resolve from
nested pages too.

## What to expect

- **`build`** loads everything, reports *all* broken files at once
  (not just the first), and writes the output only when the whole site
  parses. `public/` is wiped first — never point `-out` at a directory
  you care about.
- **`dev`** builds, serves, and polls the source directories (500 ms
  mtime check — no inotify dependency). A broken edit logs the error
  and keeps serving the last good output; fix the file and it rebuilds.
  There is no browser live-reload: refresh yourself.
- **`format`** is the pre-deploy check: frontmatter parses, required
  fields present, no slug collisions. It rewrites nothing.
- Output pages are `<dir>/index.html` files, so any static file server
  serves clean URLs without configuration.

Not included (yet, deliberately): RSS, sitemap, drafts, syntax
highlighting, pagination, live-reload, remote deploy. The structure has
room for them; they'll arrive when needed.

## Development

```
main.go              command dispatch only
internal/content     frontmatter parser, post/page model, site loader
internal/render      markdown (goldmark, quarantined) + templates
internal/build       orchestrates a full build
internal/server      dev server + polling watcher
```

`go test ./...` runs the suite — behavior tests, not implementation
tests, including a corpus test that parses every post of the real blog
when present on disk. Design notes live in `docs/superpowers/specs/`.
