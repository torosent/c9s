package system

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/skinx"
	"github.com/torosent/c9s/internal/ui/theme"
)

// DFModel is the :df sub-screen — a read-only summary table from
// `container system df`.
type DFModel struct {
	client  cli.Client
	clk     clock.Clock
	palette theme.Palette
	tbl     table.Model
	df      cli.SystemDF
	keymap  *keymap.Map
	width   int
	height  int
}

// NewDF creates a new :df sub-screen.
func NewDF(client cli.Client, clk clock.Clock, p theme.Palette) DFModel {
	km := keymap.Default()
	columns := []table.Column{
		{Title: "TYPE", Width: 16},
		{Title: "COUNT", Width: 8},
		{Title: "ACTIVE", Width: 8},
		{Title: "SIZE", Width: 12},
		{Title: "RECLAIMABLE", Width: 14},
	}
	tbl := table.New(
		table.WithColumns(columns),
		table.WithFocused(false),
		table.WithHeight(6),
	)
	tbl.SetStyles(skinx.TableStyles(p))
	return DFModel{
		client:  client,
		clk:     clk,
		palette: p,
		tbl:     tbl,
		keymap:  km,
	}
}

type dfMsg cli.SystemDF

// Init implements screens.Screen.
func (m DFModel) Init() tea.Cmd {
	return m.refreshCmd()
}

func (m DFModel) refreshCmd() tea.Cmd {
	return func() tea.Msg {
		df, err := m.client.SystemDF(context.Background())
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("system df failed: %v", err)}
		}
		return dfMsg(df)
	}
}

// Update implements screens.Screen.
func (m DFModel) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case dfMsg:
		m.df = cli.SystemDF(msg)
		m.rebuildTable()
	case tea.KeyMsg:
		if m.keymap.Matches("refresh", msg) {
			return m, m.refreshCmd()
		}
	}
	return m, nil
}

// View implements screens.Screen.
func (m DFModel) View(width, height int) string {
	help := lipgloss.NewStyle().Foreground(m.palette.Dim).Render("r: refresh")
	return m.tbl.View() + "\n" + help
}

// Title implements screens.Screen.
func (m DFModel) Title() string { return "System df" }

// Hotkeys implements screens.Screen.
func (m DFModel) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m DFModel) Summary() string {
	total := m.df.Images.SizeBytes + m.df.Containers.SizeBytes + m.df.Volumes.SizeBytes
	reclaimable := m.df.Images.ReclaimBytes + m.df.Containers.ReclaimBytes + m.df.Volumes.ReclaimBytes
	return fmt.Sprintf("total %s · reclaimable %s", formatBytes(total), formatBytes(reclaimable))
}

func (m *DFModel) rebuildTable() {
	rows := []table.Row{
		dfRow("Images", m.df.Images),
		dfRow("Containers", m.df.Containers),
		dfRow("Volumes", m.df.Volumes),
	}
	m.tbl.SetRows(rows)
}

func dfRow(label string, sec cli.DFSection) table.Row {
	return table.Row{
		label,
		fmt.Sprintf("%d", sec.Count),
		fmt.Sprintf("%d", sec.Active),
		formatBytes(sec.SizeBytes),
		formatBytes(sec.ReclaimBytes),
	}
}

// formatBytes is a small duplicate of the columns helper used elsewhere; we
// keep it package-local to avoid a cross-screen import.
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

// (no extra unused imports needed; strings and time are used elsewhere in the package)
