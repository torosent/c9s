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
	"github.com/torosent/c9s/internal/ui/theme"
)

// TestViewFitsInTerminal — root-cause regression for the "post-exec
// only the bottom of the banner is visible" bug. The containers
// screen used to size its bubbles/table viewport off the FULL
// terminal height passed via WindowSizeMsg, but the screen actually
// renders into a smaller body region (terminal minus banner + status
// bar + palette line). Result: View() returned ~88 lines for an
// 80-row terminal, bubbletea's renderer truncated the top 8 to fit,
// and the user lost the banner.
//
// The fix forwards a corrected WindowSizeMsg with Height = body
// region to the active screen, so its internal widgets size against
// the right region. View() output then exactly matches m.height.
//
// Widths are kept ≥ 130 cols because the banner has fixed-width
// columns (38 + 22 + 22 + 28 + 4 spacing = 114) that wrap at
// narrower widths — that's a separate layout bug, not the one this
// regression covers.
func TestViewFitsInTerminal(t *testing.T) {
	cases := []struct {
		name   string
		width  int
		height int
	}{
		{"actual user 120x80", 120, 80},
		{"normal", 140, 40},
		{"large", 200, 80},
		{"wide compact", 200, 24},
		{"wide tall", 160, 60},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &cli.Fake{
				VersionResp: "container CLI version 0.12.1",
				ListContainersResp: []cli.Container{
					{ID: "c1abcdef0123", ShortID: "c1abcdef0123", Image: "nginx", Status: "running"},
					{ID: "c2abcdef0123", ShortID: "c2abcdef0123", Image: "redis", Status: "stopped"},
					{ID: "c3abcdef0123", ShortID: "c3abcdef0123", Image: "alpine", Status: "running"},
				},
			}
			app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
			var m tea.Model = app
			m, _ = m.Update(tea.WindowSizeMsg{Width: tc.width, Height: tc.height})
			m, _ = m.Update(SplashDoneMsg{})
			m, _ = m.Update(state.RefreshedMsg[cli.Container]{
				Resource: cli.ResourceContainers,
				Snapshot: state.Snapshot[cli.Container]{
					Items:     fake.ListContainersResp,
					FetchedAt: time.Unix(0, 0),
				},
			})

			view := m.View().Content
			gotLines := strings.Count(view, "\n") + 1
			if gotLines != tc.height {
				t.Errorf("View() returned %d lines for terminal %dx%d; want exactly %d (otherwise bubbletea's renderer truncates and the banner gets dropped)",
					gotLines, tc.width, tc.height, tc.height)
			}

			// Every container row must be visible in the View output.
			for _, c := range fake.ListContainersResp {
				prefix := c.ShortID
				if len(prefix) > 8 {
					prefix = prefix[:8]
				}
				if !strings.Contains(view, prefix) {
					t.Errorf("View() missing container row %s for terminal %dx%d", c.ShortID, tc.width, tc.height)
				}
			}

			// Banner Context label must be present (no top truncation).
			if !strings.Contains(view, "Context:") {
				t.Errorf("View() missing banner Context: label for terminal %dx%d — top rows got truncated", tc.width, tc.height)
			}
		})
	}
}

// TestViewFitsAfterScreenSized — explicitly forces the active screen
// to receive a "raw" full-terminal WindowSizeMsg (simulating a
// SIGWINCH or the post-exec resize firework), then asserts View()
// still fits in the terminal. Without the bodyRegionHeight fix the
// screen sizes its table off the full terminal height, table.View()
// overflows the body region in BorderedBox, and View() returns more
// lines than m.height — bubbletea's renderer truncates the top
// (banner) to fit, which is the user-visible bug.
func TestViewFitsAfterScreenSized(t *testing.T) {
	const W, H = 120, 80
	fake := &cli.Fake{
		VersionResp: "container CLI version 0.12.1",
		ListContainersResp: []cli.Container{
			{ID: "c1abcdef0123", ShortID: "c1abcdef0123", Image: "nginx", Status: "running"},
			{ID: "c2abcdef0123", ShortID: "c2abcdef0123", Image: "redis", Status: "stopped"},
			{ID: "c3abcdef0123", ShortID: "c3abcdef0123", Image: "alpine", Status: "running"},
		},
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
	var m tea.Model = app
	m, _ = m.Update(tea.WindowSizeMsg{Width: W, Height: H})
	m, _ = m.Update(SplashDoneMsg{})
	m, _ = m.Update(state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: state.Snapshot[cli.Container]{Items: fake.ListContainersResp, FetchedAt: time.Unix(0, 0)},
	})

	// Simulate a SIGWINCH (or the post-exec resize the old shellExecDoneMsg
	// handler used to fire) by feeding ANOTHER full-terminal-size
	// WindowSizeMsg. The bug shows up as soon as the screen is sized
	// with the full terminal height instead of the body region.
	m, _ = m.Update(tea.WindowSizeMsg{Width: W, Height: H})

	view := m.View().Content
	gotLines := strings.Count(view, "\n") + 1
	if gotLines != H {
		t.Errorf("View() returned %d lines for %dx%d terminal after second WindowSizeMsg; want %d (the screen sized its table off the full terminal height instead of the body region)",
			gotLines, W, H, H)
	}
	if !strings.Contains(view, "Context:") {
		t.Error("View() missing banner Context: label after second WindowSizeMsg — the renderer truncated the top to fit")
	}
}

// TestViewFitsAfterShellExec — same invariant must hold after the
// shellExecDoneMsg handler runs (post-shell-exit recovery). This is
// the specific path the user hit; before the bodyRegionHeight fix
// the table was sized off the full terminal height and the banner
// was always truncated by bubbletea's renderer to fit.
func TestViewFitsAfterShellExec(t *testing.T) {
	const W, H = 140, 40
	fake := &cli.Fake{
		VersionResp: "container CLI version 0.12.1",
		ListContainersResp: []cli.Container{
			{ID: "c1abcdef0123", ShortID: "c1abcdef0123", Image: "nginx", Status: "running"},
		},
	}
	app := NewApp(fake, clock.NewFake(time.Unix(0, 0)), theme.DefaultDark(), config.Default())
	var m tea.Model = app
	m, _ = m.Update(tea.WindowSizeMsg{Width: W, Height: H})
	m, _ = m.Update(SplashDoneMsg{})
	m, _ = m.Update(state.RefreshedMsg[cli.Container]{
		Resource: cli.ResourceContainers,
		Snapshot: state.Snapshot[cli.Container]{Items: fake.ListContainersResp, FetchedAt: time.Unix(0, 0)},
	})

	// Simulate shell exec returning.
	m, _ = m.Update(shellExecDoneMsg{})

	view := m.View().Content
	gotLines := strings.Count(view, "\n") + 1
	if gotLines != H {
		t.Errorf("post-exec View() returned %d lines for %dx%d terminal; want %d", gotLines, W, H, H)
	}
	if !strings.Contains(view, "Context:") {
		t.Error("post-exec View() missing banner Context: label")
	}
	if !strings.Contains(view, "c1abcdef") {
		t.Error("post-exec View() missing container row")
	}
}
