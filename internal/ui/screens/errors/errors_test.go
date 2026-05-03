package errors_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/log"
	"github.com/torosent/c9s/internal/ui/screens/errors"
	"github.com/torosent/c9s/internal/ui/theme"
)

func TestErrorsScreen_NoFile(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))
	palette := theme.DefaultDark()

	m := errors.New(dir, clk, palette)
	if m == nil {
		t.Fatal("New() returned nil")
	}

	// Initialize
	m.Init()

	// Trigger load - Update returns Screen interface
	s, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m, _ = s.(*errors.Model)

	view := m.View(100, 40)
	if !strings.Contains(view, "TIME") {
		t.Error("view missing TIME header")
	}

	summary := m.Summary()
	if summary != "no errors" {
		t.Errorf("Summary() = %q, want %q", summary, "no errors")
	}
}

func TestErrorsScreen_WithEntries(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))
	palette := theme.DefaultDark()

	// Create log file with entries
	logPath := filepath.Join(dir, "errors-2026-05-02.log")
	f, err := os.Create(logPath)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	entries := []log.Entry{
		{Time: clk.Now(), Op: "container.start", Resource: "api", Message: "failed to start", Detail: "port busy"},
		{Time: clk.Now().Add(time.Minute), Op: "image.pull", Resource: "nginx", Message: "network timeout", Detail: ""},
	}

	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			t.Fatalf("Encode() error: %v", err)
		}
	}
	f.Close()

	// Create screen
	m := errors.New(dir, clk, palette)
	m.Init()

	// Send WindowSizeMsg to set dimensions
	s, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m, _ = s.(*errors.Model)

	// Manually trigger load (in real app, Init() returns tick command)
	s, _ = m.Update(errors.LoadEntriesMsg{})
	m, _ = s.(*errors.Model)

	view := m.View(100, 40)
	if !strings.Contains(view, "container.start") {
		t.Error("view missing first entry op")
	}
	if !strings.Contains(view, "image.pull") {
		t.Error("view missing second entry op")
	}

	summary := m.Summary()
	if summary != "2 errors" {
		t.Errorf("Summary() = %q, want %q", summary, "2 errors")
	}
}

func TestErrorsScreen_Title(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))
	palette := theme.DefaultDark()

	m := errors.New(dir, clk, palette)
	if title := m.Title(); title != "errors" {
		t.Errorf("Title() = %q, want %q", title, "errors")
	}
}

func TestErrorsScreen_Hotkeys(t *testing.T) {
	dir := t.TempDir()
	clk := clock.NewFake(time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC))
	palette := theme.DefaultDark()

	m := errors.New(dir, clk, palette)
	km := m.Hotkeys()
	if km == nil {
		t.Fatal("Hotkeys() returned nil")
	}

	// Check for key bindings
	if _, ok := km.Get("inspect"); !ok {
		t.Error("missing 'inspect' binding")
	}
	if _, ok := km.Get("copy"); !ok {
		t.Error("missing 'copy' binding")
	}
}
