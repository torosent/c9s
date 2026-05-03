package networks

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

func sample() []cli.Network {
	return []cli.Network{
		{Name: "bridge0", Driver: "bridge", Subnet: "192.168.64.0/24", Containers: []string{"api"}},
		{Name: "isolated", Driver: "bridge", Subnet: "10.0.0.0/24"},
	}
}

func feed(t *testing.T, m Model, ns []cli.Network) Model {
	t.Helper()
	snap := state.Snapshot[cli.Network]{Items: ns, FetchedAt: time.Now()}
	s, _ := m.Update(state.RefreshedMsg[cli.Network]{
		Resource: cli.ResourceNetworks,
		Snapshot: snap,
	})
	return s.(Model)
}

func TestNew(t *testing.T) {
	m := New(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "Networks" {
		t.Errorf("Title=%q", m.Title())
	}
	if m.Init() == nil {
		t.Error("Init returned nil")
	}
}

func TestRender(t *testing.T) {
	m := New(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	view := m.View(120, 30)
	if !strings.Contains(view, "bridge0") {
		t.Errorf("expected bridge0 in view: %q", view)
	}
	if !strings.Contains(m.Summary(), "2 networks") {
		t.Errorf("Summary=%q", m.Summary())
	}
}

func TestSpaceMarkAndStar(t *testing.T) {
	m := New(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = s.(Model)
	if !strings.Contains(m.Summary(), "1 selected") {
		t.Errorf("got %q", m.Summary())
	}
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = s.(Model)
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'*'}})
	m = s.(Model)
	if !strings.Contains(m.Summary(), "2 selected") {
		t.Errorf("after *: %q", m.Summary())
	}
}

func TestRefresh(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	_ = cmd()
	if len(f.Calls) == 0 {
		t.Error("no refresh call")
	}
}

func TestInspect(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	if _, ok := cmd().(screens.OpenModalMsg); !ok {
		t.Error("expected OpenModalMsg")
	}
}

func TestDeleteFlow(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	cmd()

	_, cmd = m.Update(modals.ConfirmResultMsg{Result: modals.ConfirmResult{Confirmed: true, Tag: "delete-networks"}})
	if cmd == nil {
		t.Fatal("nil cmd from confirm")
	}
	walk(t, cmd)
	if !contains(f.Calls, "DeleteNetwork") {
		t.Errorf("expected DeleteNetwork in calls: %v", f.Calls)
	}
}

func TestFilter(t *testing.T) {
	m := New(cli.NewFake(), clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = s.(Model)
	for _, r := range "iso" {
		s, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = s.(Model)
	}
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = s.(Model)
	v := m.View(120, 30)
	if !strings.Contains(v, "isolated") {
		t.Errorf("expected isolated in filtered view: %q", v)
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

func TestSortableColumnsReturnsExpected(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	cols := m.SortableColumns()
	expected := []string{"name", "driver", "subnet", "container_count"}
	if len(cols) != len(expected) {
		t.Fatalf("expected %d columns, got %d", len(expected), len(cols))
	}
	for i, exp := range expected {
		if cols[i].Key != exp {
			t.Errorf("column %d: expected %q, got %q", i, exp, cols[i].Key)
		}
	}
}

func TestApplySortByNameReorders(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	nets := []cli.Network{
		{Name: "zebra", Driver: "bridge", Subnet: "10.0.0.0/24"},
		{Name: "alpha", Driver: "host", Subnet: "192.168.0.0/24"},
	}
	m = feed(t, m, nets)
	m.ApplySort("name", false)
	if m.networks[0].Name != "alpha" {
		t.Errorf("expected first network to be alpha, got %s", m.networks[0].Name)
	}
	if m.networks[1].Name != "zebra" {
		t.Errorf("expected second network to be zebra, got %s", m.networks[1].Name)
	}
}

func TestMouseLeftClickSelectsRow(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feed(t, m, sample())
	s, _ := m.Update(tea.MouseMsg{
		X:      5,
		Y:      4,
		Button: tea.MouseButtonLeft,
	})
	m2, ok := s.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", s)
	}
	if m2.tbl.Cursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", m2.tbl.Cursor())
	}
}
