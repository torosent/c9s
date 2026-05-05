package modals

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/torosent/c9s/internal/ui/theme"
)

// ShellPickerModel is a small two-option picker for choosing the
// in-container shell (bash or sh). The host's $SHELL is irrelevant —
// what matters is what's on PATH inside the container — so we let the
// user pick rather than guess. POSIX requires /bin/sh in essentially
// every Linux container; /bin/bash is common but not universal.
type ShellPickerModel struct {
	palette     theme.Palette
	containerID string
	shortID     string
	options     []shellOption
	cursor      int
}

type shellOption struct {
	key   rune
	label string
	path  string
}

// ShellPickedMsg is emitted when a shell is selected. The containers
// screen catches it and converts it to screens.SuspendShellMsg.
type ShellPickedMsg struct {
	ID    string
	Shell string
}

// NewShellPicker creates a new shell-picker modal for the given
// container.
func NewShellPicker(containerID, shortID string, p theme.Palette) ShellPickerModel {
	return ShellPickerModel{
		palette:     p,
		containerID: containerID,
		shortID:     shortID,
		options: []shellOption{
			{key: 'b', label: "bash  (/bin/bash)", path: "/bin/bash"},
			{key: 's', label: "sh    (/bin/sh)", path: "/bin/sh"},
		},
		cursor: 0,
	}
}

// Init implements Modal.
func (m ShellPickerModel) Init() tea.Cmd { return nil }

// Update implements Modal.
func (m ShellPickerModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	if key, ok := msg.(tea.KeyPressMsg); ok {
		switch key.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			return m.pick(m.options[m.cursor])
		case "esc", "q":
			return m, func() tea.Msg { return CloseModalMsg{} }
		}
		// Direct hot-letter selection: 'b' or 's'.
		if t := key.Text; len(t) == 1 {
			r := rune(t[0])
			for _, opt := range m.options {
				if r == opt.key {
					return m.pick(opt)
				}
			}
		}
	}
	return m, nil
}

func (m ShellPickerModel) pick(opt shellOption) (Modal, tea.Cmd) {
	id := m.containerID
	path := opt.path
	return m, tea.Batch(
		func() tea.Msg { return ShellPickedMsg{ID: id, Shell: path} },
		func() tea.Msg { return CloseModalMsg{} },
	)
}

// View implements Modal.
func (m ShellPickerModel) View(width, height int) string {
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
	keyHint := lipgloss.NewStyle().Foreground(m.palette.Accent).Background(m.palette.Bg).Bold(true)

	subject := m.shortID
	if subject == "" {
		subject = "container"
	}
	header := fmt.Sprintf("Open shell in %s", subject)

	lines := []string{
		bg.Width(innerW).Render(titleStyle.Render(" " + header + " ")),
		bg.Width(innerW).Render(" "),
	}

	for i, opt := range m.options {
		hint := keyHint.Render(fmt.Sprintf("[%c] ", opt.key))
		var row string
		if i == m.cursor {
			row = selRow.Width(innerW).Render(" ▸ " + string(opt.key) + "  " + opt.label)
		} else {
			row = bg.Width(innerW).Render("   " + hint + bg.Render(opt.label))
		}
		lines = append(lines, row)
	}

	lines = append(lines,
		bg.Width(innerW).Render(" "),
		bg.Width(innerW).Render(dim.Render("b/s: pick • ↑/↓+Enter: pick • Esc: cancel")),
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
		lipgloss.WithWhitespaceStyle(lipgloss.NewStyle().Background(m.palette.Bg).Foreground(m.palette.Bg)),
	)
}

// Title implements Modal.
func (m ShellPickerModel) Title() string { return "Shell" }
