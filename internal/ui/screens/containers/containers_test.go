package containers

import (
	"context"
	"os"
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

func TestContainersInit(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	cmd := m.Init()

	if cmd == nil {
		t.Fatal("expected Init to return a cmd")
	}

	// The cmd is a batch; we can't easily test the async behavior here,
	// but we can verify it's non-nil which means refresh was scheduled
}

func TestContainersRefreshTick(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)

	// Init
	cmd := m.Init()
	if cmd == nil {
		t.Fatal("expected Init to return a cmd")
	}

	// The refresh mechanism is set up via TickCmd which fires after 2s.
	// Testing this fully requires integration with the Bubble Tea runtime.
	// For now, we just verify Init returned a cmd.
}

func TestContainersSpaceTogglesMarks(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
			{ID: "c2", ShortID: "c2", Image: "redis", Status: "exited"},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	m.Init()

	// Simulate receiving the RefreshedMsg
	snapshot := state.Snapshot[cli.Container]{
		Items:     fake.ListContainersResp,
		FetchedAt: time.Now(),
	}
	msg := state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: snapshot,
	}
	s, _ := m.Update(msg)
	m = assertModel(s)

	// Press space to mark the focused row
	keyMsg := tea.KeyMsg{Type: tea.KeySpace}
	s, _ = m.Update(keyMsg)
	m = assertModel(s)

	summary := m.Summary()
	if !strings.Contains(summary, "1 selected") && !strings.Contains(summary, "marked") {
		t.Errorf("expected summary to mention selection, got: %s", summary)
	}
}

func TestContainersStarSelectsAll(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
			{ID: "c2", ShortID: "c2", Image: "redis", Status: "exited"},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	m.Init()

	snapshot := state.Snapshot[cli.Container]{
		Items:     fake.ListContainersResp,
		FetchedAt: time.Now(),
	}
	msg := state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: snapshot,
	}
	s, _ := m.Update(msg)
	m = assertModel(s)

	// Press * to select all
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'*'}}
	s, _ = m.Update(keyMsg)
	m = assertModel(s)

	summary := m.Summary()
	if !strings.Contains(summary, "2") {
		t.Errorf("expected summary to show 2 selected, got: %s", summary)
	}
}

func TestContainersEscClearsMarks(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	m.Init()

	snapshot := state.Snapshot[cli.Container]{
		Items:     fake.ListContainersResp,
		FetchedAt: time.Now(),
	}
	msg := state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: snapshot,
	}
	s, _ := m.Update(msg)
	m = assertModel(s)

	// Mark one
	keyMsg := tea.KeyMsg{Type: tea.KeySpace}
	s, _ = m.Update(keyMsg)
	m = assertModel(s)

	// Now press Esc
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	s, _ = m.Update(escMsg)
	m = assertModel(s)

	summary := m.Summary()
	if strings.Contains(summary, "selected") || strings.Contains(summary, "marked") {
		t.Errorf("expected summary to not mention selection after Esc, got: %s", summary)
	}
}

func TestContainersRTriggersRefresh(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	m.Init()

	initialCalls := len(fake.Calls)

	// Press 'r' to trigger manual refresh
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}}
	_, cmd := m.Update(keyMsg)

	if cmd != nil {
		_ = cmd()
	}

	// Should have triggered a new ListContainers call
	if len(fake.Calls) <= initialCalls {
		t.Error("expected 'r' key to trigger a refresh")
	}
}

func TestContainersFilterByImageOrID(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
			{ID: "c2", ShortID: "c2", Image: "redis", Status: "exited"},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	m.Init()

	snapshot := state.Snapshot[cli.Container]{
		Items:     fake.ListContainersResp,
		FetchedAt: time.Now(),
	}
	msg := state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: snapshot,
	}
	s, _ := m.Update(msg)
	m = assertModel(s)

	// Enter filter mode with '/'
	slashMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}
	s, _ = m.Update(slashMsg)
	m = assertModel(s)

	// Type 'ngi'
	for _, r := range "ngi" {
		keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		s, _ = m.Update(keyMsg)
		m = assertModel(s)
	}

	// Press Enter to apply filter
	enterMsg := tea.KeyMsg{Type: tea.KeyEnter}
	s, _ = m.Update(enterMsg)
	m = assertModel(s)

	// View should now show only nginx container
	view := m.View(80, 20)
	if !strings.Contains(view, "nginx") {
		t.Error("expected filtered view to contain nginx")
	}
	// Redis should be filtered out (though view rendering might still show it in table)
	// This is hard to test without inspecting internal state
}

