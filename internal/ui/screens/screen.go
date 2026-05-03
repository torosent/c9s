package screens

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/theme"
)

// PaletteChangedMsg is broadcast to every registered screen when the user
// switches skins. Each screen handles it by updating its internal palette
// (and restyling any cached widgets that captured colors at construction
// time, e.g. a bubbles/table.Styles). Screens that don't care about the
// palette can ignore the message and Update will fall through to its
// default case.
type PaletteChangedMsg struct {
	P theme.Palette
}

// Screen represents a full-screen view in the TUI (e.g., containers, images).
// Each screen manages its own state, keybindings, and rendering.
type Screen interface {
	// Init returns the initial command to run when the screen is activated.
	Init() tea.Cmd

	// Update handles incoming messages and returns the updated screen and
	// any command to execute.
	Update(msg tea.Msg) (Screen, tea.Cmd)

	// View renders the screen content at the given dimensions.
	View(width, height int) string

	// Title returns the screen's name for display (e.g., "containers").
	Title() string

	// Hotkeys returns the screen's key binding map for the help overlay.
	Hotkeys() *keymap.Map

	// Summary returns a short status line for the screen (e.g., "15 running").
	Summary() string
}

// Sortable is an optional interface for screens that support sorting.
type Sortable interface {
	// SortableColumns returns the list of columns available for sorting.
	SortableColumns() []modals.SortColumn

	// ApplySort applies the given sort key and direction to the screen's data.
	ApplySort(key string, reverse bool)
}
