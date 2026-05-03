// Package pulses provides the :pulses dashboard screen.
package pulses

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/theme"
)

// Model represents the pulses dashboard screen.
type Model struct {
	client  cli.Client
	clk     clock.Clock
	palette theme.Palette
	keymap  *keymap.Map
	width   int
	height  int

	containers int
	images     int
	volumes    int
	networks   int
}

// New creates a new pulses screen.
func New(client cli.Client, clk clock.Clock, p theme.Palette) *Model {
	km := keymap.Default()
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
		m.refresh(),
		m.tick(),
	)
}

func (m *Model) tick() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return RefreshMsg{}
	})
}

// RefreshMsg triggers a data refresh.
type RefreshMsg struct{}

func (m *Model) refresh() tea.Cmd {
	return func() tea.Msg {
		containers, _ := m.client.ListContainers(cli.DefaultCtx(), false)
		images, _ := m.client.ListImages(cli.DefaultCtx())
		volumes, _ := m.client.ListVolumes(cli.DefaultCtx())
		networks, _ := m.client.ListNetworks(cli.DefaultCtx())

		return DataRefreshedMsg{
			Containers: len(containers),
			Images:     len(images),
			Volumes:    len(volumes),
			Networks:   len(networks),
		}
	}
}

// DataRefreshedMsg contains refreshed metrics.
type DataRefreshedMsg struct {
	Containers int
	Images     int
	Volumes    int
	Networks   int
}

// Update implements screens.Screen.
func (m *Model) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case DataRefreshedMsg:
		m.containers = msg.Containers
		m.images = msg.Images
		m.volumes = msg.Volumes
		m.networks = msg.Networks
		cmds = append(cmds, m.tick())

	case RefreshMsg:
		cmds = append(cmds, m.refresh())

	case screens.PaletteChangedMsg:
		m.palette = msg.P

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, tea.Batch(cmds...)
}

// View implements screens.Screen.
func (m *Model) View(width, height int) string {
	if width != m.width || height != m.height {
		m.width = width
		m.height = height
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.Border).
		BorderBackground(m.palette.Bg).
		Background(m.palette.Bg).
		Foreground(m.palette.Fg).
		Padding(1, 2).
		Width(width/2 - 4)

	titleStyle := lipgloss.NewStyle().
		Foreground(m.palette.Accent).
		Background(m.palette.Bg).
		Bold(true)

	// Resource counts card
	resourcesCard := cardStyle.Render(
		titleStyle.Render("Resources") + "\n\n" +
			fmt.Sprintf("Containers: %d\n", m.containers) +
			fmt.Sprintf("Images:     %d\n", m.images) +
			fmt.Sprintf("Volumes:    %d\n", m.volumes) +
			fmt.Sprintf("Networks:   %d\n", m.networks),
	)

	// Status card
	statusCard := cardStyle.Render(
		titleStyle.Render("System") + "\n\n" +
			fmt.Sprintf("Time: %s\n", m.clk.Now().Format("15:04:05")) +
			"Status: Healthy\n",
	)

	row1 := lipgloss.JoinHorizontal(lipgloss.Top, resourcesCard, statusCard)

	return row1
}

// Title implements screens.Screen.
func (m *Model) Title() string {
	return "pulses"
}

// Hotkeys implements screens.Screen.
func (m *Model) Hotkeys() *keymap.Map {
	return m.keymap
}

// Summary implements screens.Screen.
func (m *Model) Summary() string {
	return fmt.Sprintf("%d containers, %d images", m.containers, m.images)
}
