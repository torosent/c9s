package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	// Check UI defaults
	if cfg.UI.ReadOnly {
		t.Error("expected ReadOnly to be false by default")
	}
	if !cfg.UI.HeaderVisibleDefault {
		t.Error("expected HeaderVisibleDefault to be true")
	}
	if cfg.UI.RefreshIntervalSeconds <= 0 {
		t.Error("expected positive refresh interval")
	}

	// Check Stream defaults
	if cfg.Stream.LogBufferLines <= 0 {
		t.Error("expected positive log buffer lines")
	}
	if !cfg.Stream.LogFollowDefault {
		t.Error("expected log follow default true")
	}

	// Check Theme defaults
	if cfg.Theme.Name == "" {
		t.Error("expected default theme name")
	}

	// Check that collections are non-nil
	if cfg.Hotkeys == nil {
		t.Error("expected non-nil Hotkeys map")
	}
	if cfg.Aliases == nil {
		t.Error("expected non-nil Aliases map")
	}
	if cfg.Views == nil {
		t.Error("expected non-nil Views map")
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	// Create empty config file
	if err := os.WriteFile(cfgPath, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Should return defaults for empty file
	defaults := Default()
	if cfg.UI.RefreshIntervalSeconds != defaults.UI.RefreshIntervalSeconds {
		t.Error("expected default refresh interval for empty file")
	}
}

func TestLoad_PartialOverride(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	tomlContent := `
[ui]
readonly = true
refresh_interval_seconds = 10

[stream]
log_buffer_lines = 5000
`

	if err := os.WriteFile(cfgPath, []byte(tomlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Check overridden values
	if !cfg.UI.ReadOnly {
		t.Error("expected ReadOnly to be true from config")
	}
	if cfg.UI.RefreshIntervalSeconds != 10 {
		t.Errorf("expected refresh interval 10, got %d", cfg.UI.RefreshIntervalSeconds)
	}
	if cfg.Stream.LogBufferLines != 5000 {
		t.Errorf("expected log buffer 5000, got %d", cfg.Stream.LogBufferLines)
	}

	// Check non-overridden defaults are preserved
	if !cfg.UI.HeaderVisibleDefault {
		t.Error("expected default HeaderVisibleDefault to be preserved")
	}
	if !cfg.Stream.LogFollowDefault {
		t.Error("expected default LogFollowDefault to be preserved")
	}
}

func TestLoad_NonExistentFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.toml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadFromXDG_MissingFile(t *testing.T) {
	// Override XDG_CONFIG_HOME to a temp dir with no config
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	t.Setenv("HOME", tmpDir) // fallback

	cfg, err := LoadFromXDG()
	if err != nil {
		t.Fatalf("LoadFromXDG() should not error on missing file: %v", err)
	}

	// Should return defaults when file doesn't exist
	defaults := Default()
	if cfg.UI.RefreshIntervalSeconds != defaults.UI.RefreshIntervalSeconds {
		t.Error("expected default config when file missing")
	}
}

func TestLoadFromXDG_WithFile(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfgDir := filepath.Join(tmpDir, "c9s")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(cfgDir, "config.toml")
	tomlContent := `
[ui]
readonly = true
`

	if err := os.WriteFile(cfgPath, []byte(tomlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromXDG()
	if err != nil {
		t.Fatalf("LoadFromXDG() failed: %v", err)
	}

	if !cfg.UI.ReadOnly {
		t.Error("expected readonly from XDG config file")
	}
}

func TestLoad_ComplexConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")

	tomlContent := `
[ui]
readonly = false
header_visible_default = false
refresh_interval_seconds = 5

[stream]
log_buffer_lines = 10000
log_follow_default = false
build_save_logs = true
build_autoclose = false

[theme]
name = "light"

[hotkeys]
"containers.delete" = "x"
"images.delete" = "shift+x"

[aliases]
kpods = "containers"
kimg = "images"
`

	if err := os.WriteFile(cfgPath, []byte(tomlContent), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// UI checks
	if cfg.UI.ReadOnly {
		t.Error("expected ReadOnly false")
	}
	if cfg.UI.HeaderVisibleDefault {
		t.Error("expected HeaderVisibleDefault false")
	}
	if cfg.UI.RefreshIntervalSeconds != 5 {
		t.Errorf("expected refresh 5, got %d", cfg.UI.RefreshIntervalSeconds)
	}

	// Stream checks
	if cfg.Stream.LogBufferLines != 10000 {
		t.Errorf("expected log buffer 10000, got %d", cfg.Stream.LogBufferLines)
	}
	if cfg.Stream.LogFollowDefault {
		t.Error("expected LogFollowDefault false")
	}
	if !cfg.Stream.BuildSaveLogs {
		t.Error("expected BuildSaveLogs true")
	}

	// Theme checks
	if cfg.Theme.Name != "light" {
		t.Errorf("expected theme 'light', got %q", cfg.Theme.Name)
	}

	// Hotkeys checks
	if cfg.Hotkeys["containers.delete"] != "x" {
		t.Error("expected hotkey override for containers.delete")
	}
	if cfg.Hotkeys["images.delete"] != "shift+x" {
		t.Error("expected hotkey override for images.delete")
	}

	// Aliases checks
	if cfg.Aliases["kpods"] != "containers" {
		t.Error("expected alias kpods -> containers")
	}
	if cfg.Aliases["kimg"] != "images" {
		t.Error("expected alias kimg -> images")
	}
}

func TestWatch(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfgDir := filepath.Join(tmpDir, "c9s")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}

	cfgPath := filepath.Join(cfgDir, "config.toml")

	// Write initial config
	initialContent := `[ui]
refresh_interval_seconds = 5
`
	if err := os.WriteFile(cfgPath, []byte(initialContent), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := Watch(ctx, []string{cfgPath})
	if err != nil {
		t.Fatalf("Watch() failed: %v", err)
	}

	// Update config file
	updatedContent := `[ui]
refresh_interval_seconds = 10
`
	if err := os.WriteFile(cfgPath, []byte(updatedContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Wait for change notification
	select {
	case msg := <-ch:
		if msg.Config.UI.RefreshIntervalSeconds != 10 {
			t.Errorf("expected refresh interval 10, got %d", msg.Config.UI.RefreshIntervalSeconds)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for config change notification")
	}
}
