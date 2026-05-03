package images

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

func sampleImages() []cli.Image {
	return []cli.Image{
		{
			ID: "sha256:abc123def456", ShortID: "abc123def456",
			Repository: "ghcr.io/acme/api", Tag: "1.4.2",
			Reference: "ghcr.io/acme/api:1.4.2",
			Created:   time.Date(2025, 4, 1, 10, 30, 0, 0, time.UTC),
			SizeBytes: 524288000,
		},
		{
			ID: "sha256:fedcba987654", ShortID: "fedcba987654",
			Repository: "docker.io/library/nginx", Tag: "latest",
			Reference: "docker.io/library/nginx:latest",
			Created:   time.Date(2025, 3, 15, 8, 20, 0, 0, time.UTC),
			SizeBytes: 157286400,
		},
	}
}

func assertModel(t *testing.T, s screens.Screen) *Model {
	t.Helper()
	m, ok := s.(*Model)
	if !ok {
		t.Fatalf("expected Model, got %T", s)
	}
	return m
}

func feedSnapshot(t *testing.T, m *Model, imgs []cli.Image) *Model {
	t.Helper()
	snap := state.Snapshot[cli.Image]{Items: imgs, FetchedAt: time.Now()}
	s, _ := m.Update(state.RefreshedMsg[cli.Image]{
		Resource: cli.ResourceImages,
		Snapshot: snap,
	})
	return assertModel(t, s)
}

func TestNewReturnsModel(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	if m.Title() != "Images" {
		t.Errorf("Title() = %q, want Images", m.Title())
	}
	km := m.Hotkeys()
	if km == nil {
		t.Fatal("Hotkeys returned nil")
	}
	for _, name := range []string{"inspect", "tag", "push", "delete", "run", "refresh"} {
		if _, ok := km.Get(name); !ok {
			t.Errorf("expected binding %q in hotkeys", name)
		}
	}
}

func TestInitReturnsCmd(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	if cmd := m.Init(); cmd == nil {
		t.Fatal("Init returned nil cmd")
	}
}

func TestRefreshedMsgPopulatesTable(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	view := m.View(120, 30)
	if !strings.Contains(view, "ghcr.io/acme/api") {
		t.Errorf("View should contain repository, got: %q", view)
	}
	if !strings.Contains(view, "1.4.2") {
		t.Errorf("View should contain tag, got: %q", view)
	}
}

func TestSummaryShowsCount(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	if got := m.Summary(); !strings.Contains(got, "2 items") {
		t.Errorf("Summary = %q, want to contain '2 items'", got)
	}
}

func TestSpaceTogglesMark(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	s, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = assertModel(t, s)
	if !strings.Contains(m.Summary(), "1 selected") {
		t.Errorf("expected summary to mention selection, got %q", m.Summary())
	}
}

func TestStarSelectsAll(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'*'}})
	m = assertModel(t, s)
	if !strings.Contains(m.Summary(), "2 selected") {
		t.Errorf("expected '2 selected', got %q", m.Summary())
	}
}

func TestEscClearsMarks(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	s, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = assertModel(t, s)
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = assertModel(t, s)
	if strings.Contains(m.Summary(), "selected") {
		t.Errorf("Esc should have cleared marks, got %q", m.Summary())
	}
}

func TestRTriggersRefresh(t *testing.T) {
	f := cli.NewFake()
	f.ListImagesResp = sampleImages()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected cmd from r key")
	}
	_ = cmd()
	found := false
	for _, c := range f.Calls {
		if c == "ListImages" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected ListImages call, got %v", f.Calls)
	}
}

func TestDOpensInspectModal(t *testing.T) {
	f := cli.NewFake()
	f.InspectImageResp = []byte(`{"id":"sha256:abc"}`)
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if cmd == nil {
		t.Fatal("expected cmd from d key")
	}
	msg := cmd()
	if _, ok := msg.(screens.OpenModalMsg); !ok {
		t.Errorf("expected OpenModalMsg, got %T", msg)
	}
}

func TestUppercaseDOpensConfirm(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if cmd == nil {
		t.Fatal("expected cmd from D key")
	}
	msg := cmd()
	if _, ok := msg.(screens.OpenModalMsg); !ok {
		t.Errorf("expected OpenModalMsg, got %T", msg)
	}
}

func TestTKeyOpensTagModal(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	if cmd == nil {
		t.Fatal("expected cmd from t key")
	}
	msg := cmd()
	if _, ok := msg.(screens.OpenModalMsg); !ok {
		t.Fatalf("expected OpenModalMsg, got %T", msg)
	}
}

func TestTagResultCallsClient(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	_, cmd := m.Update(modals.TextInputResultMsg{Result: modals.TextInputResult{
		Label: "tag-image:ghcr.io/acme/api:1.4.2",
		Value: "ghcr.io/acme/api:latest",
	}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	cmd()
	if !containsCall(f.Calls, "TagImage") {
		t.Errorf("expected TagImage in calls: %v", f.Calls)
	}
}

func TestTagResultEmptyToasts(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	_, cmd := m.Update(modals.TextInputResultMsg{Result: modals.TextInputResult{
		Label: "tag-image:src", Value: "  ",
	}})
	if cmd == nil {
		t.Fatal("nil cmd")
	}
	msg := cmd()
	st, ok := msg.(screens.StatusMsg)
	if !ok {
		t.Fatalf("expected StatusMsg, got %T", msg)
	}
	if !strings.Contains(st.Toast, "empty") {
		t.Errorf("Toast=%q", st.Toast)
	}
}

func TestRKeyOpensRunForm(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'R'}})
	if cmd == nil {
		t.Fatal("expected cmd from R key")
	}
	msg := cmd()
	if _, ok := msg.(screens.OpenModalMsg); !ok {
		t.Fatalf("expected OpenModalMsg from R, got %T", msg)
	}
}

