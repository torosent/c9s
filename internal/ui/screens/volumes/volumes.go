// Package volumes implements the :volumes resource screen.
package volumes

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/skinx"
	"github.com/torosent/c9s/internal/ui/theme"
)

// Model represents the :volumes screen.
type Model struct {
	client      cli.Client
	clk         clock.Clock
	palette     theme.Palette
	tbl         table.Model
	volumes     []cli.Volume
	marks       map[string]bool
	filter      string
	filterMode  bool
	keymap      *keymap.Map
	width       int
	height      int
	sortKey     string
	sortReverse bool
}

// New creates a new Volumes screen model.
func New(client cli.Client, clk clock.Clock, p theme.Palette) *Model {
	km := keymap.Default()
	km.Add("inspect", keymap.Binding{
		Keys: []string{"d"}, Help: "Inspect", Description: "Inspect volume",
	})
	km.Add("delete", keymap.Binding{
		Keys: []string{"D", "shift+d"}, Help: "Delete", Description: "Delete volume",
	})

	columns := []table.Column{
		{Title: "NAME", Width: 24},
		{Title: "DRIVER", Width: 10},
		{Title: "MOUNTPOINT", Width: 36},
		{Title: "SIZE", Width: 10},
		{Title: "USED-BY", Width: 24},
	}
	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	tbl.SetStyles(skinx.TableStyles(p))

	return &Model{
		client:  client,
		clk:     clk,
		palette: p,
		tbl:     tbl,
		marks:   make(map[string]bool),
		keymap:  km,
	}
}

// Init implements screens.Screen.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		state.MakeRefreshedCmd[cli.Volume](
			context.Background(),
			func(ctx context.Context) ([]cli.Volume, error) {
				return m.client.ListVolumes(ctx)
			},
			cli.ResourceVolumes,
		),
		state.TickCmd(2*time.Second, m.clk, cli.ResourceVolumes),
	)
}

// Update implements screens.Screen.
func (m *Model) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case screens.PaletteChangedMsg:
		m.palette = msg.P
		m.tbl.SetStyles(skinx.TableStyles(msg.P))

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.height > 5 {
			m.tbl.SetHeight(m.height - 5)
		}

	case state.RefreshedMsg[cli.Volume]:
		if msg.Resource != cli.ResourceVolumes {
			break
		}
		m.volumes = msg.Snapshot.Items
		m.rebuildTable()

	case state.TickMsg:
		if msg.Resource != cli.ResourceVolumes {
			break
		}
		cmds = append(cmds,
			state.MakeRefreshedCmd[cli.Volume](
				context.Background(),
				func(ctx context.Context) ([]cli.Volume, error) {
					return m.client.ListVolumes(ctx)
				},
				cli.ResourceVolumes,
			),
			state.TickCmd(2*time.Second, m.clk, cli.ResourceVolumes),
		)

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Y >= 3 {
				row := msg.Y - 3
				visible := m.visibleVolumes()
				if row >= 0 && row < len(visible) {
					m.tbl.SetCursor(row)
				}
			}
		case tea.MouseButtonWheelUp:
			m.tbl.MoveUp(1)
		case tea.MouseButtonWheelDown:
			m.tbl.MoveDown(1)
		}

	case modals.ConfirmResultMsg:
		if msg.Result.Tag == "delete-volumes" && msg.Result.Confirmed {
			cmds = append(cmds, m.performDelete())
		}

	case tea.KeyMsg:
		if m.filterMode {
			return m.handleFilterKey(msg)
		}
		if m.keymap.Matches("escape", msg) {
			m.marks = make(map[string]bool)
			return m, nil
		}
		if m.keymap.Matches("mark", msg) {
			m.toggleMark()
			return m, nil
		}
		if m.keymap.Matches("mark_all", msg) {
			m.selectAll()
			return m, nil
		}
		if m.keymap.Matches("refresh", msg) {
			return m, state.MakeRefreshedCmd[cli.Volume](
				context.Background(),
				func(ctx context.Context) ([]cli.Volume, error) {
					return m.client.ListVolumes(ctx)
				},
				cli.ResourceVolumes,
			)
		}
		if m.keymap.Matches("filter", msg) {
			m.filterMode = true
			m.filter = ""
			return m, nil
		}
		if m.keymap.Matches("inspect", msg) {
			return m, m.inspectFocused()
		}
		if m.keymap.Matches("delete", msg) {
			return m, m.requestDelete()
		}

		var cmd tea.Cmd
		m.tbl, cmd = m.tbl.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements screens.Screen.
func (m *Model) View(width, height int) string {
	body := m.tbl.View()
	if m.filterMode {
		body = m.tbl.View() + "\n" + fmt.Sprintf("Filter: %s_", m.filter)
	}
	filter := "all"
	if m.filter != "" {
		filter = m.filter
	}
	return skinx.BorderedBox(m.palette, "Volumes", filter, len(m.volumes), width, height, body)
}

// Title implements screens.Screen.
func (m *Model) Title() string { return "Volumes" }

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	total := len(m.volumes)
	marked := len(m.marks)
	if marked > 0 {
		return fmt.Sprintf("%d volumes · %d selected", total, marked)
	}
	return fmt.Sprintf("%d volumes", total)
}

