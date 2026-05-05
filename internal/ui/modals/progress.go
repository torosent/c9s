package modals

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/jobs"
	"github.com/torosent/c9s/internal/ui/theme"
)

// ProgressModel is a modal for streaming build/pull/push progress.
type ProgressModel struct {
	jobID       string
	kind        jobs.Kind
	target      string
	stream      cli.Stream
	clock       clock.Clock
	palette     theme.Palette
	started     time.Time
	buildSteps  []cli.BuildStepEvent
	layerState  map[string]cli.LayerProgress
	rawLines    []string
	viewport    viewport.Model
	width       int
	height      int
	showRaw     bool // For build: toggle between step list and raw output
	done        bool
	exitCode    int
	err         error
	awaitCancel bool
	// cancelGen lets us drop expired cancelWindowMsg deliveries when the
	// user has already pressed Ctrl+C a second time (which advances the
	// generation) so the still-pending Tick from the previous press
	// doesn't reset state we already cleared.
	cancelGen int
	ctx       context.Context
	cancel    context.CancelFunc
}

// NewProgressModel creates a new progress modal.
func NewProgressModel(kind jobs.Kind, target string, stream cli.Stream, clk clock.Clock) *ProgressModel {
	return NewProgressModelWithPalette(kind, target, stream, clk, theme.DefaultDark())
}

// NewProgressModelWithPalette creates a progress modal styled with the given palette.
func NewProgressModelWithPalette(kind jobs.Kind, target string, stream cli.Stream, clk clock.Clock, p theme.Palette) *ProgressModel {
	ctx, cancel := context.WithCancel(context.Background())

	vp := viewport.New(80, 20)
	vp.Style = lipgloss.NewStyle().Background(p.Bg).Foreground(p.Fg)

	return &ProgressModel{
		kind:       kind,
		target:     target,
		stream:     stream,
		clock:      clk,
		palette:    p,
		started:    clk.Now(),
		layerState: make(map[string]cli.LayerProgress),
		viewport:   vp,
		showRaw:    false,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Init implements tea.Model.
func (m *ProgressModel) Init() tea.Cmd {
	return tea.Batch(
		m.waitForEvent(),
		m.tickElapsed(),
	)
}

type progressEventMsg struct {
	event cli.StreamEvent
}

type progressDoneMsg struct {
	result cli.StreamResult
}

type elapsedTickMsg struct{}

// cancelWindowMsg is delivered ~2s after a first Ctrl+C press; it clears
// awaitCancel back to false so a stale "press Ctrl+C again" hint doesn't
// linger. The gen field guards against a second Ctrl+C arriving in
// between (which advances ProgressModel.cancelGen) and being clobbered
// by the in-flight expiration.
type cancelWindowMsg struct{ gen int }

func (m *ProgressModel) waitForEvent() tea.Cmd {
	return func() tea.Msg {
		event, ok := <-m.stream.Events
		if !ok {
			result := <-m.stream.Done
			return progressDoneMsg{result: result}
		}
		return progressEventMsg{event: event}
	}
}

func (m *ProgressModel) tickElapsed() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return elapsedTickMsg{}
	})
}

// Update implements tea.Model.
func (m *ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case progressEventMsg:
		m.handleEvent(msg.event)
		cmds = append(cmds, m.waitForEvent())

	case progressDoneMsg:
		m.done = true
		m.exitCode = msg.result.ExitCode
		m.err = msg.result.Err

	case elapsedTickMsg:
		if !m.done {
			cmds = append(cmds, m.tickElapsed())
		}

	case cancelWindowMsg:
		// Only clear if this is the most recent Ctrl+C window; otherwise
		// the user already pressed Ctrl+C twice and we're done, or pressed
		// it again and a newer window is in flight.
		if msg.gen == m.cancelGen && m.awaitCancel {
			m.awaitCancel = false
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 5
		m.updateViewport()
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *ProgressModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		if m.done {
			m.cancel()
			return m, CloseModal()
		}
		// Ignore if not done
		return m, nil

	case "ctrl+c":
		if m.awaitCancel {
			// Second press within window → actually cancel
			m.stream.Cancel()
			m.cancel()
			return m, CloseModal()
		}
		// First press → set flag and arm a Tick that will clear it via
		// cancelWindowMsg routed through Update (no goroutine writes to
		// model fields, no race with the Bubble Tea event loop).
		m.awaitCancel = true
		m.cancelGen++
		gen := m.cancelGen
		return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
			return cancelWindowMsg{gen: gen}
		})

	case "ctrl+z":
		// Detach: emit message to app to background this job
		m.cancel()
		return m, func() tea.Msg {
			return JobDetachMsg{JobID: m.jobID}
		}

	case "v":
		if m.kind == jobs.KindBuild {
			m.showRaw = !m.showRaw
			m.updateViewport()
		}
		return m, nil
	}
	return m, nil
}

