package theme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

func TestImportK9sSkin(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Real k9s YAML uses nested body/info/frame blocks.
	k9sYAML := `
k9s:
  body:
    fgColor: "#d8dee9"
    bgColor: "#2e3440"
    logoColor: "#88c0d0"
  info:
    fgColor: "#88c0d0"
    sectionColor: "#616e88"
  frame:
    border:
      fgColor: "#4c566a"
      focusColor: "#5e81ac"
    title:
      fgColor: "#eceff4"
      bgColor: "#2e3440"
      highlightColor: "#88c0d0"
    crumbs:
      fgColor: "#616e88"
    status:
      newColor: "#a3be8c"
      modifyColor: "#d08770"
      addColor: "#a3be8c"
      errorColor: "#bf616a"
      pendingColor: "#ebcb8b"
      highlightColor: "#88c0d0"
`

	yamlPath := filepath.Join(tmpDir, "my-k9s-skin.yaml")
	if err := os.WriteFile(yamlPath, []byte(k9sYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	outPath, err := ImportK9sSkin(yamlPath)
	if err != nil {
		t.Fatalf("ImportK9sSkin failed: %v", err)
	}

	expectedName := "my-k9s-skin.toml"
	if !strings.HasSuffix(outPath, expectedName) {
		t.Errorf("expected output path to end with %q, got %q", expectedName, outPath)
	}

	if _, err := os.Stat(outPath); err != nil {
		t.Errorf("output file not found: %v", err)
	}

	var skin Skin
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := toml.Unmarshal(data, &skin); err != nil {
		t.Fatalf("parse generated TOML: %v", err)
	}

	// Verify mappings
	if skin.Colors.Fg != "#d8dee9" {
		t.Errorf("expected Fg=#d8dee9, got %s", skin.Colors.Fg)
	}
	if skin.Colors.Bg != "#2e3440" {
		t.Errorf("expected Bg=#2e3440, got %s", skin.Colors.Bg)
	}
	// Border maps to frame.border.focusColor (the visible focused border in k9s).
	if skin.Colors.Border != "#5e81ac" {
		t.Errorf("expected Border=#5e81ac, got %s", skin.Colors.Border)
	}
	// Accent maps to body.logoColor (the c9s/k9s logo + highlight color).
	if skin.Colors.Accent != "#88c0d0" {
		t.Errorf("expected Accent=#88c0d0, got %s", skin.Colors.Accent)
	}
	if skin.Colors.Success != "#a3be8c" {
		t.Errorf("expected Success=#a3be8c, got %s", skin.Colors.Success)
	}

	if skin.StateColors["running"] != "#a3be8c" {
		t.Errorf("expected running=#a3be8c, got %s", skin.StateColors["running"])
	}
	if skin.StateColors["exited"] != "#616e88" {
		t.Errorf("expected exited=#616e88, got %s", skin.StateColors["exited"])
	}
}

func TestImportK9sSkin_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(yamlPath, []byte("not valid yaml: [[["), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := ImportK9sSkin(yamlPath)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("expected 'parse' in error, got: %v", err)
	}
}

func TestImportK9sSkin_MissingFile(t *testing.T) {
	_, err := ImportK9sSkin("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "read") {
		t.Errorf("expected 'read' in error, got: %v", err)
	}
}
