# piuma

Minimal static site builder for paolobietolini.com. Markdown in, HTML out,
almost everything standard library.

## Commands

```
piuma build  [-dir site] [-out public]   render the site into public/
piuma dev    [-dir site] [-addr :8080]   local server, rebuilds on change
piuma format [-dir site]                 validate all content, exit 1 on problems
```

## Site layout

```
content/<category>/<slug>.md   posts; URL becomes /<category>/<slug>
templates/                     layout.html, post.html, index.html (optional,
                               embedded defaults used when missing)
static/                        copied verbatim into the output
public/                        output, wiped on every build
```

Frontmatter is the YAML subset the blog actually uses — string scalars,
lists, one nested map — parsed by a hand-written parser
(`internal/content/frontmatter.go`) to avoid a YAML dependency.

## Packages

- `internal/content` — frontmatter, post model, site loader
- `internal/render`  — markdown→HTML (goldmark, disposable, isolated here) and templates
- `internal/build`   — orchestrates a full build
- `internal/server`  — dev server + mtime-polling watcher

Design notes: `docs/superpowers/specs/2026-07-14-piuma-design.md`.

## Tests

`go test ./...` — includes a corpus test that parses every post of the real
blog when `~/code/paolobietolini.com` is present.
