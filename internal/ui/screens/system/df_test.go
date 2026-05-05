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

func TestNewDF(t *testing.T) {
	m := NewDF(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "System df" {
		t.Errorf("Title=%q", m.Title())
	}
	if m.Init() == nil {
		t.Error("Init nil")
	}
}

func TestDFRefreshAndRender(t *testing.T) {
	f := cli.NewFake()
	f.SystemDFResp = cli.SystemDF{
		Images:     cli.DFSection{Count: 12, Active: 5, SizeBytes: 1 << 30, ReclaimBytes: 524288000},
		Containers: cli.DFSection{Count: 4, Active: 2, SizeBytes: 314572800},
		Volumes:    cli.DFSection{Count: 3, Active: 1, SizeBytes: 209715200, ReclaimBytes: 104857600},
	}
	m := NewDF(f, clock.NewFake(time.Now()), theme.DefaultDark())
	cmd := m.refreshCmd()
	msg := cmd()
	s, _ := m.Update(msg)
	m = s.(DFModel)
	view := m.View(120, 30)
	if !strings.Contains(view, "Images") || !strings.Contains(view, "Containers") || !strings.Contains(view, "Volumes") {
		t.Errorf("expected all sections in view: %q", view)
	}
	if !strings.Contains(view, "1.0G") {
		t.Errorf("expected images size in view: %q", view)
	}
	if !strings.Contains(m.Summary(), "total") {
		t.Errorf("Summary=%q", m.Summary())
	}
}

func TestDFKeyRefresh(t *testing.T) {
	f := cli.NewFake()
	m := NewDF(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("r nil")
	}
	cmd()
	if !contains(f.Calls, "SystemDF") {
		t.Errorf("expected SystemDF call: %v", f.Calls)
	}
}

func TestDFFormatBytes(t *testing.T) {
	cases := []struct {
		in   int64
		want string
	}{
		{0, "-"},
		{500, "500B"},
		{2048, "2.0K"},
		{1572864, "2M"},
		{1610612736, "1.5G"},
	}
	for _, tc := range cases {
		if got := formatBytes(tc.in); got != tc.want {
			t.Errorf("formatBytes(%d)=%q want %q", tc.in, got, tc.want)
		}
	}
}
