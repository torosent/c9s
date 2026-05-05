// Package registry implements the :registry resource screen.
package registry

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/skinx"
	"github.com/torosent/c9s/internal/ui/theme"
)

// Model represents the :registry screen.
type Model struct {
	client      cli.Client
	clk         clock.Clock
	palette     theme.Palette
	tbl         table.Model
	entries     []cli.RegistryEntry
	marks       map[string]bool
	filter      string
	filterMode  bool
	keymap      *keymap.Map
	width       int
	height      int
	sortKey     string
	sortReverse bool
}

// New creates a new Registry screen.
func New(client cli.Client, clk clock.Clock, p theme.Palette) *Model {
	km := keymap.Default()
	km.Add("logout", keymap.Binding{
		Keys: []string{"D", "shift+d"}, Help: "Logout", Description: "Log out of the focused registry",
	})
	km.Add("login", keymap.Binding{
		Keys: []string{"L", "shift+l"}, Help: "Login", Description: "Open the login modal",
	})
	km.Add("set_default", keymap.Binding{
		Keys: []string{"*"}, Help: "Default", Description: "Set focused registry as default",
	})

	columns := []table.Column{
		{Title: "HOST", Width: 32},
		{Title: "USER", Width: 18},
		{Title: "DEFAULT", Width: 8},
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
		m.refreshCmd(),
		state.TickCmd(2*time.Second, m.clk, cli.ResourceRegistry),
	)
}

func (m *Model) refreshCmd() tea.Cmd {
	return state.MakeRefreshedCmd[cli.RegistryEntry](
		cli.DefaultCtx(),
		func(ctx context.Context) ([]cli.RegistryEntry, error) {
			return m.client.ListRegistries(ctx)
		},
		cli.ResourceRegistry,
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

	case state.RefreshedMsg[cli.RegistryEntry]:
		if msg.Resource != cli.ResourceRegistry {
			break
		}
		m.entries = msg.Snapshot.Items
		m.rebuildTable()

	case state.TickMsg:
		if msg.Resource != cli.ResourceRegistry {
			break
		}
		cmds = append(cmds, m.refreshCmd(), state.TickCmd(2*time.Second, m.clk, cli.ResourceRegistry))

	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			if msg.Y >= 3 {
				row := msg.Y - 3
				visible := m.visibleEntries()
				if row >= 0 && row < len(visible) {
					m.tbl.SetCursor(row)
				}
			}
		case tea.MouseButtonWheelUp:
			m.tbl.MoveUp(1)
		case tea.MouseButtonWheelDown:
			m.tbl.MoveDown(1)
		}

	case modals.LoginResultMsg:
		cmds = append(cmds, m.performLogin(msg.Result))

	case modals.ConfirmResultMsg:
		if msg.Result.Tag == "registry-logout" && msg.Result.Confirmed {
			cmds = append(cmds, m.performLogout())
		}

	case tea.KeyMsg:
		if m.filterMode {
			return m.handleFilterKey(msg)
		}

		// Note: set_default is handled before the global mark_all binding to
		// avoid the `*` collision (mark_all is meaningless on registry rows).
		if m.keymap.Matches("set_default", msg) {
			return m, m.requestSetDefault()
		}
		if m.keymap.Matches("escape", msg) {
			m.marks = make(map[string]bool)
			return m, nil
		}
		if m.keymap.Matches("mark", msg) {
			m.toggleMark()
			return m, nil
		}
		if m.keymap.Matches("refresh", msg) {
			return m, m.refreshCmd()
		}
		if m.keymap.Matches("filter", msg) {
			m.filterMode = true
			m.filter = ""
			return m, nil
		}
		if m.keymap.Matches("login", msg) {
			return m, m.requestLogin()
		}
		if m.keymap.Matches("logout", msg) {
			return m, m.requestLogout()
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
	return skinx.BorderedBox(m.palette, "Registries", filter, len(m.entries), width, height, body)
}

// Title implements screens.Screen.
func (m *Model) Title() string { return "Registry" }

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	total := len(m.entries)
	def := ""
	for _, e := range m.entries {
		if e.Default {
			def = e.Host
			break
		}
	}
	if def != "" {
		return fmt.Sprintf("%d registries · default %s", total, def)
	}
	return fmt.Sprintf("%d registries", total)
}

