package modals

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/torosent/c9s/internal/cli"
	"github.com/torosent/c9s/internal/clock"
	"github.com/torosent/c9s/internal/jobs"
)

func TestProgressModel_Smoke(t *testing.T) {
	// Build variant
	stream := cli.Stream{
		Events: make(chan cli.StreamEvent),
		Done:   make(chan cli.StreamResult),
		Cancel: func() {},
	}
	clk := clock.NewFake(time.Now())

	m := NewProgressModel(jobs.KindBuild, "/path", stream, clk)

	if m == nil {
		t.Fatal("NewProgressModel returned nil")
	}

	// Init
	m.Init()

	// View should not panic
	_ = m.View()
}

func TestProgressModel_PullVariant(t *testing.T) {
	// Pull variant
	stream := cli.Stream{
		Events: make(chan cli.StreamEvent),
		Done:   make(chan cli.StreamResult),
		Cancel: func() {},
	}
	clk := clock.NewFake(time.Now())

	m := NewProgressModel(jobs.KindPull, "myimage:latest", stream, clk)

	if m == nil {
		t.Fatal("NewProgressModel returned nil")
	}

	// Init
	m.Init()

	// View should not panic
	_ = m.View()
}

// Test 1: Init returns waitForEvent cmd
func TestProgressModel_InitReturnsCmd(t *testing.T) {
	stream := cli.Stream{
		Events: make(chan cli.StreamEvent),
		Done:   make(chan cli.StreamResult),
		Cancel: func() {},
	}
	clk := clock.NewFake(time.Now())
	m := NewProgressModel(jobs.KindBuild, "/path", stream, clk)

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() returned nil cmd, expected non-nil")
	}
}

// Test 2: Build variant: Update with BuildStepEvent appends to step list
func TestProgressModel_BuildStepEvent(t *testing.T) {
	stream := cli.Stream{
		Events: make(chan cli.StreamEvent),
		Done:   make(chan cli.StreamResult),
		Cancel: func() {},
	}
	clk := clock.NewFake(time.Now())
	m := NewProgressModel(jobs.KindBuild, "/path", stream, clk)

	// Inject a BuildStepEvent
	msg := progressEventMsg{
		event: cli.BuildStepEvent{
			Index:    1,
			Stage:    "builder",
			Step:     "RUN apt-get update",
			Duration: "2.3s",
			Status:   "done",
		},
	}

	m.Update(msg)

	// Render and check output
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	output := m.View()
	if !strings.Contains(output, "RUN apt-get update") {
		t.Errorf("View() missing build step text, got: %s", output)
	}
}

// Test 3: `v` toggles raw view
func TestProgressModel_VTogglesRaw(t *testing.T) {
	stream := cli.Stream{
		Events: make(chan cli.StreamEvent),
		Done:   make(chan cli.StreamResult),
		Cancel: func() {},
	}
	clk := clock.NewFake(time.Now())
	m := NewProgressModel(jobs.KindBuild, "/path", stream, clk)

	if m.showRaw {
		t.Fatal("showRaw should be false initially")
	}

	// Press v
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	if !m.showRaw {
		t.Errorf("showRaw = %v, want true after 'v'", m.showRaw)
	}

	// Press v again
	m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	if m.showRaw {
		t.Errorf("showRaw = %v, want false after second 'v'", m.showRaw)
	}
}

// Test 4: Pull variant: LayerProgressEvent updates layer table
func TestProgressModel_LayerProgressEvent(t *testing.T) {
	stream := cli.Stream{
		Events: make(chan cli.StreamEvent),
		Done:   make(chan cli.StreamResult),
		Cancel: func() {},
	}
	clk := clock.NewFake(time.Now())
	m := NewProgressModel(jobs.KindPull, "myimage:latest", stream, clk)

	// Inject a LayerProgressEvent
	msg := progressEventMsg{
		event: cli.LayerProgressEvent{
			Layers: []cli.LayerProgress{
				{
					Digest:     "sha256:abc123def456",
					State:      "downloading",
					BytesDone:  1024,
					BytesTotal: 2048,
					Mounted:    false,
				},
			},
		},
	}

	m.Update(msg)

	// Render and check output
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	output := m.View()
	if !strings.Contains(output, "abc123def456") {
		t.Errorf("View() missing layer digest in output")
	}
	if !strings.Contains(output, "downloading") {
		t.Errorf("View() missing layer state in output")
	}
}

// Test 5: Single Ctrl+C sets confirm pending; second emits cancel
func TestProgressModel_DoubleCtrlCCancels(t *testing.T) {
	stream := cli.Stream{
		Events: make(chan cli.StreamEvent),
		Done:   make(chan cli.StreamResult),
		Cancel: func() {},
	}
	clk := clock.NewFake(time.Now())
	m := NewProgressModel(jobs.KindBuild, "/path", stream, clk)
	m.jobID = "test-job-123"

	// First Ctrl+C
	m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !m.awaitCancel {
		t.Fatal("awaitCancel should be true after first Ctrl+C")
	}

	// Check footer shows "press Ctrl+C again"
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	output := m.View()
	if !strings.Contains(output, "Press Ctrl+C again") {
		t.Errorf("View() missing cancel confirmation text")
	}

	// Second Ctrl+C
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("second Ctrl+C should return CloseModal cmd")
	}

	// Execute cmd
	msg := cmd()
	if _, ok := msg.(CloseModalMsg); !ok {
		t.Errorf("cmd() returned %T, want CloseModalMsg", msg)
	}
}

// Test 6: Ctrl+Z emits JobDetachMsg + CloseModalMsg
func TestProgressModel_CtrlZDetaches(t *testing.T) {
	stream := cli.Stream{
		Events: make(chan cli.StreamEvent),
		Done:   make(chan cli.StreamResult),
		Cancel: func() {},
	}
	clk := clock.NewFake(time.Now())
	m := NewProgressModel(jobs.KindBuild, "/path", stream, clk)
	m.jobID = "test-job-456"

	// Press Ctrl+Z
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlZ})
	if cmd == nil {
		t.Fatal("Ctrl+Z should return a cmd")
	}

	// Execute cmd
	msg := cmd()
	detachMsg, ok := msg.(JobDetachMsg)
	if !ok {
		t.Fatalf("cmd() returned %T, want JobDetachMsg", msg)
	}
	if detachMsg.JobID != "test-job-456" {
		t.Errorf("JobDetachMsg.JobID = %q, want %q", detachMsg.JobID, "test-job-456")
	}
}

// Test 7: Done event with non-zero exit colors header red
func TestProgressModel_NonZeroExitCode(t *testing.T) {
	stream := cli.Stream{
		Events: make(chan cli.StreamEvent),
		Done:   make(chan cli.StreamResult),
		Cancel: func() {},
	}
	clk := clock.NewFake(time.Now())
	m := NewProgressModel(jobs.KindBuild, "/path", stream, clk)

	// Inject a done event with exit code 1 and error
	msg := progressDoneMsg{
		result: cli.StreamResult{
			ExitCode: 1,
			Err:      fmt.Errorf("exit status 1"),
		},
	}

	m.Update(msg)

	// Render and check header contains "exit 1" or failure indicator
	m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	output := m.View()
	if !strings.Contains(output, "exit 1") && !strings.Contains(output, "✗") {
		t.Errorf("View() missing failure indicator in output for failed build, got: %s", output)
	}
}
