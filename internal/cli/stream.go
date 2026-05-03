package cli

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"sync"
	"syscall"
	"time"
)

// StreamEvent is the base interface for all streaming events.
type StreamEvent interface {
	streamEvent()
}

// RawLine represents an unparsed line of output.
type RawLine struct {
	Text string
}

func (RawLine) streamEvent() {}

// StreamResult contains the final result of a streaming command.
type StreamResult struct {
	ExitCode int
	Err      error
}

// Stream represents an active streaming command.
type Stream struct {
	Events <-chan StreamEvent
	Done   <-chan StreamResult
	Cancel context.CancelFunc
}

// ParseFunc converts a line of output into zero or more StreamEvents.
type ParseFunc func(line string) []StreamEvent

// runStream executes a command and streams parsed events through channels.
// It handles context cancellation by sending SIGINT first, then SIGKILL
// after a 2-second grace period (or immediately if the process has already
// exited gracefully). Both channels are closed before returning.
func runStream(ctx context.Context, cmd *exec.Cmd, parse ParseFunc) (Stream, error) {
	eventsCh := make(chan StreamEvent, 100)
	doneCh := make(chan StreamResult, 1)

	ctx, cancel := context.WithCancel(ctx)

	// Merge stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		close(eventsCh)
		close(doneCh)
		return Stream{}, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		close(eventsCh)
		close(doneCh)
		return Stream{}, err
	}

	if err := cmd.Start(); err != nil {
		cancel()
		close(eventsCh)
		close(doneCh)
		return Stream{}, err
	}

	stream := Stream{
		Events: eventsCh,
		Done:   doneCh,
		Cancel: cancel,
	}

	// finished is closed by the reader goroutine once cmd.Wait() has returned.
	// Both the reader and the cancel-watcher use it to coordinate cleanup so
	// that:
	//   - if the command exits naturally, the cancel-watcher unblocks via
	//     `finished` instead of leaking on `<-ctx.Done()` forever; and
	//   - if cancellation is requested, the watcher waits up to 2 s for a
	//     SIGINT to take effect rather than always sleeping a full 2 s and
	//     issuing an unconditional Kill.
	finished := make(chan struct{})

	// Cancel watcher: SIGINT-then-SIGKILL on cancellation, exit early on
	// natural completion.
	go func() {
		select {
		case <-finished:
			return
		case <-ctx.Done():
			if cmd.Process == nil {
				return
			}
			_ = cmd.Process.Signal(syscall.SIGINT)
			select {
			case <-finished:
			case <-time.After(2 * time.Second):
				_ = cmd.Process.Kill()
			}
		}
	}()

	// Reader/parser goroutine: scans output, emits events, waits for the
	// command to exit and reports the StreamResult, then signals `finished`.
	go func() {
		defer close(eventsCh)
		defer func() {
			result := StreamResult{}

			waitErr := cmd.Wait()
			if waitErr != nil {
				result.Err = waitErr
				if exitErr, ok := waitErr.(*exec.ExitError); ok {
					result.ExitCode = exitErr.ExitCode()
				} else {
					result.ExitCode = -1
				}
			} else {
				result.ExitCode = 0
			}

			close(finished)

			doneCh <- result
			close(doneCh)
		}()

		// Merge stdout and stderr into a single reader
		merged := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(merged)

		var wg sync.WaitGroup
		lineCh := make(chan string, 100)

		// Scanner goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer close(lineCh)
			for scanner.Scan() {
				select {
				case lineCh <- scanner.Text():
				case <-ctx.Done():
					return
				}
			}
		}()

		// Parser goroutine
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range lineCh {
				events := parse(line)
				for _, event := range events {
					select {
					case eventsCh <- event:
					case <-ctx.Done():
						return
					}
				}
			}
		}()

		wg.Wait()
	}()

	return stream, nil
}
