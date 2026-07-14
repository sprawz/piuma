// Package server runs the dev loop: serve the built site and rebuild
// when sources change.
package server

import (
	"context"
	"fmt"
	"io/fs"
	"path/filepath"
	"time"
)

// Watch polls dirs at the given interval and calls fn whenever any file
// under them is added, removed, or modified. It blocks until ctx is done.
// Polling keeps us on the standard library; at blog scale a scan is cheap.
func Watch(ctx context.Context, dirs []string, interval time.Duration, fn func()) {
	last := fingerprint(dirs)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if cur := fingerprint(dirs); cur != last {
				last = cur
				fn()
			}
		}
	}
}

// fingerprint reduces the watched trees to a comparable string of
// path, size and mtime lines. WalkDir is deterministic, so equal trees
// produce equal strings.
func fingerprint(dirs []string) string {
	var s string
	for _, dir := range dirs {
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if info, err := d.Info(); err == nil {
				s += fmt.Sprintf("%s %d %d\n", path, info.Size(), info.ModTime().UnixNano())
			}
			return nil
		})
	}
	return s
}
