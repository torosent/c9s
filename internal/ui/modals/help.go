package modals

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/theme"
)

// HelpModel is a modal that displays keybindings.
type HelpModel struct {
	keymap      *keymap.Map
	screenTitle string
	palette     theme.Palette
}

// NewHelp creates a new help modal.
func NewHelp(km *keymap.Map, screenTitle string, p theme.Palette) HelpModel {
	return HelpModel{
		keymap:      km,
		screenTitle: screenTitle,
		palette:     p,
	}
}

// Init implements Modal.
func (m HelpModel) Init() tea.Cmd {
	return nil
}

// Update implements Modal.
func (m HelpModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	switch msg.(type) {
	case tea.KeyMsg:
		// Any key closes the help modal
		return m, func() tea.Msg {
			return CloseModalMsg{}
		}
	}
	return m, nil
}

// View implements Modal.
func (m HelpModel) View(width, height int) string {
	title := fmt.Sprintf("%s — keybinds", m.screenTitle)

	var content strings.Builder
	content.WriteString(title)
	content.WriteString("\n\n")

	// Get all bindings
	names := m.keymap.Names()

	// Determine layout based on width
	useTwoColumns := width >= 80

	if useTwoColumns {
		// Two column layout
		midpoint := (len(names) + 1) / 2
		leftCol := names[:midpoint]
		rightCol := names[midpoint:]

		for i := 0; i < midpoint; i++ {
			// Left column
			name := leftCol[i]
			b, _ := m.keymap.Get(name)
			left := fmt.Sprintf("%-30s %s", b.Help, formatKeys(b.Keys))

			// Right column (if available)
			right := ""
			if i < len(rightCol) {
				name := rightCol[i]
				b, _ := m.keymap.Get(name)
				right = fmt.Sprintf("%-30s %s", b.Help, formatKeys(b.Keys))
			}

			content.WriteString(fmt.Sprintf("%-40s  %s\n", left, right))
		}
	} else {
		// Single column layout
		for _, name := range names {
			b, _ := m.keymap.Get(name)
			content.WriteString(fmt.Sprintf("%-30s %s\n", b.Help, formatKeys(b.Keys)))
		}
	}

	content.WriteString("\nPress any key to close")

	// Style with border
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		BorderBackground(m.palette.Bg).
		Background(m.palette.Bg).
		Foreground(m.palette.Fg).
		Padding(1, 2)

	return boxStyle.Render(content.String())
}

// Title implements Modal.
func (m HelpModel) Title() string {
	return "Help"
}

// formatKeys formats the key list for display.
func formatKeys(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	return "[" + strings.Join(keys, ", ") + "]"
}
