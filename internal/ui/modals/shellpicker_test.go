package modals

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/torosent/c9s/internal/ui/theme"
)

func TestShellPicker_HotkeyB_PicksBash(t *testing.T) {
	picker := NewShellPicker("c1", "c1", theme.DefaultDark())
	_, cmd := picker.Update(tea.KeyPressMsg{Code: 'b', Text: "b"})
	if cmd == nil {
		t.Fatal("expected 'b' to return a cmd")
	}
	got := drainShellPickerBatch(cmd)
	if got.shell != "/bin/bash" {
		t.Errorf("Shell = %q, want /bin/bash", got.shell)
	}
	if got.id != "c1" {
		t.Errorf("ID = %q, want c1", got.id)
	}
	if !got.closed {
		t.Error("expected modal to also emit CloseModalMsg")
	}
}

func TestShellPicker_HotkeyS_PicksSh(t *testing.T) {
	picker := NewShellPicker("c1", "c1", theme.DefaultDark())
	_, cmd := picker.Update(tea.KeyPressMsg{Code: 's', Text: "s"})
	if cmd == nil {
		t.Fatal("expected 's' to return a cmd")
	}
	got := drainShellPickerBatch(cmd)
	if got.shell != "/bin/sh" {
		t.Errorf("Shell = %q, want /bin/sh", got.shell)
	}
}

func TestShellPicker_EnterPicksCursor(t *testing.T) {
	picker := NewShellPicker("c1", "c1", theme.DefaultDark())
	// cursor starts at 0 (bash); arrow down to sh
	pickerModel, _ := picker.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	picker = pickerModel.(ShellPickerModel)
	_, cmd := picker.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected enter to return a cmd")
	}
	got := drainShellPickerBatch(cmd)
	if got.shell != "/bin/sh" {
		t.Errorf("Shell = %q, want /bin/sh after Down+Enter", got.shell)
	}
}

func TestShellPicker_EscClosesWithoutPick(t *testing.T) {
	picker := NewShellPicker("c1", "c1", theme.DefaultDark())
	_, cmd := picker.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("expected esc to return a cmd")
	}
	if _, ok := cmd().(CloseModalMsg); !ok {
		t.Errorf("expected CloseModalMsg, got %T", cmd())
	}
}

func TestShellPicker_ViewMentionsContainerAndOptions(t *testing.T) {
	picker := NewShellPicker("c1", "abc1234567ab", theme.DefaultDark())
	view := picker.View(80, 24)
	for _, want := range []string{"abc1234567ab", "bash", "sh"} {
		if !strings.Contains(view, want) {
			t.Errorf("view should mention %q; got:\n%s", want, view)
		}
	}
}

type pickResult struct {
	id     string
	shell  string
	closed bool
}

func drainShellPickerBatch(cmd tea.Cmd) pickResult {
	out := pickResult{}
	if cmd == nil {
		return out
	}
	msg := cmd()
	batch, ok := msg.(tea.BatchMsg)
	if !ok {
		// Single-message cmd (e.g., direct ShellPickedMsg)
		switch m := msg.(type) {
		case ShellPickedMsg:
			out.id = m.ID
			out.shell = m.Shell
		case CloseModalMsg:
			out.closed = true
		}
		return out
	}
	for _, c := range batch {
		if c == nil {
			continue
		}
		switch m := c().(type) {
		case ShellPickedMsg:
			out.id = m.ID
			out.shell = m.Shell
		case CloseModalMsg:
			out.closed = true
		}
	}
	return out
}
