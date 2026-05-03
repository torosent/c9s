package jobs

import (
	"testing"
	"time"

	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/jobs"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestJobsModel_Implements(t *testing.T) {
	mgr := jobs.New(clock.NewFake(time.Now()))
	clk := clock.NewFake(time.Now())

	m := New(mgr, clk, theme.DefaultDark())

	// Init
	_ = m.Init()

	// Title
	title := m.Title()
	if title != "Jobs" {
		t.Errorf("Title() = %q, want Jobs", title)
	}

	// View
	_ = m.View(80, 24)

	// Hotkeys
	km := m.Hotkeys()
	if km == nil {
		t.Error("Hotkeys() returned nil")
	}

	// Summary
	summary := m.Summary()
	if summary == "" {
		t.Error("Summary() returned empty string")
	}

	// SortableColumns
	cols := m.SortableColumns()
	if len(cols) == 0 {
		t.Error("SortableColumns() returned empty list")
	}

	// ApplySort (just verify it doesn't panic)
	m.ApplySort("id", false)
	m.ApplySort("kind", true)
}
