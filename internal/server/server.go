package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/paolobietolini/piuma/internal/build"
)

const pollInterval = 500 * time.Millisecond

// Handler serves the built site from outDir.
func Handler(outDir string) http.Handler {
	return http.FileServer(http.Dir(outDir))
}

// Serve builds the site, then serves it on addr while rebuilding on any
// source change. A failed rebuild logs the error and keeps the last good
// output. Blocks until the server stops.
func Serve(addr string, cfg build.Config) error {
	rebuild := func() {
		if err := build.Build(cfg); err != nil {
			log.Printf("build failed, serving last good output:\n%v", err)
			return
		}
		log.Printf("rebuilt")
	}
	if err := build.Build(cfg); err != nil {
		// First build may legitimately fail while drafting; dev keeps going.
		log.Printf("initial build failed:\n%v", err)
	}
	watched := []string{cfg.ContentDir, cfg.TemplateDir, cfg.StaticDir}
	go Watch(context.Background(), watched, pollInterval, rebuild)
	log.Printf("serving %s on http://localhost%s", cfg.OutDir, addr)
	return http.ListenAndServe(addr, Handler(cfg.OutDir))
}
