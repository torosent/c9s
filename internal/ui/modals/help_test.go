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

// TestHelpModel_TwoColumnAlignment is a regression test for the v0.1.3
// help-modal alignment bug: every right-side cell ended at a different
// position because the layout used "%-40s" to pad the combined "label +
// keys" string, but the keys segment varied in length. The fix renders
// each cell as label-padded + gap + keys-padded so columns are stable.
//
// The assertion strips ANSI styling and looks for a stable column where
// the right-side keys all begin (every right cell's "[" should sit at
// the same column after the cross-cell separator).
func TestHelpModel_TwoColumnAlignment(t *testing.T) {
	km := keymap.Default()
	km.Add("shell", keymap.Binding{Keys: []string{"s"}, Help: "Shell"})
	km.Add("logs", keymap.Binding{Keys: []string{"l", "enter"}, Help: "Logs"})
	km.Add("inspect", keymap.Binding{Keys: []string{"d"}, Help: "Details"})
	km.Add("stop", keymap.Binding{Keys: []string{"x"}, Help: "Stop"})
	km.Add("kill", keymap.Binding{Keys: []string{"shift+k", "K"}, Help: "Kill"})
	km.Add("restart", keymap.Binding{Keys: []string{"shift+r", "R"}, Help: "Restart"})
	km.Add("delete", keymap.Binding{Keys: []string{"shift+d", "D"}, Help: "Delete"})
	km.Add("prune", keymap.Binding{Keys: []string{"shift+p", "P"}, Help: "Prune"})

	m := NewHelp(km, "Containers", theme.DefaultDark())
	out := stripAnsiInTest(m.View(140, 30))

	// Find each "[" character on every line and bucket by line. With
	// proper alignment, each line in the two-column body has exactly two
	// "[" characters and they sit at two stable columns.
	firstBracketCol, secondBracketCol := -1, -1
	for _, line := range strings.Split(out, "\n") {
		// Skip border / title / dismiss-hint rows.
		if !strings.Contains(line, "[") {
			continue
		}
		first := strings.Index(line, "[")
		second := strings.Index(line[first+1:], "[")
		if second < 0 {
			continue // single-column row
		}
		second += first + 1

		if firstBracketCol == -1 {
			firstBracketCol = first
		} else if first != firstBracketCol {
			t.Errorf("left-column [ shifted: was at col %d, now at col %d on line %q",
				firstBracketCol, first, line)
		}
		if secondBracketCol == -1 {
			secondBracketCol = second
		} else if second != secondBracketCol {
			t.Errorf("right-column [ shifted: was at col %d, now at col %d on line %q",
				secondBracketCol, second, line)
		}
	}

	if firstBracketCol == -1 {
		t.Fatal("no two-column body rows found in help output")
	}
}

func TestHelpModel_IncludesSortAndScreenSwitch(t *testing.T) {
	m := NewHelp(keymap.Default(), "Test", theme.DefaultDark())
	out := stripAnsiInTest(m.View(140, 30))
	for _, want := range []string{"sort", "shift+s, S", "switch screen", "1-9, 0"} {
		if !strings.Contains(out, want) {
			t.Errorf("help output missing %q; full text:\n%s", want, out)
		}
	}
}

func stripAnsiInTest(s string) string {
	var b strings.Builder
	skip := false
	for _, r := range s {
		if r == '\x1b' {
			skip = true
			continue
		}
		if skip {
			if r == 'm' {
				skip = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
