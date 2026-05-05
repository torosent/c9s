package jobs

import (
	"fmt"
	"sort"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/jobs"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/skinx"
	"github.com/torosent/c9s/internal/ui/theme"
)

// Model is the jobs screen showing background streaming jobs.
type Model struct {
	manager     *jobs.Manager
	clock       clock.Clock
	palette     theme.Palette
	table       table.Model
	keymap      *keymap.Map
	width       int
	height      int
	sortKey     string
	sortReverse bool
	jobsList    []*jobs.Job
}

// New creates a new jobs screen.
func New(mgr *jobs.Manager, clk clock.Clock, p theme.Palette) *Model {
	// Build keymap with screen-specific bindings
	km := keymap.Default()
	km.Add("reattach", keymap.Binding{
		Keys:        []string{"enter"},
		Help:        "Re-attach",
		Description: "Re-attach to selected job",
	})
	km.Add("cancel", keymap.Binding{
		Keys:        []string{"ctrl+c"},
		Help:        "Cancel",
		Description: "Cancel selected job",
	})
	km.Add("clear-done", keymap.Binding{
		Keys:        []string{"shift+d", "D"},
		Help:        "Clear done",
		Description: "Clear completed jobs",
	})

	columns := []table.Column{
		{Title: "ID", Width: 6},
		{Title: "KIND", Width: 8},
		{Title: "TARGET", Width: 30},
		{Title: "STATE", Width: 12},
		{Title: "ELAPSED", Width: 10},
		{Title: "LINES", Width: 8},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)
	t.SetStyles(skinx.TableStyles(p))

	return &Model{
		manager: mgr,
		clock:   clk,
		palette: p,
		table:   t,
		keymap:  km,
	}
}

// Init implements screens.Screen.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshJobs(),
		m.tickRefresh(),
	)
}

type refreshJobsMsg struct{}

func (m *Model) refreshJobs() tea.Cmd {
	return func() tea.Msg {
		return refreshJobsMsg{}
	}
}

func (m *Model) tickRefresh() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return refreshJobsMsg{}
	})
}

// Update implements screens.Screen.
func (m *Model) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			cmds = append(cmds, m.reattachJob())
		case "ctrl+c":
			cmds = append(cmds, m.cancelJob())
		case "D":
			cmds = append(cmds, m.clearDone())
		}

	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft {
			// Compute row index (assuming table starts at Y=3 after title+header)
			if msg.Y >= 3 {
				row := msg.Y - 3
				if row >= 0 && row < len(m.jobsList) {
					m.table.SetCursor(row)
				}
			}
		}
	case tea.MouseWheelMsg:
		switch msg.Button {
		case tea.MouseWheelUp:
			m.table.MoveUp(1)
		case tea.MouseWheelDown:
			m.table.MoveDown(1)
		}

	case refreshJobsMsg:
		m.updateTable()
		cmds = append(cmds, m.tickRefresh())

	case screens.PaletteChangedMsg:
		m.palette = msg.P
		m.table.SetStyles(skinx.TableStyles(msg.P))

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

func (m *Model) updateTable() {
	m.jobsList = m.manager.List()
	m.sortJobs()

	rows := []table.Row{}
	for _, job := range m.jobsList {
		elapsed := job.Elapsed(m.clock).Round(time.Second).String()
		rows = append(rows, table.Row{
			job.ID,
			string(job.Kind),
			job.Target,
			string(job.State),
			elapsed,
			fmt.Sprintf("%d", len(job.Lines)),
		})
	}

	m.table.SetRows(rows)
}

func (m *Model) reattachJob() tea.Cmd {
	row := m.table.SelectedRow()
	if len(row) == 0 {
		return nil
	}
	jobID := row[0]

	job, ok := m.manager.Get(jobID)
	if !ok {
		return nil
	}

	// Emit OpenModalMsg to re-attach to this job
	// The modal will be created by app.go based on job.Kind
	return func() tea.Msg {
		return OpenModalMsg{JobID: jobID, Kind: job.Kind}
	}
}

func (m *Model) cancelJob() tea.Cmd {
	row := m.table.SelectedRow()
	if len(row) == 0 {
		return nil
	}
	jobID := row[0]

	_ = m.manager.Cancel(jobID)
	return m.refreshJobs()
}

func (m *Model) clearDone() tea.Cmd {
	// Remove done/failed/cancelled jobs from manager
	// For now, just refresh (manager doesn't have ClearDone yet)
	// TODO: implement Manager.ClearDone() if needed
	return m.refreshJobs()
}

// View implements screens.Screen.
func (m *Model) View(width, height int) string {
	m.table.SetWidth(width)
	if width > 0 && height > 0 {
		m.table.SetWidth(width - 4)
		m.table.SetHeight(height - 4)
	}
	return skinx.BorderedBox(m.palette, "Jobs", "all", len(m.jobsList), width, height, m.table.View())
}

// Title implements screens.Screen.
func (m *Model) Title() string {
	return "Jobs"
}

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map {
	return m.keymap
}

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	running := 0
	done := 0
	for _, job := range m.jobsList {
		if job.State == jobs.StateRunning {
			running++
		} else {
			done++
		}
	}
	return fmt.Sprintf("%d jobs · %d running, %d done", len(m.jobsList), running, done)
}

// OpenModalMsg signals that a modal should be opened for a job.
type OpenModalMsg struct {
	JobID string
	Kind  jobs.Kind
}

// SortableColumns implements screens.Sortable.
func (m *Model) SortableColumns() []modals.SortColumn {
	return []modals.SortColumn{
		{Key: "id", Label: "ID"},
		{Key: "kind", Label: "Kind"},
		{Key: "target", Label: "Target"},
		{Key: "state", Label: "State"},
		{Key: "started", Label: "Started"},
	}
}

// ApplySort implements screens.Sortable.
func (m *Model) ApplySort(key string, reverse bool) {
	m.sortKey = key
	m.sortReverse = reverse
	m.sortJobs()
	m.updateTable()
}

// sortJobs sorts the jobsList slice in place.
func (m *Model) sortJobs() {
	if m.sortKey == "" {
		return
	}

	sort.Slice(m.jobsList, func(i, j int) bool {
		less := false
		switch m.sortKey {
		case "id":
			less = m.jobsList[i].ID < m.jobsList[j].ID
		case "kind":
			less = string(m.jobsList[i].Kind) < string(m.jobsList[j].Kind)
		case "target":
			less = m.jobsList[i].Target < m.jobsList[j].Target
		case "state":
			less = string(m.jobsList[i].State) < string(m.jobsList[j].State)
		case "started":
			less = m.jobsList[i].Started.Before(m.jobsList[j].Started)
		default:
			return false
		}
		if m.sortReverse {
			return !less
		}
		return less
	})
}
