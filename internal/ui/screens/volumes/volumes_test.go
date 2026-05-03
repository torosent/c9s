package volumes

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

func sampleVolumes() []cli.Volume {
	return []cli.Volume{
		{Name: "api-data", Driver: "local", Mountpoint: "/mnt/api", SizeBytes: 104857600, UsedBy: []string{"api"}},
		{Name: "cache", Driver: "local", Mountpoint: "/mnt/cache", SizeBytes: 0},
	}
}

func assertModel(t *testing.T, s screens.Screen) Model {
	t.Helper()
	m, ok := s.(Model)
	if !ok {
		t.Fatalf("expected Model, got %T", s)
	}
	return m
}

func feedSnap(t *testing.T, m Model, vs []cli.Volume) Model {
	t.Helper()
	snap := state.Snapshot[cli.Volume]{Items: vs, FetchedAt: time.Now()}
	s, _ := m.Update(state.RefreshedMsg[cli.Volume]{
		Resource: cli.ResourceVolumes,
		Snapshot: snap,
	})
	return assertModel(t, s)
}

func TestNew(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "Volumes" {
		t.Errorf("Title = %q", m.Title())
	}
	if m.Init() == nil {
		t.Error("Init returned nil")
	}
}

func TestRender(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnap(t, m, sampleVolumes())
	view := m.View(120, 30)
	if !strings.Contains(view, "api-data") {
		t.Errorf("expected api-data in view, got %q", view)
	}
	if !strings.Contains(m.Summary(), "2 volumes") {
		t.Errorf("Summary = %q", m.Summary())
	}
}

func TestSpaceMark(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnap(t, m, sampleVolumes())
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = assertModel(t, s)
	if !strings.Contains(m.Summary(), "1 selected") {
		t.Errorf("expected 1 selected, got %q", m.Summary())
	}
}

func TestStarMarksAll(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnap(t, m, sampleVolumes())
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'*'}})
	m = assertModel(t, s)
	if !strings.Contains(m.Summary(), "2 selected") {
		t.Errorf("got %q", m.Summary())
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
		t.Errorf("expected ListVolumes call, got %v", f.Calls)
	}
}

func TestInspectOpensModal(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnap(t, m, sampleVolumes())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	if _, ok := cmd().(screens.OpenModalMsg); !ok {
		t.Errorf("expected OpenModalMsg")
	}
}

func TestDeleteFlow(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnap(t, m, sampleVolumes())
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	if _, ok := cmd().(screens.OpenModalMsg); !ok {
		t.Errorf("expected confirm modal")
	}
	_, cmd = m.Update(modals.ConfirmResultMsg{Result: modals.ConfirmResult{Confirmed: true, Tag: "delete-volumes"}})
	if cmd != nil {
		runCmd(t, cmd)
	}
	if !contains(f.Calls, "DeleteVolume") {
		t.Errorf("expected DeleteVolume in calls: %v", f.Calls)
	}
}

func TestFilter(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnap(t, m, sampleVolumes())
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = assertModel(t, s)
	for _, r := range "cache" {
		s, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = assertModel(t, s)
	}
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = assertModel(t, s)
	v := m.View(120, 30)
	if !strings.Contains(v, "cache") {
		t.Errorf("expected cache row visible: %q", v)
	}
}

func TestEscClearsMarks(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnap(t, m, sampleVolumes())
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = assertModel(t, s)
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = assertModel(t, s)
	if strings.Contains(m.Summary(), "selected") {
		t.Errorf("Esc should clear marks, got %q", m.Summary())
	}
}

func TestFormatBytesCol(t *testing.T) {
	if formatBytes(0) != "-" {
		t.Errorf("got %q", formatBytes(0))
	}
	if formatBytes(2048) != "2.0K" {
		t.Errorf("got %q", formatBytes(2048))
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

func runCmd(t *testing.T, cmd tea.Cmd) {
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
			runCmd(t, c)
		}
	}
}

func TestSortableColumnsReturnsExpected(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	cols := m.SortableColumns()
	expected := []string{"name", "driver", "size", "used_by_count"}
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
	vols := []cli.Volume{
		{Name: "zebra", Driver: "local", SizeBytes: 100},
		{Name: "alpha", Driver: "local", SizeBytes: 200},
	}
	m = feedSnap(t, m, vols)
	m.ApplySort("name", false)
	if m.volumes[0].Name != "alpha" {
		t.Errorf("expected first volume to be alpha, got %s", m.volumes[0].Name)
	}
	if m.volumes[1].Name != "zebra" {
		t.Errorf("expected second volume to be zebra, got %s", m.volumes[1].Name)
	}
}

func TestMouseLeftClickSelectsRow(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnap(t, m, sampleVolumes())
	s, _ := m.Update(tea.MouseMsg{
		X:      5,
		Y:      4,
		Button: tea.MouseButtonLeft,
	})
	m = assertModel(t, s)
	if m.tbl.Cursor() != 1 {
		t.Errorf("expected cursor at 1, got %d", m.tbl.Cursor())
	}
}