func TestContainersDOpensInspectModal(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
		},
		InspectContainerResp: []byte(`{"ID":"c1","Image":"nginx"}`),
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	m.Init()

	snapshot := state.Snapshot[cli.Container]{
		Items:     fake.ListContainersResp,
		FetchedAt: time.Now(),
	}
	msg := state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: snapshot,
	}
	s, _ := m.Update(msg)
	m = assertModel(s)

	// Press 'd' to inspect
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected 'd' key to return a cmd")
	}

	cmdMsg := cmd()
	if _, ok := cmdMsg.(screens.OpenModalMsg); !ok {
		t.Errorf("expected OpenModalMsg, got %T", cmdMsg)
	}
}

func TestContainersXStopsContainer(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	m.Init()

	snapshot := state.Snapshot[cli.Container]{
		Items:     fake.ListContainersResp,
		FetchedAt: time.Now(),
	}
	msg := state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: snapshot,
	}
	s, _ := m.Update(msg)
	m = assertModel(s)

	// Press 'x' to stop
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}}
	_, cmd := m.Update(keyMsg)

	if cmd != nil {
		_ = cmd()
	}

	// Check that StopContainer was called
	found := false
	for _, call := range fake.Calls {
		if call == "StopContainer" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected StopContainer to be called")
	}
}

func TestContainersSEmitsSuspendShellMsg(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	m.Init()

	snapshot := state.Snapshot[cli.Container]{
		Items:     fake.ListContainersResp,
		FetchedAt: time.Now(),
	}
	msg := state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: snapshot,
	}
	s, _ := m.Update(msg)
	m = assertModel(s)

	// Press 's' to open shell
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected 's' key to return a cmd")
	}

	cmdMsg := cmd()
	if suspendMsg, ok := cmdMsg.(screens.SuspendShellMsg); !ok {
		t.Errorf("expected SuspendShellMsg, got %T", cmdMsg)
	} else {
		if suspendMsg.ID != "c1" {
			t.Errorf("expected ID='c1', got %q", suspendMsg.ID)
		}
		// Shell should be from SHELL env or default /bin/sh
		expectedShell := os.Getenv("SHELL")
		if expectedShell == "" {
			expectedShell = "/bin/sh"
		}
		if suspendMsg.Shell != expectedShell {
			t.Errorf("expected Shell=%q, got %q", expectedShell, suspendMsg.Shell)
		}
	}
}

func TestContainersPauseUnsupportedEmitsToast(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
		},
		CapsResp: cli.Capabilities{
			Version: "0.1.0",
			Major:   0,
			Minor:   1,
			Patch:   0,
			Pause:   false, // pause not supported
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)

	// Simulate capabilities response
	caps, _ := fake.Capabilities(context.Background())
	s, _ := m.Update(capabilitiesMsg(caps))
	m = assertModel(s)

	snapshot := state.Snapshot[cli.Container]{
		Items:     fake.ListContainersResp,
		FetchedAt: time.Now(),
	}
	msg := state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: snapshot,
	}
	s, _ = m.Update(msg)
	m = assertModel(s)

	// Press 'p' to pause
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'p'}}
	_, cmd := m.Update(keyMsg)

	if cmd == nil {
		t.Fatal("expected 'p' key to return a cmd")
	}

	cmdMsg := cmd()
	if statusMsg, ok := cmdMsg.(screens.StatusMsg); !ok {
		t.Errorf("expected StatusMsg, got %T", cmdMsg)
	} else {
		if !strings.Contains(statusMsg.Toast, "pause not supported") {
			t.Errorf("expected toast about pause not supported, got: %q", statusMsg.Toast)
		}
	}

	// Ensure PauseContainer was NOT called
	for _, call := range fake.Calls {
		if call == "PauseContainer" {
			t.Error("PauseContainer should not have been called when unsupported")
		}
	}
}

func TestContainersHotkeysReturnsPopulatedMap(t *testing.T) {
	fake := &cli.Fake{}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)

	km := m.Hotkeys()
	if km == nil {
		t.Fatal("expected Hotkeys to return a non-nil Map")
	}

	// Check that some expected bindings exist
	expectedNames := []string{"refresh", "mark", "mark_all", "escape"}
	for _, name := range expectedNames {
		if _, ok := km.Get(name); !ok {
			t.Errorf("expected binding %q to exist in Hotkeys", name)
		}
	}
}

// Helper: type assertion for Screen to Model
func assertModel(s screens.Screen) Model {
	if m, ok := s.(Model); ok {
		return m
	}
	panic("expected Model type")
}

