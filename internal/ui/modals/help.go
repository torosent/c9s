package modals

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
	case tea.KeyPressMsg:
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

	// Build [(label, keys)] entries in stable name-sorted order so help
	// stays predictable across runs and does not jitter between screens.
	names := m.keymap.Names()
	type entry struct{ label, keys string }
	entries := make([]entry, 0, len(names))
	for _, n := range names {
		b, ok := m.keymap.Get(n)
		if !ok {
			continue
		}
		entries = append(entries, entry{label: b.Help, keys: formatKeys(b.Keys)})
	}

	// Compute column widths from data so labels and key strings each have
	// a stable left edge regardless of which screen is active.
	maxLabel, maxKeys := 0, 0
	for _, e := range entries {
		if l := visibleLen(e.label); l > maxLabel {
			maxLabel = l
		}
		if l := visibleLen(e.keys); l > maxKeys {
			maxKeys = l
		}
	}
	const labelKeyGap = 2
	const colSep = 4
	pairW := maxLabel + labelKeyGap + maxKeys
	totalTwoColW := pairW + colSep + pairW

	// Two columns when there's room; otherwise one. Account for the
	// rounded border + padding (4 cols) added by the box style below.
	useTwoCols := width-4 >= totalTwoColW

	renderRow := func(e entry) string {
		return padRight(e.label, maxLabel) + strings.Repeat(" ", labelKeyGap) + padRight(e.keys, maxKeys)
	}

	if useTwoCols {
		mid := (len(entries) + 1) / 2
		left := entries[:mid]
		right := entries[mid:]
		gap := strings.Repeat(" ", colSep)
		for i := 0; i < mid; i++ {
			l := renderRow(left[i])
			r := ""
			if i < len(right) {
				r = renderRow(right[i])
			}
			content.WriteString(l)
			if r != "" {
				content.WriteString(gap)
				content.WriteString(r)
			}
			content.WriteString("\n")
		}
	} else {
		for _, e := range entries {
			content.WriteString(renderRow(e))
			content.WriteString("\n")
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

// visibleLen returns the visible-rune count of s. It's intentionally
// simple — help labels and keys are ASCII / common punctuation, so a
// rune count matches the on-screen column count.
func visibleLen(s string) int {
	return len([]rune(s))
}

// padRight returns s padded with trailing spaces so its visible length
// equals at least n columns.
func padRight(s string, n int) string {
	if l := visibleLen(s); l < n {
		return s + strings.Repeat(" ", n-l)
	}
	return s
}
