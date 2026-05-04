package ui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/config"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestAppShowsSplashThenContainersThenQuits(t *testing.T) {
	fake := &cli.Fake{
		VersionResp: "container CLI version 0.12.1",
		ListContainersResp: []cli.Container{
			{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"},
			{ID: "c2", ShortID: "c2", Image: "redis", Status: "exited"},
		},
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
	tm := teatest.NewTestModel(t, app, teatest.WithInitialTermSize(120, 40))

	// Frame 1: splash visible
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(string(b), "c9s")
	}, teatest.WithDuration(2*time.Second))

	// Press any key to dismiss the splash
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	// Frame 2: containers screen visible with table headers
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		s := string(b)
		return strings.Contains(s, "SHORT-ID") || strings.Contains(s, "IMAGE") || strings.Contains(s, "STATE")
	}, teatest.WithDuration(2*time.Second))

	// Test :images command — should switch to the new images screen
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'i'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'m'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Should show the images table headers (REPOSITORY/TAG/SIZE)
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		s := string(b)
		return strings.Contains(s, "REPOSITORY") || strings.Contains(s, "Images")
	}, teatest.WithDuration(2*time.Second))

	// Type ":" then "q" then Enter to quit
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	tm.WaitFinished(t, teatest.WithFinalTimeout(2*time.Second))

	if !contains(fake.Calls, "Capabilities") && !contains(fake.Calls, "ListContainers") {
		t.Errorf("Fake.Calls = %v, expected Capabilities and ListContainers", fake.Calls)
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
// We exercise Update directly (rather than via teatest) so the assertion
// targets exactly the splash-gate code path, with no async Cmd
// goroutines racing the test.
func TestAppForwardsInitMessagesDuringSplash(t *testing.T) {
	fake := &cli.Fake{
		VersionResp: "container CLI version 0.12.1",
		// Intentionally empty: only the synthesized RefreshedMsg below
		// supplies the data, so the test fails cleanly if that message
		// is dropped during the splash.
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())

	// Drive Init so screens get their initial state. Discard the
	// returned Cmd; we don't run the goroutines for this test.
	mdl, _ := app, app.Init()
	_ = mdl
	var m tea.Model = app

	// Sized so the table renders rows.
	m, _ = m.Update(tea.WindowSizeMsg{Width: 140, Height: 40})

	// Splash is showing. Synthesize the RefreshedMsg the containers
	// screen Init would emit. Without the fix this is dropped on the
	// floor by the `if !m.showSplash` gate in app.Update.
	m, _ = m.Update(state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: state.Snapshot[cli.Container]{
			Items: []cli.Container{
				{ID: "abc123demo", ShortID: "abc123demo", Image: "ghcr.io/example/api:1.0", Status: "running"},
			},
			FetchedAt: time.Unix(0, 0),
		},
	})

	// Dismiss the splash.
	m, _ = m.Update(SplashDoneMsg{})

	view := m.View()
	if !strings.Contains(view, "abc123demo") || !strings.Contains(view, "ghcr.io/example/api") {
		t.Fatalf("expected container row to be visible after splash dismissal; got:\n%s", view)
	}
}

func TestAppCtrlETogglesHeader(t *testing.T) {
	fake := &cli.Fake{
		VersionResp:        "container CLI version 0.12.1",
		ListContainersResp: []cli.Container{{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"}},
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
	tm := teatest.NewTestModel(t, app, teatest.WithInitialTermSize(120, 40))

	// Dismiss splash
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(string(b), "SHORT-ID")
	}, teatest.WithDuration(2*time.Second))

	// Press Ctrl+E to toggle header
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlE})

	// Give it a moment to process (no specific visual change to wait for, just ensure no crash)
	time.Sleep(50 * time.Millisecond)

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(1*time.Second))
}

func TestAppRunCommandUnknown(t *testing.T) {
	fake := &cli.Fake{
		VersionResp:        "container CLI version 0.12.1",
		ListContainersResp: []cli.Container{{ID: "c1", ShortID: "c1", Image: "nginx", Status: "running"}},
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
	tm := teatest.NewTestModel(t, app, teatest.WithInitialTermSize(120, 40))

	// Dismiss splash
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(string(b), "SHORT-ID")
	}, teatest.WithDuration(2*time.Second))

	// Type :foo (unknown command)
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{':'}})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("foo")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})

	// Should show "unknown" toast
	teatest.WaitFor(t, tm.Output(), func(b []byte) bool {
		return strings.Contains(string(b), "unknown")
	}, teatest.WithDuration(1*time.Second))

	// Quit
	tm.Send(tea.KeyMsg{Type: tea.KeyCtrlC})
	tm.WaitFinished(t, teatest.WithFinalTimeout(1*time.Second))
}
