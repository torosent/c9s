package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/torosent/c9s/internal/ui/theme"
)

func TestSplashViewMentionsVersion(t *testing.T) {
	s := NewSplash(theme.DefaultDark(), "c9s 0.1.0 (commit abc, built 2026-05-02)")
	out := stripANSI(s.View())
	if !strings.Contains(out, "c9s") {
		t.Errorf("View missing 'c9s': %q", out)
	}
	if !strings.Contains(out, "0.1.0") {
		t.Errorf("View missing version 0.1.0: %q", out)
	}
}

func TestSplashAnyKeyEmitsDoneMsg(t *testing.T) {
	s := NewSplash(theme.DefaultDark(), "c9s 0.1.0")
	_, cmd := s.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd == nil {
		t.Fatal("expected a tea.Cmd")
	}
	msg := cmd()
	if _, ok := msg.(SplashDoneMsg); !ok {
		t.Errorf("expected SplashDoneMsg, got %T", msg)
	}
}
