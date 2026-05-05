package theme

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/BurntSushi/toml"
	"charm.land/lipgloss/v2"
)

//go:embed skins/*.toml
var bundledSkins embed.FS

// Skin represents a color theme loaded from TOML.
type Skin struct {
	Colors struct {
		Fg          string `toml:"fg"`
		Bg          string `toml:"bg"`
		Border      string `toml:"border"`
		Accent      string `toml:"accent"`
		Dim         string `toml:"dim"`
		Success     string `toml:"success"`
		Warning     string `toml:"warning"`
		Error       string `toml:"error"`
		SelectionFg string `toml:"selection_fg"`
		SelectionBg string `toml:"selection_bg"`
		HeaderFg    string `toml:"header_fg"`
		HeaderBg    string `toml:"header_bg"`
	} `toml:"colors"`
	StateColors map[string]string `toml:"state_colors"`
}

// ListSkins returns the names of all available skins (bundled + user-installed).
// Names exclude the .toml extension.
func ListSkins() []string {
	seen := map[string]bool{}
	var names []string

	// Bundled
	entries, _ := bundledSkins.ReadDir("skins")
	for _, e := range entries {
		name := strings.TrimSuffix(e.Name(), ".toml")
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}

	// User
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			configDir = filepath.Join(home, ".config")
		}
	}
	if configDir != "" {
		userDir := filepath.Join(configDir, "c9s", "skins")
		userEntries, _ := os.ReadDir(userDir)
		for _, e := range userEntries {
			if filepath.Ext(e.Name()) != ".toml" {
				continue
			}
			name := strings.TrimSuffix(e.Name(), ".toml")
			if seen[name] {
				continue
			}
			seen[name] = true
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

// LoadSkin loads a skin by name, first checking bundled skins (embedded),
// then ~/.config/c9s/skins/<name>.toml.
func LoadSkin(name string) (Palette, error) {
	// Try bundled first
	bundledPath := fmt.Sprintf("skins/%s.toml", name)
	data, err := bundledSkins.ReadFile(bundledPath)
	if err == nil {
		return parseSkin(data)
	}

	// Try user config directory
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Palette{}, fmt.Errorf("skin %q not found", name)
		}
		configDir = filepath.Join(home, ".config")
	}

	userPath := filepath.Join(configDir, "c9s", "skins", name+".toml")
	data, err = os.ReadFile(userPath)
	if err != nil {
		return Palette{}, fmt.Errorf("skin %q not found (tried bundled and %s)", name, userPath)
	}

	return parseSkin(data)
}

func parseSkin(data []byte) (Palette, error) {
	var skin Skin
	if err := toml.Unmarshal(data, &skin); err != nil {
		return Palette{}, fmt.Errorf("parse skin: %w", err)
	}

	// Convert Skin to Palette
	p := Palette{
		Fg:          lipgloss.Color(skin.Colors.Fg),
		Bg:          lipgloss.Color(skin.Colors.Bg),
		Border:      lipgloss.Color(skin.Colors.Border),
		Accent:      lipgloss.Color(skin.Colors.Accent),
		Dim:         lipgloss.Color(skin.Colors.Dim),
		Success:     lipgloss.Color(skin.Colors.Success),
		Warning:     lipgloss.Color(skin.Colors.Warning),
		Error:       lipgloss.Color(skin.Colors.Error),
		SelectionFg: lipgloss.Color(skin.Colors.SelectionFg),
		SelectionBg: lipgloss.Color(skin.Colors.SelectionBg),
		HeaderFg:    lipgloss.Color(skin.Colors.HeaderFg),
		HeaderBg:    lipgloss.Color(skin.Colors.HeaderBg),
		State:       make(map[string]lipgloss.Color),
	}

	// Convert state colors
	for state, color := range skin.StateColors {
		p.State[state] = lipgloss.Color(color)
	}

	// Fill in missing state colors with defaults
	defaults := DefaultDark()
	for _, state := range []string{"running", "exited", "paused", "stopping", "created"} {
		if _, ok := p.State[state]; !ok {
			p.State[state] = defaults.State[state]
		}
	}

	return p, nil
}
