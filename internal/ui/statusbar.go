package ui

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/torosent/c9s/internal/ui/theme"
)

// StatusUpdate is the message a screen sends to refresh the status bar.
type StatusUpdate struct {
	Screen  string
	Summary string
	Hint    string
	Toast   string // optional; if non-empty, replaces Hint with red styling
}

// StatusBar is the bottom-line status component.
type StatusBar struct {
	palette theme.Palette
	state   StatusUpdate
}

// NewStatusBar returns a StatusBar styled with the given palette.
func NewStatusBar(p theme.Palette) StatusBar {
	return StatusBar{palette: p}
}

// Update returns a new StatusBar with the updated content.
func (s StatusBar) Update(u StatusUpdate) StatusBar {
	s.state = u
	return s
}

// View renders the status bar to a single line of `width` columns,
// truncating with an ellipsis if necessary.
func (s StatusBar) View(width int, readonly bool) string {
	if width <= 0 {
		return ""
	}
	p := s.palette

	base := lipgloss.NewStyle().Foreground(p.Fg).Background(p.Bg)
	accent := lipgloss.NewStyle().Foreground(p.Accent).Background(p.Bg).Bold(true)
	toastStyle := lipgloss.NewStyle().Background(p.Error).Foreground(lipgloss.Color("#ffffff")).Bold(true)
	dim := lipgloss.NewStyle().Foreground(p.Dim).Background(p.Bg)
	warning := lipgloss.NewStyle().Foreground(p.Warning).Background(p.Bg).Bold(true)

	var rightSlot string
	if s.state.Toast != "" {
		rightSlot = toastStyle.Render(" " + s.state.Toast + " ")
	} else if s.state.Hint != "" {
		hint := s.state.Hint
		if readonly {
			hint += " [READONLY]"
		}
		rightSlot = dim.Render(hint)
		if readonly {
			rightSlot = dim.Render(s.state.Hint) + " " + warning.Render("[READONLY]")
		}
	} else if readonly {
		rightSlot = warning.Render("[READONLY]")
	}

	left := accent.Render(s.state.Screen) + base.Render("  "+s.state.Summary)
	gap := base.Render("  ")
	line := lipgloss.JoinHorizontal(lipgloss.Top, left, gap, rightSlot)
	line = truncateToWidth(line, width)
	// Fill the whole width with the skin's bg so light themes don't fall
	// through to the terminal's default bg under the status bar.
	return lipgloss.NewStyle().
		Width(width).
		Foreground(p.Fg).
		Background(p.Bg).
		Render(line)
}

// truncateToWidth shortens a (possibly ANSI-styled) string to at most
// `width` visible columns. Uses ANSI-aware truncation so escape
// sequences don't count toward the visible width — bubbletea v2's
// lipgloss emits longer escape sequences than v1, which made the
// rune-count-based truncator drop visible content prematurely.
func truncateToWidth(s string, width int) string {
	if ansi.StringWidth(s) <= width {
		return s
	}
	if width <= 1 {
		return ansi.Truncate(s, width, "")
	}
	return ansi.Truncate(s, width, "…")
}
