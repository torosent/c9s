package registry

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/theme"
)

func sample() []cli.RegistryEntry {
	return []cli.RegistryEntry{
		{Host: "ghcr.io", User: "torosent", Default: true},
		{Host: "docker.io", User: "alice"},
	}
}

func feed(t *testing.T, m *Model, es []cli.RegistryEntry) *Model {
	t.Helper()
	snap := state.Snapshot[cli.RegistryEntry]{Items: es, FetchedAt: time.Now()}
	s, _ := m.Update(state.RefreshedMsg[cli.RegistryEntry]{
		Resource: cli.ResourceRegistry,
		Snapshot: snap,
	})
	return s.(*Model)
}

func TestNew(t *testing.T) {
	m := New(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "Registry" {
		t.Errorf("Title=%q", m.Title())
	}
}

func TestRender(t *testing.T) {
	m := New(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	v := m.View(120, 30)
	if !strings.Contains(v, "ghcr.io") {
		t.Errorf("expected ghcr.io in view: %q", v)
	}
	sum := m.Summary()
	if !strings.Contains(sum, "default ghcr.io") {
		t.Errorf("Summary=%q", sum)
	}
}

func TestSummaryNoDefault(t *testing.T) {
	m := New(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, []cli.RegistryEntry{{Host: "x", User: "u"}})
	if !strings.Contains(m.Summary(), "1 registries") {
		t.Errorf("got %q", m.Summary())
	}
}

func TestLoginKeyOpensModal(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'L'}})
	if cmd == nil {
		t.Fatal("L returned nil")
	}
	if _, ok := cmd().(screens.OpenModalMsg); !ok {
		t.Error("expected OpenModalMsg")
	}
}

func TestLoginResultCallsClient(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(modals.LoginResultMsg{Result: modals.LoginRequest{Host: "ghcr.io", Username: "u", Password: "p"}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	walk(t, cmd)
	if f.RegistryLoginLastHost != "ghcr.io" || f.RegistryLoginLastUser != "u" || f.RegistryLoginLastPass != "p" {
		t.Errorf("RegistryLogin args wrong: host=%q user=%q pass=%q",
			f.RegistryLoginLastHost, f.RegistryLoginLastUser, f.RegistryLoginLastPass)
	}
}

func TestLogoutFlow(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if cmd == nil {
		t.Fatal("D nil")
	}
	if _, ok := cmd().(screens.OpenModalMsg); !ok {
		t.Error("expected confirm modal")
	}
	_, cmd = m.Update(modals.ConfirmResultMsg{Result: modals.ConfirmResult{Confirmed: true, Tag: "registry-logout"}})
	if cmd == nil {
		t.Fatal("nil cmd from confirm")
	}
	walk(t, cmd)
	if !contains(f.Calls, "RegistryLogout") {
		t.Errorf("expected RegistryLogout: %v", f.Calls)
	}
}

func TestSetDefault(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'*'}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	walk(t, cmd)
	if !contains(f.Calls, "RegistrySetDefault") {
		t.Errorf("expected RegistrySetDefault: %v", f.Calls)
	}
}

func TestRefresh(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("r nil")
	}
	cmd()
	if !contains(f.Calls, "ListRegistries") {
		t.Errorf("expected ListRegistries: %v", f.Calls)
	}
}

func TestFilter(t *testing.T) {
	m := New(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = s.(*Model)
	for _, r := range "docker" {
		s, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = s.(*Model)
	}
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = s.(*Model)
	v := m.View(120, 30)
	if !strings.Contains(v, "docker.io") {
		t.Errorf("expected docker.io in filtered view: %q", v)
	}
}

func contains(arr []string, s string) bool {
	for _, e := range arr {
		if e == s {
			return true
		}
	}
	return false
}

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
