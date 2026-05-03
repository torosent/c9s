package system

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/ui/keymap"
	"github.com/torosent/c9s/internal/ui/screens"
	"github.com/torosent/c9s/internal/ui/theme"
)

// KernelModel is the :kernel sub-screen — a read-only viewer of the
// current kernel configuration sourced from cli.ListSystemProperties
// and filtered to keys with the "kernel." prefix.
type KernelModel struct {
	client   cli.Client
	clk      clock.Clock
	palette  theme.Palette
	props    []cli.SystemProperty
	viewport viewport.Model
	keymap   *keymap.Map
	width    int
	height   int
}

type kernelMsg []cli.SystemProperty

// NewKernel creates a new :kernel sub-screen.
func NewKernel(client cli.Client, clk clock.Clock, p theme.Palette) KernelModel {
	km := keymap.Default()
	vp := viewport.New(80, 18)
	vp.Style = lipgloss.NewStyle()
	return KernelModel{
		client:   client,
		clk:      clk,
		palette:  p,
		viewport: vp,
		keymap:   km,
	}
}

// Init implements screens.Screen.
func (m KernelModel) Init() tea.Cmd { return m.refreshCmd() }

func (m KernelModel) refreshCmd() tea.Cmd {
	c := m.client
	return func() tea.Msg {
		props, err := c.ListSystemProperties(context.Background())
		if err != nil {
			return screens.StatusMsg{Toast: fmt.Sprintf("kernel read failed: %v", err)}
		}
		filtered := make([]cli.SystemProperty, 0, len(props))
		for _, p := range props {
			if strings.HasPrefix(p.Key, "kernel.") {
				filtered = append(filtered, p)
			}
		}
		return kernelMsg(filtered)
	}
}

// Update implements screens.Screen.
func (m KernelModel) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		if msg.Height > 4 {
			m.viewport.Height = msg.Height - 4
		}
		m.rebuild()
	case kernelMsg:
		m.props = []cli.SystemProperty(msg)
		m.rebuild()
	case tea.KeyMsg:
		if m.keymap.Matches("refresh", msg) {
			return m, m.refreshCmd()
		}
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View implements screens.Screen.
func (m KernelModel) View(width, height int) string {
	help := lipgloss.NewStyle().Foreground(m.palette.Dim).Render("r: refresh · j/k: scroll")
	return m.viewport.View() + "\n" + help
}

// Title implements screens.Screen.
func (m KernelModel) Title() string { return "System kernel" }

// Hotkeys implements screens.Screen.
func (m KernelModel) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m KernelModel) Summary() string {
	if len(m.props) == 0 {
		return "no kernel properties"
	}
	return fmt.Sprintf("%d kernel properties", len(m.props))
}

func (m *KernelModel) rebuild() {
	if len(m.props) == 0 {
		m.viewport.SetContent("(no kernel.* properties found)")
		return
	}
	body := ""
	for _, p := range m.props {
		ro := ""
		if p.ReadOnly {
			ro = " (ro)"
		}
		body += fmt.Sprintf("%s = %s%s\n", p.Key, p.Value, ro)
	}
	m.viewport.SetContent(body)
}
