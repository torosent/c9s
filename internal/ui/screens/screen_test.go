package screens

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/ui/keymap"
)

// noopScreen is a stub implementation for testing.
type noopScreen struct{}

func (n *noopScreen) Init() tea.Cmd {
	return nil
}

func (n *noopScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	return n, nil
}

func (n *noopScreen) View(width, height int) string {
	return "noop screen"
}

func (n *noopScreen) Title() string {
	return "noop"
}

func (n *noopScreen) Hotkeys() *keymap.Map {
	return keymap.Default()
}

func (n *noopScreen) Summary() string {
	return "noop screen summary"
}

// Compile-time assertion that noopScreen implements Screen.
var _ Screen = (*noopScreen)(nil)

func TestNoopScreenImplementsInterface(t *testing.T) {
	s := &noopScreen{}

	// Test that all methods return non-empty values
	if s.Title() == "" {
		t.Error("Title() returned empty string")
	}

	if s.Summary() == "" {
		t.Error("Summary() returned empty string")
	}

	if s.Hotkeys() == nil {
		t.Error("Hotkeys() returned nil")
	}

	view := s.View(80, 24)
	if view == "" {
		t.Error("View() returned empty string")
	}
}
