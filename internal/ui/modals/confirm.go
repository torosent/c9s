package modals

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/torosent/c9s/internal/ui/theme"
)

// ConfirmModel is a modal that asks the user to confirm an action.
type ConfirmModel struct {
	title   string
	body    string
	lines   []string
	tag     string
	palette theme.Palette
}

// NewConfirm creates a new confirmation modal.
func NewConfirm(title, body string, lines []string, tag string, p theme.Palette) ConfirmModel {
	return ConfirmModel{
		title:   title,
		body:    body,
		lines:   lines,
		tag:     tag,
		palette: p,
	}
}

// Init implements Modal.
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update implements Modal.
func (m ConfirmModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyRunes:
			if len(msg.Runes) > 0 {
				switch msg.Runes[0] {
				case 'y', 'Y':
					return m, tea.Batch(
						func() tea.Msg {
							return ConfirmResultMsg{Result: ConfirmResult{Confirmed: true, Tag: m.tag}}
						},
						func() tea.Msg {
							return CloseModalMsg{}
						},
					)
				case 'n', 'N':
					return m, tea.Batch(
						func() tea.Msg {
							return ConfirmResultMsg{Result: ConfirmResult{Confirmed: false, Tag: m.tag}}
						},
						func() tea.Msg {
							return CloseModalMsg{}
						},
					)
				}
			}
		case tea.KeyEsc:
			return m, tea.Batch(
				func() tea.Msg {
					return ConfirmResultMsg{Result: ConfirmResult{Confirmed: false, Tag: m.tag}}
				},
				func() tea.Msg {
					return CloseModalMsg{}
				},
			)
		}
	}
	return m, nil
}

// View implements Modal.
func (m ConfirmModel) View(width, height int) string {
	// Build content
	var content strings.Builder

	if m.title != "" {
		content.WriteString(m.title)
		content.WriteString("\n\n")
	}

	if m.body != "" {
		content.WriteString(m.body)
		content.WriteString("\n")
	}

	// Add bullet points for lines
	for _, line := range m.lines {
		content.WriteString("  • ")
		content.WriteString(line)
		content.WriteString("\n")
	}

	content.WriteString("\n")
	content.WriteString("[y] yes  [n/Esc] cancel")

	// Style with border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		BorderBackground(m.palette.Bg).
		Background(m.palette.Bg).
		Foreground(m.palette.Fg).
		Padding(1, 2).
		Width(width - 4) // Account for padding and border

	// Truncate if needed
	text := content.String()
	maxWidth := width - 8 // Leave room for border and padding
	if maxWidth < 20 {
		maxWidth = 20
	}

	lines := strings.Split(text, "\n")
	var truncated []string
	for _, line := range lines {
		runes := []rune(line)
		if len(runes) > maxWidth {
			runes = runes[:maxWidth-1]
			truncated = append(truncated, string(runes)+"…")
		} else {
			truncated = append(truncated, line)
		}
	}

	return boxStyle.Render(strings.Join(truncated, "\n"))
}

// Title implements Modal.
func (m ConfirmModel) Title() string {
	return m.title
}
