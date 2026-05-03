package modals

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"github.com/torosent/c9s/internal/ui/theme"
)

// styleTextInput paints a bubbles textinput with the active skin's
// palette so the prompt, value, placeholder, completion suggestion and
// cursor cells all show the skin's bg/fg — preventing the terminal's
// default (often black) bg from leaking through inside themed modals.
func styleTextInput(t *textinput.Model, p theme.Palette) {
	t.PromptStyle = lipgloss.NewStyle().Foreground(p.Dim).Background(p.Bg)
	t.TextStyle = lipgloss.NewStyle().Foreground(p.Fg).Background(p.Bg)
	t.PlaceholderStyle = lipgloss.NewStyle().Foreground(p.Dim).Background(p.Bg)
	t.CompletionStyle = lipgloss.NewStyle().Foreground(p.Dim).Background(p.Bg)
	t.Cursor.Style = lipgloss.NewStyle().Foreground(p.Bg).Background(p.Accent)
	t.Cursor.TextStyle = lipgloss.NewStyle().Foreground(p.Fg).Background(p.Bg)
}

// renderTextInput renders a textinput.Model and pads/wraps it to
// `width` cells with the active skin's bg. bubbles/textinput.View
// emits its own unstyled trailing spaces when the input's Width > 0
// (placeholderView appends strings.Repeat(" ", availWidth) AFTER the
// styled placeholder). Lipgloss cannot retroactively paint those
// embedded raw spaces with bg via an outer Width()/Background() wrap
// because the inner ANSI resets break propagation. So we render the
// input with its internal Width = 0 (no embedded padding) and then
// pad ourselves with bg-styled space cells.
func renderTextInput(t textinput.Model, p theme.Palette, width int) string {
	t.Width = 0
	bg := lipgloss.NewStyle().Foreground(p.Fg).Background(p.Bg)
	return bg.Width(width).Render(t.View())
}
