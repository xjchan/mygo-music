package app

import (
	"errors"
	"log"
	"sync/atomic"

	"mygo-music/internal/music"
	"mygo-music/internal/plugin"
)

// fallbackPlugins are the plugins that can play a song another source
// couldn't, in the order to ask them. The player asks off the main thread, so
// the list is a snapshot that enabledChanged replaces whole.
type fallbackPlugins struct {
	on atomic.Pointer[[]*plugin.Plugin]
}

// set takes the enabled plugins that can answer for another source.
func (f *fallbackPlugins) set(on []*plugin.Plugin) {
	on = append([]*plugin.Plugin(nil), on...)
	f.on.Store(&on)
}

// stream asks each plugin in turn for audio of t. A plugin that finds nothing
// says ErrUnavailable, and one without the export isn't in the list; either
// way the next is asked. If none answers, ErrUnsupported leaves the player to
// show the source's own error.
func (f *fallbackPlugins) stream(t music.Track, quality, reason string) (*music.Stream, error) {
	on := f.on.Load()
	if on == nil {
		return nil, music.ErrUnsupported
	}
	for _, p := range *on {
		// The plugin whose own stream failed needn't answer the same song
		// again, and one that does could loop.
		if p.ID() == t.Source {
			continue
		}
		st, err := p.FallbackStream(t, quality, reason)
		if err == nil && st != nil && st.URL != "" {
			return st, nil
		}
		if err != nil && !errors.Is(err, music.ErrUnsupported) && !errors.Is(err, music.ErrUnavailable) {
			log.Printf("fallback plugin %s: %v", p.ID(), err)
		}
	}
	return nil, music.ErrUnsupported
}
