package plugin

import "mygo-music/internal/music"

// FallbackStream asks the plugin for audio of a track another source couldn't
// play. It gets the track's metadata, as the plugin's own IDs mean nothing to
// another source, and the quality asked for. A plugin without the export
// answers ErrUnsupported; null says it found no song sure enough to be that
// one.
func (p *Plugin) FallbackStream(t music.Track, quality, reason string) (*music.Stream, error) {
	var st *music.Stream
	params := map[string]any{
		"track":   t,
		"quality": quality,
		"reason":  reason,
	}
	if err := p.call("fallbackStream", params, &st); err != nil {
		return nil, err
	}
	if st == nil || st.URL == "" {
		return nil, music.ErrUnavailable
	}
	return st, nil
}

// ID is the plugin's ID, as plugin.json says.
func (p *Plugin) ID() string { return p.manifest.ID }
