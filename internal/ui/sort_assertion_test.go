package ui

import (
	"testing"
	"time"

	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/config"
	"github.com/torosent/c9s/internal/ui/theme"
)

// TestSortAssertion_AllScreens is the regression test for C1 of the
// v0.1.0 review. The sort modal dispatch in app.go uses
//
//	scr, ok := m.screens[m.active].(interface{ ApplySort(string, bool) })
//
// to decide whether to forward the chosen column to the screen. Before
// the fix, every screen's New() returned a value type but ApplySort was
// declared on the pointer receiver, so the type assertion always
// failed silently and Shift+S did nothing. This test asserts that every
// sortable screen registered in app.go satisfies that anonymous
// interface when stored as screens.Screen.
func TestSortAssertion_AllScreens(t *testing.T) {
	fake := cli.NewFake()
	clk := clock.NewFake(time.Unix(0, 0))
	app := NewApp(fake, clk, theme.DefaultDark(), config.Default())

	type sortable interface {
		ApplySort(key string, reverse bool)
	}

	// Screens that document SortableColumns / ApplySort. Pulled from
	// the screens.Sortable users in the codebase. Static drift is
	// intentional: if a new sortable screen is added, this test should
	// be updated to include it (or vice versa).
	want := []string{
		"containers",
		"images",
		"volumes",
		"networks",
		"registry",
		"errors",
		"jobs",
		"pinned",
	}

	for _, id := range want {
		scr, ok := app.screens[id]
		if !ok {
			t.Errorf("screen %q not registered", id)
			continue
		}
		if _, ok := scr.(sortable); !ok {
			t.Errorf("screen %q does not satisfy ApplySort interface — Shift+S sort will silently no-op (C1 regression)", id)
		}
	}
}
