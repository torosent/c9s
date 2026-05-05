package modals

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestNewTextInput(t *testing.T) {
	m := NewTextInput("create-dns", "DNS domain to create:", "internal.local", theme.DefaultDark())
	if m.Title() != "DNS domain to create:" {
		t.Errorf("Title=%q", m.Title())
	}
	v := m.View(80, 20)
	if !strings.Contains(v, "DNS domain to create:") {
		t.Errorf("expected prompt: %q", v)
	}
	if !strings.Contains(v, "internal.local") {
		t.Errorf("expected initial value: %q", v)
	}
}

func TestTextInputEnterEmitsResult(t *testing.T) {
	m := NewTextInput("create-dns", "Name?", "", theme.DefaultDark())
	for _, r := range "myzone.local" {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = m2.(TextInputModel)
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter returned nil")
	}
	gotResult := false
	gotClose := false
	collect(cmd, func(msg tea.Msg) {
		if r, ok := msg.(TextInputResultMsg); ok {
			gotResult = true
			if r.Result.Label != "create-dns" || r.Result.Value != "myzone.local" {
				t.Errorf("unexpected result: %+v", r.Result)
			}
		}
		if _, ok := msg.(CloseModalMsg); ok {
			gotClose = true
		}
	})
	if !gotResult || !gotClose {
		t.Errorf("expected both result and close, got result=%v close=%v", gotResult, gotClose)
	}
}

func TestTextInputEscEmitsCancel(t *testing.T) {
	m := NewTextInput("save", "Path?", "", theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc returned nil")
	}
	gotCancel := false
	collect(cmd, func(msg tea.Msg) {
		if c, ok := msg.(TextInputCancelledMsg); ok {
			gotCancel = true
			if c.Label != "save" {
				t.Errorf("Label=%q", c.Label)
			}
		}
	})
	if !gotCancel {
		t.Error("expected TextInputCancelledMsg")
	}
}

func TestTextInputValidatorBlocksSubmit(t *testing.T) {
	m := NewTextInput("tag", "Tag?", "", theme.DefaultDark()).
		WithValidator(func(v string) string {
			if v == "" {
				return "tag must not be empty"
			}
			return ""
		})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		// Should not produce a result/close pair if validation failed
		seenResult := false
		collect(cmd, func(msg tea.Msg) {
			if _, ok := msg.(TextInputResultMsg); ok {
				seenResult = true
			}
		})
		if seenResult {
			t.Error("expected no result when validator returns non-empty error")
		}
	}
}

func TestTextInputValidatorPassesAfterTyping(t *testing.T) {
	m := NewTextInput("tag", "Tag?", "", theme.DefaultDark()).
		WithValidator(func(v string) string {
			if v == "" {
				return "empty"
			}
			return ""
		})
	for _, r := range "v1" {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = m2.(TextInputModel)
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter cmd nil")
	}
	gotResult := false
	collect(cmd, func(msg tea.Msg) {
		if _, ok := msg.(TextInputResultMsg); ok {
			gotResult = true
		}
	})
	if !gotResult {
		t.Error("expected result after passing validation")
	}
}

func TestTextInputViewShowsValidatorMessage(t *testing.T) {
	m := NewTextInput("x", "x?", "", theme.DefaultDark()).
		WithValidator(func(v string) string {
			return "must include letters"
		})
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = m2.(TextInputModel)
	v := m.View(80, 20)
	if !strings.Contains(v, "must include letters") {
		t.Errorf("expected validator message in view: %q", v)
	}
}
