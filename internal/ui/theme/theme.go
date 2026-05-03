// Package theme defines the color palette c9s uses across screens.
// Plan 5 will add TOML-driven custom skins; v0.1.0 ships DefaultDark only.
package theme

import "github.com/charmbracelet/lipgloss"

// Palette is the resolved set of colors a screen renders with.
type Palette struct {
	Fg, Bg              lipgloss.Color
	Border, Accent, Dim lipgloss.Color
	Success             lipgloss.Color
	Warning             lipgloss.Color
	Error               lipgloss.Color
	SelectionFg         lipgloss.Color
	SelectionBg         lipgloss.Color
	HeaderFg, HeaderBg  lipgloss.Color
	State               map[string]lipgloss.Color // running/exited/paused/stopping/created
}

// DefaultDark returns the default dark palette.
func DefaultDark() Palette {
	return Palette{
		Fg:          "#c9d1d9",
		Bg:          "#0d1117",
		Border:      "#30363d",
		Accent:      "#58a6ff",
		Dim:         "#6e7681",
		Success:     "#3fb950",
		Warning:     "#d29922",
		Error:       "#f85149",
		SelectionFg: "#ffffff",
		SelectionBg: "#1f6feb",
		HeaderFg:    "#f0f6fc",
		HeaderBg:    "#161b22",
		State: map[string]lipgloss.Color{
			"running":  "#3fb950",
			"exited":   "#6e7681",
			"paused":   "#d29922",
			"stopping": "#f85149",
			"created":  "#58a6ff",
		},
	}
}

// Accent2 returns a secondary accent for k9s-style label colors. Falls back
// to Warning, Success, then Accent if the secondary is empty.
func (p Palette) Accent2() lipgloss.Color {
	if p.Warning != "" {
		return p.Warning
	}
	if p.Success != "" {
		return p.Success
	}
	return p.Accent
}

// SourceColors is a palette of colors for multi-source log viewers.
// Used in stable hash → color assignment.
var SourceColors = []lipgloss.Color{
	"#58a6ff", // blue
	"#3fb950", // green
	"#d29922", // yellow
	"#f85149", // red
	"#a371f7", // purple
	"#ff7b72", // orange
	"#56d4dd", // cyan
	"#ffa657", // amber
}
