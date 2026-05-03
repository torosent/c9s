// Package errors provides the :errors screen for viewing logged errors.
package errors

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/log"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/skinx"
	"github.com/torosent/c9s/internal/ui/theme"
)

// Model represents the errors screen.
type Model struct {
	logDir      string
	clock       clock.Clock
	palette     theme.Palette
	table       table.Model
	entries     []log.Entry
	keymap      *keymap.Map
	width       int
	height      int
	sortKey     string
	sortReverse bool
}

// New creates a new errors screen.
func New(logDir string, clk clock.Clock, p theme.Palette) *Model {
	km := keymap.Default()
	km.Add("inspect", keymap.Binding{
		Keys:        []string{"enter"},
		Help:        "Details",
		Description: "View error details",
	})
	km.Add("copy", keymap.Binding{
		Keys:        []string{"y"},
		Help:        "Copy",
		Description: "Copy as markdown",
	})

	columns := []table.Column{
		{Title: "TIME", Width: 19},
		{Title: "OP", Width: 20},
		{Title: "RESOURCE", Width: 25},
		{Title: "MESSAGE", Width: 50},
	}

	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)

	s := skinx.TableStyles(p)
	tbl.SetStyles(s)

	return &Model{
		logDir:  logDir,
		clock:   clk,
		palette: p,
		table:   tbl,
		keymap:  km,
	}
}

// Init implements screens.Screen.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadEntries(),
		m.tickRefresh(),
	)
}

// LoadEntriesMsg triggers a refresh of error entries from disk.
type LoadEntriesMsg struct{}

func (m *Model) loadEntries() tea.Cmd {
	return func() tea.Msg {
		return LoadEntriesMsg{}
	}
}

func (m *Model) tickRefresh() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return LoadEntriesMsg{}
	})
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
				if row >= 0 && row < len(m.entries) {
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
		case m.keymap.Matches("inspect", msg):
			row := m.table.SelectedRow()
			if len(row) > 0 {
				// Find corresponding entry
				idx := m.table.Cursor()
				if idx < len(m.entries) {
					entry := m.entries[idx]
					// Build detail JSON
					data, _ := json.MarshalIndent(entry, "", "  ")
					modal := modals.NewInspect("Error Detail", data, m.palette)
					return m, func() tea.Msg {
						return screens.OpenModalMsg{Modal: modal}
					}
				}
			}

		case m.keymap.Matches("copy", msg):
			row := m.table.SelectedRow()
			if len(row) > 0 {
				idx := m.table.Cursor()
				if idx < len(m.entries) {
					entry := m.entries[idx]
					md := fmt.Sprintf("```\nTime: %s\nOp: %s\nResource: %s\nMessage: %s\nDetail: %s\n```",
						entry.Time.Format(time.RFC3339),
						entry.Op,
						entry.Resource,
						entry.Message,
						entry.Detail)
					if err := clipboard.WriteAll(md); err == nil {
						return m, func() tea.Msg {
							return screens.StatusMsg{Toast: "Copied to clipboard"}
						}
					}
				}
			}
		}

	case LoadEntriesMsg:
		m.refreshEntries()
		cmds = append(cmds, m.tickRefresh())

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetWidth(msg.Width)
		m.table.SetHeight(msg.Height - 5)
	}

	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// refreshEntries reads log files and updates the table.
func (m *Model) refreshEntries() {
	// Read today's file
	today := m.clock.Now().Format("2006-01-02")
	path := filepath.Join(m.logDir, fmt.Sprintf("errors-%s.log", today))

	entries := []log.Entry{}

	f, err := os.Open(path)
	if err != nil {
		// No log file yet is OK
		m.entries = entries
		m.updateTable()
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var entry log.Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	// Sort by time descending (newest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Time.After(entries[j].Time)
	})

	m.entries = entries
	m.updateTable()
}

func (m *Model) updateTable() {
	m.sortEntries()
	rows := []table.Row{}
	for _, e := range m.entries {
		timeStr := e.Time.Format("15:04:05")
		msg := e.Message
		if len(msg) > 50 {
			msg = msg[:47] + "..."
		}
		rows = append(rows, table.Row{
			timeStr,
			truncate(e.Op, 20),
			truncate(e.Resource, 25),
			msg,
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
	return skinx.BorderedBox(m.palette, "Errors", "all", len(m.entries), width, height, m.table.View())
}

// Title implements screens.Screen.
func (m *Model) Title() string {
	return "errors"
}

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map {
	return m.keymap
}

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	count := len(m.entries)
	if count == 0 {
		return "no errors"
	}
	if count == 1 {
		return "1 error"
	}
	return fmt.Sprintf("%d errors", count)
}

// SortableColumns implements screens.Sortable.
func (m *Model) SortableColumns() []modals.SortColumn {
	return []modals.SortColumn{
		{Key: "time", Label: "Time"},
		{Key: "op", Label: "Operation"},
		{Key: "resource", Label: "Resource"},
	}
}

// ApplySort implements screens.Sortable.
func (m *Model) ApplySort(key string, reverse bool) {
	m.sortKey = key
	m.sortReverse = reverse
	m.sortEntries()
	m.updateTable()
}

// sortEntries sorts the entries slice in place.
func (m *Model) sortEntries() {
	if m.sortKey == "" {
		return
	}

	sort.Slice(m.entries, func(i, j int) bool {
		less := false
		switch m.sortKey {
		case "time":
			less = m.entries[i].Time.Before(m.entries[j].Time)
		case "op":
			less = m.entries[i].Op < m.entries[j].Op
		case "resource":
			less = m.entries[i].Resource < m.entries[j].Resource
		default:
			return false
		}
		if m.sortReverse {
			return !less
		}
		return less
	})
}
