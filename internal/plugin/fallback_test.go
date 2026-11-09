package plugin

import (
	"errors"
	"testing"

	"mygo-music/internal/music"
)

// The fallback gets the failing track's metadata, quality and why it failed,
// and its answer is the stream to play.
func TestFallbackStream(t *testing.T) {
	p := load(t, plain, `
export function fallbackStream({ track, quality, reason }) {
  return {
    url: "https://example.com/" + track.name,
    type: "flac",
    quality,
    headers: { "x-reason": reason, "x-source": track.source },
  };
}`)
	st, err := p.FallbackStream(music.Track{
		ID: "1", Name: "晴天", Source: "other",
		Artists: []music.Artist{{Name: "周杰伦"}},
	}, "lossless", "unavailable")
	if err != nil {
		t.Fatal(err)
	}
	if st.URL != "https://example.com/晴天" || st.Type != "flac" || st.Quality != "lossless" {
		t.Errorf("stream %+v", st)
	}
	if st.Headers["x-reason"] != "unavailable" || st.Headers["x-source"] != "other" {
		t.Errorf("headers %v", st.Headers)
	}
}

// Null is no match, not an empty stream to play.
func TestFallbackStreamNoMatch(t *testing.T) {
	p := load(t, plain, `export function fallbackStream() { return null; }`)
	if _, err := p.FallbackStream(music.Track{Name: "x"}, "high", "error"); !errors.Is(err, music.ErrUnavailable) {
		t.Fatalf("got %v", err)
	}
}

// A plugin without the export isn't asked to start for it.
func TestFallbackStreamUnsupported(t *testing.T) {
	p := load(t, plain, `export function home() { return {}; }`)
	if _, err := p.FallbackStream(music.Track{Name: "x"}, "high", "error"); !errors.Is(err, music.ErrUnsupported) {
		t.Fatalf("got %v", err)
	}
}
