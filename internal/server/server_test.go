package server

import (
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sprawz/piuma/internal/build"
)

func fixtureSite(t *testing.T) build.Config {
	t.Helper()
	root := t.TempDir()
	post := filepath.Join(root, "content/hello.md")
	if err := os.MkdirAll(filepath.Dir(post), 0o755); err != nil {
		t.Fatal(err)
	}
	src := "---\npublishDate: 2026-07-07T16:00:00Z\ntitle: 'Hello'\n---\nBody one.\n"
	if err := os.WriteFile(post, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	return build.DefaultConfig(root)
}

func get(t *testing.T, ts *httptest.Server, path string) (int, string) {
	t.Helper()
	res, err := ts.Client().Get(ts.URL + path)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return res.StatusCode, string(body)
}

func TestHandlerServesBuiltSite(t *testing.T) {
	cfg := fixtureSite(t)
	if err := build.Build(cfg); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(Handler(cfg.OutDir))
	defer ts.Close()

	if code, body := get(t, ts, "/blog/hello/"); code != 200 || !strings.Contains(body, "Body one.") {
		t.Errorf("post page: code=%d body=%q", code, body)
	}
	if code, body := get(t, ts, "/"); code != 200 || !strings.Contains(body, "Hello") {
		t.Errorf("index: code=%d body=%q", code, body)
	}
	if code, _ := get(t, ts, "/nope"); code != 404 {
		t.Errorf("missing page: code=%d, want 404", code)
	}
}
