package modals

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestHelpListsAllBindings(t *testing.T) {
	km := keymap.Default()
	p := theme.DefaultDark()
	m := NewHelp(km, "Containers", p)

	v := m.View(120, 40)
	stripped := stripANSI(v)

	// Check title
	if !strings.Contains(stripped, "Containers") {
		t.Errorf("expected 'Containers' in title")
	}
	if !strings.Contains(stripped, "keybinds") {
		t.Errorf("expected 'keybinds' in title")
	}

	// Check that global bindings appear
	expectedBindings := []string{
		"quit", "help", "filter", "palette", "mark", "mark_all",
		"refresh", "up", "down", "escape", "header_toggle",
	}

	for _, name := range expectedBindings {
		b, ok := km.Get(name)
		if !ok {
			continue // Skip if binding doesn't exist
		}
		// Check that either the Help text or at least one of the Keys appears
		found := strings.Contains(stripped, b.Help)
		if !found && len(b.Keys) > 0 {
			found = strings.Contains(stripped, b.Keys[0])
		}
		if !found {
			t.Errorf("expected binding %q (Help: %q, Keys: %v) to appear in help view", name, b.Help, b.Keys)
		}
	}
}

func TestHelpAnyKeyCloses(t *testing.T) {
	km := keymap.Default()
	p := theme.DefaultDark()
	m := NewHelp(km, "Test", p)

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected Update to return a cmd on keypress")
	}

	msg := cmd()
	if _, ok := msg.(CloseModalMsg); !ok {
		t.Errorf("expected CloseModalMsg, got %T", msg)
	}
}
