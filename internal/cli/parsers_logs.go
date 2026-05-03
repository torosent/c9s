package cli

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// LogLine represents a parsed log line with optional level detection.
type LogLine struct {
	Raw   string
	Level string // INFO|WARN|ERROR|DEBUG|""
}

func (LogLine) streamEvent() {}

var logLevelRegex = regexp.MustCompile(`^(?i)(INFO|WARN|ERROR|DEBUG)[\s:]`)

// parseLogLine detects log level at the start of the line and returns a LogLine event.
func parseLogLine(line string) []StreamEvent {
	level := ""
	if matches := logLevelRegex.FindStringSubmatch(line); matches != nil {
		level = strings.ToUpper(matches[1])
	}
	return []StreamEvent{LogLine{Raw: line, Level: level}}
}

// StreamLogs streams logs from a container. If follow is true, it tails the logs.
func (c *DefaultClient) StreamLogs(ctx context.Context, id string, follow bool) (Stream, error) {
	args := []string{"logs", id}
	if follow {
		args = append(args, "--follow")
	}

	//nolint:gosec // c.bin is controlled by c9s, not user input
	cmd := exec.CommandContext(ctx, c.bin, args...)

	return runStream(ctx, cmd, parseLogLine)
}

// StreamLogs implements Client for Fake.
func (f *Fake) StreamLogs(_ context.Context, id string, follow bool) (Stream, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, fmt.Sprintf("StreamLogs(%s,follow=%t)", id, follow))
	events := f.logStreamEvents
	exitCode := f.logStreamExitCode
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

// ReplayLogStream configures the Fake to replay the given events and exit code.
func (f *Fake) ReplayLogStream(events []StreamEvent, exitCode int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.logStreamEvents = events
	f.logStreamExitCode = exitCode
}
