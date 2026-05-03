package pinned_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/torosent/c9s/internal/pinned"
)

func TestLoad_NoFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pinned.toml")
	s, err := pinned.Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if s == nil {
		t.Fatal("Load() returned nil store")
	}
	pins := s.List()
	if len(pins) != 0 {
		t.Errorf("List() = %d pins, want 0", len(pins))
	}
}

func TestPin_RoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pinned.toml")
	s, err := pinned.Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	now := time.Date(2026, 5, 2, 10, 0, 0, 0, time.UTC)
	p := pinned.Pin{
		Resource: "containers",
		ID:       "abc123",
		Display:  "api-server",
		Added:    now,
	}

	if err := s.Pin(p); err != nil {
		t.Fatalf("Pin() error: %v", err)
	}

	if !s.Has("containers", "abc123") {
		t.Error("Has() = false after Pin()")
	}

	// Reload from disk
	s2, err := pinned.Load(path)
	if err != nil {
		t.Fatalf("Load() after Pin() error: %v", err)
	}

	pins := s2.List()
	if len(pins) != 1 {
		t.Fatalf("List() = %d pins, want 1", len(pins))
	}
	if pins[0].Resource != "containers" {
		t.Errorf("pins[0].Resource = %q, want %q", pins[0].Resource, "containers")
	}
	if pins[0].ID != "abc123" {
		t.Errorf("pins[0].ID = %q, want %q", pins[0].ID, "abc123")
	}
}

func TestUnpin(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pinned.toml")
	s, err := pinned.Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	now := time.Now()
	p := pinned.Pin{Resource: "images", ID: "nginx", Display: "nginx:latest", Added: now}
	if err := s.Pin(p); err != nil {
		t.Fatalf("Pin() error: %v", err)
	}

	if err := s.Unpin("images", "nginx"); err != nil {
		t.Fatalf("Unpin() error: %v", err)
	}

	if s.Has("images", "nginx") {
		t.Error("Has() = true after Unpin()")
	}

	pins := s.List()
	if len(pins) != 0 {
		t.Errorf("List() = %d pins after Unpin(), want 0", len(pins))
	}
}

func TestList_SortedByAddedDesc(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pinned.toml")
	s, err := pinned.Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	t1 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	t3 := time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC)

	s.Pin(pinned.Pin{Resource: "c", ID: "1", Display: "c1", Added: t2})
	s.Pin(pinned.Pin{Resource: "c", ID: "2", Display: "c2", Added: t1})
	s.Pin(pinned.Pin{Resource: "c", ID: "3", Display: "c3", Added: t3})

	pins := s.List()
	if len(pins) != 3 {
		t.Fatalf("List() = %d pins, want 3", len(pins))
	}

	// Should be sorted newest first
	if !pins[0].Added.Equal(t3) {
		t.Errorf("pins[0].Added = %v, want %v", pins[0].Added, t3)
	}
	if !pins[1].Added.Equal(t2) {
		t.Errorf("pins[1].Added = %v, want %v", pins[1].Added, t2)
	}
	if !pins[2].Added.Equal(t1) {
		t.Errorf("pins[2].Added = %v, want %v", pins[2].Added, t1)
	}
}

func TestPin_Persistence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pinned.toml")

	// Create and pin
	s1, _ := pinned.Load(path)
	s1.Pin(pinned.Pin{Resource: "volumes", ID: "vol1", Display: "data", Added: time.Now()})

	// Reload
	s2, err := pinned.Load(path)
	if err != nil {
		t.Fatalf("Load() after Pin() error: %v", err)
	}

	if !s2.Has("volumes", "vol1") {
		t.Error("Has() = false after reload")
	}

	// Check file exists
	if _, err := os.Stat(path); err != nil {
		t.Errorf("pinned.toml not created: %v", err)
	}
}
