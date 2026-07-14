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

	"github.com/paolobietolini/piuma/internal/build"
	"github.com/paolobietolini/piuma/internal/content"
	"github.com/paolobietolini/piuma/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch cmd := os.Args[1]; cmd {
	case "build":
		err = runBuild(os.Args[2:])
	case "dev":
		err = runDev(os.Args[2:])
	case "format":
		err = runFormat(os.Args[2:])
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: piuma <build|dev|format> [flags]")
	os.Exit(2)
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
	cfg := build.DefaultConfig(siteFlags(fs, args))
	if *out != "" {
		cfg.OutDir = *out
	}
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

// runFormat validates the content tree: every post must parse and carry
// its required frontmatter. It rewrites nothing — the useful part of a
// formatter here is the check, not canonicalized prose.
func runFormat(args []string) error {
	fs := flag.NewFlagSet("format", flag.ExitOnError)
	cfg := build.DefaultConfig(siteFlags(fs, args))
	site, err := content.LoadSite(cfg.ContentDir)
	if err != nil {
		return err
	}
	fmt.Printf("ok: %d posts\n", len(site.Posts))
	return nil
}
