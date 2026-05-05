package keymap

import (
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
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
// In v2, tea.KeyMsg is an interface; we only match key presses (not
// releases), so we type-assert to tea.KeyPressMsg.
func (m *Map) Matches(name string, msg tea.KeyMsg) bool {
	b, ok := m.Get(name)
	if !ok {
		return false
	}
	press, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return false
	}

	for _, keyStr := range b.Keys {
		if matchesKey(keyStr, press) {
			return true
		}
	}
	return false
}

// matchesKey checks if a key string matches a KeyPressMsg using v2's
// standardized String() format. v2's String() returns keys like "esc",
// "enter", "space", "ctrl+c", "shift+P", "a", etc., which directly
// matches the Keys[] entries in our Binding definitions.
func matchesKey(keyStr string, msg tea.KeyPressMsg) bool {
	got := msg.String()
	want := keyStr

	// Direct match (covers "esc", "enter", "space", "ctrl+c", etc.)
	if got == want {
		return true
	}

	// Tolerate case differences for special-key aliases.
	if strings.EqualFold(got, want) {
		return true
	}

	// Common aliases.
	if (want == "escape" && got == "esc") ||
		(want == "return" && got == "enter") {
		return true
	}

	// "shift+x" where x is lowercase: v2's String() reports "shift+X"
	// (with the shifted character). Handle that.
	if strings.HasPrefix(strings.ToLower(want), "shift+") {
		suffix := want[len("shift+"):]
		if len(suffix) == 1 {
			// got might be "shift+X" or just "X"
			if got == "shift+"+strings.ToUpper(suffix) {
				return true
			}
			if got == strings.ToUpper(suffix) {
				return true
			}
		}
	}

	// Single uppercase letter — v2's String() returns just the
	// shifted character (e.g. "P"); old code treated this as "shift+p".
	if len(keyStr) == 1 {
		c := keyStr[0]
		if c >= 'A' && c <= 'Z' && got == string(c) {
			return true
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

	m.Add("sort", Binding{
		Keys:        []string{"shift+s", "S"},
		Help:        "sort",
		Description: "Open the column sort picker (sortable screens)",
	})

	m.Add("screen_switch", Binding{
		Keys:        []string{"1-9", "0"},
		Help:        "switch screen",
		Description: "Quick-switch (1 containers · 2 images · 3 volumes · 4 networks · 5 builder · 6 registry · 7 system · 8 pulses · 9 xray · 0 pinned)",
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
