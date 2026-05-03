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

// LogsModel is the :logs sub-screen — a streaming viewer of
// `container system logs --follow`.
type LogsModel struct {
	client   cli.Client
	clk      clock.Clock
	palette  theme.Palette
	stream   *cli.Stream
	cancel   context.CancelFunc
	buffer   []string
	maxLines int
	viewport viewport.Model
	follow   bool
	keymap   *keymap.Map
	width    int
	height   int
	started  bool
	done     bool
}

// NewLogs creates a new :logs sub-screen.
func NewLogs(client cli.Client, clk clock.Clock, p theme.Palette) *LogsModel {
	km := keymap.Default()
	vp := viewport.New(80, 18)
	vp.Style = lipgloss.NewStyle()
	return &LogsModel{
		client:   client,
		clk:      clk,
		palette:  p,
		buffer:   []string{},
		maxLines: 5000,
		viewport: vp,
		follow:   true,
		keymap:   km,
	}
}

type logLineMsg struct{ event cli.StreamEvent }

type logsDoneMsg struct{ result cli.StreamResult }

// Init implements screens.Screen.
func (m *LogsModel) Init() tea.Cmd {
	return m.start()
}

func (m *LogsModel) start() tea.Cmd {
	if m.started {
		return nil
	}
	ctx, cancel := context.WithCancel(cli.DefaultCtx())
	stream, err := m.client.StreamSystemLogs(ctx, true)
	if err != nil {
		cancel()
		return func() tea.Msg {
			return screens.StatusMsg{Toast: fmt.Sprintf("system logs failed: %v", err)}
		}
	}
	m.stream = &stream
	m.cancel = cancel
	m.started = true
	return m.waitEvent()
}

func (m *LogsModel) waitEvent() tea.Cmd {
	stream := m.stream
	if stream == nil {
		return nil
	}
	return func() tea.Msg {
		ev, ok := <-stream.Events
		if !ok {
			res := <-stream.Done
			return logsDoneMsg{result: res}
		}
		return logLineMsg{event: ev}
	}
}

// Update implements screens.Screen.
func (m *LogsModel) Update(msg tea.Msg) (screens.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case screens.PaletteChangedMsg:
		m.palette = msg.P

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		if msg.Height > 4 {
			m.viewport.Height = msg.Height - 4
		}
		m.rebuild()

	case logLineMsg:
		m.appendEvent(msg.event)
		m.rebuild()
		return m, m.waitEvent()

	case logsDoneMsg:
		m.done = true
		m.appendLine(fmt.Sprintf("[stream ended: exit %d]", msg.result.ExitCode))
		m.rebuild()

	case tea.KeyMsg:
		switch msg.String() {
		case "G":
			m.follow = true
			m.viewport.GotoBottom()
		case "g":
			m.follow = false
			m.viewport.GotoTop()
		case "j", "down":
			m.viewport.ScrollDown(1)
		case "k", "up":
			m.viewport.ScrollUp(1)
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View implements screens.Screen.
func (m *LogsModel) View(width, height int) string {
	header := lipgloss.NewStyle().Bold(true).Foreground(m.palette.HeaderFg).
		Render("system logs")
	if m.follow {
		header += lipgloss.NewStyle().Foreground(m.palette.Dim).Render(" [follow]")
	}
	if m.done {
		header += lipgloss.NewStyle().Foreground(m.palette.Warning).Render(" [stopped]")
	}
	help := lipgloss.NewStyle().Foreground(m.palette.Dim).Render("g/G: top/tail · j/k: scroll · q: leave")
	return header + "\n" + m.viewport.View() + "\n" + help
}

// Title implements screens.Screen.
func (m *LogsModel) Title() string { return "System logs" }

// Hotkeys implements screens.Screen.
func (m *LogsModel) Hotkeys() *keymap.Map { return m.keymap }

// Summary implements screens.Screen.
func (m *LogsModel) Summary() string {
	if m.done {
		return fmt.Sprintf("%d lines · stopped", len(m.buffer))
	}
	return fmt.Sprintf("%d lines · streaming", len(m.buffer))
}

func (m *LogsModel) appendEvent(ev cli.StreamEvent) {
	switch e := ev.(type) {
	case cli.LogLine:
		line := e.Raw
		if e.Level != "" {
			line = colorizeLevel(line, e.Level)
		}
		m.appendLine(line)
	case cli.RawLine:
		m.appendLine(e.Text)
	default:
		m.appendLine(fmt.Sprintf("%v", ev))
	}
}

func (m *LogsModel) appendLine(line string) {
	m.buffer = append(m.buffer, line)
	if len(m.buffer) > m.maxLines {
		m.buffer = m.buffer[len(m.buffer)-m.maxLines:]
	}
}

func (m *LogsModel) rebuild() {
	m.viewport.SetContent(strings.Join(m.buffer, "\n"))
	if m.follow {
		m.viewport.GotoBottom()
	}
}

// Cancel terminates the active stream. Called by the app when the screen
// is replaced.
func (m *LogsModel) Cancel() {
	if m.cancel != nil {
		m.cancel()
	}
}

func colorizeLevel(line, level string) string {
	var color lipgloss.Color
	switch level {
	case "INFO":
		color = lipgloss.Color("86")
	case "WARN":
		color = lipgloss.Color("226")
	case "ERROR":
		color = lipgloss.Color("196")
	case "DEBUG":
		color = lipgloss.Color("240")
	default:
		return line
	}
	return lipgloss.NewStyle().Foreground(color).Render(line)
}
