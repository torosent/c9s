package system

import (
	"context"
	"fmt"
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

// DNSResource is the (logical) resource ID for the DNS sub-screen, reusing
// ResourceSystem so refresh tags do not collide with services.
const dnsResource cli.Resource = "system-dns"

// DNSModel is the :dns sub-screen.
type DNSModel struct {
	client     cli.Client
	clk        clock.Clock
	palette    theme.Palette
	tbl        table.Model
	domains    []cli.DNSDomain
	filter     string
	filterMode bool
	keymap     *keymap.Map
	width      int
	height     int
}

// NewDNS creates a new :dns sub-screen.
func NewDNS(client cli.Client, clk clock.Clock, p theme.Palette) DNSModel {
	km := keymap.Default()
	km.Add("create", keymap.Binding{
		Keys: []string{"c"}, Help: "Create", Description: "Create DNS domain",
	})
	km.Add("delete", keymap.Binding{
		Keys: []string{"D", "shift+d"}, Help: "Delete", Description: "Delete DNS domain",
	})
	km.Add("set_default", keymap.Binding{
		Keys: []string{"*"}, Help: "Default", Description: "Set focused domain as default",
	})

	columns := []table.Column{
		{Title: "NAME", Width: 32},
		{Title: "DEFAULT", Width: 8},
	}
	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	tbl.SetStyles(skinx.TableStyles(p))

	return DNSModel{
		client:  client,
		clk:     clk,
		palette: p,
		tbl:     tbl,
		keymap:  km,
	}
}

// Init implements screens.Screen.
func (m DNSModel) Init() tea.Cmd {
	return tea.Batch(
		m.refreshCmd(),
		state.TickCmd(2*time.Second, m.clk, dnsResource),
	)
}

func (m DNSModel) refreshCmd() tea.Cmd {
	return state.MakeRefreshedCmd[cli.DNSDomain](
		cli.DefaultCtx(),
		func(ctx context.Context) ([]cli.DNSDomain, error) {
			return m.client.ListDNSDomains(ctx)
		},
		dnsResource,
	)
}

// Update implements screens.Screen.
func (m DNSModel) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
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
	case state.RefreshedMsg[cli.DNSDomain]:
		if msg.Resource != dnsResource {
			break
		}
		m.domains = msg.Snapshot.Items
		m.rebuildTable()
	case modals.ConfirmResultMsg:
		if msg.Result.Tag == "delete-dns" && msg.Result.Confirmed {
			cmds = append(cmds, m.performDelete())
		}
	case modals.TextInputResultMsg:
		if msg.Result.Label == "create-dns" {
			cmds = append(cmds, m.performCreate(msg.Result.Value))
		}
	case tea.KeyMsg:
		if m.filterMode {
			return m.handleFilterKey(msg)
		}
		if m.keymap.Matches("set_default", msg) {
			return m, m.requestSetDefault()
		}
		if m.keymap.Matches("escape", msg) {
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
		if m.keymap.Matches("create", msg) {
			return m, m.requestCreate()
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
func (m DNSModel) View(width, height int) string {
	if m.filterMode {
		return m.tbl.View() + "\n" + fmt.Sprintf("Filter: %s_", m.filter)
	}
	return m.tbl.View()
}

// Title implements screens.Screen.
func (m DNSModel) Title() string { return "System DNS" }

// Hotkeys implements screens.Screen.
func (m DNSModel) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m DNSModel) Summary() string {
	def := ""
	for _, d := range m.domains {
		if d.Default {
			def = d.Name
			break
		}
	}
	if def != "" {
		return fmt.Sprintf("%d domains · default %s", len(m.domains), def)
	}
	return fmt.Sprintf("%d domains", len(m.domains))
}

func (m *DNSModel) rebuildTable() {
	visible := m.visible()
	rows := make([]table.Row, 0, len(visible))
	for _, d := range visible {
		def := ""
		if d.Default {
			def = "*"
		}
		rows = append(rows, table.Row{d.Name, def})
	}
	m.tbl.SetRows(rows)
}

func (m *DNSModel) visible() []cli.DNSDomain {
	if m.filter == "" {
		return m.domains
	}
	needle := strings.ToLower(m.filter)
	out := make([]cli.DNSDomain, 0, len(m.domains))
	for _, d := range m.domains {
		if strings.Contains(strings.ToLower(d.Name), needle) {
			out = append(out, d)
		}
	}
	return out
}

func (m *DNSModel) focused() *cli.DNSDomain {
	idx := m.tbl.Cursor()
	visible := m.visible()
	if idx < 0 || idx >= len(visible) {
		return nil
	}
	return &visible[idx]
}

func (m *DNSModel) requestCreate() tea.Cmd {
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewTextInput(
			"create-dns", "DNS domain to create:", "", m.palette,
		)}
	}
}

func (m *DNSModel) performCreate(name string) tea.Cmd {
	name = strings.TrimSpace(name)
	if name == "" {
		return func() tea.Msg {
			return screens.StatusMsg{Toast: "create dns: empty name"}
		}
	}
	return func() tea.Msg {
		err := m.client.CreateDNSDomain(cli.DefaultCtx(), name)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("create dns %s failed: %v", name, err)}
		}
		return screens.StatusMsg{Toast: fmt.Sprintf("created dns domain %s", name)}
	}
}

func (m *DNSModel) requestDelete() tea.Cmd {
	d := m.focused()
	if d == nil {
		return nil
	}
	name := d.Name
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewConfirm(
			"Delete DNS domain", "This will remove:",
			[]string{name}, "delete-dns", m.palette,
		)}
	}
}

func (m *DNSModel) performDelete() tea.Cmd {
	d := m.focused()
	if d == nil {
		return nil
	}
	name := d.Name
	return func() tea.Msg {
		err := m.client.DeleteDNSDomain(cli.DefaultCtx(), name)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("delete dns %s failed: %v", name, err)}
		}
		return screens.StatusMsg{Toast: fmt.Sprintf("deleted dns domain %s", name)}
	}
}

func (m *DNSModel) requestSetDefault() tea.Cmd {
	d := m.focused()
	if d == nil {
		return nil
	}
	name := d.Name
	return func() tea.Msg {
		err := m.client.SetDefaultDNSDomain(cli.DefaultCtx(), name)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("default dns %s failed: %v", name, err)}
		}
		return screens.StatusMsg{Toast: fmt.Sprintf("default dns set to %s", name)}
	}
}

func (m DNSModel) handleFilterKey(msg tea.KeyMsg) (screens.Screen, tea.Cmd) {
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
