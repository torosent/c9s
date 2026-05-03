package buildstats

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSaveRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "stats.toml")

	// Create stats
	stats := Stats{
		Stats: []Stat{
			{Key: "image1", DurationSeconds: 45.5},
			{Key: "image2", DurationSeconds: 120.3},
		},
	}

	// Save
	if err := stats.Save(path); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify
	if len(loaded.Stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(loaded.Stats))
	}

	if loaded.Stats[0].Key != "image1" || loaded.Stats[0].DurationSeconds != 45.5 {
		t.Errorf("stat 0 mismatch: %+v", loaded.Stats[0])
	}

	if loaded.Stats[1].Key != "image2" || loaded.Stats[1].DurationSeconds != 120.3 {
		t.Errorf("stat 1 mismatch: %+v", loaded.Stats[1])
	}
}

func TestLoadNonexistent(t *testing.T) {
	stats, err := Load("/nonexistent/path.toml")
	if err != nil {
		t.Errorf("Load on nonexistent file should not error, got: %v", err)
	}
	if len(stats.Stats) != 0 {
		t.Errorf("expected empty stats, got %d", len(stats.Stats))
	}
}

func TestGet(t *testing.T) {
	stats := Stats{
		Stats: []Stat{
			{Key: "image1", DurationSeconds: 45.5},
			{Key: "image2", DurationSeconds: 120.3},
		},
	}

	if got := stats.Get("image1"); got != 45.5 {
		t.Errorf("Get(image1) = %f, want 45.5", got)
	}

	if got := stats.Get("image2"); got != 120.3 {
		t.Errorf("Get(image2) = %f, want 120.3", got)
	}

	if got := stats.Get("nonexistent"); got != 0 {
		t.Errorf("Get(nonexistent) = %f, want 0", got)
	}
}

func TestUpdateExisting(t *testing.T) {
	stats := Stats{
		Stats: []Stat{
			{Key: "image1", DurationSeconds: 100.0},
		},
	}

	// Update with alpha=0.3
	// new = 0.3 * 90 + 0.7 * 100 = 27 + 70 = 97
	stats.Update("image1", 90.0, 0.3)

	got := stats.Get("image1")
	expected := 97.0
	if math.Abs(got-expected) > 0.01 {
		t.Errorf("Update EWMA: got %f, want %f", got, expected)
	}
}

func TestUpdateNew(t *testing.T) {
	stats := Stats{}

	// Add new key
	stats.Update("new-image", 55.5, 0.3)

	if got := stats.Get("new-image"); got != 55.5 {
		t.Errorf("Update new key: got %f, want 55.5", got)
	}

	if len(stats.Stats) != 1 {
		t.Errorf("expected 1 stat, got %d", len(stats.Stats))
	}
}

func TestUpdateInvalidAlpha(t *testing.T) {
	stats := Stats{
		Stats: []Stat{
			{Key: "image1", DurationSeconds: 100.0},
		},
	}

	// Invalid alpha should default to 0.3
	// new = 0.3 * 80 + 0.7 * 100 = 24 + 70 = 94
	stats.Update("image1", 80.0, 0)

	got := stats.Get("image1")
	expected := 94.0
	if math.Abs(got-expected) > 0.01 {
		t.Errorf("Update with invalid alpha: got %f, want %f", got, expected)
	}
}

func TestUpdateMultipleTimes(t *testing.T) {
	stats := Stats{}

	// Simulate multiple builds
	stats.Update("image1", 100.0, 0.3)
	stats.Update("image1", 90.0, 0.3)
	stats.Update("image1", 110.0, 0.3)

	// After first: 100
	// After second: 0.3*90 + 0.7*100 = 27 + 70 = 97
	// After third: 0.3*110 + 0.7*97 = 33 + 67.9 = 100.9
	got := stats.Get("image1")
	expected := 100.9
	if math.Abs(got-expected) > 0.01 {
		t.Errorf("Update multiple times: got %f, want %f", got, expected)
	}
}

func TestLoadFromXDG(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	statsDir := filepath.Join(tmpDir, "c9s")
	if err := os.MkdirAll(statsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create a test stats file
	stats := Stats{
		Stats: []Stat{
			{Key: "test-image", DurationSeconds: 42.0},
		},
	}

	if err := stats.SaveToXDG(); err != nil {
		t.Fatalf("SaveToXDG failed: %v", err)
	}

	// Load it back
	loaded, err := LoadFromXDG()
	if err != nil {
		t.Fatalf("LoadFromXDG failed: %v", err)
	}

	if len(loaded.Stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(loaded.Stats))
	}

	if loaded.Stats[0].Key != "test-image" {
		t.Errorf("expected key 'test-image', got %q", loaded.Stats[0].Key)
	}
}
