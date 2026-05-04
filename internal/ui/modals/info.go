package modals

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/torosent/c9s/internal/ui/theme"
)

// InfoModel is a non-blocking informational modal: title plus a body of
// arbitrary lines. Dismissed by any key (Enter, Esc, q, space). Useful
// for surfacing the result of operations whose output doesn't fit the
// one-line status-bar toast (e.g. ":acr-login" success, ":install-docker-shim"
// install confirmation with detected-docker warnings).
type InfoModel struct {
	title   string
	body    []string
	level   InfoLevel
	palette theme.Palette
}

// InfoLevel selects the title-bar accent color.
type InfoLevel int

const (
	// InfoOK is the default level (palette accent).
	InfoOK InfoLevel = iota
	// InfoWarning highlights the title in the warning color.
	InfoWarning
	// InfoError highlights the title in the error color.
	InfoError
)

// NewInfo constructs a modal with a title and body lines.
func NewInfo(title string, body []string, level InfoLevel, p theme.Palette) InfoModel {
	return InfoModel{title: title, body: body, level: level, palette: p}
}

// Init implements Modal.
func (m InfoModel) Init() tea.Cmd { return nil }

// Update implements Modal. Any key dismisses the modal.
func (m InfoModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	if _, ok := msg.(tea.KeyMsg); ok {
		return m, func() tea.Msg { return CloseModalMsg{} }
	}
	return m, nil
}

// View implements Modal.
func (m InfoModel) View(width, height int) string {
	p := m.palette

	titleColor := p.Accent
	switch m.level {
	case InfoWarning:
		titleColor = p.Warning
	case InfoError:
		titleColor = p.Error
	}

	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Background(p.Bg).Bold(true)
	bodyStyle := lipgloss.NewStyle().Foreground(p.Fg).Background(p.Bg)
	hintStyle := lipgloss.NewStyle().Foreground(p.Dim).Background(p.Bg).Italic(true)

	var content strings.Builder
	if m.title != "" {
		content.WriteString(titleStyle.Render(m.title))
		content.WriteString("\n\n")
	}
	for _, line := range m.body {
		content.WriteString(bodyStyle.Render(line))
		content.WriteString("\n")
	}
	content.WriteString("\n")
	content.WriteString(hintStyle.Render("press any key to dismiss"))

	boxWidth := width - 4
	if boxWidth < 40 {
		boxWidth = 40
	}
	if boxWidth > 100 {
		boxWidth = 100
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(p.Border).
		BorderBackground(p.Bg).
		Background(p.Bg).
		Foreground(p.Fg).
		Padding(1, 2).
		Width(boxWidth)

	return box.Render(content.String())
}

// Title implements Modal.
func (m InfoModel) Title() string { return m.title }
