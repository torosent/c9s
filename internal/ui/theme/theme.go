// Package theme defines the color palette c9s uses across screens.
// Plan 5 will add TOML-driven custom skins; v0.1.0 ships DefaultDark only.
package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Palette is the resolved set of colors a screen renders with.
type Palette struct {
	Fg, Bg              color.Color
	Border, Accent, Dim color.Color
	Success             color.Color
	Warning             color.Color
	Error               color.Color
	SelectionFg         color.Color
	SelectionBg         color.Color
	HeaderFg, HeaderBg  color.Color
	State               map[string]color.Color // running/exited/paused/stopping/created
}

// DefaultDark returns the default dark palette.
func DefaultDark() Palette {
	return Palette{
		Fg:          lipgloss.Color("#c9d1d9"),
		Bg:          lipgloss.Color("#0d1117"),
		Border:      lipgloss.Color("#30363d"),
		Accent:      lipgloss.Color("#58a6ff"),
		Dim:         lipgloss.Color("#6e7681"),
		Success:     lipgloss.Color("#3fb950"),
		Warning:     lipgloss.Color("#d29922"),
		Error:       lipgloss.Color("#f85149"),
		SelectionFg: lipgloss.Color("#ffffff"),
		SelectionBg: lipgloss.Color("#1f6feb"),
		HeaderFg:    lipgloss.Color("#f0f6fc"),
		HeaderBg:    lipgloss.Color("#161b22"),
		State: map[string]color.Color{
			"running":  lipgloss.Color("#3fb950"),
			"exited":   lipgloss.Color("#6e7681"),
			"paused":   lipgloss.Color("#d29922"),
			"stopping": lipgloss.Color("#f85149"),
			"created":  lipgloss.Color("#58a6ff"),
		},
	}
}

// Accent2 returns a secondary accent for k9s-style label colors. Falls back
// to Warning, Success, then Accent if the secondary is empty.
func (p Palette) Accent2() color.Color {
	if p.Warning != nil {
		return p.Warning
	}
	if p.Success != nil {
		return p.Success
	}
	return p.Accent
}

// SourceColors is a palette of colors for multi-source log viewers.
// Used in stable hash → color assignment.
var SourceColors = []color.Color{
	lipgloss.Color("#58a6ff"), // blue
	lipgloss.Color("#3fb950"), // green
	lipgloss.Color("#d29922"), // yellow
	lipgloss.Color("#f85149"), // red
	lipgloss.Color("#a371f7"), // purple
	lipgloss.Color("#ff7b72"), // orange
	lipgloss.Color("#56d4dd"), // cyan
	lipgloss.Color("#ffa657"), // amber
}
