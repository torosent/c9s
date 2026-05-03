// Package palette provides command history for the palette.
package palette

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// History manages command history.
type History struct {
	path    string
	entries []string
	idx     int
	mu      sync.Mutex
}

// Load loads command history from disk.
func Load(path string) (*History, error) {
	h := &History{
		path:    path,
		entries: []string{},
		idx:     -1,
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return h, nil
		}
		return nil, fmt.Errorf("open history: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			h.entries = append(h.entries, line)
		}
	}

	h.idx = len(h.entries)
	return h, nil
}

// Append adds a command to history.
func (h *History) Append(cmd string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if cmd == "" {
		return nil
	}

	// Dedup consecutive
	if len(h.entries) > 0 && h.entries[len(h.entries)-1] == cmd {
		return nil
	}

	h.entries = append(h.entries, cmd)

	// Cap at 1000
	if len(h.entries) > 1000 {
		h.entries = h.entries[len(h.entries)-1000:]
	}

	h.idx = len(h.entries)

	return h.save()
}

// Up navigates backward in history.
func (h *History) Up() string {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.entries) == 0 {
		return ""
	}

	if h.idx > 0 {
		h.idx--
	}

	if h.idx < len(h.entries) {
		return h.entries[h.idx]
	}

	return ""
}

// Down navigates forward in history.
func (h *History) Down() string {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.entries) == 0 {
		return ""
	}

	if h.idx < len(h.entries)-1 {
		h.idx++
		return h.entries[h.idx]
	}

	h.idx = len(h.entries)
	return ""
}

// Reset resets the navigation index.
func (h *History) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.idx = len(h.entries)
}

func (h *History) save() error {
	dir := filepath.Dir(h.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create history dir: %w", err)
	}

	f, err := os.Create(h.path)
	if err != nil {
		return fmt.Errorf("create history file: %w", err)
	}
	defer f.Close()

	for _, entry := range h.entries {
		if _, err := f.WriteString(entry + "\n"); err != nil {
			return fmt.Errorf("write history: %w", err)
		}
	}

	return nil
}
