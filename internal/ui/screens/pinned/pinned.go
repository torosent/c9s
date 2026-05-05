// Package pinned provides the :pinned screen for viewing bookmarked resources.
package pinned

import (
	"fmt"
	"sort"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/pinned"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/skinx"
	"github.com/torosent/c9s/internal/ui/theme"
)

// Model represents the pinned screen.
type Model struct {
	store       *pinned.Store
	palette     theme.Palette
	table       table.Model
	pins        []pinned.Pin
	keymap      *keymap.Map
	width       int
	height      int
	sortKey     string
	sortReverse bool
}

// New creates a new pinned screen.
func New(store *pinned.Store, p theme.Palette) *Model {
	km := keymap.Default()
	km.Add("jump", keymap.Binding{
		Keys:        []string{"enter"},
		Help:        "Jump",
		Description: "Jump to resource screen",
	})
	km.Add("unpin", keymap.Binding{
		Keys:        []string{"shift+d", "D"},
		Help:        "Unpin",
		Description: "Remove bookmark",
	})

	columns := []table.Column{
		{Title: "RESOURCE", Width: 15},
		{Title: "ID", Width: 20},
		{Title: "DISPLAY", Width: 30},
		{Title: "ADDED", Width: 19},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := skinx.TableStyles(p)
	tbl.SetStyles(s)

	return &Model{
		store:   store,
		palette: p,
		table:   tbl,
		keymap:  km,
	}
}

// Init implements screens.Screen.
func (m *Model) Init() tea.Cmd {
	m.refresh()
	return nil
}

// Update implements screens.Screen.
func (m *Model) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Y >= 3 {
				row := msg.Y - 3
				if row >= 0 && row < len(m.pins) {
					m.table.SetCursor(row)
				}
			}
		case tea.MouseButtonWheelUp:
			m.table.MoveUp(1)
		case tea.MouseButtonWheelDown:
			m.table.MoveDown(1)
		}

	case tea.KeyMsg:
		switch {
		case m.keymap.Matches("jump", msg):
			row := m.table.SelectedRow()
			if len(row) > 0 {
				idx := m.table.Cursor()
				if idx < len(m.pins) {
					p := m.pins[idx]
					// Emit message to switch to resource screen
					return m, func() tea.Msg {
						return JumpToPinMsg{Resource: p.Resource, ID: p.ID}
					}
				}
			}

		case m.keymap.Matches("unpin", msg):
			row := m.table.SelectedRow()
			if len(row) > 0 {
				idx := m.table.Cursor()
				if idx < len(m.pins) {
					p := m.pins[idx]
					if err := m.store.Unpin(p.Resource, p.ID); err == nil {
						m.refresh()
						return m, func() tea.Msg {
							return screens.StatusMsg{Toast: "Unpinned"}
						}
					}
				}
			}
		}

	case screens.PaletteChangedMsg:
		m.palette = msg.P
		m.table.SetStyles(skinx.TableStyles(msg.P))

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 5)

	case RefreshPinsMsg:
		m.refresh()
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// RefreshPinsMsg triggers a refresh of the pin list.
type RefreshPinsMsg struct{}

// JumpToPinMsg requests switching to a pinned resource's screen.
type JumpToPinMsg struct {
	Resource string
	ID       string
}

func (m *Model) refresh() {
	m.pins = m.store.List()
	m.updateTable()
}

func (m *Model) updateTable() {
	m.sortPins()
	rows := []table.Row{}
	for _, p := range m.pins {
		addedStr := p.Added.Format(time.RFC3339)
		rows = append(rows, table.Row{
			p.Resource,
			truncate(p.ID, 20),
			truncate(p.Display, 30),
			addedStr,
		})
	}
	m.table.SetRows(rows)
}

func truncate(s string, max int) string {
	if len(s) > max {
		return s[:max-3] + "..."
	}
	return s
}

// View implements screens.Screen.
func (m *Model) View(width, height int) string {
	if width != m.width || height != m.height {
		m.width = width
		m.height = height
		m.table.SetWidth(width - 4)
		m.table.SetHeight(height - 4)
	}
	return skinx.BorderedBox(m.palette, "Pinned", "all", len(m.pins), width, height, m.table.View())
}

// Title implements screens.Screen.
func (m *Model) Title() string {
	return "pinned"
}

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map {
	return m.keymap
}

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	count := len(m.pins)
	if count == 0 {
		return "no pins"
	}
	if count == 1 {
		return "1 pin"
	}
	return fmt.Sprintf("%d pins", count)
}

// SortableColumns implements screens.Sortable.
func (m *Model) SortableColumns() []modals.SortColumn {
	return []modals.SortColumn{
		{Key: "resource", Label: "Resource"},
		{Key: "id", Label: "ID"},
		{Key: "added", Label: "Added"},
	}
}

// ApplySort implements screens.Sortable.
func (m *Model) ApplySort(key string, reverse bool) {
	m.sortKey = key
	m.sortReverse = reverse
	m.sortPins()
	m.updateTable()
}

// sortPins sorts the pins slice in place.
func (m *Model) sortPins() {
	if m.sortKey == "" {
		return
	}

	sort.Slice(m.pins, func(i, j int) bool {
		less := false
		switch m.sortKey {
		case "resource":
			less = m.pins[i].Resource < m.pins[j].Resource
		case "id":
			less = m.pins[i].ID < m.pins[j].ID
		case "added":
			less = m.pins[i].Added.Before(m.pins[j].Added)
		default:
			return false
		}
		if m.sortReverse {
			return !less
		}
		return less
	})
}
