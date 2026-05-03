package modals

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestConfirmRendersTitleBodyLines(t *testing.T) {
	p := theme.DefaultDark()
	m := NewConfirm("Delete containers", "This will permanently remove:", []string{"container-a", "container-b"}, "delete", p)
	v := m.View(80, 20)

	// Remove ANSI codes to check content
	stripped := stripANSI(v)

	if !strings.Contains(stripped, "Delete containers") {
		t.Errorf("expected title 'Delete containers' in view")
	}
	if !strings.Contains(stripped, "This will permanently remove:") {
		t.Errorf("expected body text in view")
	}
	if !strings.Contains(stripped, "container-a") {
		t.Errorf("expected line 'container-a' in view")
	}
	if !strings.Contains(stripped, "container-b") {
		t.Errorf("expected line 'container-b' in view")
	}
	if !strings.Contains(stripped, "[y] yes") {
		t.Errorf("expected '[y] yes' in footer")
	}
	if !strings.Contains(stripped, "[n/Esc] cancel") {
		t.Errorf("expected '[n/Esc] cancel' in footer")
	}
}

func TestConfirmYesEmitsConfirmedTrue(t *testing.T) {
	p := theme.DefaultDark()
	m := NewConfirm("Delete", "Sure?", []string{}, "delete", p)

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'y'}}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected Update to return a cmd on 'y' keypress")
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", msg)
	}

	var foundConfirmResult, foundClose bool
	for _, c := range batch {
		msg := c()
		switch m := msg.(type) {
		case ConfirmResultMsg:
			foundConfirmResult = true
			if !m.Result.Confirmed {
				t.Errorf("expected Confirmed=true, got false")
			}
			if m.Result.Tag != "delete" {
				t.Errorf("expected Tag='delete', got %q", m.Result.Tag)
			}
		case CloseModalMsg:
			foundClose = true
		}
	}
	if !foundConfirmResult {
		t.Error("expected ConfirmResultMsg in batch")
	}
	if !foundClose {
		t.Error("expected CloseModalMsg in batch")
	}
}

func TestConfirmNoEmitsConfirmedFalse(t *testing.T) {
	p := theme.DefaultDark()
	m := NewConfirm("Cancel test", "", []string{}, "delete", p)

	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected Update to return a cmd on 'n' keypress")
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", msg)
	}

	var foundConfirmResult bool
	for _, c := range batch {
		msg := c()
		if m, ok := msg.(ConfirmResultMsg); ok {
			foundConfirmResult = true
			if m.Result.Confirmed {
				t.Errorf("expected Confirmed=false, got true")
			}
		}
	}
	if !foundConfirmResult {
		t.Error("expected ConfirmResultMsg in batch")
	}
}

func TestConfirmEscEmitsConfirmedFalse(t *testing.T) {
	p := theme.DefaultDark()
	m := NewConfirm("Cancel test", "", []string{}, "delete", p)

	keyMsg := tea.KeyMsg{Type: tea.KeyEsc}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected Update to return a cmd on Esc keypress")
	}

	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		t.Fatalf("expected tea.BatchMsg, got %T", msg)
	}

	var foundConfirmResult bool
	for _, c := range batch {
		msg := c()
		if m, ok := msg.(ConfirmResultMsg); ok {
			foundConfirmResult = true
			if m.Result.Confirmed {
				t.Errorf("expected Confirmed=false on Esc, got true")
			}
		}
	}
	if !foundConfirmResult {
		t.Error("expected ConfirmResultMsg in batch")
	}
}

func TestConfirmTruncatesNarrowWidth(t *testing.T) {
	p := theme.DefaultDark()
	m := NewConfirm("Very long title that should be truncated", "Body text", []string{"line1", "line2"}, "test", p)

	v := m.View(40, 10)

	stripped := stripANSI(v)
	lines := strings.Split(stripped, "\n")
	for _, line := range lines {
		visibleLen := len([]rune(line))
		if visibleLen > 45 { // Allow a bit of margin for borders
			t.Errorf("line too wide (%d runes): %q", visibleLen, line)
		}
	}
}

// stripANSI removes ANSI escape codes from a string.
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for _, ch := range s {
		if ch == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(ch)
	}
	return result.String()
}
