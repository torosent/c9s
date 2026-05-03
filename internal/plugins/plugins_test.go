package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sample plugin files
	plugin1 := `
name = "edit"
description = "Edit container with $EDITOR"
scope = "container"
key = "ctrl+e"
confirm = false
background = false
command = ["sh", "-c", "$EDITOR /var/lib/docker/containers/$RESOURCE_ID/config.json"]
`

	plugin2 := `
name = "logs-less"
description = "View logs in less"
scope = "container"
key = "ctrl+l"
confirm = false
background = true
command = ["docker", "logs", "-f", "$RESOURCE_NAME"]
`

	if err := os.WriteFile(filepath.Join(tmpDir, "edit.toml"), []byte(plugin1), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "logs.toml"), []byte(plugin2), 0o644); err != nil {
		t.Fatal(err)
	}

	// Also create a non-plugin file to verify filtering
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Plugins"), 0o644); err != nil {
		t.Fatal(err)
	}

	plugins, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}

	// Verify plugin1
	if plugins[0].Name != "edit" && plugins[1].Name != "edit" {
		t.Error("expected one plugin named 'edit'")
	}

	// Verify plugin2
	if plugins[0].Name != "logs-less" && plugins[1].Name != "logs-less" {
		t.Error("expected one plugin named 'logs-less'")
	}
}

func TestLoadNonexistentDir(t *testing.T) {
	plugins, err := Load("/nonexistent/path")
	if err != nil {
		t.Errorf("Load on nonexistent dir should return nil error, got: %v", err)
	}
	if len(plugins) != 0 {
		t.Errorf("expected 0 plugins from nonexistent dir, got %d", len(plugins))
	}
}

func TestLoadFromXDG(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	pluginsDir := filepath.Join(tmpDir, "c9s", "plugins")
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		t.Fatal(err)
	}

	plugin := `
name = "test"
description = "Test plugin"
scope = "container"
key = "ctrl+t"
command = ["echo", "test"]
`

	if err := os.WriteFile(filepath.Join(pluginsDir, "test.toml"), []byte(plugin), 0o644); err != nil {
		t.Fatal(err)
	}

	plugins, err := LoadFromXDG()
	if err != nil {
		t.Fatalf("LoadFromXDG failed: %v", err)
	}

	if len(plugins) != 1 {
		t.Fatalf("expected 1 plugin, got %d", len(plugins))
	}

	if plugins[0].Name != "test" {
		t.Errorf("expected plugin name 'test', got %q", plugins[0].Name)
	}
}

func TestMatch(t *testing.T) {
	plugins := []Plugin{
		{Name: "edit", Scope: "container", Key: "ctrl+e"},
		{Name: "logs", Scope: "container", Key: "ctrl+l"},
		{Name: "pull", Scope: "image", Key: "ctrl+p"},
	}

	// Match found
	p, found := Match(plugins, "container", "ctrl+e")
	if !found {
		t.Error("expected to find container/ctrl+e")
	}
	if p.Name != "edit" {
		t.Errorf("expected plugin 'edit', got %q", p.Name)
	}

	// Match not found
	_, found = Match(plugins, "container", "ctrl+z")
	if found {
		t.Error("should not find container/ctrl+z")
	}

	// Match different scope
	p, found = Match(plugins, "image", "ctrl+p")
	if !found {
		t.Error("expected to find image/ctrl+p")
	}
	if p.Name != "pull" {
		t.Errorf("expected plugin 'pull', got %q", p.Name)
	}
}

func TestSubstitute(t *testing.T) {
	cmd := []string{"docker", "exec", "$RESOURCE_NAME", "sh", "-c", "echo $RESOURCE_ID"}
	vars := map[string]string{
		"RESOURCE_NAME": "my-container",
		"RESOURCE_ID":   "abc123",
	}

	result := Substitute(cmd, vars)

	expected := []string{"docker", "exec", "my-container", "sh", "-c", "echo abc123"}
	if len(result) != len(expected) {
		t.Fatalf("expected %d args, got %d", len(expected), len(result))
	}

	for i, arg := range result {
		if arg != expected[i] {
			t.Errorf("arg[%d]: expected %q, got %q", i, expected[i], arg)
		}
	}
}

func TestSubstituteNoVars(t *testing.T) {
	cmd := []string{"echo", "hello"}
	vars := map[string]string{}

	result := Substitute(cmd, vars)

	if len(result) != 2 {
		t.Fatalf("expected 2 args, got %d", len(result))
	}
	if result[0] != "echo" || result[1] != "hello" {
		t.Errorf("command should be unchanged: got %v", result)
	}
}
