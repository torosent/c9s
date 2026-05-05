package modals

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestInfoModel_AnyKeyDismisses(t *testing.T) {
	m := NewInfo("Title", []string{"line one", "line two"}, InfoOK, theme.DefaultDark())
	for _, key := range []tea.KeyPressMsg{
		tea.KeyPressMsg{Code: tea.KeyEnter},
		tea.KeyPressMsg{Code: tea.KeyEsc},
		tea.KeyPressMsg{Code: 'q', Text: "q"},
		tea.KeyPressMsg{Code: ' ', Text: " "},
	} {
		_, cmd := m.Update(key)
		if cmd == nil {
			t.Fatalf("key %v should dismiss but no cmd returned", key)
		}
		msg := cmd()
		if _, ok := msg.(CloseModalMsg); !ok {
			t.Errorf("key %v: expected CloseModalMsg, got %T", key, msg)
		}
	}
}

func TestInfoModel_RendersTitleAndBody(t *testing.T) {
	m := NewInfo("Hello", []string{"alpha", "beta"}, InfoWarning, theme.DefaultDark())
	out := m.View(120, 20)
	if !strings.Contains(out, "Hello") {
		t.Errorf("View missing title: %q", out)
	}
	if !strings.Contains(out, "alpha") || !strings.Contains(out, "beta") {
		t.Errorf("View missing body lines: %q", out)
	}
	if !strings.Contains(out, "press any key to dismiss") {
		t.Errorf("View missing dismiss hint: %q", out)
	}
}

func TestInfoModel_NonKeyMessageDoesNotDismiss(t *testing.T) {
	m := NewInfo("Title", []string{"body"}, InfoOK, theme.DefaultDark())
	_, cmd := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if cmd != nil {
		t.Errorf("WindowSizeMsg should not dismiss; got cmd=%v", cmd)
	}
}
