package cli

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// LayerProgress represents the progress of a single layer during pull/push.
type LayerProgress struct {
	Digest     string
	State      string // waiting|downloading|extracting|pushing|verifying|done|mounted
	BytesDone  int64
	BytesTotal int64
	Mounted    bool // for push: layer already exists in registry
}

// LayerProgressEvent contains a snapshot of all tracked layers.
type LayerProgressEvent struct {
	Layers []LayerProgress
}

func (LayerProgressEvent) streamEvent() {}

// layerParser maintains state for parsing layer progress lines.
type layerParser struct {
	layers map[string]*LayerProgress
}

func newLayerParser() *layerParser {
	return &layerParser{
		layers: make(map[string]*LayerProgress),
	}
}

// Regex patterns for layer progress lines
// Examples:
//
//	sha256:abc123 downloading 1024 / 10240
//	sha256:def456 extracting
//	sha256:ghi789 done
//	sha256:xyz mounted
//	sha256:abc pushing 500 / 1000
var (
	layerProgressRegex = regexp.MustCompile(`^(sha256:[a-z0-9]+)\s+(\w+)(?:\s+(\d+)\s+/\s+(\d+))?`)
)

func (p *layerParser) parseLine(line string) []StreamEvent {
	matches := layerProgressRegex.FindStringSubmatch(line)
	if matches == nil {
		return []StreamEvent{RawLine{Text: line}}
	}

	digest := matches[1]
	state := strings.ToLower(matches[2])

	var bytesDone, bytesTotal int64
	if matches[3] != "" && matches[4] != "" {
		bytesDone, _ = strconv.ParseInt(matches[3], 10, 64)
		bytesTotal, _ = strconv.ParseInt(matches[4], 10, 64)
	}

	// Get or create layer
	layer, exists := p.layers[digest]
	if !exists {
		layer = &LayerProgress{Digest: digest}
		p.layers[digest] = layer
	}

	// Update layer state
	layer.State = state
	if bytesDone > 0 {
		layer.BytesDone = bytesDone
	}
	if bytesTotal > 0 {
		layer.BytesTotal = bytesTotal
	}
	if state == "mounted" {
		layer.Mounted = true
	}

	// Return snapshot of all layers
	snapshot := make([]LayerProgress, 0, len(p.layers))
	for _, l := range p.layers {
		snapshot = append(snapshot, *l)
	}

	return []StreamEvent{LayerProgressEvent{Layers: snapshot}}
}

// StreamPull streams output from `container image pull`.
func (c *DefaultClient) StreamPull(ctx context.Context, ref string) (Stream, error) {
	args := []string{"image", "pull", ref}

	//nolint:gosec // c.bin is controlled by c9s, not user input
	cmd := exec.CommandContext(ctx, c.bin, args...)

	parser := newLayerParser()
	return runStream(ctx, cmd, parser.parseLine)
}

// StreamPush streams output from `container image push`.
func (c *DefaultClient) StreamPush(ctx context.Context, ref string) (Stream, error) {
	args := []string{"image", "push", ref}

	//nolint:gosec // c.bin is controlled by c9s, not user input
	cmd := exec.CommandContext(ctx, c.bin, args...)

	parser := newLayerParser()
	return runStream(ctx, cmd, parser.parseLine)
}

// StreamPull implements Client for Fake.
func (f *Fake) StreamPull(_ context.Context, ref string) (Stream, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, fmt.Sprintf("StreamPull(%s)", ref))
	events := f.layerStreamEvents
	exitCode := f.layerStreamExitCode
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

// StreamPush implements Client for Fake.
func (f *Fake) StreamPush(_ context.Context, ref string) (Stream, error) {
	f.mu.Lock()
	f.Calls = append(f.Calls, fmt.Sprintf("StreamPush(%s)", ref))
	events := f.layerStreamEvents
	exitCode := f.layerStreamExitCode
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

// ReplayLayerStream configures the Fake to replay the given events and exit code.
func (f *Fake) ReplayLayerStream(events []StreamEvent, exitCode int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.layerStreamEvents = events
	f.layerStreamExitCode = exitCode
}
