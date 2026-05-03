// Package config provides the configuration system for c9s.
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
)

// Config is the top-level configuration container.
type Config struct {
	UI      UI                    `toml:"ui"`
	Stream  Stream                `toml:"stream"`
	Theme   Theme                 `toml:"theme"`
	Hotkeys map[string]string     `toml:"hotkeys"`
	Aliases map[string]string     `toml:"aliases"`
	Views   map[string]ViewConfig `toml:"views"`
}

// UI holds UI-related settings.
type UI struct {
	ReadOnly               bool `toml:"readonly"`
	HeaderVisibleDefault   bool `toml:"header_visible_default"`
	RefreshIntervalSeconds int  `toml:"refresh_interval_seconds"`
}

// Stream holds streaming/logs settings.
type Stream struct {
	LogBufferLines   int  `toml:"log_buffer_lines"`
	LogFollowDefault bool `toml:"log_follow_default"`
	BuildSaveLogs    bool `toml:"build_save_logs"`
	BuildAutoclose   bool `toml:"build_autoclose"`
}

// Theme holds theme configuration.
type Theme struct {
	Name      string            `toml:"name"`
	Overrides map[string]string `toml:"overrides"`
}

// ViewConfig holds per-screen column configuration.
type ViewConfig struct {
	Columns map[string]ColumnConfig `toml:"columns"`
}

// ColumnConfig defines visibility, order, and width for a table column.
type ColumnConfig struct {
	Visible bool `toml:"visible"`
	Width   int  `toml:"width"`
	Order   int  `toml:"order"`
}

// Default returns the default configuration.
func Default() Config {
	return Config{
		UI: UI{
			ReadOnly:               false,
			HeaderVisibleDefault:   true,
			RefreshIntervalSeconds: 3,
		},
		Stream: Stream{
			LogBufferLines:   10000,
			LogFollowDefault: true,
			BuildSaveLogs:    false,
			BuildAutoclose:   true,
		},
		Theme: Theme{
			Name:      "dark",
			Overrides: make(map[string]string),
		},
		Hotkeys: make(map[string]string),
		Aliases: make(map[string]string),
		Views:   make(map[string]ViewConfig),
	}
}

// Load parses a TOML config file and merges it atop the defaults.
func Load(path string) (Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	// Use a pointer-based struct for selective merging
	type uiPartial struct {
		ReadOnly               *bool `toml:"readonly"`
		HeaderVisibleDefault   *bool `toml:"header_visible_default"`
		RefreshIntervalSeconds *int  `toml:"refresh_interval_seconds"`
	}
	type streamPartial struct {
		LogBufferLines   *int  `toml:"log_buffer_lines"`
		LogFollowDefault *bool `toml:"log_follow_default"`
		BuildSaveLogs    *bool `toml:"build_save_logs"`
		BuildAutoclose   *bool `toml:"build_autoclose"`
	}
	type themePartial struct {
		Name      *string           `toml:"name"`
		Overrides map[string]string `toml:"overrides"`
	}
	type partialConfig struct {
		UI      *uiPartial            `toml:"ui"`
		Stream  *streamPartial        `toml:"stream"`
		Theme   *themePartial         `toml:"theme"`
		Hotkeys map[string]string     `toml:"hotkeys"`
		Aliases map[string]string     `toml:"aliases"`
		Views   map[string]ViewConfig `toml:"views"`
	}

	var partial partialConfig
	if _, err := toml.Decode(string(data), &partial); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	// Merge UI
	if partial.UI != nil {
		if partial.UI.ReadOnly != nil {
			cfg.UI.ReadOnly = *partial.UI.ReadOnly
		}
		if partial.UI.HeaderVisibleDefault != nil {
			cfg.UI.HeaderVisibleDefault = *partial.UI.HeaderVisibleDefault
		}
		if partial.UI.RefreshIntervalSeconds != nil {
			cfg.UI.RefreshIntervalSeconds = *partial.UI.RefreshIntervalSeconds
		}
	}

	// Merge Stream
	if partial.Stream != nil {
		if partial.Stream.LogBufferLines != nil {
			cfg.Stream.LogBufferLines = *partial.Stream.LogBufferLines
		}
		if partial.Stream.LogFollowDefault != nil {
			cfg.Stream.LogFollowDefault = *partial.Stream.LogFollowDefault
		}
		if partial.Stream.BuildSaveLogs != nil {
			cfg.Stream.BuildSaveLogs = *partial.Stream.BuildSaveLogs
		}
		if partial.Stream.BuildAutoclose != nil {
			cfg.Stream.BuildAutoclose = *partial.Stream.BuildAutoclose
		}
	}

	// Merge Theme
	if partial.Theme != nil {
		if partial.Theme.Name != nil {
			cfg.Theme.Name = *partial.Theme.Name
		}
		if len(partial.Theme.Overrides) > 0 {
			cfg.Theme.Overrides = partial.Theme.Overrides
		}
	}

	// Merge maps
	if len(partial.Hotkeys) > 0 {
		cfg.Hotkeys = partial.Hotkeys
	}
	if len(partial.Aliases) > 0 {
		cfg.Aliases = partial.Aliases
	}
	if len(partial.Views) > 0 {
		cfg.Views = partial.Views
	}

	return cfg, nil
}

