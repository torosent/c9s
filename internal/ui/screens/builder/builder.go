// Package builder implements the :builder resource screen — a single-card
// view (not a table) that shows the runtime build subsystem state.
package builder

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/state"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/modals"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/theme"
)

// Model represents the :builder screen.
type Model struct {
	client  cli.Client
	clk     clock.Clock
	palette theme.Palette
	status  cli.BuilderStatus
	keymap  *keymap.Map
	width   int
	height  int
}

// New creates a new Builder screen.
func New(client cli.Client, clk clock.Clock, p theme.Palette) *Model {
	km := keymap.Default()
	km.Add("start", keymap.Binding{
		Keys: []string{"S", "shift+s"}, Help: "Start", Description: "Start the builder",
	})
	km.Add("stop", keymap.Binding{
		Keys: []string{"X", "shift+x"}, Help: "Stop", Description: "Stop the builder",
	})
	km.Add("delete", keymap.Binding{
		Keys: []string{"D", "shift+d"}, Help: "Delete", Description: "Delete the builder",
	})
	return &Model{
		client:  client,
		clk:     clk,
		palette: p,
		keymap:  km,
	}
}

// Init implements screens.Screen.
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshCmd(),
		state.TickCmd(2*time.Second, m.clk, cli.ResourceBuilder),
	)
}

// Update implements screens.Screen.
func (m *Model) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case screens.PaletteChangedMsg:
		m.palette = msg.P

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case statusMsg:
		m.status = cli.BuilderStatus(msg)

	case state.TickMsg:
		if msg.Resource != cli.ResourceBuilder {
			break
		}
		cmds = append(cmds, m.refreshCmd(), state.TickCmd(2*time.Second, m.clk, cli.ResourceBuilder))

	case modals.ConfirmResultMsg:
		if msg.Result.Tag == "delete-builder" && msg.Result.Confirmed {
			cmds = append(cmds, m.deleteBuilder())
		}

	case tea.KeyMsg:
		if m.keymap.Matches("refresh", msg) {
			return m, m.refreshCmd()
		}
		if m.keymap.Matches("start", msg) {
			return m, m.startBuilder()
		}
		if m.keymap.Matches("stop", msg) {
			return m, m.stopBuilder()
		}
		if m.keymap.Matches("delete", msg) {
			return m, m.requestDelete()
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements screens.Screen.
func (m *Model) View(width, height int) string {
	return m.renderCard(width)
}

func (m *Model) renderCard(width int) string {
	state := m.status.State
	if state == "" {
		state = "unknown"
	}
	stateColor := lipgloss.Color("#999999")
	switch state {
	case "running":
		stateColor = m.palette.Success
	case "stopped":
		stateColor = m.palette.Warning
	case "error":
		stateColor = m.palette.Error
	}

	memory := formatBytes(m.status.MemoryBytes)
	uptime := formatUptime(time.Duration(m.status.UptimeSec) * time.Second)
	cpus := fmt.Sprintf("%d", m.status.CPUs)

	cardWidth := 40
	if width-4 < cardWidth {
		cardWidth = width - 4
	}
	if cardWidth < 20 {
		cardWidth = 20
	}

	// Inner content area = card width - border (2) - horizontal padding (4).
	innerW := cardWidth - 6
	if innerW < 10 {
		innerW = 10
	}

	bg := lipgloss.NewStyle().Foreground(m.palette.Fg).Background(m.palette.Bg)
	dim := lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg)
	stateStyle := lipgloss.NewStyle().Foreground(stateColor).Background(m.palette.Bg).Bold(true)

	// Each row: dim label + value, then padded out to innerW with bg spaces.
	mkRow := func(label, value string, valueStyle lipgloss.Style) string {
		text := dim.Render(label) + valueStyle.Render(value)
		pad := innerW - lipgloss.Width(text)
		if pad < 0 {
			pad = 0
		}
		return text + bg.Render(strings.Repeat(" ", pad))
	}

	rows := []string{
		mkRow("STATE   ", state, stateStyle),
		mkRow("CPU     ", cpus, bg),
		mkRow("MEM     ", memory, bg),
		mkRow("UPTIME  ", uptime, bg),
	}
	body := strings.Join(rows, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		BorderBackground(m.palette.Bg).
		Background(m.palette.Bg).
		Foreground(m.palette.Fg).
		Padding(1, 2).
		Width(cardWidth).
		Render(body)

	help := lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg).Render("S: start · X: stop · D: delete · r: refresh")
	return box + "\n" + bg.Render(" ") + "\n" + help
}

// renderCard ends here.

// Title implements screens.Screen.
func (m *Model) Title() string { return "Builder" }

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	state := m.status.State
	if state == "" {
		state = "unknown"
	}
	return fmt.Sprintf("builder %s", state)
}

func (m *Model) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		st, err := m.client.BuilderStatus(cli.DefaultCtx())
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("builder status failed: %v", err)}
		}
		return statusMsg(st)
	}
}

func (m *Model) startBuilder() tea.Cmd {
	return func() tea.Msg {
		err := m.client.BuilderStart(cli.DefaultCtx())
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("builder start failed: %v", err)}
		}
		return screens.StatusMsg{Toast: "builder started"}
	}
}

func (m *Model) stopBuilder() tea.Cmd {
	return func() tea.Msg {
		err := m.client.BuilderStop(cli.DefaultCtx())
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("builder stop failed: %v", err)}
		}
		return screens.StatusMsg{Toast: "builder stopped"}
	}
}

func (m *Model) requestDelete() tea.Cmd {
	return func() tea.Msg {
		return screens.OpenModalMsg{Modal: modals.NewConfirm(
			"Delete builder",
			"This will permanently destroy the builder VM:",
			[]string{"all build cache and state will be lost"},
			"delete-builder",
			m.palette,
		)}
	}
}

func (m *Model) deleteBuilder() tea.Cmd {
	return func() tea.Msg {
		err := m.client.BuilderDelete(cli.DefaultCtx())
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("builder delete failed: %v", err)}
		}
		return screens.StatusMsg{Toast: "builder deleted"}
	}
}

type statusMsg cli.BuilderStatus

// formatBytes renders a byte count as a short human-readable string.
func formatBytes(b int64) string {
	if b == 0 {
		return "-"
	}
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1fG", float64(b)/gb)
	case b >= mb:
		return fmt.Sprintf("%.0fM", float64(b)/mb)
	case b >= kb:
		return fmt.Sprintf("%.1fK", float64(b)/kb)
	default:
		return fmt.Sprintf("%dB", b)
	}
}

// formatUptime renders a duration as a short human-readable string.
func formatUptime(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm", m)
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
