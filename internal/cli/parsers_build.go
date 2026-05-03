package cli

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// BuildStepEvent represents a parsed BuildKit step line.
type BuildStepEvent struct {
	Index    int
	Stage    string
	Step     string
	Duration string
	Status   string // "running"|"done"|"cached"|""
}

func (BuildStepEvent) streamEvent() {}

// BuildOpts contains options for building an image.
type BuildOpts struct {
	ContextPath       string
	Tag               string
	ContainerfilePath string
	Platform          string
	ExtraArgs         []string
}

// args converts BuildOpts to command-line arguments.
func (opts BuildOpts) args() []string {
	args := []string{"build", opts.ContextPath}
	if opts.Tag != "" {
		args = append(args, "-t", opts.Tag)
	}
	if opts.ContainerfilePath != "" {
		args = append(args, "-f", opts.ContainerfilePath)
	}
	if opts.Platform != "" {
		args = append(args, "--platform", opts.Platform)
	}
	args = append(args, opts.ExtraArgs...)
	return args
}

// Tolerant regex for BuildKit step lines: #N [stage] step   [duration]
// Example: "#5 [internal] load build context  done"
// Example: "#3 [stage 1/2] COPY . /app    0.5s"
var buildStepRegex = regexp.MustCompile(`^#(\d+)\s+(?:\[([^\]]+)\]\s+)?(.+?)\s*$`)

// parseBuildLine parses a BuildKit-style step line or returns RawLine for unparsed output.
func parseBuildLine(line string) []StreamEvent {
	matches := buildStepRegex.FindStringSubmatch(line)
	if matches == nil {
		return []StreamEvent{RawLine{Text: line}}
	}

	index, _ := strconv.Atoi(matches[1])
	stage := strings.TrimSpace(matches[2])
	rest := strings.TrimSpace(matches[3])

	// Extract duration/status from the end
	duration := ""
	status := "running"

	// Check for common endings
	if strings.HasSuffix(rest, " done") {
		duration = "done"
		status = "done"
		rest = strings.TrimSuffix(rest, " done")
	} else if strings.HasSuffix(rest, " CACHED") {
		duration = "CACHED"
		status = "cached"
		rest = strings.TrimSuffix(rest, " CACHED")
	} else {
		// Check for time duration like "0.5s"
		parts := strings.Fields(rest)
		if len(parts) > 1 {
			lastPart := parts[len(parts)-1]
			if strings.HasSuffix(lastPart, "s") || strings.HasSuffix(lastPart, "ms") {
				duration = lastPart
				rest = strings.TrimSuffix(rest, " "+lastPart)
			}
		}
	}

	return []StreamEvent{
		BuildStepEvent{
			Index:    index,
			Stage:    stage,
			Step:     strings.TrimSpace(rest),
			Duration: duration,
			Status:   status,
		},
	}
}

// StreamBuild streams build output from `container build`.
func (c *DefaultClient) StreamBuild(ctx context.Context, opts BuildOpts) (Stream, error) {
	args := opts.args()

	//nolint:gosec // c.bin is controlled by c9s, not user input
	cmd := exec.CommandContext(ctx, c.bin, args...)

	return runStream(ctx, cmd, parseBuildLine)
}

// StreamBuild implements Client for Fake.
func (f *Fake) StreamBuild(_ context.Context, opts BuildOpts) (Stream, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, fmt.Sprintf("StreamBuild(context=%s,tag=%s)", opts.ContextPath, opts.Tag))
	events := f.buildStreamEvents
	exitCode := f.buildStreamExitCode
	f.mu.Unlock()

	eventsCh := make(chan StreamEvent, len(events)+1)
	doneCh := make(chan StreamResult, 1)
	ctx, cancel := context.WithCancel(context.Background())

	// Send events and close
	go func() {
		defer close(eventsCh)
		defer func() {
			doneCh <- StreamResult{ExitCode: exitCode}
			close(doneCh)
		}()

		for _, event := range events {
			select {
			case eventsCh <- event:
			case <-ctx.Done():
				return
			}
		}
	}()

	return Stream{
		Events: eventsCh,
		Done:   doneCh,
		Cancel: cancel,
	}, nil
}

// ReplayBuildStream configures the Fake to replay the given events and exit code.
func (f *Fake) ReplayBuildStream(events []StreamEvent, exitCode int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.buildStreamEvents = events
	f.buildStreamExitCode = exitCode
}
