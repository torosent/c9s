package theme

import (
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
)

// sameColor compares two colors by their RGBA components, allowing for
// lipgloss's parser-internal type variations (RGBA, ANSIColor, etc.).
func sameColor(a, b color.Color) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if a == nil {
		return true
	}
	ar, ag, ab, aa := a.RGBA()
	br, bg, bb, ba := b.RGBA()
	return ar == br && ag == bg && ab == bb && aa == ba
}

func TestLoadSkin_BundledDark(t *testing.T) {
	p, err := LoadSkin("dark")
	if err != nil {
		t.Fatalf("LoadSkin(dark) failed: %v", err)
	}

	if p.Fg == nil {
		t.Error("expected Fg to be set")
	}
	if p.Bg == nil {
		t.Error("expected Bg to be set")
	}
	if len(p.State) == 0 {
		t.Error("expected State colors to be set")
	}
}

func TestLoadSkin_BundledLight(t *testing.T) {
	p, err := LoadSkin("light")
	if err != nil {
		t.Fatalf("LoadSkin(light) failed: %v", err)
	}

	// Light theme should have different colors than dark
	if sameColor(p.Bg, lipgloss.Color("#0d1117")) {
		t.Error("light theme should not have dark background")
	}
}

func TestLoadSkin_K9sDark(t *testing.T) {
	p, err := LoadSkin("k9s-dark")
	if err != nil {
		t.Fatalf("LoadSkin(k9s-dark) failed: %v", err)
	}

	if p.Accent == nil {
		t.Error("expected Accent to be set")
	}
}

func TestLoadSkin_K9sLight(t *testing.T) {
	p, err := LoadSkin("k9s-light")
	if err != nil {
		t.Fatalf("LoadSkin(k9s-light) failed: %v", err)
	}

	if p.Fg == nil {
		t.Error("expected Fg to be set")
	}
}

func TestLoadSkin_Nonexistent(t *testing.T) {
	_, err := LoadSkin("nonexistent-theme")
	if err == nil {
		t.Error("expected error for nonexistent skin")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestLoadSkin_UserCustom(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	skinsDir := filepath.Join(tmpDir, "c9s", "skins")
	if err := os.MkdirAll(skinsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	customSkin := `
[colors]
fg = "#ff0000"
bg = "#000000"
border = "#333333"
accent = "#00ff00"
dim = "#666666"
success = "#0f0"
warning = "#ff0"
error = "#f00"
selection_fg = "#fff"
selection_bg = "#00f"
header_fg = "#eee"
header_bg = "#111"

[state_colors]
running = "#0f0"
exited = "#666"
paused = "#ff0"
stopping = "#f00"
created = "#00f"
`

	skinPath := filepath.Join(skinsDir, "custom.toml")
	if err := os.WriteFile(skinPath, []byte(customSkin), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := LoadSkin("custom")
	if err != nil {
		t.Fatalf("LoadSkin(custom) failed: %v", err)
	}

	if !sameColor(p.Fg, lipgloss.Color("#ff0000")) {
		t.Errorf("expected custom Fg=#ff0000, got %v", p.Fg)
	}
	if !sameColor(p.Accent, lipgloss.Color("#00ff00")) {
		t.Errorf("expected custom Accent=#00ff00, got %v", p.Accent)
	}
}

func TestLoadSkin_StateColorDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	skinsDir := filepath.Join(tmpDir, "c9s", "skins")
	if err := os.MkdirAll(skinsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Skin with partial state colors
	partialSkin := `
[colors]
fg = "#ffffff"
bg = "#000000"
border = "#333333"
accent = "#00ff00"
dim = "#666666"
success = "#0f0"
warning = "#ff0"
error = "#f00"
selection_fg = "#fff"
selection_bg = "#00f"
header_fg = "#eee"
header_bg = "#111"

[state_colors]
running = "#0f0"
# Missing exited, paused, stopping, created
`

	skinPath := filepath.Join(skinsDir, "partial.toml")
	if err := os.WriteFile(skinPath, []byte(partialSkin), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := LoadSkin("partial")
	if err != nil {
		t.Fatalf("LoadSkin(partial) failed: %v", err)
	}

	// Should have all state colors (some from defaults)
	for _, state := range []string{"running", "exited", "paused", "stopping", "created"} {
		if _, ok := p.State[state]; !ok {
			t.Errorf("expected State[%q] to be set (from defaults if needed)", state)
		}
	}
}
