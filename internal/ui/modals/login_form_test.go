package modals

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestNewLoginPrefillsHost(t *testing.T) {
	m := NewLogin("ghcr.io", theme.DefaultDark())
	if !strings.Contains(m.View(80, 20), "ghcr.io") {
		t.Errorf("expected prefilled host in view")
	}
}

func TestLoginPasswordIsMaskedInRender(t *testing.T) {
	m := NewLogin("", theme.DefaultDark())
	// type into host first
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab}) // move to user
	m = m2.(LoginModel)
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // move to password
	m = m2.(LoginModel)
	for _, r := range "topsecret" {
		m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = m2.(LoginModel)
	}
	view := m.View(80, 20)
	if strings.Contains(view, "topsecret") {
		t.Errorf("expected password to be masked, got: %q", view)
	}
	if !strings.Contains(view, "*") {
		t.Errorf("expected '*' masking in view: %q", view)
	}
}

func TestLoginEnterCascadesToSubmit(t *testing.T) {
	m := NewLogin("", theme.DefaultDark())
	for _, r := range "ghcr.io" {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = m2.(LoginModel)
	}
	m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = m2.(LoginModel)
	for _, r := range "alice" {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = m2.(LoginModel)
	}
	m2, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = m2.(LoginModel)
	for _, r := range "passw0rd" {
		m2, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = m2.(LoginModel)
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd from final Enter")
	}
	gotResult := false
	gotClose := false
	collect(cmd, func(msg tea.Msg) {
		if r, ok := msg.(LoginResultMsg); ok {
			gotResult = true
			if r.Result.Host != "ghcr.io" || r.Result.Username != "alice" || r.Result.Password != "passw0rd" {
				t.Errorf("unexpected result: %+v", r.Result)
			}
		}
		if _, ok := msg.(CloseModalMsg); ok {
			gotClose = true
		}
	})
	if !gotResult || !gotClose {
		t.Errorf("expected both LoginResultMsg and CloseModalMsg, got result=%v close=%v", gotResult, gotClose)
	}
}

func TestLoginEscEmitsCancel(t *testing.T) {
	m := NewLogin("", theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if cmd == nil {
		t.Fatal("Esc returned nil cmd")
	}
	gotCancel := false
	collect(cmd, func(msg tea.Msg) {
		if _, ok := msg.(LoginCancelledMsg); ok {
			gotCancel = true
		}
	})
	if !gotCancel {
		t.Error("expected LoginCancelledMsg")
	}
}

func TestLoginShiftTabCycles(t *testing.T) {
	m := NewLogin("", theme.DefaultDark())
	for _, kt := range []tea.KeyType{tea.KeyShiftTab, tea.KeyTab, tea.KeyShiftTab} {
		m2, _ := m.Update(tea.KeyMsg{Type: kt})
		m = m2.(LoginModel)
	}
	// Just ensure nothing panics; render once
	_ = m.View(80, 20)
}

func TestLoginInitReturnsBlinker(t *testing.T) {
	m := NewLogin("", theme.DefaultDark())
	if m.Init() == nil {
		t.Error("expected non-nil cmd from Init")
	}
}

func TestLoginTitleAndView(t *testing.T) {
	m := NewLogin("", theme.DefaultDark())
	if m.Title() != "Registry login" {
		t.Errorf("Title=%q", m.Title())
	}
	v := m.View(80, 20)
	if !strings.Contains(v, "Host:") || !strings.Contains(v, "User:") || !strings.Contains(v, "Password:") {
		t.Errorf("view missing labels: %q", v)
	}
}

// collect drains a tea.Cmd recursively, invoking visit on every produced message.
func collect(cmd tea.Cmd, visit func(tea.Msg)) {
	if cmd == nil {
		return
	}
	msg := cmd()
	if msg == nil {
		return
	}
	if b, ok := msg.(tea.BatchMsg); ok {
		for _, c := range b {
			collect(c, visit)
		}
		return
	}
	visit(msg)
}
