package modals

import (
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"

	"github.com/torosent/c9s/internal/ui/theme"
)

// styleTextInput paints a bubbles textinput with the active skin's
// palette so the prompt, value, placeholder, completion suggestion and
// cursor cells all show the skin's bg/fg — preventing the terminal's
// default (often black) bg from leaking through inside themed modals.
func styleTextInput(t *textinput.Model, p theme.Palette) {
	styles := t.Styles()
	prompt := lipgloss.NewStyle().Foreground(p.Dim).Background(p.Bg)
	textStyle := lipgloss.NewStyle().Foreground(p.Fg).Background(p.Bg)
	placeholder := lipgloss.NewStyle().Foreground(p.Dim).Background(p.Bg)
	suggestion := lipgloss.NewStyle().Foreground(p.Dim).Background(p.Bg)
	styles.Focused.Prompt = prompt
	styles.Focused.Text = textStyle
	styles.Focused.Placeholder = placeholder
	styles.Focused.Suggestion = suggestion
	styles.Blurred.Prompt = prompt
	styles.Blurred.Text = textStyle
	styles.Blurred.Placeholder = placeholder
	styles.Blurred.Suggestion = suggestion
	styles.Cursor.Color = p.Accent
	t.SetStyles(styles)
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
	t.SetWidth(0)
	bg := lipgloss.NewStyle().Foreground(p.Fg).Background(p.Bg)
	return bg.Width(width).Render(t.View())
}
