package system

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestNewLogs(t *testing.T) {
	m := NewLogs(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "System logs" {
		t.Errorf("Title=%q", m.Title())
	}
}

func TestLogsStreamsAndAppends(t *testing.T) {
	f := cli.NewFake()
	f.ReplaySystemLogStream([]cli.StreamEvent{
		cli.LogLine{Raw: "INFO: started", Level: "INFO"},
		cli.RawLine{Text: "ready"},
	}, 0)
	m := NewLogs(f, clock.NewFake(time.Now()), theme.DefaultDark())
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init nil")
	}
	// Drain events through the message pipeline
	for i := 0; i < 5; i++ {
		msg := cmd()
		if msg == nil {
			break
		}
		s, next := m.Update(msg)
		m = s.(*LogsModel)
		if next == nil {
			break
		}
		cmd = next
	}
	view := m.View(80, 24)
	if !strings.Contains(view, "started") {
		t.Errorf("expected log content in view: %q", view)
	}
	if !strings.Contains(m.Summary(), "lines") {
		t.Errorf("Summary=%q", m.Summary())
	}
	m.Cancel()
}

func TestLogsKeyboardScroll(t *testing.T) {
	f := cli.NewFake()
	m := NewLogs(f, clock.NewFake(time.Now()), theme.DefaultDark())
	for _, key := range []string{"j", "k", "g", "G"} {
		s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
		m = s.(*LogsModel)
	}
	_ = m.View(80, 24)
}

func TestColorizeLevel(t *testing.T) {
	// In CI without a TTY, lipgloss may return the input unchanged.
	// We just exercise the function for all branches and verify it
	// doesn't panic + returns a non-empty string for known/unknown
	// levels.
	for _, lvl := range []string{"", "INFO", "WARN", "ERROR", "DEBUG", "UNKNOWN"} {
		out := colorizeLevel("hi", lvl)
		if out == "" {
			t.Errorf("colorizeLevel(%q) returned empty string", lvl)
		}
	}
}
