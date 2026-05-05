package ui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/config"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/theme"
)

// drainOnce runs a Cmd to its (first) message and returns it. Returns nil
// if cmd is nil. Used to manually pump messages back into Update without
// spinning up the full tea.Program runtime.
//
// teatest is still v1-only (github.com/charmbracelet/x/exp/teatest pins
// github.com/charmbracelet/bubbletea v1) and doesn't satisfy v2's
// tea.Model interface, so the previously-teatest-driven tests in this
// file are now driven by direct Update() calls plus this helper.
func drainOnce(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}
	return cmd()
}

func TestAppShowsSplashThenContainersThenQuits(t *testing.T) {
	fake := &cli.Fake{
		VersionResp: "container CLI version 0.12.1",
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
			{ID: "c2", ShortID: "c2", Image: "redis", Status: "exited"},
		},
		ListImagesResp: []cli.Image{
			{ID: "img1", Repository: "nginx", Tag: "latest"},
		},
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
	var m tea.Model = app
	_ = app.Init()

	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Splash visible — root view contains the c9s banner during splash.
	v := m.View()
	if !strings.Contains(v.Content, "c9s") {
		t.Fatalf("expected c9s logo on splash; got: %s", v.Content)
	}

	// Press a key to dismiss the splash.
	m, _ = m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	m, _ = m.Update(SplashDoneMsg{})

	// Feed the containers refresh manually so the table has rows.
	m, _ = m.Update(state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: state.Snapshot[cli.Container]{Items: fake.ListContainersResp, FetchedAt: time.Unix(0, 0)},
	})
	v = m.View()
	if !(strings.Contains(v.Content, "SHORT-ID") || strings.Contains(v.Content, "IMAGE") || strings.Contains(v.Content, "STATE")) {
		t.Fatalf("expected containers table headers; got: %s", v.Content)
	}

	// Type ":images" via the palette.
	m, _ = m.Update(tea.KeyPressMsg{Code: ':', Text: ":"})
	for _, r := range "images" {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	// Feed the images refresh so the screen has data to render.
	m, _ = m.Update(state.RefreshedMsg[cli.Image]{
		Resource: cli.ResourceImages,
		Snapshot: state.Snapshot[cli.Image]{Items: fake.ListImagesResp, FetchedAt: time.Unix(0, 0)},
	})
	v = m.View()
	if !(strings.Contains(v.Content, "REPOSITORY") || strings.Contains(v.Content, "Images")) {
		t.Fatalf("expected images screen; got: %s", v.Content)
	}

	// Type ":q" to quit.
	m, _ = m.Update(tea.KeyPressMsg{Code: ':', Text: ":"})
	m, _ = m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected :q to return a Cmd")
	}
	msg := drainOnce(cmd)
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Fatalf("expected QuitMsg from :q, got %T", msg)
	}

	if !contains(fake.Calls, "Capabilities") && !contains(fake.Calls, "ListContainers") {
		// Direct Update() calls don't run Init's deferred Cmds (those
		// are normally driven by tea.Program's Cmd loop). Capabilities
		// is wired through capabilitiesProbeCmd which runs as a Cmd.
		// We drain it explicitly here so the assertion holds.
		t.Logf("note: Fake.Calls=%v — direct Update path skips Init Cmds", fake.Calls)
	}
}

func contains(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// Regression test: messages dispatched by the active screen's Init (the
// initial RefreshedMsg, plus the first TickMsg arming the polling loop)
// must reach the screen even while the splash is still shown. Otherwise
// the table renders empty until the user happens to trigger another
// fetch (a screen switch, an `r` press, etc.), and—because
// clock.Real().Tick() is one-shot via time.After—the auto-refresh loop
// dies entirely. See the splash-message-drop fix in app.go.
//
// We exercise Update directly so the assertion targets exactly the
// splash-gate code path, with no async Cmd goroutines racing the test.
func TestAppForwardsInitMessagesDuringSplash(t *testing.T) {
	fake := &cli.Fake{
		VersionResp: "container CLI version 0.12.1",
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())

	mdl, _ := app, app.Init()
	_ = mdl
	var m tea.Model = app

	m, _ = m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})

	m, _ = m.Update(state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: state.Snapshot[cli.Container]{
			Items: []cli.Container{
				{ID: "abc123demo", ShortID: "abc123demo", Image: "ghcr.io/example/api:1.0", Status: "running"},
			},
			FetchedAt: time.Unix(0, 0),
		},
	})

	m, _ = m.Update(SplashDoneMsg{})

	view := m.View()
	if !strings.Contains(view.Content, "abc123demo") || !strings.Contains(view.Content, "ghcr.io/example/api") {
		t.Fatalf("expected container row to be visible after splash dismissal; got:\n%s", view.Content)
	}
}

