package app

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"mygo-music/internal/music"
	"mygo-music/internal/plugin"
)

func fallbackPlugin(t *testing.T, id, source string) *plugin.Plugin {
	t.Helper()
	dir := t.TempDir()
	manifest := `{"id":"` + id + `","name":"` + id + `"}`
	os.WriteFile(filepath.Join(dir, "plugin.json"), []byte(manifest), 0o600)
	os.WriteFile(filepath.Join(dir, "index.js"), []byte(source), 0o600)
	p, err := plugin.Load(dir, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(p.Close)
	return p
}

// The first fallback plugin that answers is used; one that finds nothing lets
// the next be asked.
func TestFallbackPluginsTryInOrder(t *testing.T) {
	none := fallbackPlugin(t, "none", `export function fallbackStream() { return null; }`)
	yes := fallbackPlugin(t, "yes", `
export function fallbackStream({ track, quality, reason }) {
  return { url: "https://example.com/" + track.name, type: "mp3", quality, headers: { "x-reason": reason } };
}`)
	f := &fallbackPlugins{}
	f.set([]*plugin.Plugin{none, yes})
	st, err := f.stream(music.Track{ID: "1", Name: "晴天", Source: "other"}, "high", "unavailable")
	if err != nil {
		t.Fatal(err)
	}
	if st.URL != "https://example.com/晴天" || st.Headers["x-reason"] != "unavailable" {
		t.Errorf("stream %+v", st)
	}
}

// The plugin whose own stream failed is skipped: it needn't answer the same
// song again, and one that did could loop.
func TestFallbackPluginsSkipTheirSource(t *testing.T) {
	p := fallbackPlugin(t, "self", `export function fallbackStream() { return { url: "https://example.com/self", type: "mp3" }; }`)
	f := &fallbackPlugins{}
	f.set([]*plugin.Plugin{p})
	if _, err := f.stream(music.Track{Name: "x", Source: "self"}, "high", "error"); !errors.Is(err, music.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}

// With no plugin that can answer, the player keeps the source's own error.
func TestFallbackPluginsWithoutOne(t *testing.T) {
	f := &fallbackPlugins{}
	if _, err := f.stream(music.Track{Name: "x"}, "high", "error"); !errors.Is(err, music.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}
