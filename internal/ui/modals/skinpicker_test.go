package modals

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestSkinPicker_ListsSkins(t *testing.T) {
	p := theme.DefaultDark()
	m := NewSkinPicker([]string{"dark", "light", "k9s-dracula"}, p)
	out := m.View(80, 24)
	for _, name := range []string{"dark", "light", "k9s-dracula"} {
		if !strings.Contains(out, name) {
			t.Errorf("View() should contain skin %q\n%s", name, out)
		}
	}
	if !strings.Contains(out, "Pick a skin") {
		t.Errorf("View() missing title 'Pick a skin'\n%s", out)
	}
}

func TestSkinPicker_TitleIsSkins(t *testing.T) {
	p := theme.DefaultDark()
	m := NewSkinPicker([]string{"dark"}, p)
	if got := m.Title(); got != "Skins" {
		t.Errorf("Title()=%q, want Skins", got)
	}
}

func TestSkinPicker_EnterEmitsSkinPickedMsg(t *testing.T) {
	p := theme.DefaultDark()
	m := NewSkinPicker([]string{"dark", "k9s-dracula"}, p)
	// Trigger size so the underlying list is laid out
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(SkinPickerModel)
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("Enter should produce a Cmd")
	}
	msg := cmd()
	picked, ok := msg.(SkinPickedMsg)
	if !ok {
		t.Fatalf("Enter should emit SkinPickedMsg, got %T", msg)
	}
	if picked.Name != "dark" {
		t.Errorf("expected first skin 'dark' picked, got %q", picked.Name)
	}
}

func TestSkinPicker_EscClosesModal(t *testing.T) {
	p := theme.DefaultDark()
	m := NewSkinPicker([]string{"dark"}, p)
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc should produce a Cmd")
	}
	if _, ok := cmd().(CloseModalMsg); !ok {
		t.Fatalf("Esc should emit CloseModalMsg, got %T", cmd())
	}
}

func TestSkinPicker_InitReturnsNil(t *testing.T) {
	m := NewSkinPicker(nil, theme.DefaultDark())
	if m.Init() != nil {
		t.Error("Init() should return nil")
	}
}
