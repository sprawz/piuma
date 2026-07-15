// piuma is a minimal static site builder: markdown in, HTML out.
//
//	piuma build [-dir site] [-out public]   render the site
//	piuma dev   [-dir site] [-addr :8080]   serve with live rebuild
//	piuma format [-dir site]                validate all content
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sprawz/piuma/internal/build"
	"github.com/sprawz/piuma/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		usage(os.Stderr)
		os.Exit(2)
	}
	var err error
	switch cmd := os.Args[1]; cmd {
	case "build":
		err = runBuild(os.Args[2:])
	case "dev":
		err = runDev(os.Args[2:])
	case "format":
		err = runFormat(os.Args[2:])
	case "help", "-h", "-help", "--help":
		usage(os.Stdout)
	default:
		fmt.Fprintf(os.Stderr, "piuma: unknown command %q\n\n", cmd)
		usage(os.Stderr)
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage(w *os.File) {
	fmt.Fprint(w, `piuma — minimal static site builder: markdown in, HTML out.

usage: piuma <command> [flags]

Commands:
  build    render the site into the output directory
  dev      serve the site locally, rebuilding on every change
  format   validate all content (frontmatter, required fields, slug
           collisions); exits 1 listing every problem found
  help     show this help

Flags (all commands):
  -dir string    site root directory (default ".")

Flags (build):
  -out string    output directory (default <dir>/public);
                 wiped on every build. Refused if it contains
                 the site's source directories.
  -base string   absolute site URL (https://example.com); when set,
                 sitemap.xml, atom.xml and llms.txt are generated
                 at the output root

Flags (dev):
  -addr string   listen address (default ":8080")

Run "piuma <command> -h" for that command's flags.
`)
}

// siteFlags defines the flags shared by every command and returns the
// site root after parsing.
func siteFlags(fs *flag.FlagSet, args []string) string {
	dir := fs.String("dir", ".", "site root directory")
	fs.Parse(args)
	return *dir
}

func runBuild(args []string) error {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	out := fs.String("out", "", "output directory (default <dir>/public)")
	base := fs.String("base", "", "absolute site URL; enables sitemap.xml, atom.xml, llms.txt")
	cfg := build.DefaultConfig(siteFlags(fs, args))
	if *out != "" {
		cfg.OutDir = *out
	}
	cfg.BaseURL = *base
	if err := build.Build(cfg); err != nil {
		return err
	}
	fmt.Println("built", cfg.OutDir)
	return nil
}

func runDev(args []string) error {
	fs := flag.NewFlagSet("dev", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "listen address")
	cfg := build.DefaultConfig(siteFlags(fs, args))
	return server.Serve(*addr, cfg)
}

// runFormat validates the site: every post and page must parse and
// carry its required frontmatter, and the templates must compile. It
// rewrites nothing — the useful part of a formatter here is the check,
// not canonicalized prose.
func runFormat(args []string) error {
	fs := flag.NewFlagSet("format", flag.ExitOnError)
	cfg := build.DefaultConfig(siteFlags(fs, args))
	site, err := build.Validate(cfg)
	if err != nil {
		return err
	}
	fmt.Printf("ok: %d posts, %d pages\n", len(site.Posts), len(site.Pages))
	return nil
}
