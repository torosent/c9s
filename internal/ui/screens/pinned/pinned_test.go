package pinned_test

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/pinned"
	"github.com/torosent/c9s/internal/ui/screens"
	pinnedscreen "github.com/torosent/c9s/internal/ui/screens/pinned"
	"github.com/torosent/c9s/internal/ui/theme"
)

func newStore(t *testing.T) *pinned.Store {
	t.Helper()
	store, err := pinned.Load(filepath.Join(t.TempDir(), "pinned.toml"))
	if err != nil {
		t.Fatalf("pinned.Load: %v", err)
	}
	return store
}

func TestPinnedScreen_Empty(t *testing.T) {
	m := pinnedscreen.New(newStore(t), theme.DefaultDark())
	if m == nil {
		t.Fatal("New() returned nil")
	}
	m.Init()
	if got := m.Title(); got != "pinned" {
		t.Errorf("Title() = %q, want %q", got, "pinned")
	}
	if got := m.Summary(); got != "no pins" {
		t.Errorf("Summary() with no pins = %q, want %q", got, "no pins")
	}
	if hk := m.Hotkeys(); hk == nil {
		t.Error("Hotkeys() returned nil")
	}
	view := m.View(100, 40)
	if !strings.Contains(view, "RESOURCE") {
		t.Errorf("view missing RESOURCE header; got: %q", view)
	}
}

func TestPinnedScreen_PopulatedSortAndSummary(t *testing.T) {
	store := newStore(t)
	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	if err := store.Pin(pinned.Pin{Resource: "containers", ID: "abc123", Display: "redis", Added: now}); err != nil {
		t.Fatalf("Pin: %v", err)
	}
	if err := store.Pin(pinned.Pin{Resource: "images", ID: "img:1", Display: "alpine", Added: now.Add(time.Minute)}); err != nil {
		t.Fatalf("Pin: %v", err)
	}

	m := pinnedscreen.New(store, theme.DefaultDark())
	m.Init()

	if got := m.Summary(); got != "2 pins" {
		t.Errorf("Summary() with 2 pins = %q, want %q", got, "2 pins")
	}

	cols := m.SortableColumns()
	if len(cols) == 0 {
		t.Fatal("SortableColumns() returned no columns")
	}

	// Exercise ApplySort through both directions.
	m.ApplySort("resource", false)
	m.ApplySort("added", true)

	// Trigger a window-size update + render to exercise Update path.
	s, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	if _, ok := s.(*pinnedscreen.Model); !ok {
		t.Fatalf("Update returned wrong concrete type %T", s)
	}
	if view := m.View(120, 40); view == "" {
		t.Error("View returned empty string")
	}
}

// TestPinnedScreen_SortableInterface asserts the screen satisfies the same
// anonymous interface that app.go uses to dispatch sort changes. This is a
// regression guard for the kind of pointer/value-receiver mismatch described
// in C1 of the v0.1.0 review (the sort modal silently no-op'd on every
// value-returning screen).
func TestPinnedScreen_SortableInterface(t *testing.T) {
	var s screens.Screen = pinnedscreen.New(newStore(t), theme.DefaultDark())
	if _, ok := s.(interface {
		ApplySort(key string, reverse bool)
	}); !ok {
		t.Fatal("pinned screen does not satisfy ApplySort via screens.Screen interface")
	}
}