func (m *Model) rebuildTable() {
	m.sortEntries()
	visible := m.visible()
	rows := make([]table.Row, 0, len(visible))
	for _, e := range visible {
		def := ""
		if e.Default {
			def = "*"
		}
		rows = append(rows, table.Row{e.Host, e.User, def})
	}
	m.tbl.SetRows(rows)
}

func (m *Model) visible() []cli.RegistryEntry {
	if m.filter == "" {
		return m.entries
	}
	needle := strings.ToLower(m.filter)
	visible := make([]cli.RegistryEntry, 0, len(m.entries))
	for _, e := range m.entries {
		hay := strings.ToLower(e.Host + " " + e.User)
		if strings.Contains(hay, needle) {
			visible = append(visible, e)
		}
	}
	return visible
}

func (m *Model) toggleMark() {
	e := m.focused()
	if e == nil {
		return
	}
	if m.marks[e.Host] {
		delete(m.marks, e.Host)
		return
	}
	m.marks[e.Host] = true
}

func (m *Model) focused() *cli.RegistryEntry {
	idx := m.tbl.Cursor()
	visible := m.visible()
	if idx < 0 || idx >= len(visible) {
		return nil
	}
	return &visible[idx]
}

func (m *Model) requestLogin() tea.Cmd {
	host := ""
	if e := m.focused(); e != nil {
		host = e.Host
	}
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewLogin(host, m.palette)}
	}
}

func (m *Model) performLogin(req modals.LoginRequest) tea.Cmd {
	return func() tea.Msg {
		err := m.client.RegistryLogin(cli.DefaultCtx(), req.Host, req.Username, req.Password)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("login %s failed: %v", req.Host, err)}
		}
		return screens.StatusMsg{Toast: fmt.Sprintf("logged in to %s", req.Host)}
	}
}

func (m *Model) requestLogout() tea.Cmd {
	e := m.focused()
	if e == nil {
		return nil
	}
	host := e.Host
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewConfirm(
			"Log out of registry",
			"You will need to re-authenticate to push or pull from:",
			[]string{host},
			"registry-logout",
			m.palette,
		)}
	}
}

func (m *Model) performLogout() tea.Cmd {
	e := m.focused()
	if e == nil {
		return nil
	}
	host := e.Host
	return func() tea.Msg {
		err := m.client.RegistryLogout(cli.DefaultCtx(), host)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("logout %s failed: %v", host, err)}
		}
		return screens.StatusMsg{Toast: fmt.Sprintf("logged out of %s", host)}
	}
}

func (m *Model) requestSetDefault() tea.Cmd {
	e := m.focused()
	if e == nil {
		return nil
	}
	host := e.Host
	return func() tea.Msg {
		err := m.client.RegistrySetDefault(cli.DefaultCtx(), host)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("set-default %s failed: %v", host, err)}
		}
		return screens.StatusMsg{Toast: fmt.Sprintf("default registry set to %s", host)}
	}
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
		{Key: "host", Label: "Host"},
		{Key: "user", Label: "User"},
		{Key: "default", Label: "Default"},
	}
}

// ApplySort implements screens.Sortable.
func (m *Model) ApplySort(key string, reverse bool) {
	m.sortKey = key
	m.sortReverse = reverse
	m.sortEntries()
	m.rebuildTable()
}

// sortEntries sorts the entries slice in place.
func (m *Model) sortEntries() {
	if m.sortKey == "" {
		return
	}

	sort.Slice(m.entries, func(i, j int) bool {
		less := false
		switch m.sortKey {
		case "host":
			less = m.entries[i].Host < m.entries[j].Host
		case "user":
			less = m.entries[i].User < m.entries[j].User
		case "default":
			less = m.entries[i].Default && !m.entries[j].Default
		default:
			return false
		}
		if m.sortReverse {
			return !less
		}
		return less
	})
}

// visibleEntries is an alias for visible() for mouse handler consistency.
func (m *Model) visibleEntries() []cli.RegistryEntry {
	return m.visible()
}