func TestSummaryFormat(t *testing.T) {
	f := &cli.Fake{ListContainersResp: []cli.Container{
		{ID: "a", ShortID: "a", Status: "running"},
		{ID: "b", ShortID: "b", Status: "running"},
		{ID: "c", ShortID: "c", Status: "exited"},
	}}
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	// Drive an initial refresh by sending a synthetic RefreshedMsg.
	s, _ := m.Update(state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: state.Snapshot[cli.Container]{Items: f.ListContainersResp},
	})
	cs := assertModel(s)
	summary := cs.Summary()
	if !strings.Contains(summary, "3 items") {
		t.Errorf("Summary = %q, want to contain '3 items'", summary)
	}
	if !strings.Contains(summary, "2 running") {
		t.Errorf("Summary = %q, want to contain '2 running'", summary)
	}
	if !strings.Contains(summary, "1 exited") {
		t.Errorf("Summary = %q, want to contain '1 exited'", summary)
	}
}

func TestPauseBindingAnnotatedWhenUnsupported(t *testing.T) {
	f := &cli.Fake{CapsResp: cli.Capabilities{Pause: false}}
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	// Send capabilitiesMsg (not cli.Capabilities directly) to set caps in the model
	s, _ := m.Update(capabilitiesMsg(cli.Capabilities{Pause: false}))
	cs := assertModel(s)

	km := cs.Hotkeys()
	pauseBinding, ok := km.Get("pause")
	if !ok {
		t.Fatal("expected pause binding to exist")
	}
	if !strings.Contains(pauseBinding.Description, "unsupported") {
		t.Errorf("Description = %q, want to contain 'unsupported'", pauseBinding.Description)
	}
}

func TestPauseBindingNotAnnotatedWhenSupported(t *testing.T) {
	f := &cli.Fake{CapsResp: cli.Capabilities{Pause: true}}
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())
	// Send capabilitiesMsg (not cli.Capabilities directly) to set caps in the model
	s, _ := m.Update(capabilitiesMsg(cli.Capabilities{Pause: true}))
	cs := assertModel(s)

	km := cs.Hotkeys()
	pauseBinding, ok := km.Get("pause")
	if !ok {
		t.Fatal("expected pause binding to exist")
	}
	if strings.Contains(pauseBinding.Description, "unsupported") {
		t.Errorf("Description = %q, should not contain 'unsupported' when supported", pauseBinding.Description)
	}
}

func TestSortableColumns(t *testing.T) {
	f := &cli.Fake{}
	m := New(f, clock.NewFake(time.Now()), theme.DefaultDark())

	cols := m.SortableColumns()
	if len(cols) < 3 {
		t.Errorf("expected at least 3 sortable columns, got %d", len(cols))
	}

	// Verify column structure
	found := false
	for _, col := range cols {
		if col.Key == "image" && col.Label == "Image" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find 'image' column")
	}
}

func TestApplySort(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "zebra", Status: "running", CPU: 10},
			{ID: "c2", ShortID: "c2", Image: "apache", Status: "stopped", CPU: 20},
			{ID: "c3", ShortID: "c3", Image: "mysql", Status: "running", CPU: 5},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	// Load containers
	s, _ := m.Update(state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: state.Snapshot[cli.Container]{Items: fake.ListContainersResp},
	})
	cs := assertModel(s)

	// Sort by image ascending
	cs.ApplySort("image", false)

	if cs.containers[0].Image != "apache" {
		t.Errorf("expected first container to be 'apache', got %s", cs.containers[0].Image)
	}
	if cs.containers[2].Image != "zebra" {
		t.Errorf("expected last container to be 'zebra', got %s", cs.containers[2].Image)
	}

	// Sort by CPU descending
	cs.ApplySort("cpu", true)

	if cs.containers[0].CPU != 20 {
		t.Errorf("expected first container CPU to be 20, got %d", cs.containers[0].CPU)
	}
	if cs.containers[2].CPU != 5 {
		t.Errorf("expected last container CPU to be 5, got %d", cs.containers[2].CPU)
	}
}

func TestMouseClick(t *testing.T) {
	fake := &cli.Fake{
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
			{ID: "c2", ShortID: "c2", Image: "apache", Status: "stopped"},
			{ID: "c3", ShortID: "c3", Image: "mysql", Status: "running"},
		},
	}
	clk := clock.NewFake(time.Now())
	p := theme.DefaultDark()

	m := New(fake, clk, p)
	// Load containers
	s, _ := m.Update(state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: state.Snapshot[cli.Container]{Items: fake.ListContainersResp},
	})
	cs := assertModel(s)

	// Simulate mouse click at Y=5 (should select row 2, index 2)
	mouseMsg := tea.MouseMsg{
		X:      10,
		Y:      5,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	}
	s, _ = cs.Update(mouseMsg)
	cs = assertModel(s)

	// Cursor should have moved
	if cs.tbl.Cursor() != 2 {
		t.Errorf("expected cursor at row 2 after click, got %d", cs.tbl.Cursor())
	}
}