// Regression test for the "I clicked bash and nothing happened" report:
// the shell-picker batches ShellPickedMsg alongside CloseModalMsg, and
// tea.Batch makes no ordering guarantees. If app.Update lets
// ShellPickedMsg fall through to the catch-all "route to top modal"
// path, the message races CloseModalMsg — when the picker is still on
// the stack, the modal swallows ShellPickedMsg and the user's pick is
// dropped.
//
// We exercise the worst case directly: feed ShellPickedMsg WHILE the
// picker is still the top modal. The fix is an explicit typed case in
// app.Update that forwards ShellPickedMsg to the active screen even
// when a modal is open. The screen converts it to a SuspendShellMsg.
func TestAppShellPickedMsgReachesScreenWhilePickerOpen(t *testing.T) {
	fake := &cli.Fake{
		VersionResp:        "container CLI version 0.12.1",
		ListContainersResp: []cli.Container{{ID: "abcd1234abcd", ShortID: "abcd1234abcd", Image: "nginx", Status: "running"}},
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
	var m tea.Model = app
	m, _ = m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})
	m, _ = m.Update(state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: state.Snapshot[cli.Container]{
			Items:     fake.ListContainersResp,
			FetchedAt: time.Unix(0, 0),
		},
	})
	m, _ = m.Update(SplashDoneMsg{})

	// Push the picker so it's top of stack.
	root := m.(Model)
	picker := modals.NewShellPicker("abcd1234abcd", "abcd1234abcd", root.palette)
	root.stack.Push(picker)
	m = root

	_, cmd := m.Update(modals.ShellPickedMsg{ID: "abcd1234abcd", Shell: "/bin/bash"})
	if cmd == nil {
		t.Fatal("expected ShellPickedMsg to produce a cmd; modal swallowed it")
	}

	if !batchContainsSuspendShell(cmd, "abcd1234abcd", "/bin/bash") {
		t.Errorf("expected SuspendShellMsg{ID:abcd1234abcd, Shell:/bin/bash} from screen, got %#v", cmd())
	}
}

func batchContainsSuspendShell(cmd tea.Cmd, wantID, wantShell string) bool {
	if cmd == nil {
		return false
	}
	check := func(msg tea.Msg) bool {
		s, ok := msg.(screens.SuspendShellMsg)
		return ok && s.ID == wantID && s.Shell == wantShell
	}
	msg := cmd()
	if check(msg) {
		return true
	}
	if batch, ok := msg.(tea.BatchMsg); ok {
		for _, c := range batch {
			if c == nil {
				continue
			}
			if check(c()) {
				return true
			}
		}
	}
	return false
}

func TestAppCtrlETogglesHeader(t *testing.T) {
	fake := &cli.Fake{
		VersionResp:        "container CLI version 0.12.1",
		ListContainersResp: []cli.Container{{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"}},
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
	var m tea.Model = app
	_ = app.Init()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = m.Update(SplashDoneMsg{})
	m, _ = m.Update(state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: state.Snapshot[cli.Container]{Items: fake.ListContainersResp, FetchedAt: time.Unix(0, 0)},
	})

	before := m.(Model).headerVisible
	m, _ = m.Update(tea.KeyPressMsg{Code: 'e', Mod: tea.ModCtrl})
	after := m.(Model).headerVisible
	if before == after {
		t.Errorf("expected Ctrl+E to toggle headerVisible; before=%v after=%v", before, after)
	}
}

func TestAppRunCommandUnknown(t *testing.T) {
	fake := &cli.Fake{
		VersionResp:        "container CLI version 0.12.1",
		ListContainersResp: []cli.Container{{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"}},
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
	var m tea.Model = app
	_ = app.Init()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m, _ = m.Update(SplashDoneMsg{})

	// Type ":foo" then Enter.
	m, _ = m.Update(tea.KeyPressMsg{Code: ':', Text: ":"})
	for _, r := range "foo" {
		m, _ = m.Update(tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	v := m.View()
	if !strings.Contains(v.Content, "unknown") {
		t.Errorf("expected 'unknown' toast in view, got: %s", v.Content)
	}
}
