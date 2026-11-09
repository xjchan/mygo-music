package player

import (
	"path/filepath"
	"testing"

	"mygo-music/internal/music"
)

func seekStream(t *testing.T) *music.Stream {
	t.Helper()
	path, err := filepath.Abs("testdata/seek.mp3")
	if err != nil {
		t.Fatal(err)
	}
	return &music.Stream{URL: "file://" + filepath.ToSlash(path), Type: "mp3"}
}

// A song its source says isn't available plays from the fallback, and the
// fallback is asked once.
func TestFallbackPlaysWhenSourceCant(t *testing.T) {
	p, ch, q := pickyQueue()
	defer p.stop()
	asked := 0
	p.Fallback = func(music.Track, string, string) (*music.Stream, error) {
		asked++
		return seekStream(t), nil
	}
	p.PlayTrack(q, 1)
	settle(t, ch, p)
	if asked != 1 {
		t.Errorf("fallback asked %d times", asked)
	}
	if p.Current().ID != "b" || !p.Playing() || p.Err() != "" {
		t.Fatalf("on %s, playing %v, err %q", p.Current().ID, p.Playing(), p.Err())
	}
}

// A stream whose audio can't be opened, as a missing file, is what the
// fallback replaces too.
func TestFallbackPlaysWhenAudioCantOpen(t *testing.T) {
	ch := make(chan func(), 64)
	p := New(func(fn func()) { ch <- fn })
	p.SetVolume(0)
	src := missing{}
	p.Source = func(string) music.Source { return src }
	p.Fallback = func(music.Track, string, string) (*music.Stream, error) {
		return seekStream(t), nil
	}
	defer p.stop()
	p.PlayTrack([]music.Track{{ID: "a", Source: "s"}}, 0)
	settle(t, ch, p)
	if !p.Playing() || p.Err() != "" {
		t.Fatalf("playing %v, err %q", p.Playing(), p.Err())
	}
}

// A source that wants a sign-in keeps saying so: the fallback isn't a way
// around the listener's account.
func TestFallbackLeavesSignInAlone(t *testing.T) {
	ch := make(chan func(), 64)
	p := New(func(fn func()) { ch <- fn })
	p.SetVolume(0)
	src := signing{}
	p.Source = func(string) music.Source { return src }
	asked := 0
	p.Fallback = func(music.Track, string, string) (*music.Stream, error) {
		asked++
		return seekStream(t), nil
	}
	defer p.stop()
	p.PlayTrack([]music.Track{{ID: "a", Source: "s"}}, 0)
	settle(t, ch, p)
	if asked != 0 {
		t.Errorf("fallback asked %d times", asked)
	}
	if p.Playing() || p.Err() == "" {
		t.Errorf("playing %v, err %q", p.Playing(), p.Err())
	}
}

// missing answers a stream that isn't there on disk.
type missing struct{ nothing }

func (missing) Info() music.Info { return music.Info{ID: "s", Playback: music.PlaybackStream} }
func (missing) Stream(string, string) (*music.Stream, error) {
	return &music.Stream{URL: "file:///no/such/song.mp3", Type: "mp3"}, nil
}

// signing answers the sign-in error a source gives before it can play.
type signing struct{ nothing }

func (signing) Info() music.Info { return music.Info{ID: "s", Playback: music.PlaybackStream} }
func (signing) Stream(string, string) (*music.Stream, error) {
	return nil, music.ErrSignIn
}
