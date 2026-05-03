// Package modals provides modal dialogs for c9s.
package modals

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/torosent/c9s/internal/ui/theme"
)

// SortColumn describes a sortable column.
type SortColumn struct {
	Key   string
	Label string
}

// SortPickerModel is a modal for selecting sort column. We use a
// hand-rolled list (rather than bubbles/list) so every cell is painted
// with the active skin's bg — bubbles/list emits ANSI resets between
// item titles and the surrounding cells, which leak terminal-default
// (black) bg through skinned modals.
type SortPickerModel struct {
	palette theme.Palette
	columns []SortColumn
	cursor  int
	reverse bool
	width   int
	height  int
}

// NewSortPicker creates a new sort picker modal.
func NewSortPicker(columns []SortColumn, p theme.Palette) SortPickerModel {
	return SortPickerModel{
		palette: p,
		columns: columns,
		cursor:  0,
		reverse: false,
	}
}

// Init implements Modal.
func (m SortPickerModel) Init() tea.Cmd {
	return nil
}

// SortPickedMsg is emitted when a sort column is picked.
type SortPickedMsg struct {
	Key     string
	Reverse bool
}

// Update implements Modal.
func (m SortPickerModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.columns)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			if m.cursor >= 0 && m.cursor < len(m.columns) {
				col := m.columns[m.cursor]
				rev := m.reverse
				return m, func() tea.Msg {
					return SortPickedMsg{Key: col.Key, Reverse: rev}
				}
			}
		case "r", "R":
			m.reverse = !m.reverse
			return m, nil
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
func (m SortPickerModel) View(width, height int) string {
	innerW := 44
	if width < innerW+8 {
		innerW = width - 8
		if innerW < 24 {
			innerW = 24
		}
	}

	bg := lipgloss.NewStyle().Foreground(m.palette.Fg).Background(m.palette.Bg)
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(m.palette.HeaderFg).Background(m.palette.Accent).Padding(0, 1)
	accentText := lipgloss.NewStyle().Foreground(m.palette.Accent).Background(m.palette.Bg).Bold(true)
	dim := lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg)
	selRow := lipgloss.NewStyle().Foreground(m.palette.Bg).Background(m.palette.Accent).Bold(true)

	direction := "ascending"
	if m.reverse {
		direction = "descending"
	}

	lines := []string{
		bg.Width(innerW).Render(accentText.Render("Sort Direction: ") + bg.Render(direction)),
		bg.Width(innerW).Render(" "),
		bg.Width(innerW).Render(titleStyle.Render(" Select Sort Column ")),
		bg.Width(innerW).Render(" "),
	}

	for i, col := range m.columns {
		var line string
		if i == m.cursor {
			line = selRow.Width(innerW).Render(" ▸ " + col.Label)
		} else {
			line = bg.Width(innerW).Render("   " + col.Label)
		}
		lines = append(lines, line)
	}

	lines = append(lines,
		bg.Width(innerW).Render(" "),
		bg.Width(innerW).Render(dim.Render("↑/↓: select • Enter: apply • r: reverse • Esc: cancel")),
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
func (m SortPickerModel) Title() string {
	return "Sort"
}