func TestPushKeyEmitsPushRequest(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'P'}})
	if cmd == nil {
		t.Fatal("expected cmd from P key")
	}
	msg := cmd()
	if _, ok := msg.(PushRequestMsg); !ok {
		t.Fatalf("expected PushRequestMsg from P, got %T", msg)
	}
}

func TestSlashEntersFilterMode(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = assertModel(t, s)
	for _, r := range "ngin" {
		s, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = assertModel(t, s)
	}
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = assertModel(t, s)
	view := m.View(120, 30)
	if !strings.Contains(view, "nginx") {
		t.Errorf("expected filtered view to contain nginx, got %q", view)
	}
}

func TestWindowSizeIsApplied(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	s, _ := m.Update(tea.WindowSizeMsg{Width: 200, Height: 60})
	m = assertModel(t, s)
	_ = m.View(200, 60)
}

func TestConfirmDeleteFiresDelete(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	// Mark first image
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = assertModel(t, s)

	// Send the confirmation result message directly
	confirm := modals.ConfirmResultMsg{Result: modals.ConfirmResult{Confirmed: true, Tag: "delete-images"}}
	_, cmd := m.Update(confirm)
	if cmd == nil {
		t.Fatal("expected a cmd from ConfirmResultMsg")
	}
	// Execute the cmd batch
	_ = cmd()
	// Use a fresh path: invoke deleteSelected then ConfirmResult
	_, cmd2 := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'D'}})
	if cmd2 != nil {
		_ = cmd2()
	}
	_, cmd3 := m.Update(modals.ConfirmResultMsg{Result: modals.ConfirmResult{Confirmed: true, Tag: "delete-images"}})
	if cmd3 != nil {
		runCmd(t, cmd3)
	}
	if !containsCall(f.Calls, "DeleteImage") {
		t.Errorf("expected DeleteImage in calls: %v", f.Calls)
	}
}

func TestConfirmDeleteRejectedDoesNothing(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	_, cmd := m.Update(modals.ConfirmResultMsg{Result: modals.ConfirmResult{Confirmed: false, Tag: "delete-images"}})
	if cmd != nil {
		runCmd(t, cmd)
	}
	if containsCall(f.Calls, "DeleteImage") {
		t.Errorf("DeleteImage should not be called when confirmation is rejected: %v", f.Calls)
	}
}

func TestFilterEscRestoresAll(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	// enter filter, type, then esc
	s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = assertModel(t, s)
	for _, r := range "ngin" {
		s, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = assertModel(t, s)
	}
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = assertModel(t, s)

	view := m.View(120, 30)
	if !strings.Contains(view, "ghcr.io/acme/api") {
		t.Errorf("expected restored view to contain api repo, got: %q", view)
	}
}

func TestFilterBackspaceShrinks(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())

	s, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	m = assertModel(t, s)
	for _, r := range "ng" {
		s, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = assertModel(t, s)
	}
	s, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = assertModel(t, s)
	view := m.View(120, 30)
	// Only one char left "n" — both nginx and api don't contain just "n" actually api doesn't contain n
	if !strings.Contains(view, "nginx") {
		t.Errorf("expected nginx to remain after partial filter: %q", view)
	}
}

func containsCall(calls []string, name string) bool {
	for _, c := range calls {
		if c == name {
			return true
		}
	}
	return false
}

// runCmd unwraps tea.Batch results recursively until base messages are produced.
func runCmd(t *testing.T, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		return
	}
	msg := cmd()
	if msg == nil {
		return
	}
	switch v := msg.(type) {
	case tea.BatchMsg:
		for _, c := range v {
			runCmd(t, c)
		}
	case tea.Cmd:
		runCmd(t, v)
	}
}

func TestSortableColumnsReturnsExpected(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	cols := m.SortableColumns()
	expected := []string{"id", "repository", "tag", "created", "size"}
	if len(cols) != len(expected) {
		t.Fatalf("expected %d columns, got %d", len(expected), len(cols))
	}
	for i, exp := range expected {
		if cols[i].Key != exp {
			t.Errorf("column %d: expected %q, got %q", i, exp, cols[i].Key)
		}
	}
}

func TestApplySortByRepositoryReorders(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	imgs := []cli.Image{
		{ID: "1", Repository: "zebra", Tag: "latest", Created: time.Now(), SizeBytes: 100},
		{ID: "2", Repository: "alpha", Tag: "v1", Created: time.Now(), SizeBytes: 200},
	}
	m = feedSnapshot(t, m, imgs)
	m.ApplySort("repository", false)
	if m.images[0].Repository != "alpha" {
		t.Errorf("expected first image to be alpha, got %s", m.images[0].Repository)
	}
	if m.images[1].Repository != "zebra" {
		t.Errorf("expected second image to be zebra, got %s", m.images[1].Repository)
	}
}

func TestMouseLeftClickSelectsRow(t *testing.T) {
	f := cli.NewFake()
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	m = feedSnapshot(t, m, sampleImages())
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
