// Package networks implements the :networks resource screen.
package networks

import (
	"context"
	"encoding/json"
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

// Model represents the :networks screen.
type Model struct {
	client      cli.Client
	clk         clock.Clock
	palette     theme.Palette
	tbl         table.Model
	networks    []cli.Network
	marks       map[string]bool
	filter      string
	filterMode  bool
	keymap      *keymap.Map
	width       int
	height      int
	sortKey     string
	sortReverse bool
}

// New creates a new Networks screen.
func New(client cli.Client, clk clock.Clock, p theme.Palette) *Model {
	km := keymap.Default()
	km.Add("inspect", keymap.Binding{
		Keys: []string{"d"}, Help: "Inspect", Description: "Inspect network",
	})
	km.Add("delete", keymap.Binding{
		Keys: []string{"D", "shift+d"}, Help: "Delete", Description: "Delete network",
	})

	columns := []table.Column{
		{Title: "NAME", Width: 24},
		{Title: "DRIVER", Width: 10},
		{Title: "SUBNET", Width: 18},
		{Title: "CONTAINERS", Width: 36},
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
		state.MakeRefreshedCmd[cli.Network](
			cli.DefaultCtx(),
			func(ctx context.Context) ([]cli.Network, error) {
				return m.client.ListNetworks(ctx)
			},
			cli.ResourceNetworks,
		),
		state.TickCmd(2*time.Second, m.clk, cli.ResourceNetworks),
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

	case state.RefreshedMsg[cli.Network]:
		if msg.Resource != cli.ResourceNetworks {
			break
		}
		m.networks = msg.Snapshot.Items
		m.rebuildTable()

	case state.TickMsg:
		if msg.Resource != cli.ResourceNetworks {
			break
		}
		cmds = append(cmds,
			state.MakeRefreshedCmd[cli.Network](
				cli.DefaultCtx(),
				func(ctx context.Context) ([]cli.Network, error) {
					return m.client.ListNetworks(ctx)
				},
				cli.ResourceNetworks,
			),
			state.TickCmd(2*time.Second, m.clk, cli.ResourceNetworks),
		)

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Y >= 3 {
				row := msg.Y - 3
				visible := m.visible()
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
		if msg.Result.Tag == "delete-networks" && msg.Result.Confirmed {
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
			return m, state.MakeRefreshedCmd[cli.Network](
				cli.DefaultCtx(),
				func(ctx context.Context) ([]cli.Network, error) {
					return m.client.ListNetworks(ctx)
				},
				cli.ResourceNetworks,
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
	return skinx.BorderedBox(m.palette, "Networks", filter, len(m.networks), width, height, body)
}

// Title implements screens.Screen.
func (m *Model) Title() string { return "Networks" }

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	total := len(m.networks)
	marked := len(m.marks)
	if marked > 0 {
		return fmt.Sprintf("%d networks · %d selected", total, marked)
	}
	return fmt.Sprintf("%d networks", total)
}

func (m *Model) rebuildTable() {
	m.sortNetworks()
	visible := m.visible()
	rows := make([]table.Row, 0, len(visible))
	for _, n := range visible {
		ctrs := strings.Join(n.Containers, ",")
		if ctrs == "" {
			ctrs = "-"
		}
		subnet := n.Subnet
		if subnet == "" {
			subnet = "-"
		}
		rows = append(rows, table.Row{
			n.Name,
			n.Driver,
			subnet,
			ctrs,
		})
	}
	m.tbl.SetRows(rows)
}

func (m *Model) visible() []cli.Network {
	if m.filter == "" {
		return m.networks
	}
	needle := strings.ToLower(m.filter)
	visible := make([]cli.Network, 0, len(m.networks))
	for _, n := range m.networks {
		hay := strings.ToLower(n.Name + " " + n.Driver + " " + n.Subnet)
		if strings.Contains(hay, needle) {
			visible = append(visible, n)
		}
	}
	return visible
}

func (m *Model) toggleMark() {
	n := m.focused()
	if n == nil {
		return
	}
	if m.marks[n.Name] {
		delete(m.marks, n.Name)
		return
	}
	m.marks[n.Name] = true
}

func (m *Model) selectAll() {
	for _, n := range m.visible() {
		m.marks[n.Name] = true
	}
}

func (m *Model) focused() *cli.Network {
	idx := m.tbl.Cursor()
	visible := m.visible()
	if idx < 0 || idx >= len(visible) {
		return nil
	}
	return &visible[idx]
}

func (m *Model) targets() []cli.Network {
	if len(m.marks) == 0 {
		if n := m.focused(); n != nil {
			return []cli.Network{*n}
		}
		return nil
	}
	out := make([]cli.Network, 0, len(m.marks))
	for _, n := range m.visible() {
		if m.marks[n.Name] {
			out = append(out, n)
		}
	}
	return out
}

func (m *Model) inspectFocused() tea.Cmd {
	n := m.focused()
	if n == nil {
		return nil
	}
	name := n.Name
	body, _ := json.MarshalIndent(n, "", "  ")
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewInspect(
			fmt.Sprintf("Network %s", name), body, m.palette,
		)}
	}
}

func (m *Model) requestDelete() tea.Cmd {
	ns := m.targets()
	if len(ns) == 0 {
		return nil
	}
	lines := make([]string, len(ns))
	for i, n := range ns {
		lines[i] = n.Name
	}
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewConfirm(
			"Delete networks",
			"This will permanently remove:",
			lines,
			"delete-networks",
			m.palette,
		)}
	}
}

func (m *Model) performDelete() tea.Cmd {
	ns := m.targets()
	if len(ns) == 0 {
		return nil
	}
	cmds := make([]tea.Cmd, 0, len(ns))
	for _, n := range ns {
		name := n.Name
		cmds = append(cmds, func() tea.Msg {
			err := m.client.DeleteNetwork(cli.DefaultCtx(), name)
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
		{Key: "subnet", Label: "Subnet"},
		{Key: "container_count", Label: "Container Count"},
	}
}

// ApplySort implements screens.Sortable.
func (m *Model) ApplySort(key string, reverse bool) {
	m.sortKey = key
	m.sortReverse = reverse
	m.sortNetworks()
	m.rebuildTable()
}

// sortNetworks sorts the networks slice in place.
func (m *Model) sortNetworks() {
	if m.sortKey == "" {
		return
	}

	sort.Slice(m.networks, func(i, j int) bool {
		less := false
		switch m.sortKey {
		case "name":
			less = m.networks[i].Name < m.networks[j].Name
		case "driver":
			less = m.networks[i].Driver < m.networks[j].Driver
		case "subnet":
			less = m.networks[i].Subnet < m.networks[j].Subnet
		case "container_count":
			less = len(m.networks[i].Containers) < len(m.networks[j].Containers)
		default:
			return false
		}
		if m.sortReverse {
			return !less
		}
		return less
	})
}
