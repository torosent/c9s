// Package plugins provides plugin loading and execution for c9s.
package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Plugin represents a user-defined command plugin.
type Plugin struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Scope       string   `toml:"scope"`      // "container", "image", "volume", "network", etc.
	Key         string   `toml:"key"`        // Hotkey binding (e.g., "ctrl+e")
	Confirm     bool     `toml:"confirm"`    // Require confirmation before execution
	Background  bool     `toml:"background"` // Run in background (don't block TUI)
	Command     []string `toml:"command"`    // Command template with variables
}

// Load reads all *.toml plugin files from the given directory.
func Load(dir string) ([]Plugin, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No plugins dir is OK
		}
		return nil, fmt.Errorf("read plugins dir: %w", err)
	}

	var plugins []Plugin
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".toml") {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read plugin %s: %w", entry.Name(), err)
		}

		var p Plugin
		if err := toml.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("parse plugin %s: %w", entry.Name(), err)
		}

		plugins = append(plugins, p)
	}

	return plugins, nil
}

// LoadFromXDG loads plugins from ~/.config/c9s/plugins/
func LoadFromXDG() ([]Plugin, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, nil // Can't determine home, no plugins
		}
		configDir = filepath.Join(home, ".config")
	}

	pluginsDir := filepath.Join(configDir, "c9s", "plugins")
	return Load(pluginsDir)
}

// Match finds a plugin matching the given scope and key.
// Returns the plugin and true if found, or a zero Plugin and false if not found.
func Match(plugins []Plugin, scope, key string) (Plugin, bool) {
	for _, p := range plugins {
		if p.Scope == scope && p.Key == key {
			return p, true
		}
	}
	return Plugin{}, false
}

// Substitute replaces template variables in the command with actual values.
// Supported variables:
//
//	$RESOURCE_NAME - Name of the selected resource
//	$RESOURCE_ID   - ID of the selected resource
//	$NAMESPACE     - Namespace (k8s compatibility, always empty for Docker)
func Substitute(command []string, vars map[string]string) []string {
	result := make([]string, len(command))
	for i, arg := range command {
		result[i] = arg
		for k, v := range vars {
			placeholder := "$" + k
			result[i] = strings.ReplaceAll(result[i], placeholder, v)
		}
	}
	return result
}
