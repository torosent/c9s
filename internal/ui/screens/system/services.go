// Package system implements the :system family of screens — a services
// table on `:system`, plus sub-screens reached via :df, :dns, :property,
// :kernel, and :logs.
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
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/skinx"
	"github.com/torosent/c9s/internal/ui/theme"
)

// ServicesModel is the default :system screen — a services table.
type ServicesModel struct {
	client     cli.Client
	clk        clock.Clock
	palette    theme.Palette
	tbl        table.Model
	services   []cli.SystemService
	filter     string
	filterMode bool
	keymap     *keymap.Map
	width      int
	height     int
}

// NewServices creates a new system services screen.
func NewServices(client cli.Client, clk clock.Clock, p theme.Palette) ServicesModel {
	km := keymap.Default()
	km.Add("start_all", keymap.Binding{
		Keys: []string{"S", "shift+s"}, Help: "Start all", Description: "Start all services",
	})
	km.Add("stop_all", keymap.Binding{
		Keys: []string{"X", "shift+x"}, Help: "Stop all", Description: "Stop all services",
	})

	columns := []table.Column{
		{Title: "SERVICE", Width: 32},
		{Title: "STATE", Width: 10},
		{Title: "PID", Width: 8},
		{Title: "UPTIME", Width: 10},
	}
	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	tbl.SetStyles(skinx.TableStyles(p))

	return ServicesModel{
		client:  client,
		clk:     clk,
		palette: p,
		tbl:     tbl,
		keymap:  km,
	}
}

// Init implements screens.Screen.
func (m ServicesModel) Init() tea.Cmd {
	return tea.Batch(
		state.MakeRefreshedCmd[cli.SystemService](
			cli.DefaultCtx(),
			func(ctx context.Context) ([]cli.SystemService, error) {
				return m.client.ListSystemServices(ctx)
			},
			cli.ResourceSystem,
		),
		state.TickCmd(2*time.Second, m.clk, cli.ResourceSystem),
	)
}

// Update implements screens.Screen.
func (m ServicesModel) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
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

	case state.RefreshedMsg[cli.SystemService]:
		if msg.Resource != cli.ResourceSystem {
			break
		}
		m.services = msg.Snapshot.Items
		m.rebuildTable()

	case state.TickMsg:
		if msg.Resource != cli.ResourceSystem {
			break
		}
		cmds = append(cmds,
			state.MakeRefreshedCmd[cli.SystemService](
				cli.DefaultCtx(),
				func(ctx context.Context) ([]cli.SystemService, error) {
					return m.client.ListSystemServices(ctx)
				},
				cli.ResourceSystem,
			),
			state.TickCmd(2*time.Second, m.clk, cli.ResourceSystem),
		)

	case tea.KeyMsg:
		if m.filterMode {
			return m.handleFilterKey(msg)
		}
		if m.keymap.Matches("escape", msg) {
			return m, nil
		}
		if m.keymap.Matches("refresh", msg) {
			return m, state.MakeRefreshedCmd[cli.SystemService](
				cli.DefaultCtx(),
				func(ctx context.Context) ([]cli.SystemService, error) {
					return m.client.ListSystemServices(ctx)
				},
				cli.ResourceSystem,
			)
		}
		if m.keymap.Matches("filter", msg) {
			m.filterMode = true
			m.filter = ""
			return m, nil
		}
		if m.keymap.Matches("start_all", msg) {
			return m, m.startAll()
		}
		if m.keymap.Matches("stop_all", msg) {
			return m, m.stopAll()
		}

		var cmd tea.Cmd
		m.tbl, cmd = m.tbl.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements screens.Screen.
func (m ServicesModel) View(width, height int) string {
	if m.filterMode {
		return m.tbl.View() + "\n" + fmt.Sprintf("Filter: %s_", m.filter)
	}
	return m.tbl.View()
}

// Title implements screens.Screen.
func (m ServicesModel) Title() string { return "System services" }

// Hotkeys implements screens.Screen.
func (m ServicesModel) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m ServicesModel) Summary() string {
	total := len(m.services)
	running := 0
	for _, s := range m.services {
		if s.State == "running" {
			running++
		}
	}
	return fmt.Sprintf("%d services · %d running", total, running)
}

func (m *ServicesModel) rebuildTable() {
	visible := m.visible()
	rows := make([]table.Row, 0, len(visible))
	for _, s := range visible {
		uptime := formatUptime(time.Duration(s.UptimeSec) * time.Second)
		pid := "-"
		if s.PID > 0 {
			pid = fmt.Sprintf("%d", s.PID)
		}
		rows = append(rows, table.Row{s.Name, s.State, pid, uptime})
	}
	m.tbl.SetRows(rows)
}

func (m *ServicesModel) visible() []cli.SystemService {
	if m.filter == "" {
		return m.services
	}
	needle := strings.ToLower(m.filter)
	visible := make([]cli.SystemService, 0, len(m.services))
	for _, s := range m.services {
		if strings.Contains(strings.ToLower(s.Name+" "+s.State), needle) {
			visible = append(visible, s)
		}
	}
	return visible
}

func (m *ServicesModel) startAll() tea.Cmd {
	return func() tea.Msg {
		err := m.client.SystemStartAll(cli.DefaultCtx())
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("system start failed: %v", err)}
		}
		return screens.StatusMsg{Toast: "all system services started"}
	}
}

func (m *ServicesModel) stopAll() tea.Cmd {
	return func() tea.Msg {
		err := m.client.SystemStopAll(cli.DefaultCtx())
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("system stop failed: %v", err)}
		}
		return screens.StatusMsg{Toast: "all system services stopped"}
	}
}

func (m ServicesModel) handleFilterKey(msg tea.KeyMsg) (screens.Screen, tea.Cmd) {
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

// formatUptime renders a duration as a short human-readable string.
func formatUptime(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	h := int(d.Hours())
	mn := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, mn)
	}
	if mn > 0 {
		return fmt.Sprintf("%dm", mn)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
