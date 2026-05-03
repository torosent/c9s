// Package pinned provides persistent bookmarking of resources across screens.
package pinned

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

// Pin represents a bookmarked resource.
type Pin struct {
	Resource string    `toml:"resource"`
	ID       string    `toml:"id"`
	Display  string    `toml:"display"`
	Added    time.Time `toml:"added"`
}

// Store manages the pinned resources collection.
type Store struct {
	path string
	mu   sync.RWMutex
	pins map[string]Pin // key: resource:id
}

type tomlFormat struct {
	Pins []Pin `toml:"pins"`
}

// Load reads the pinned store from disk.
func Load(path string) (*Store, error) {
	s := &Store{
		path: path,
		pins: make(map[string]Pin),
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, fmt.Errorf("read pinned: %w", err)
	}

	var tf tomlFormat
	if err := toml.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("parse pinned: %w", err)
	}

	for _, p := range tf.Pins {
		s.pins[p.Resource+":"+p.ID] = p
	}

	return s, nil
}

// Pin adds a bookmark.
func (s *Store) Pin(p Pin) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := p.Resource + ":" + p.ID
	s.pins[key] = p
	return s.save()
}

// Unpin removes a bookmark.
func (s *Store) Unpin(resource, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.pins, resource+":"+id)
	return s.save()
}

// List returns all pins sorted by added time (newest first).
func (s *Store) List() []Pin {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pins := make([]Pin, 0, len(s.pins))
	for _, p := range s.pins {
		pins = append(pins, p)
	}

	// Sort descending by added time
	for i := 0; i < len(pins); i++ {
		for j := i + 1; j < len(pins); j++ {
			if pins[i].Added.Before(pins[j].Added) {
				pins[i], pins[j] = pins[j], pins[i]
			}
		}
	}

	return pins
}

// Has checks if a resource is pinned.
func (s *Store) Has(resource, id string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.pins[resource+":"+id]
	return ok
}

func (s *Store) save() error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create pinned dir: %w", err)
	}

	pins := make([]Pin, 0, len(s.pins))
	for _, p := range s.pins {
		pins = append(pins, p)
	}

	tf := tomlFormat{Pins: pins}
	f, err := os.Create(s.path)
	if err != nil {
		return fmt.Errorf("create pinned file: %w", err)
	}
	defer f.Close()

	if err := toml.NewEncoder(f).Encode(tf); err != nil {
		return fmt.Errorf("write pinned: %w", err)
	}

	return nil
}
