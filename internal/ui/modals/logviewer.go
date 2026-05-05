package modals

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/ui/theme"
)

// LogSource represents a source of log lines.
type LogSource struct {
	Name       string
	ColorIndex int
	Stream     cli.Stream
}

// LogViewerModel is a modal for viewing streaming logs.
type LogViewerModel struct {
	sources      []LogSource
	buffer       []string
	maxLines     int
	viewport     viewport.Model
	palette      theme.Palette
	width        int
	height       int
	filter       string
	filterActive bool
	filterInput  string
	followTail   bool
	showTime     bool
	showRelTime  bool
	userScrolled bool
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewLogViewer creates a new log viewer modal.
func NewLogViewer(sources []LogSource) *LogViewerModel {
	return NewLogViewerWithPalette(sources, theme.DefaultDark())
}

// NewLogViewerWithPalette creates a log viewer styled with the given palette.
func NewLogViewerWithPalette(sources []LogSource, p theme.Palette) *LogViewerModel {
	ctx, cancel := context.WithCancel(context.Background())

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(24))
	vp.Style = lipgloss.NewStyle().Background(p.Bg).Foreground(p.Fg)

	m := &LogViewerModel{
		sources:      sources,
		buffer:       []string{},
		maxLines:     5000,
		viewport:     vp,
		palette:      p,
		followTail:   true,
		showTime:     false,
		showRelTime:  false,
		userScrolled: false,
		ctx:          ctx,
		cancel:       cancel,
	}

	return m
}

// Init implements tea.Model.
func (m *LogViewerModel) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	for _, source := range m.sources {
		cmds = append(cmds, m.waitForEvent(source))
	}
	return tea.Batch(cmds...)
}

type logEventMsg struct {
	sourceName string
	event      cli.StreamEvent
}

type logDoneMsg struct {
	sourceName string
	result     cli.StreamResult
}

func (m *LogViewerModel) waitForEvent(source LogSource) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-source.Stream.Events
		if !ok {
			// Stream closed, wait for Done
			result := <-source.Stream.Done
			return logDoneMsg{sourceName: source.Name, result: result}
		}
		return logEventMsg{sourceName: source.Name, event: event}
	}
}

// Update implements tea.Model.
func (m *LogViewerModel) Update(msg tea.Msg) (Modal, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.filterActive {
			return m.handleFilterInput(msg)
		}
		return m.handleKey(msg)

	case logEventMsg:
		m.appendEvent(msg.sourceName, msg.event)
		m.updateViewport()
		// Continue waiting for next event from this source
		for _, src := range m.sources {
			if src.Name == msg.sourceName {
				cmds = append(cmds, m.waitForEvent(src))
				break
			}
		}

	case logDoneMsg:
		// Source finished, just note it
		m.appendLine(fmt.Sprintf("[%s: stream ended]", msg.sourceName))
		m.updateViewport()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.SetWidth(msg.Width)
		m.viewport.SetHeight(msg.Height - 3) // Reserve space for header/footer
		m.updateViewport()
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *LogViewerModel) handleKey(msg tea.KeyPressMsg) (Modal, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.cancel()
		return m, CloseModal()
	case "/":
		m.filterActive = true
		m.filterInput = ""
		return m, nil
	case "G":
		m.userScrolled = false
		m.followTail = true
		m.viewport.GotoBottom()
		return m, nil
	case "g":
		m.userScrolled = true
		m.followTail = false
		m.viewport.GotoTop()
		return m, nil
	case "t":
		m.showTime = !m.showTime
		m.updateViewport()
		return m, nil
	case "T":
		m.showRelTime = !m.showRelTime
		m.updateViewport()
		return m, nil
	case "ctrl+s":
		return m, m.saveToFile()
	case "up", "k":
		m.userScrolled = true
		m.followTail = false
		m.viewport.ScrollUp(1)
		return m, nil
	case "down", "j":
		// Check if we're at bottom
		if m.viewport.AtBottom() {
			m.followTail = true
			m.userScrolled = false
		}
		m.viewport.ScrollDown(1)
		return m, nil
	case "pgup":
		m.userScrolled = true
		m.followTail = false
		m.viewport.PageUp()
		return m, nil
	case "pgdown":
		if m.viewport.AtBottom() {
			m.followTail = true
			m.userScrolled = false
		}
		m.viewport.PageDown()
		return m, nil
	}
	return m, nil
}

func (m *LogViewerModel) handleFilterInput(msg tea.KeyPressMsg) (Modal, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.filter = m.filterInput
		m.filterActive = false
		m.updateViewport()
		return m, nil
	case "esc":
		m.filterActive = false
		m.filterInput = ""
		return m, nil
	case "backspace":
		if len(m.filterInput) > 0 {
			m.filterInput = m.filterInput[:len(m.filterInput)-1]
		}
		return m, nil
	default:
		if msg.Text != "" {
			m.filterInput += msg.Text
		}
		return m, nil
	}
}

