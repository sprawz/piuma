package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchDetectsChange(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "a.md")
	if err := os.WriteFile(file, []byte("one"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changed := make(chan struct{}, 1)
	go Watch(ctx, []string{dir}, 10*time.Millisecond, func() {
		changed <- struct{}{}
	})

	time.Sleep(50 * time.Millisecond) // let the watcher take its baseline
	if err := os.WriteFile(file, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	select {
	case <-changed:
	case <-time.After(2 * time.Second):
		t.Fatal("change not detected")
	}
}

func TestWatchIgnoresQuiet(t *testing.T) {
	dir := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	changed := make(chan struct{}, 1)
	go Watch(ctx, []string{dir}, 10*time.Millisecond, func() {
		changed <- struct{}{}
	})
	select {
	case <-changed:
		t.Fatal("spurious change reported")
	case <-time.After(100 * time.Millisecond):
	}
}
