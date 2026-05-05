package modals

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/torosent/c9s/internal/ui/theme"
)

// SkinPickedMsg is emitted when a skin is selected from the skin picker.
type SkinPickedMsg struct {
	Name string
}

// SkinPickerModel is a modal for selecting a skin from the available
// list. We render items by hand (rather than via bubbles/list) so every
// cell paints with the active skin's bg.
type SkinPickerModel struct {
	palette theme.Palette
	skins   []string
	cursor  int
	width   int
	height  int
}

// NewSkinPicker creates a skin picker modal listing the given skin names.
func NewSkinPicker(skins []string, p theme.Palette) SkinPickerModel {
	return SkinPickerModel{
		palette: p,
		skins:   skins,
		cursor:  0,
	}
}

// Init implements Modal.
func (m SkinPickerModel) Init() tea.Cmd { return nil }

// Update implements Modal.
func (m SkinPickerModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.skins)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			if m.cursor >= 0 && m.cursor < len(m.skins) {
				name := m.skins[m.cursor]
				return m, func() tea.Msg {
					return SkinPickedMsg{Name: name}
				}
			}
		case "esc", "q":
			return m, func() tea.Msg { return CloseModalMsg{} }
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View implements Modal.
func (m SkinPickerModel) View(width, height int) string {
	innerW := 44
	if width < innerW+8 {
		innerW = width - 8
		if innerW < 24 {
			innerW = 24
		}
	}

	bg := lipgloss.NewStyle().Foreground(m.palette.Fg).Background(m.palette.Bg)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(m.palette.HeaderFg).Background(m.palette.Accent).Padding(0, 1)
	dim := lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg)
	selRow := lipgloss.NewStyle().Foreground(m.palette.Bg).Background(m.palette.Accent).Bold(true)

	lines := []string{
		bg.Width(innerW).Render(titleStyle.Render(" Pick a skin ")),
		bg.Width(innerW).Render(" "),
	}

	for i, name := range m.skins {
		var line string
		if i == m.cursor {
			line = selRow.Width(innerW).Render(" ▸ " + name)
		} else {
			line = bg.Width(innerW).Render("   " + name)
		}
		lines = append(lines, line)
	}

	lines = append(lines,
		bg.Width(innerW).Render(" "),
		bg.Width(innerW).Render(dim.Render("Enter: apply · ↑/↓: select · Esc: cancel")),
	)

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		BorderBackground(m.palette.Bg).
		Background(m.palette.Bg).
		Foreground(m.palette.Fg).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(
		width, height,
		lipgloss.Center, lipgloss.Center,
		box,
		lipgloss.WithWhitespaceBackground(m.palette.Bg),
		lipgloss.WithWhitespaceForeground(m.palette.Bg),
	)
}

// Title implements Modal.
func (m SkinPickerModel) Title() string { return "Skins" }
