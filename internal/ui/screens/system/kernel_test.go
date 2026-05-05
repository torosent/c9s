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

func TestNewKernel(t *testing.T) {
	m := NewKernel(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "System kernel" {
		t.Errorf("Title=%q", m.Title())
	}
	if m.Init() == nil {
		t.Error("Init nil")
	}
}

func TestKernelRefreshFiltersByPrefix(t *testing.T) {
	f := cli.NewFake()
	f.ListSystemPropertiesResp = []cli.SystemProperty{
		{Key: "kernel.version", Value: "5.10", ReadOnly: true},
		{Key: "kernel.cpus", Value: "4"},
		{Key: "registry.default", Value: "ghcr.io"},
	}
	m := NewKernel(f, clock.NewFake(time.Now()), theme.DefaultDark())
	cmd := m.refreshCmd()
	msg := cmd()
	s, _ := m.Update(msg)
	m = s.(KernelModel)
	v := m.View(80, 24)
	if !strings.Contains(v, "kernel.version") {
		t.Errorf("expected kernel.version in view: %q", v)
	}
	if !strings.Contains(v, "(ro)") {
		t.Errorf("expected (ro) marker for read-only: %q", v)
	}
	if strings.Contains(v, "registry.default") {
		t.Errorf("did not expect non-kernel property: %q", v)
	}
	if !strings.Contains(m.Summary(), "2 kernel properties") {
		t.Errorf("Summary=%q", m.Summary())
	}
}

func TestKernelEmpty(t *testing.T) {
	m := NewKernel(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	s, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = s.(KernelModel)
	v := m.View(80, 30)
	if !strings.Contains(v, "no kernel.*") {
		t.Errorf("expected empty hint: %q", v)
	}
	if !strings.Contains(m.Summary(), "no kernel properties") {
		t.Errorf("Summary=%q", m.Summary())
	}
}

func TestKernelRefreshKey(t *testing.T) {
	f := cli.NewFake()
	m := NewKernel(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("r nil")
	}
	cmd()
	if !contains(f.Calls, "ListSystemProperties") {
		t.Errorf("expected ListSystemProperties: %v", f.Calls)
	}
}
