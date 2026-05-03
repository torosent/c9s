package system

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

func sampleProps() []cli.SystemProperty {
	return []cli.SystemProperty{
		{Key: "registry.default", Value: "ghcr.io"},
		{Key: "build.cache", Value: "/var/cache"},
		{Key: "version", Value: "0.4.0", ReadOnly: true},
	}
}

func feedProps(t *testing.T, m PropertyModel, ps []cli.SystemProperty) PropertyModel {
	t.Helper()
	snap := state.Snapshot[cli.SystemProperty]{Items: ps, FetchedAt: time.Now()}
	s, _ := m.Update(state.RefreshedMsg[cli.SystemProperty]{
		Resource: propertyResource,
		Snapshot: snap,
	})
	return s.(PropertyModel)
}

func TestNewProperty(t *testing.T) {
	m := NewProperty(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "System properties" {
		t.Errorf("Title=%q", m.Title())
	}
	if m.Init() == nil {
		t.Error("Init nil")
	}
}

func TestPropertyRender(t *testing.T) {
	m := NewProperty(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedProps(t, m, sampleProps())
	v := m.View(120, 30)
	if !strings.Contains(v, "registry.default") {
		t.Errorf("expected key in view: %q", v)
	}
	if !strings.Contains(m.Summary(), "3 properties") {
		t.Errorf("Summary=%q", m.Summary())
	}
}

func TestPropertyEditFlow(t *testing.T) {
	f := cli.NewFake()
	m := NewProperty(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedProps(t, m, sampleProps())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatal("e nil")
	}
	if _, ok := cmd().(screens.OpenModalMsg); !ok {
		t.Error("expected OpenModalMsg from e")
	}

	_, cmd = m.Update(modals.TextInputResultMsg{Result: modals.TextInputResult{
		Label: "set-property:registry.default", Value: "docker.io",
	}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	cmd()
	if !contains(f.Calls, "SetSystemProperty") {
		t.Errorf("expected SetSystemProperty: %v", f.Calls)
	}
}

func TestPropertyEditReadOnlyToast(t *testing.T) {
	f := cli.NewFake()
	m := NewProperty(f, clock.NewFake(time.Now()), theme.DefaultDark())
	// Make only one row, RO
	m = feedProps(t, m, []cli.SystemProperty{{Key: "version", Value: "0.4.0", ReadOnly: true}})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if cmd == nil {
		t.Fatal("e nil")
	}
	msg := cmd()
	st, ok := msg.(screens.StatusMsg)
	if !ok {
		t.Fatalf("got %T", msg)
	}
	if !strings.Contains(st.Toast, "read-only") {
		t.Errorf("Toast=%q", st.Toast)
	}
}

func TestPropertyResetFlow(t *testing.T) {
	f := cli.NewFake()
	m := NewProperty(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedProps(t, m, sampleProps())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if cmd == nil {
		t.Fatal("D nil")
	}
	if _, ok := cmd().(screens.OpenModalMsg); !ok {
		t.Error("expected confirm modal")
	}

	_, cmd = m.Update(modals.ConfirmResultMsg{Result: modals.ConfirmResult{Confirmed: true, Tag: "reset-property"}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	walk(t, cmd)
	if !contains(f.Calls, "ResetSystemProperty") {
		t.Errorf("expected ResetSystemProperty: %v", f.Calls)
	}
}

func TestPropertyResetReadOnlyBlocked(t *testing.T) {
	f := cli.NewFake()
	m := NewProperty(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedProps(t, m, []cli.SystemProperty{{Key: "version", Value: "x", ReadOnly: true}})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	msg := cmd()
	st, ok := msg.(screens.StatusMsg)
	if !ok {
		t.Fatalf("got %T", msg)
	}
	if !strings.Contains(st.Toast, "read-only") {
		t.Errorf("Toast=%q", st.Toast)
	}
}

func TestPropertyRefreshAndFilter(t *testing.T) {
	f := cli.NewFake()
	m := NewProperty(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("r nil")
	}
	cmd()
	if !contains(f.Calls, "ListSystemProperties") {
		t.Errorf("expected ListSystemProperties: %v", f.Calls)
	}
	m = feedProps(t, m, sampleProps())
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = s.(PropertyModel)
	for _, r := range "build" {
		s, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = s.(PropertyModel)
	}
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = s.(PropertyModel)
	v := m.View(120, 30)
	if !strings.Contains(v, "build.cache") {
		t.Errorf("expected build.cache in view: %q", v)
	}
}