// Watch watches config files for changes and sends ChangedMsg on updates.
// It returns a channel that will receive messages when any watched file changes.
// The context can be used to cancel the watch.
func Watch(ctx context.Context, paths []string) (<-chan ChangedMsg, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}

	// Add all paths to watch
	for _, path := range paths {
		// Watch parent directory since files might be recreated
		dir := filepath.Dir(path)
		if err := watcher.Add(dir); err != nil {
			watcher.Close()
			return nil, fmt.Errorf("watch %s: %w", dir, err)
		}
	}

	ch := make(chan ChangedMsg, 1)

	go func() {
		defer watcher.Close()
		defer close(ch)

		// Debounce rapid changes
		var debounceTimer *time.Timer
		debounceDuration := 100 * time.Millisecond

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				// Check if this event is for one of our watched files
				matched := false
				for _, path := range paths {
					if event.Name == path {
						matched = true
						break
					}
				}
				if !matched {
					continue
				}

				// Only react to write or create events
				if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}

				// Debounce: reset timer on each event
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDuration, func() {
					// Reload config from XDG (assuming most common case)
					cfg, err := LoadFromXDG()
					if err != nil {
						// Silently ignore errors - don't break on invalid config
						return
					}
					select {
					case ch <- ChangedMsg{Config: cfg}:
					case <-ctx.Done():
					}
				})

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				// Log error but continue watching
				_ = err
			}
		}
	}()

	return ch, nil
}

// LoadFromXDG loads config from XDG_CONFIG_HOME or ~/.config/c9s/config.toml.
// Returns Default() if the file doesn't exist.
func LoadFromXDG() (Config, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Default(), nil
		}
		configDir = filepath.Join(home, ".config")
	}

	cfgPath := filepath.Join(configDir, "c9s", "config.toml")

	// If file doesn't exist, return defaults
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		return Default(), nil
	}

	return Load(cfgPath)
}

// XDGConfigPath returns the path c9s reads/writes its config to.
// Creates parent directories on demand.
func XDGConfigPath() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("home dir: %w", err)
		}
		configDir = filepath.Join(home, ".config")
	}
	dir := filepath.Join(configDir, "c9s")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return filepath.Join(dir, "config.toml"), nil
}

// SaveSkin persists the active skin name to the user's config.toml so it
// survives across c9s sessions. Other config fields are left untouched if
// the file already exists; otherwise a minimal config is written.
func SaveSkin(name string) error {
	path, err := XDGConfigPath()
	if err != nil {
		return err
	}

	cfg := Default()
	if existing, err := Load(path); err == nil {
		cfg = existing
	}
	cfg.Theme.Name = name

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(&cfg)
}
