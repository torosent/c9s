package keymap

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Binding represents a keyboard shortcut with its documentation.
type Binding struct {
	Keys        []string // e.g. ["q"], ["shift+k"], ["ctrl+c"]
	Help        string   // shown in `?` overlay
	Description string   // longer form for docs
}

// Map holds a collection of named key bindings.
type Map struct {
	entries map[string]Binding
}

// Add inserts or replaces a binding.
func (m *Map) Add(name string, b Binding) {
	if m.entries == nil {
		m.entries = make(map[string]Binding)
	}
	m.entries[name] = b
}

// Get retrieves a binding by name.
func (m *Map) Get(name string) (Binding, bool) {
	b, ok := m.entries[name]
	return b, ok
}

// Names returns all binding names in sorted order.
func (m *Map) Names() []string {
	names := make([]string, 0, len(m.entries))
	for name := range m.entries {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Matches checks if the given key message matches the named binding.
func (m *Map) Matches(name string, msg tea.KeyMsg) bool {
	b, ok := m.Get(name)
	if !ok {
		return false
	}

	for _, keyStr := range b.Keys {
		if matchesKey(keyStr, msg) {
			return true
		}
	}
	return false
}

// matchesKey checks if a key string matches a KeyMsg.
func matchesKey(keyStr string, msg tea.KeyMsg) bool {
	// Don't lowercase yet - we need to preserve case for uppercase letters

	// Handle special keys (case-insensitive)
	lowerKey := strings.ToLower(keyStr)
	switch lowerKey {
	case "esc", "escape":
		return msg.Type == tea.KeyEsc
	case "enter", "return":
		return msg.Type == tea.KeyEnter
	case "space":
		return msg.Type == tea.KeySpace
	case "backspace":
		return msg.Type == tea.KeyBackspace
	case "tab":
		return msg.Type == tea.KeyTab
	case "up":
		return msg.Type == tea.KeyUp
	case "down":
		return msg.Type == tea.KeyDown
	case "left":
		return msg.Type == tea.KeyLeft
	case "right":
		return msg.Type == tea.KeyRight
	}

	// Handle ctrl+key
	if strings.HasPrefix(lowerKey, "ctrl+") {
		key := strings.TrimPrefix(lowerKey, "ctrl+")
		switch key {
		case "c":
			return msg.Type == tea.KeyCtrlC
		case "e":
			return msg.Type == tea.KeyCtrlE
		case "d":
			return msg.Type == tea.KeyCtrlD
		}
	}

	// Handle shift+key
	if strings.HasPrefix(lowerKey, "shift+") {
		key := strings.TrimPrefix(lowerKey, "shift+")
		if len(key) == 1 {
			// For single character, match the uppercase rune
			upperRune := []rune(strings.ToUpper(key))[0]
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == upperRune {
				return true
			}
		}
	}

	// Handle uppercase single character (treated as shift+key)
	if len(keyStr) == 1 {
		char := keyStr[0]
		if char >= 'A' && char <= 'Z' {
			// Uppercase letter - match exactly
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == rune(char) {
				return true
			}
		} else if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
			// Lowercase or other - match exactly
			return msg.Runes[0] == rune(char)
		}
	}

	return false
}

// Default returns a Map with global key bindings.
func Default() *Map {
	m := &Map{entries: make(map[string]Binding)}

	m.Add("quit", Binding{
		Keys:        []string{"q"},
		Help:        "quit",
		Description: "Quit the application",
	})

	m.Add("help", Binding{
		Keys:        []string{"?"},
		Help:        "help",
		Description: "Show help overlay",
	})

	m.Add("filter", Binding{
		Keys:        []string{"/"},
		Help:        "filter",
		Description: "Filter items",
	})

	m.Add("palette", Binding{
		Keys:        []string{":"},
		Help:        "palette",
		Description: "Open command palette",
	})

	m.Add("mark", Binding{
		Keys:        []string{" ", "space"},
		Help:        "mark",
		Description: "Mark/unmark item",
	})

	m.Add("mark_all", Binding{
		Keys:        []string{"*"},
		Help:        "mark all",
		Description: "Mark/unmark all items",
	})

	m.Add("refresh", Binding{
		Keys:        []string{"r"},
		Help:        "refresh",
		Description: "Refresh data",
	})

	m.Add("up", Binding{
		Keys:        []string{"up", "k"},
		Help:        "up",
		Description: "Move cursor up",
	})

	m.Add("down", Binding{
		Keys:        []string{"down", "j"},
		Help:        "down",
		Description: "Move cursor down",
	})

	m.Add("top", Binding{
		Keys:        []string{"g"},
		Help:        "top",
		Description: "Jump to top",
	})

	m.Add("bottom", Binding{
		Keys:        []string{"G", "shift+g"},
		Help:        "bottom",
		Description: "Jump to bottom",
	})

	m.Add("escape", Binding{
		Keys:        []string{"esc"},
		Help:        "escape",
		Description: "Cancel/go back",
	})

	m.Add("header_toggle", Binding{
		Keys:        []string{"ctrl+e"},
		Help:        "toggle header",
		Description: "Toggle header visibility",
	})

	m.Add("interrupt", Binding{
		Keys:        []string{"ctrl+c"},
		Help:        "interrupt",
		Description: "Interrupt operation",
	})

	return m
}

// Apply creates a new Map with the given key overrides applied.
// The overrides map keys are binding names (e.g., "quit", "filter")
// and values are the new key strings (e.g., "Q", "f").
// The original Map is not modified; a new Map is returned.
func Apply(m *Map, overrides map[string]string) *Map {
	// Create a new map with copied entries
	newMap := &Map{entries: make(map[string]Binding)}

	// Copy all existing bindings
	for name, binding := range m.entries {
		newMap.entries[name] = binding
	}

	// Apply overrides
	for name, newKey := range overrides {
		// Only override if the binding exists
		if binding, ok := newMap.entries[name]; ok {
			// Preserve Help and Description, only change Keys
			binding.Keys = []string{newKey}
			newMap.entries[name] = binding
		}
		// Silently ignore nonexistent bindings
	}

	return newMap
}
