package builder

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestNew(t *testing.T) {
	m := New(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "Builder" {
		t.Errorf("Title=%q", m.Title())
	}
	km := m.Hotkeys()
	for _, name := range []string{"start", "stop", "delete"} {
		if _, ok := km.Get(name); !ok {
			t.Errorf("expected binding %q", name)
		}
	}
	if m.Init() == nil {
		t.Error("Init returned nil")
	}
}

func TestStatusMsgUpdates(t *testing.T) {
	m := New(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	s, _ := m.Update(statusMsg{State: "running", CPUs: 2, MemoryBytes: 1 << 30, UptimeSec: 3600})
	m = s.(*Model)
	view := m.View(80, 24)
	if !strings.Contains(view, "running") {
		t.Errorf("view should include running state: %q", view)
	}
	if !strings.Contains(view, "1.0G") {
		t.Errorf("view should include MEM: %q", view)
	}
	if !strings.Contains(view, "1h0m") {
		t.Errorf("view should include UPTIME: %q", view)
	}
	if !strings.Contains(m.Summary(), "running") {
		t.Errorf("Summary=%q", m.Summary())
	}
}

func TestStartStopRefreshKeys(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("r returned nil")
	}
	cmd()

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	if cmd == nil {
		t.Fatal("S returned nil")
	}
	cmd()

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	if cmd == nil {
		t.Fatal("X returned nil")
	}
	cmd()

	want := map[string]bool{"BuilderStatus": false, "BuilderStart": false, "BuilderStop": false}
	for _, c := range f.Calls {
		if _, ok := want[c]; ok {
			want[c] = true
		}
	}
	for k, v := range want {
		if !v {
			t.Errorf("expected call %q (calls=%v)", k, f.Calls)
		}
	}
}

func TestDeleteFlow(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if cmd == nil {
		t.Fatal("D returned nil")
	}
	if _, ok := cmd().(screens.OpenModalMsg); !ok {
		t.Error("expected confirm modal")
	}
	_, cmd = m.Update(modals.ConfirmResultMsg{Result: modals.ConfirmResult{Confirmed: true, Tag: "delete-builder"}})
	if cmd == nil {
		t.Fatal("nil cmd from confirm")
	}
	walk(t, cmd)
	found := false
	for _, c := range f.Calls {
		if c == "BuilderDelete" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected BuilderDelete: %v", f.Calls)
	}
}

func TestDeleteRejectedDoesNothing(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(modals.ConfirmResultMsg{Result: modals.ConfirmResult{Confirmed: false, Tag: "delete-builder"}})
	if cmd != nil {
		walk(t, cmd)
	}
	for _, c := range f.Calls {
		if c == "BuilderDelete" {
			t.Errorf("BuilderDelete should not run on cancel")
		}
	}
}

func TestErrorPropagation(t *testing.T) {
	f := cli.NewFake()
	f.BuilderStartErr = errString("boom")
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	if cmd == nil {
		t.Fatal("S nil")
	}
	msg := cmd()
	st, ok := msg.(screens.StatusMsg)
	if !ok {
		t.Fatalf("expected StatusMsg, got %T", msg)
	}
	if !strings.Contains(st.Toast, "failed") {
		t.Errorf("toast = %q", st.Toast)
	}
}

func TestFormatBytesAndUptime(t *testing.T) {
	if formatBytes(0) != "-" {
		t.Errorf("got %q", formatBytes(0))
	}
	if formatBytes(1<<30) != "1.0G" {
		t.Errorf("got %q", formatBytes(1<<30))
	}
	if formatBytes(2048) != "2.0K" {
		t.Errorf("got %q", formatBytes(2048))
	}
	if formatBytes(500) != "500B" {
		t.Errorf("got %q", formatBytes(500))
	}
	if formatUptime(0) != "-" {
		t.Errorf("got %q", formatUptime(0))
	}
	if formatUptime(time.Hour+30*time.Minute) != "1h30m" {
		t.Errorf("got %q", formatUptime(time.Hour+30*time.Minute))
	}
	if formatUptime(45*time.Second) != "45s" {
		t.Errorf("got %q", formatUptime(45*time.Second))
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func walk(t *testing.T, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		return
	}
	msg := cmd()
	if msg == nil {
		return
	}
	if b, ok := msg.(tea.BatchMsg); ok {
		for _, c := range b {
			walk(t, c)
		}
	}
}
