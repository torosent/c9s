// Package buildstats provides build time estimation with EWMA persistence.
package buildstats

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Stat represents a single build statistic with EWMA smoothing.
type Stat struct {
	Key             string  `toml:"key"`
	DurationSeconds float64 `toml:"duration_seconds"`
}

// Stats is the top-level structure for the TOML file.
type Stats struct {
	Stats []Stat `toml:"stats"`
}

// Load reads build stats from the given path.
// Returns empty Stats if file doesn't exist (not an error).
func Load(path string) (Stats, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Stats{}, nil
		}
		return Stats{}, fmt.Errorf("read stats: %w", err)
	}

	var stats Stats
	if err := toml.Unmarshal(data, &stats); err != nil {
		return Stats{}, fmt.Errorf("parse stats: %w", err)
	}

	return stats, nil
}

// Save writes build stats to the given path.
func (s Stats) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create stats dir: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create stats file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(s); err != nil {
		return fmt.Errorf("encode stats: %w", err)
	}

	return nil
}

// Get retrieves the estimated duration for a key.
// Returns 0 if key not found.
func (s Stats) Get(key string) float64 {
	for _, stat := range s.Stats {
		if stat.Key == key {
			return stat.DurationSeconds
		}
	}
	return 0
}

// Update applies EWMA smoothing to update the duration for a key.
// If the key doesn't exist, it's added with the current value.
// alpha is the smoothing factor (0 < alpha <= 1), typically 0.3.
// new = alpha * current + (1 - alpha) * prev
func (s *Stats) Update(key string, currentSeconds float64, alpha float64) {
	if alpha <= 0 || alpha > 1 {
		alpha = 0.3 // Default smoothing factor
	}

	for i, stat := range s.Stats {
		if stat.Key == key {
			// Apply EWMA: new = alpha * current + (1 - alpha) * prev
			s.Stats[i].DurationSeconds = alpha*currentSeconds + (1-alpha)*stat.DurationSeconds
			return
		}
	}

	// Key not found, add it
	s.Stats = append(s.Stats, Stat{
		Key:             key,
		DurationSeconds: currentSeconds,
	})
}

// LoadFromXDG loads build stats from ~/.local/share/c9s/build-stats.toml
func LoadFromXDG() (Stats, error) {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return Stats{}, nil // Can't determine home, return empty
		}
		dataDir = filepath.Join(home, ".local", "share")
	}

	path := filepath.Join(dataDir, "c9s", "build-stats.toml")
	return Load(path)
}

// SaveToXDG saves build stats to ~/.local/share/c9s/build-stats.toml
func (s Stats) SaveToXDG() error {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("get home dir: %w", err)
		}
		dataDir = filepath.Join(home, ".local", "share")
	}

	path := filepath.Join(dataDir, "c9s", "build-stats.toml")
	return s.Save(path)
}
