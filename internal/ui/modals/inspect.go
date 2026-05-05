package modals

import (
	"bytes"
	"encoding/json"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/torosent/c9s/internal/ui/theme"
)

// InspectModel is a modal that displays JSON content in a scrollable viewport.
type InspectModel struct {
	title    string
	content  string
	viewport viewport.Model
	palette  theme.Palette
}

// NewInspect creates a new inspect modal.
func NewInspect(title string, jsonBytes []byte, p theme.Palette) InspectModel {
	var buf bytes.Buffer
	err := json.Indent(&buf, jsonBytes, "", "  ")

	content := ""
	if err == nil {
		content = buf.String()
	} else {
		content = string(jsonBytes)
	}

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(24))
	vp.SetContent(content)

	return InspectModel{
		title:    title,
		content:  content,
		viewport: vp,
		palette:  p,
	}
}

// Init implements Modal.
func (m InspectModel) Init() tea.Cmd {
	return nil
}

// Update implements Modal.
func (m InspectModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			return m, func() tea.Msg {
				return CloseModalMsg{}
			}
		case "q", "Q":
			return m, func() tea.Msg {
				return CloseModalMsg{}
			}
		}

		// Let viewport handle other keys (j, k, up, down, pgup, pgdn, etc.)
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View implements Modal.
func (m InspectModel) View(width, height int) string {
	m.viewport.SetWidth(width - 8)
	m.viewport.SetHeight(height - 8)

	// Build the view
	helpText := lipgloss.NewStyle().
		Foreground(m.palette.Dim).
		Render("j/k/↑/↓/PgUp/PgDn: scroll · q/Esc: close")

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		BorderBackground(m.palette.Bg).
		Background(m.palette.Bg).
		Foreground(m.palette.Fg).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.palette.HeaderFg)

	content := titleStyle.Render(m.title) + "\n\n" +
		m.viewport.View() + "\n\n" +
		helpText

	return boxStyle.Render(content)
}

// Title implements Modal.
func (m InspectModel) Title() string {
	return m.title
}
