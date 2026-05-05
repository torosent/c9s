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

const propertyResource cli.Resource = "system-property"

// PropertyModel is the :property sub-screen.
type PropertyModel struct {
	client     cli.Client
	clk        clock.Clock
	palette    theme.Palette
	tbl        table.Model
	props      []cli.SystemProperty
	pendingKey string
	filter     string
	filterMode bool
	keymap     *keymap.Map
	width      int
	height     int
}

// NewProperty creates a new :property sub-screen.
func NewProperty(client cli.Client, clk clock.Clock, p theme.Palette) PropertyModel {
	km := keymap.Default()
	km.Add("edit", keymap.Binding{
		Keys: []string{"e"}, Help: "Edit", Description: "Edit property value",
	})
	km.Add("delete", keymap.Binding{
		Keys: []string{"D", "shift+d"}, Help: "Reset", Description: "Reset property to default",
	})

	columns := []table.Column{
		{Title: "KEY", Width: 32},
		{Title: "VALUE", Width: 36},
		{Title: "RO", Width: 4},
	}
	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	tbl.SetStyles(skinx.TableStyles(p))

	return PropertyModel{
		client:  client,
		clk:     clk,
		palette: p,
		tbl:     tbl,
		keymap:  km,
	}
}

// Init implements screens.Screen.
func (m PropertyModel) Init() tea.Cmd {
	return tea.Batch(
		m.refreshCmd(),
		state.TickCmd(2*time.Second, m.clk, propertyResource),
	)
}

func (m PropertyModel) refreshCmd() tea.Cmd {
	return state.MakeRefreshedCmd[cli.SystemProperty](
		cli.DefaultCtx(),
		func(ctx context.Context) ([]cli.SystemProperty, error) {
			return m.client.ListSystemProperties(ctx)
		},
		propertyResource,
	)
}

// Update implements screens.Screen.
func (m PropertyModel) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
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

	case state.RefreshedMsg[cli.SystemProperty]:
		if msg.Resource != propertyResource {
			break
		}
		m.props = msg.Snapshot.Items
		m.rebuildTable()

	case modals.ConfirmResultMsg:
		if msg.Result.Tag == "reset-property" && msg.Result.Confirmed {
			cmds = append(cmds, m.performReset())
		}

	case modals.TextInputResultMsg:
		if strings.HasPrefix(msg.Result.Label, "set-property:") {
			key := strings.TrimPrefix(msg.Result.Label, "set-property:")
			cmds = append(cmds, m.performSet(key, msg.Result.Value))
		}

	case tea.KeyPressMsg:
		if m.filterMode {
			return m.handleFilterKey(msg)
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
		if m.keymap.Matches("edit", msg) {
			return m, m.requestEdit()
		}
		if m.keymap.Matches("delete", msg) {
			return m, m.requestReset()
		}
		var cmd tea.Cmd
		m.tbl, cmd = m.tbl.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

// View implements screens.Screen.
func (m PropertyModel) View(width, height int) string {
	(&m.tbl).SetWidth(width)
	if m.filterMode {
		return m.tbl.View() + "\n" + fmt.Sprintf("Filter: %s_", m.filter)
	}
	return m.tbl.View()
}

// Title implements screens.Screen.
func (m PropertyModel) Title() string { return "System properties" }

// Hotkeys implements screens.Screen.
func (m PropertyModel) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m PropertyModel) Summary() string {
	return fmt.Sprintf("%d properties", len(m.props))
}

func (m *PropertyModel) rebuildTable() {
	visible := m.visible()
	rows := make([]table.Row, 0, len(visible))
	for _, p := range visible {
		ro := ""
		if p.ReadOnly {
			ro = "*"
		}
		rows = append(rows, table.Row{p.Key, p.Value, ro})
	}
	m.tbl.SetRows(rows)
}

func (m *PropertyModel) visible() []cli.SystemProperty {
	if m.filter == "" {
		return m.props
	}
	needle := strings.ToLower(m.filter)
	out := make([]cli.SystemProperty, 0, len(m.props))
	for _, p := range m.props {
		if strings.Contains(strings.ToLower(p.Key+" "+p.Value), needle) {
			out = append(out, p)
		}
	}
	return out
}

func (m *PropertyModel) focused() *cli.SystemProperty {
	idx := m.tbl.Cursor()
	visible := m.visible()
	if idx < 0 || idx >= len(visible) {
		return nil
	}
	return &visible[idx]
}

func (m *PropertyModel) requestEdit() tea.Cmd {
	p := m.focused()
	if p == nil {
		return nil
	}
	if p.ReadOnly {
		return func() tea.Msg {
			return screens.StatusMsg{Toast: fmt.Sprintf("%s is read-only", p.Key)}
		}
	}
	key := p.Key
	value := p.Value
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewTextInput(
			"set-property:"+key,
			fmt.Sprintf("New value for %s:", key),
			value,
			m.palette,
		)}
	}
}

func (m *PropertyModel) performSet(key, value string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.SetSystemProperty(cli.DefaultCtx(), key, value)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("set %s failed: %v", key, err)}
		}
		return screens.StatusMsg{Toast: fmt.Sprintf("set %s = %s", key, value)}
	}
}

func (m *PropertyModel) requestReset() tea.Cmd {
	p := m.focused()
	if p == nil {
		return nil
	}
	if p.ReadOnly {
		return func() tea.Msg {
			return screens.StatusMsg{Toast: fmt.Sprintf("%s is read-only", p.Key)}
		}
	}
	mref := m
	mref.pendingKey = p.Key
	key := p.Key
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewConfirm(
			"Reset property", "Reset to default value?",
			[]string{key}, "reset-property", m.palette,
		)}
	}
}

func (m *PropertyModel) performReset() tea.Cmd {
	p := m.focused()
	if p == nil {
		return nil
	}
	key := p.Key
	return func() tea.Msg {
		err := m.client.ResetSystemProperty(cli.DefaultCtx(), key)
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("reset %s failed: %v", key, err)}
		}
		return screens.StatusMsg{Toast: fmt.Sprintf("reset %s", key)}
	}
}

func (m PropertyModel) handleFilterKey(msg tea.KeyPressMsg) (screens.Screen, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.filterMode = false
		m.rebuildTable()
		return m, nil
	case "esc":
		m.filterMode = false
		m.filter = ""
		m.rebuildTable()
		return m, nil
	case "backspace":
		if len(m.filter) > 0 {
			m.filter = m.filter[:len(m.filter)-1]
			m.rebuildTable()
		}
		return m, nil
	default:
		if msg.Text != "" {
			m.filter += msg.Text
			m.rebuildTable()
			return m, nil
		}
	}
	return m, nil
}
