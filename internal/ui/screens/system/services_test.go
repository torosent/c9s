package system

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/theme"
)

func sampleServices() []cli.SystemService {
	return []cli.SystemService{
		{Name: "container-runtime", State: "running", PID: 1234, UptimeSec: 3600},
		{Name: "container-network", State: "running", PID: 1235, UptimeSec: 3600},
		{Name: "container-builder", State: "stopped"},
	}
}

func feedServices(t *testing.T, m ServicesModel, ss []cli.SystemService) ServicesModel {
	t.Helper()
	snap := state.Snapshot[cli.SystemService]{Items: ss, FetchedAt: time.Now()}
	s, _ := m.Update(state.RefreshedMsg[cli.SystemService]{
		Resource: cli.ResourceSystem,
		Snapshot: snap,
	})
	return s.(ServicesModel)
}

func TestNewServices(t *testing.T) {
	m := NewServices(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "System services" {
		t.Errorf("Title=%q", m.Title())
	}
	if m.Init() == nil {
		t.Error("Init nil")
	}
}

func TestServicesRender(t *testing.T) {
	m := NewServices(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedServices(t, m, sampleServices())
	v := m.View(120, 30)
	if !strings.Contains(v, "container-runtime") {
		t.Errorf("expected runtime in view: %q", v)
	}
	if !strings.Contains(m.Summary(), "3 services · 2 running") {
		t.Errorf("Summary=%q", m.Summary())
	}
}

func TestServicesStartStop(t *testing.T) {
	f := cli.NewFake()
	m := NewServices(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	if cmd == nil {
		t.Fatal("S nil")
	}
	if msg := cmd(); msg == nil {
		t.Errorf("expected status msg")
	}
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
	if cmd == nil {
		t.Fatal("X nil")
	}
	cmd()
	wants := []string{"SystemStartAll", "SystemStopAll"}
	for _, w := range wants {
		if !contains(f.Calls, w) {
			t.Errorf("expected %s in calls: %v", w, f.Calls)
		}
	}
}

func TestServicesRefresh(t *testing.T) {
	f := cli.NewFake()
	m := NewServices(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("r nil")
	}
	cmd()
	if !contains(f.Calls, "ListSystemServices") {
		t.Errorf("expected ListSystemServices call: %v", f.Calls)
	}
}

func TestServicesFilter(t *testing.T) {
	m := NewServices(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedServices(t, m, sampleServices())
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = s.(ServicesModel)
	for _, r := range "build" {
		s, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = s.(ServicesModel)
	}
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = s.(ServicesModel)
	v := m.View(120, 30)
	if !strings.Contains(v, "container-builder") {
		t.Errorf("expected builder in filtered view: %q", v)
	}
}

func TestServicesStartAllErr(t *testing.T) {
	f := cli.NewFake()
	f.SystemStartAllErr = errString("nope")
	m := NewServices(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'S'}})
	msg := cmd()
	st, ok := msg.(screens.StatusMsg)
	if !ok {
		t.Fatalf("got %T", msg)
	}
	if !strings.Contains(st.Toast, "failed") {
		t.Errorf("Toast=%q", st.Toast)
	}
}

func TestFormatUptimeSystem(t *testing.T) {
	if formatUptime(0) != "-" {
		t.Errorf("got %q", formatUptime(0))
	}
	if formatUptime(time.Hour+5*time.Minute) != "1h5m" {
		t.Errorf("got %q", formatUptime(time.Hour+5*time.Minute))
	}
	if formatUptime(45*time.Second) != "45s" {
		t.Errorf("got %q", formatUptime(45*time.Second))
	}
	if formatUptime(2*time.Minute) != "2m" {
		t.Errorf("got %q", formatUptime(2*time.Minute))
	}
}

type errString string

func (e errString) Error() string { return string(e) }

func contains(arr []string, s string) bool {
	for _, e := range arr {
		if e == s {
			return true
		}
	}
	return false
}