func (m *LogViewerModel) appendEvent(sourceName string, event cli.StreamEvent) {
	var line string
	var level string

	switch e := event.(type) {
	case cli.LogLine:
		line = e.Raw
		level = e.Level
	case cli.RawLine:
		line = e.Text
	default:
		line = fmt.Sprintf("%v", event)
	}

	// Multi-source: prefix with [name]
	if len(m.sources) > 1 {
		// Find color index
		var colorIdx int
		for _, src := range m.sources {
			if src.Name == sourceName {
				colorIdx = src.ColorIndex
				break
			}
		}
		color := theme.SourceColors[colorIdx%len(theme.SourceColors)]
		prefix := lipgloss.NewStyle().Foreground(color).Render(fmt.Sprintf("[%s]", sourceName))
		line = prefix + " " + line
	}

	// Apply level coloring if level detected
	if level != "" {
		line = m.colorizeLevel(line, level)
	}

	m.appendLine(line)
}

func (m *LogViewerModel) appendLine(line string) {
	m.buffer = append(m.buffer, line)
	if len(m.buffer) > m.maxLines {
		m.buffer = m.buffer[len(m.buffer)-m.maxLines:]
	}
}

func (m *LogViewerModel) updateViewport() {
	var lines []string
	for _, line := range m.buffer {
		if m.filter == "" || strings.Contains(line, m.filter) {
			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n")
	m.viewport.SetContent(content)

	if m.followTail && !m.userScrolled {
		m.viewport.GotoBottom()
	}
}

func (m *LogViewerModel) colorizeLevel(line, level string) string {
	var color color.Color
	switch level {
	case "INFO":
		color = lipgloss.Color("86") // cyan
	case "WARN":
		color = lipgloss.Color("226") // yellow
	case "ERROR":
		color = lipgloss.Color("196") // red
	case "DEBUG":
		color = lipgloss.Color("240") // dim gray
	default:
		return line
	}
	return lipgloss.NewStyle().Foreground(color).Render(line)
}

func (m *LogViewerModel) saveToFile() tea.Cmd {
	return func() tea.Msg {
		dataDir := filepath.Join(os.Getenv("HOME"), ".local", "share", "c9s", "logs")
		if err := os.MkdirAll(dataDir, 0o755); err != nil {
			return nil // Ignore error
		}

		sourceName := "log"
		if len(m.sources) > 0 {
			sourceName = m.sources[0].Name
		}

		filename := fmt.Sprintf("%s-%d.log", sourceName, time.Now().Unix())
		path := filepath.Join(dataDir, filename)

		content := strings.Join(m.buffer, "\n")
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			return nil // Ignore error
		}

		return StatusMsg(fmt.Sprintf("Saved to %s", path))
	}
}

// View implements Modal.
func (m *LogViewerModel) View(width, height int) string {
	if width > 0 && (m.viewport.Width() != width || m.viewport.Height() != height-3) {
		m.viewport.SetWidth(width)
		m.viewport.SetHeight(height - 3)
		m.updateViewport()
	}
	header := m.renderHeader()
	footer := m.renderFooter()
	body := m.viewport.View()

	bg := lipgloss.NewStyle().Background(m.palette.Bg).Foreground(m.palette.Fg)
	out := lipgloss.JoinVertical(
		lipgloss.Left,
		bg.Width(width).Render(header),
		bg.Width(width).Render(body),
		bg.Width(width).Render(footer),
	)
	return bg.Width(width).Render(out)
}

// Title implements Modal.
func (m *LogViewerModel) Title() string {
	if len(m.sources) == 1 {
		return "Logs: " + m.sources[0].Name
	}
	return fmt.Sprintf("Logs (%d sources)", len(m.sources))
}

func (m *LogViewerModel) renderHeader() string {
	title := "Logs"
	if len(m.sources) == 1 {
		title = fmt.Sprintf("Logs: %s", m.sources[0].Name)
	} else if len(m.sources) > 1 {
		title = fmt.Sprintf("Logs: %d sources", len(m.sources))
	}

	badges := []string{}
	if m.followTail {
		badges = append(badges, "follow")
	}
	if m.filter != "" {
		badges = append(badges, fmt.Sprintf("filter:%s", m.filter))
	}
	if m.showTime {
		badges = append(badges, "time")
	}
	if m.showRelTime {
		badges = append(badges, "rel-time")
	}

	badgeStr := ""
	if len(badges) > 0 {
		badgeStr = " [" + strings.Join(badges, " ") + "]"
	}

	return lipgloss.NewStyle().Bold(true).Background(m.palette.Bg).Foreground(m.palette.HeaderFg).Render(title + badgeStr)
}

func (m *LogViewerModel) renderFooter() string {
	if m.filterActive {
		return fmt.Sprintf("Filter: %s_", m.filterInput)
	}

	help := "q:quit  /:filter  G:tail  t:time  T:rel-time  ctrl+s:save"
	return lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg).Render(help)
}