func (m *Model) rebuildTable() {
	m.sortVolumes()
	visible := m.visibleVolumes()
	rows := make([]table.Row, 0, len(visible))
	for _, v := range visible {
		used := strings.Join(v.UsedBy, ",")
		if used == "" {
			used = "-"
		}
		rows = append(rows, table.Row{
			v.Name,
			v.Driver,
			v.Mountpoint,
			formatBytes(v.SizeBytes),
			used,
		})
	}
	m.tbl.SetRows(rows)
}

func (m *Model) visibleVolumes() []cli.Volume {
	if m.filter == "" {
		return m.volumes
	}
	needle := strings.ToLower(m.filter)
	visible := make([]cli.Volume, 0, len(m.volumes))
	for _, v := range m.volumes {
		hay := strings.ToLower(v.Name + " " + v.Driver + " " + v.Mountpoint)
		if strings.Contains(hay, needle) {
			visible = append(visible, v)
		}
	}
	return visible
}

func (m *Model) toggleMark() {
	v := m.focused()
	if v == nil {
		return
	}
	if m.marks[v.Name] {
		delete(m.marks, v.Name)
		return
	}
	m.marks[v.Name] = true
}

func (m *Model) selectAll() {
	for _, v := range m.visibleVolumes() {
		m.marks[v.Name] = true
	}
}

func (m *Model) focused() *cli.Volume {
	idx := m.tbl.Cursor()
	visible := m.visibleVolumes()
	if idx < 0 || idx >= len(visible) {
		return nil
	}
	return &visible[idx]
}

func (m *Model) targetVolumes() []cli.Volume {
	if len(m.marks) == 0 {
		if v := m.focused(); v != nil {
			return []cli.Volume{*v}
		}
		return nil
	}
	out := make([]cli.Volume, 0, len(m.marks))
	for _, v := range m.visibleVolumes() {
		if m.marks[v.Name] {
			out = append(out, v)
		}
	}
	return out
}

func (m *Model) inspectFocused() tea.Cmd {
	v := m.focused()
	if v == nil {
		return nil
	}
	name := v.Name
	body, _ := jsonOf(v)
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewInspect(
			fmt.Sprintf("Volume %s", name), body, m.palette,
		)}
	}
}

func (m *Model) requestDelete() tea.Cmd {
	vols := m.targetVolumes()
	if len(vols) == 0 {
		return nil
	}
	lines := make([]string, len(vols))
	for i, v := range vols {
		lines[i] = v.Name
	}
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewConfirm(
			"Delete volumes",
			"This will permanently remove:",
			lines,
			"delete-volumes",
			m.palette,
		)}
	}
}

func (m *Model) performDelete() tea.Cmd {
	vols := m.targetVolumes()
	if len(vols) == 0 {
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(vols))
	for _, v := range vols {
		name := v.Name
		cmds = append(cmds, func() tea.Msg {
			err := m.client.DeleteVolume(context.Background(), name)
			if err != nil {
				return screens.StatusMsg{Toast: fmt.Sprintf("delete %s failed: %v", name, err)}
			}
			return screens.StatusMsg{Toast: fmt.Sprintf("deleted %s", name)}
		})
	}
	m.marks = make(map[string]bool)
	return tea.Batch(cmds...)
}

func (m *Model) handleFilterKey(msg tea.KeyMsg) (screens.Screen, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		m.filterMode = false
		m.rebuildTable()
		return m, nil
	case tea.KeyEsc:
		m.filterMode = false
		m.filter = ""
		m.rebuildTable()
		return m, nil
	case tea.KeyBackspace:
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.rebuildTable()
		}
		return m, nil
	case tea.KeyRunes:
		m.filter += string(msg.Runes)
		m.rebuildTable()
		return m, nil
	}
	return m, nil
}

// SortableColumns implements screens.Sortable.
func (m *Model) SortableColumns() []modals.SortColumn {
	return []modals.SortColumn{
		{Key: "name", Label: "Name"},
		{Key: "driver", Label: "Driver"},
		{Key: "size", Label: "Size"},
		{Key: "used_by_count", Label: "Used By Count"},
	}
}

// ApplySort implements screens.Sortable.
func (m *Model) ApplySort(key string, reverse bool) {
	m.sortKey = key
	m.sortReverse = reverse
	m.sortVolumes()
	m.rebuildTable()
}

// sortVolumes sorts the volumes slice in place.
func (m *Model) sortVolumes() {
	if m.sortKey == "" {
		return
	}

	sort.Slice(m.volumes, func(i, j int) bool {
		less := false
		switch m.sortKey {
		case "name":
			less = m.volumes[i].Name < m.volumes[j].Name
		case "driver":
			less = m.volumes[i].Driver < m.volumes[j].Driver
		case "size":
			less = m.volumes[i].SizeBytes < m.volumes[j].SizeBytes
		case "used_by_count":
			less = len(m.volumes[i].UsedBy) < len(m.volumes[j].UsedBy)
		default:
			return false
		}
		if m.sortReverse {
			return !less
		}
		return less
	})
}