func (m *ProgressModel) handleEvent(event cli.StreamEvent) {
	switch e := event.(type) {
	case cli.BuildStepEvent:
		m.buildSteps = append(m.buildSteps, e)
	case cli.LayerProgressEvent:
		for _, layer := range e.Layers {
			m.layerState[layer.Digest] = layer
		}
	case cli.RawLine:
		m.rawLines = append(m.rawLines, e.Text)
	case cli.LogLine:
		m.rawLines = append(m.rawLines, e.Raw)
	}
	m.updateViewport()
}

func (m *ProgressModel) updateViewport() {
	var content string

	switch m.kind {
	case jobs.KindBuild:
		if m.showRaw {
			content = m.renderRaw()
		} else {
			content = m.renderBuildSteps()
		}
	case jobs.KindPull, jobs.KindPush:
		content = m.renderLayers()
	default:
		content = m.renderRaw()
	}

	m.viewport.SetContent(content)
}

func (m *ProgressModel) renderBuildSteps() string {
	if len(m.buildSteps) == 0 {
		return "Waiting for build steps..."
	}

	var lines []string
	for _, step := range m.buildSteps {
		var statusIcon string
		if step.Status == "done" {
			statusIcon = "✓"
		} else if step.Status == "cached" {
			statusIcon = "⚡"
		} else {
			statusIcon = "●"
		}

		stage := step.Stage
		if stage == "" {
			stage = "—"
		}

		duration := step.Duration
		if duration == "" {
			duration = "..."
		}

		line := fmt.Sprintf("%s #%d [%s] %s  %s", statusIcon, step.Index, stage, step.Step, duration)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m *ProgressModel) renderLayers() string {
	if len(m.layerState) == 0 {
		action := "Pulling"
		if m.kind == jobs.KindPush {
			action = "Pushing"
		}
		return fmt.Sprintf("%s %s...", action, m.target)
	}

	var lines []string
	for digest, layer := range m.layerState {
		shortDigest := digest
		if len(digest) > 20 {
			shortDigest = digest[:20]
		}

		stateStr := layer.State
		if layer.Mounted {
			stateStr = "mounted"
		}

		progressBar := ""
		if layer.BytesTotal > 0 {
			pct := float64(layer.BytesDone) / float64(layer.BytesTotal)
			barWidth := 20
			filled := int(pct * float64(barWidth))
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
			progressBar = fmt.Sprintf(" %s %3.0f%%", bar, pct*100)
		}

		line := fmt.Sprintf("%s  %s%s", shortDigest, stateStr, progressBar)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

func (m *ProgressModel) renderRaw() string {
	if len(m.rawLines) == 0 {
		return "(no output yet)"
	}
	return strings.Join(m.rawLines, "\n")
}

// View implements tea.Model.
func (m *ProgressModel) View() string {
	header := m.renderHeader()
	body := m.viewport.View()
	footer := m.renderFooter()

	bg := lipgloss.NewStyle().Background(m.palette.Bg).Foreground(m.palette.Fg)
	return bg.Render(lipgloss.JoinVertical(lipgloss.Left, header, body, footer))
}

func (m *ProgressModel) renderHeader() string {
	kindStr := string(m.kind)
	elapsed := m.clock.Now().Sub(m.started).Round(time.Second)

	status := "running"
	if m.done {
		if m.err != nil {
			status = fmt.Sprintf("✗ exit %d", m.exitCode)
		} else {
			status = "✓ done"
		}
	}

	title := fmt.Sprintf("%s: %s  ·  %s  ·  %s", kindStr, m.target, status, elapsed)

	if m.kind == jobs.KindBuild && !m.showRaw {
		title += "  [v:toggle raw]"
	}

	return lipgloss.NewStyle().Bold(true).Background(m.palette.Bg).Foreground(m.palette.HeaderFg).Render(title)
}

func (m *ProgressModel) renderFooter() string {
	if m.awaitCancel {
		return lipgloss.NewStyle().Foreground(m.palette.Error).Background(m.palette.Bg).Render("Press Ctrl+C again to cancel")
	}

	if m.done {
		return lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg).Render("q:quit")
	}

	help := "ctrl+c:cancel  ctrl+z:detach"
	if m.kind == jobs.KindBuild {
		help += "  v:toggle view"
	}
	return lipgloss.NewStyle().Foreground(m.palette.Dim).Background(m.palette.Bg).Render(help)
}

// JobDetachMsg signals that a job should be detached to background.
type JobDetachMsg struct {
	JobID string
}
